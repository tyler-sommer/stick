package stick

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/tyler-sommer/stick/parse"
)

// Type state represents the internal state of a template execution.
//
// state implements the exported Context interface.
type state struct {
	out  io.Writer  // Output.
	node parse.Node // Current node.

	name string    // The name of the template.
	meta *metadata // Additional template metadata.

	blocks []map[string]*parse.BlockNode // Block scopes.
	macros map[string]*parse.MacroNode   // Imported macros.

	env   *Env        // The configured Stick environment.
	scope *scopeStack // Handles execution scope.
}

// newState creates a new template execution state, ready for use.
func newState(name string, out io.Writer, ctx map[string]Value, env *Env) *state {
	return &state{
		out:  out,
		node: nil,

		name: name,
		meta: &metadata{make(map[string]string)},

		blocks: make([]map[string]*parse.BlockNode, 0),
		macros: make(map[string]*parse.MacroNode),

		env:   env,
		scope: &scopeStack{[]map[string]Value{ctx}},
	}
}

func (s *state) Env() *Env {
	return s.env
}

func (s *state) Name() string {
	return s.name
}

func (s *state) Scope() ContextScope {
	return s.scope
}

func (s *state) Meta() ContextMetadata {
	return s.meta
}

// noexport satisfies the Context interface.
func (s *state) noexport() {}

// metadata contains extra information about the state.
//
// metadata implements the exported ContextMetadata interface.
type metadata struct {
	attr map[string]string
}

func (m *metadata) All() map[string]string {
	ret := make(map[string]string, len(m.attr))
	for k, v := range m.attr {
		ret[k] = v
	}
	return ret
}

func (m *metadata) Set(name, val string) {
	m.attr[name] = val
}

func (m *metadata) Get(name string) (string, bool) {
	ret, ok := m.attr[name]
	return ret, ok
}

// noexport satisfies the ContextMetadata interface.
func (m *metadata) noexport() {}

// A scopeStack is a structure that represents local
// and parent scopes.
//
// scopeStack implements the exported ContextScope interface.
type scopeStack struct {
	scopes []map[string]Value
}

// push adds a scope on top of the stack.
func (s *scopeStack) push() {
	s.scopes = append(s.scopes, make(map[string]Value))
}

// pop removes the top-most scope.
func (s *scopeStack) pop() {
	s.scopes = s.scopes[0 : len(s.scopes)-1]
}

// All returns a flat map of the current scope.
func (s *scopeStack) All() map[string]Value {
	res := make(map[string]Value)
	for _, scope := range s.scopes {
		for k, v := range scope {
			res[k] = v
		}
	}
	return res
}

// noexport satisfies the ContextScope interface.
func (s *scopeStack) noexport() {}

// Get returns a value in the scope.
//
// This function works from the current (local) scope
// upward. The second parameter returns false if the
// value was not found-- this can be used to distinguish
// a non-existent value and a valid nil value.
func (s *scopeStack) Get(name string) (Value, bool) {
	for i := len(s.scopes); i > 0; i-- {
		scope := s.scopes[i-1]
		if v, ok := scope[name]; ok {
			return v, true
		}
	}
	return nil, false
}

// Set sets the value in the scope.
//
// This function will work from the top-most scope downward,
// looking for a scope with name defined. The value is set
// on the scope that it was originally defined on, otherwise
// on the local (last) scope.
func (s *scopeStack) Set(name string, val Value) {
	for _, scope := range s.scopes {
		if _, ok := scope[name]; ok {
			scope[name] = val
			return
		}
	}
	s.scopes[len(s.scopes)-1][name] = val
}

// setLocal explicitly sets the value in the local scope.
//
// This is useful when a new scope is created, such as
// a macro call, and you need to override a local variable
// without destroying the value in the parent scope.
//
//	fnParam := // an function argument's name
// 	s.scope.push()
//	defer s.scope.pop()
//	s.scope.setLocal(fnParam, "some value")
func (s *scopeStack) setLocal(name string, val Value) {
	s.scopes[len(s.scopes)-1][name] = val
}

// Method getBlock iterates through each set of blocks, returning the first
// block with the given name.
func (s *state) getBlock(name string) *parse.BlockNode {
	for _, blocks := range s.blocks {
		if block, ok := blocks[name]; ok {
			return block
		}
	}

	return nil
}

// Method walk is the main entry-point into template execution.
func (s *state) walk(node parse.Node) error {
	switch node := node.(type) {
	case *parse.ModuleNode:
		if p := node.Parent; p != nil {
			tplName, err := s.evalExpr(p.Tpl)
			if err != nil {
				return err
			}
			name := CoerceString(tplName)
			tree, err := s.env.load(name)
			if err != nil {
				return err
			}
			defer func(name string) {
				s.name = name
			}(s.name)
			s.name = name
			s.blocks = append(s.blocks, tree.Blocks())
			err = s.walkChild(node.BodyNode)
			if err != nil {
				return err
			}
			return s.walk(tree.Root())
		}
		return s.walk(node.BodyNode)
	case *parse.BodyNode:
		for _, c := range node.All() {
			err := s.walk(c)
			if err != nil {
				return err
			}
		}
	case *parse.MacroNode:
		return nil
	case *parse.TextNode:
		io.WriteString(s.out, node.Data)
	case *parse.PrintNode:
		v, err := s.evalExpr(node.X)
		if err != nil {
			return err
		}
		io.WriteString(s.out, CoerceString(v))
	case *parse.BlockNode:
		name := node.Name
		if block := s.getBlock(name); block != nil {
			if block.Origin != "" {
				defer func(name string) {
					s.name = name
				}(s.name)
				s.name = block.Origin
			}
			return s.walk(block.Body)
		}
		// TODO: It seems this should never occur.
		return errors.New("Unable to locate block " + name)
	case *parse.IfNode:
		v, err := s.evalExpr(node.Cond)
		if err != nil {
			return err
		}
		if CoerceBool(v) {
			s.walk(node.Body)
		} else {
			s.walk(node.Else)
		}
	case *parse.IncludeNode:
		tpl, ctx, err := s.walkIncludeNode(node)
		if err != nil {
			return err
		}
		err = execute(tpl, s.out, ctx, s.env)
		if err != nil {
			return err
		}
	case *parse.EmbedNode:
		tpl, ctx, err := s.walkIncludeNode(node.IncludeNode)
		if err != nil {
			return err
		}
		si := newState(tpl, s.out, ctx, s.env)
		tree, err := s.env.load(tpl)
		if err != nil {
			return err
		}
		si.blocks = append(s.blocks, node.Blocks, tree.Blocks())
		err = si.walk(tree.Root())
		if err != nil {
			return err
		}
	case *parse.UseNode:
		return s.walkUseNode(node)
	case *parse.ForNode:
		return s.walkForNode(node)
	case *parse.SetNode:
		return s.walkSetNode(node)
	case *parse.DoNode:
		return s.walkDoNode(node)
	case *parse.FilterNode:
		return s.walkFilterNode(node)
	case *parse.ImportNode:
		return s.walkImportNode(node)
	case *parse.FromNode:
		return s.walkFromNode(node)
	case *parse.CommentNode:
		// Nothing.
	default:
		return errors.New("Unknown node " + node.String())
	}
	return nil
}

// walkChild only executes a subset of nodes, intended to be used on child templates.
func (s *state) walkChild(node parse.Node) error {
	switch node := node.(type) {
	case *parse.BodyNode:
		for _, c := range node.All() {
			err := s.walkChild(c)
			if err != nil {
				return err
			}
		}
	case *parse.UseNode:
		return s.walkUseNode(node)
	}
	return nil
}

func (s *state) walkForNode(node *parse.ForNode) error {
	res, err := s.evalExpr(node.X)
	if err != nil {
		return err
	}
	kn := node.Key
	vn := node.Val
	ct, err := Iterate(res, func(k Value, v Value, l Loop) (bool, error) {
		s.scope.push()
		defer s.scope.pop()

		if kn != "" {
			s.scope.Set(kn, k)
		}
		s.scope.Set(vn, v)
		s.scope.Set("loop", l)

		err := s.walk(node.Body)
		if err != nil {
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	if ct == 0 {
		return s.walk(node.Else)
	}
	return nil
}

// Method walkInclude determines the necessary parameters for including or embedding a template.
func (s *state) walkIncludeNode(node *parse.IncludeNode) (tpl string, ctx map[string]Value, err error) {
	ctx = make(map[string]Value)
	v, err := s.evalExpr(node.Tpl)
	if err != nil {
		return "", nil, err
	}
	tpl = CoerceString(v)
	var with Value
	if n := node.With; n != nil {
		with, err = s.evalExpr(n)
		// TODO: Assert "with" is a hash?
		if err != nil {
			return "", nil, err
		}
	}
	if !node.Only {
		ctx = s.scope.All()
	}
	if with != nil {
		if with, ok := with.(map[string]Value); ok {
			for k, v := range with {
				ctx[k] = v
			}
		}
	}
	return tpl, ctx, err
}

func (s *state) walkUseNode(node *parse.UseNode) error {
	v, err := s.evalExpr(node.Tpl)
	if err != nil {
		return err
	}
	tpl := CoerceString(v)
	tree, err := s.env.load(tpl)
	if err != nil {
		return err
	}
	blocks := tree.Blocks()
	for orig, alias := range node.Aliases {
		v, ok := blocks[orig]
		if !ok {
			return errors.New("Unable to locate block with name \"" + orig + "\"")
		}
		blocks[alias] = v
	}
	l := len(s.blocks)
	lb := s.blocks[l-1]
	s.blocks = append(append(s.blocks[:l-1], blocks), lb)
	return nil
}

func (s *state) walkSetNode(node *parse.SetNode) error {
	v, err := s.evalExpr(node.X)
	if err != nil {
		return err
	}
	s.scope.Set(node.Name, v)
	return nil
}

func (s *state) walkDoNode(node *parse.DoNode) error {
	_, err := s.evalExpr(node.X)
	if err != nil {
		return err
	}
	return nil
}

func (s *state) walkFilterNode(node *parse.FilterNode) error {
	prevBuf := s.out
	defer func() {
		s.out = prevBuf
	}()
	buf := &bytes.Buffer{}
	s.out = buf
	err := s.walk(node.Body)
	if err != nil {
		return err
	}
	val := string(buf.Bytes())
	for _, v := range node.Filters {
		f, ok := s.env.Filters[v]
		if !ok {
			return errors.New("undefined filter \"" + v + "\".")
		}
		val = CoerceString(f(s, val))
	}
	io.WriteString(prevBuf, val)
	return nil
}

func (s *state) walkImportNode(node *parse.ImportNode) error {
	tpl, err := s.evalExpr(node.Tpl)
	if err != nil {
		return err
	}
	tree, err := s.env.load(CoerceString(tpl))
	if err != nil {
		return err
	}
	macros := make(map[string]macroDef)
	for name, def := range tree.Macros() {
		macros[name] = macroDef{def}
	}
	s.scope.Set(node.Alias, macroSet{macros})
	return nil
}

func (s *state) walkFromNode(node *parse.FromNode) error {
	tpl, err := s.evalExpr(node.Tpl)
	if err != nil {
		return err
	}
	tree, err := s.env.load(CoerceString(tpl))
	if err != nil {
		return err
	}
	macros := tree.Macros()
	for name, alias := range node.Imports {
		def, ok := macros[name]
		if !ok {
			return errors.New("undefined macro " + name)
		}
		s.macros[alias] = def
	}
	return nil
}

// Method evalExpr evaluates the given expression, returning a Value or error.
func (s *state) evalExpr(exp parse.Expr) (v Value, e error) {
	switch exp := exp.(type) {
	case *parse.NullExpr:
		return nil, nil
	case *parse.BoolExpr:
		return exp.Value, nil
	case *parse.NameExpr:
		if val, ok := s.scope.Get(exp.Name); ok {
			v = val
		} else {
			e = errors.New("undefined variable \"" + exp.Name + "\"")
		}
	case *parse.NumberExpr:
		num, err := strconv.ParseFloat(exp.Value, 64)
		if err != nil {
			return nil, err
		}
		return num, nil
	case *parse.StringExpr:
		return exp.Text, nil
	case *parse.GroupExpr:
		return s.evalExpr(exp.X)
	case *parse.UnaryExpr:
		in, err := s.evalExpr(exp.X)
		if err != nil {
			return nil, err
		}
		switch exp.Op {
		case parse.OpUnaryNot:
			return !CoerceBool(in), nil
		case parse.OpUnaryPositive:
			// no-op, +1 = 1, +(-1) = -1, +(false) = 0
			return CoerceNumber(in), nil
		case parse.OpUnaryNegative:
			return -CoerceNumber(in), nil
		}
	case *parse.BinaryExpr:
		left, err := s.evalExpr(exp.Left)
		if err != nil {
			return nil, err
		}
		right, err := s.evalExpr(exp.Right)
		if err != nil {
			return nil, err
		}
		switch exp.Op {
		case parse.OpBinaryAdd:
			return CoerceNumber(left) + CoerceNumber(right), nil
		case parse.OpBinarySubtract:
			return CoerceNumber(left) - CoerceNumber(right), nil
		case parse.OpBinaryMultiply:
			return CoerceNumber(left) * CoerceNumber(right), nil
		case parse.OpBinaryDivide:
			return CoerceNumber(left) / CoerceNumber(right), nil
		case parse.OpBinaryFloorDiv:
			return math.Floor(CoerceNumber(left) / CoerceNumber(right)), nil
		case parse.OpBinaryModulo:
			return float64(int(CoerceNumber(left)) % int(CoerceNumber(right))), nil
		case parse.OpBinaryPower:
			return math.Pow(CoerceNumber(left), CoerceNumber(right)), nil
		case parse.OpBinaryConcat:
			return CoerceString(left) + CoerceString(right), nil
		case parse.OpBinaryEndsWith:
			return strings.HasSuffix(CoerceString(left), CoerceString(right)), nil
		case parse.OpBinaryStartsWith:
			return strings.HasPrefix(CoerceString(left), CoerceString(right)), nil
		case parse.OpBinaryIn:
			return Contains(right, left)
		case parse.OpBinaryNotIn:
			res, err := Contains(right, left)
			if err != nil {
				return false, err
			}
			return !res, nil
		case parse.OpBinaryIs:
			if fn, ok := right.(func(v Value) bool); ok {
				return fn(left), nil
			}
			return nil, errors.New("right operand was of unexpected type")
		case parse.OpBinaryIsNot:
			if fn, ok := right.(func(v Value) bool); ok {
				return !fn(left), nil
			}
			return nil, errors.New("right operand was of unexpected type")
		case parse.OpBinaryMatches:
			reg, err := regexp.Compile(CoerceString(right))
			if err != nil {
				return nil, err
			}
			return reg.MatchString(CoerceString(left)), nil
		case parse.OpBinaryEqual:
			return Equal(left, right), nil
		case parse.OpBinaryNotEqual:
			return !Equal(left, right), nil
		case parse.OpBinaryGreaterEqual:
			return CoerceNumber(left) >= CoerceNumber(right), nil
		case parse.OpBinaryGreaterThan:
			return CoerceNumber(left) > CoerceNumber(right), nil
		case parse.OpBinaryLessEqual:
			return CoerceNumber(left) <= CoerceNumber(right), nil
		case parse.OpBinaryLessThan:
			return CoerceNumber(left) < CoerceNumber(right), nil
		case parse.OpBinaryRange:
			l, r := CoerceNumber(left), CoerceNumber(right)
			res := make([]float64, uint(math.Ceil(r-l))+1)
			for i, k := 0, l; k <= r; i, k = i+1, k+1 {
				res[i] = k
			}
			return res, nil
		case parse.OpBinaryBitwiseAnd:
			return int(CoerceNumber(left)) & int(CoerceNumber(right)), nil
		case parse.OpBinaryBitwiseOr:
			return int(CoerceNumber(left)) | int(CoerceNumber(right)), nil
		case parse.OpBinaryBitwiseXor:
			return int(CoerceNumber(left)) ^ int(CoerceNumber(right)), nil
		case parse.OpBinaryAnd:
			return CoerceBool(left) && CoerceBool(right), nil
		case parse.OpBinaryOr:
			return CoerceBool(left) && CoerceBool(right), nil
		}
	case *parse.FuncExpr:
		return s.evalFunction(exp)
	case *parse.FilterExpr:
		return s.evalFilter(exp)
	case *parse.GetAttrExpr:
		c, err := s.evalExpr(exp.Cont)
		if err != nil {
			return nil, err
		}
		k, err := s.evalExpr(exp.Attr)
		if err != nil {
			return nil, err
		}
		exargs := exp.Args
		args := make([]Value, len(exargs))
		for k, e := range exargs {
			v, err := s.evalExpr(e)
			if err != nil {
				return nil, err
			}
			args[k] = v
		}
		if set, ok := c.(macroSet); ok {
			if macro, ok := set.defs[CoerceString(k)]; ok {
				return s.callMacro(macro, args...)
			}
			return nil, errors.New("undefined macro: " + CoerceString(k))
		}
		v, err = GetAttr(c, k, args...)
		if err != nil {
			return nil, err
		}
	case *parse.TestExpr:
		if tfn, ok := s.env.Tests[exp.Name]; ok {
			eargs := exp.Args
			args := make([]Value, len(eargs))
			for i, e := range eargs {
				v, err := s.evalExpr(e)
				if err != nil {
					return nil, err
				}
				args[i] = v
			}
			return func(v Value) bool {
				return tfn(s, v, args...)
			}, nil
		}
		return nil, fmt.Errorf(`unknown test "%v"`, exp.Name)
	case *parse.TernaryIfExpr:
		cond, err := s.evalExpr(exp.Cond)
		if err != nil {
			return nil, err
		}
		if CoerceBool(cond) == true {
			return s.evalExpr(exp.TrueX)
		}
		return s.evalExpr(exp.FalseX)

	case *parse.HashExpr:
		vals := make(map[string]Value)
		for _, v := range exp.Elements {
			var key Value
			var err error
			if k, ok := v.Key.(*parse.NameExpr); ok {
				key = k.Name
			} else {
				key, err = s.evalExpr(v.Key)
				if err != nil {
					return nil, err
				}
			}
			val, err := s.evalExpr(v.Value)
			if err != nil {
				return nil, err
			}
			vals[CoerceString(key)] = val
		}
		return vals, nil

	case *parse.ArrayExpr:
		vals := make([]Value, len(exp.Elements))
		for i, v := range exp.Elements {
			val, err := s.evalExpr(v)
			if err != nil {
				return nil, err
			}
			vals[i] = val
		}
		return vals, nil
	}

	return v, nil
}

func (s *state) evalFunction(exp *parse.FuncExpr) (Value, error) {
	fnName := exp.Name
	switch fnName {
	case "block":
		eargs := exp.Args
		if len(eargs) != 1 {
			return nil, errors.New("block expects one parameter")
		}
		val, err := s.evalExpr(eargs[0])
		if err != nil {
			return nil, err
		}
		name := CoerceString(val)
		if blk := s.getBlock(name); blk != nil {
			pout := s.out
			buf := &bytes.Buffer{}
			s.out = buf
			err = s.walk(blk.Body)
			if err != nil {
				return nil, err
			}
			s.out = pout
			return buf.String(), nil
		}
		return nil, errors.New("Unable to locate block \"" + name + "\"")
	}
	if macro, ok := s.macros[fnName]; ok {
		eargs := exp.Args
		args := make([]Value, len(eargs))
		for i, e := range eargs {
			v, err := s.evalExpr(e)
			if err != nil {
				return nil, err
			}
			args[i] = v
		}
		return s.callMacro(macroDef{macro}, args...)
	}
	if fn, ok := s.env.Functions[fnName]; ok {
		eargs := exp.Args
		args := make([]Value, len(eargs))
		for i, e := range eargs {
			v, err := s.evalExpr(e)
			if err != nil {
				return nil, err
			}
			args[i] = v
		}
		return fn(s, args...), nil
	}
	return nil, errors.New("Undeclared function \"" + fnName + "\"")
}

func (s *state) evalFilter(exp *parse.FilterExpr) (Value, error) {
	ftName := exp.Name
	if fn, ok := s.env.Filters[ftName]; ok {
		eargs := exp.Args
		if len(eargs) == 0 {
			return nil, errors.New("Filter call must receive at least one argument")
		}
		args := make([]Value, len(eargs))
		for i, e := range eargs {
			v, err := s.evalExpr(e)
			if err != nil {
				return nil, err
			}
			args[i] = v
		}
		return fn(s, args[0], args[1:]...), nil
	}
	return nil, errors.New("Undeclared filter \"" + ftName + "\"")
}

type macroDef struct {
	*parse.MacroNode
}

type macroSet struct {
	defs map[string]macroDef
}

func (s *state) callMacro(macro macroDef, args ...Value) (Value, error) {
	s.scope.push()
	defer s.scope.pop()
	for i, name := range macro.Args {
		if i >= len(args) {
			s.scope.setLocal(name, nil)
		} else {
			s.scope.setLocal(name, args[i])
		}
	}
	defer func(buf io.Writer) {
		s.out = buf
	}(s.out)
	buf := &bytes.Buffer{}
	s.out = buf
	if macro.Origin != "" {
		defer func(name string) {
			s.name = name
		}(s.name)
		s.name = macro.Origin
	}
	err := s.walk(macro.Body)
	if err != nil {
		return nil, err
	}
	return buf.String(), nil
}

// execute kicks off execution of the given template.
func execute(name string, out io.Writer, ctx map[string]Value, env *Env) error {
	if ctx == nil {
		ctx = make(map[string]Value)
	}
	s := newState(name, out, ctx, env)
	tree, err := s.env.load(name)
	if err != nil {
		return err
	}
	s.blocks = append(s.blocks, tree.Blocks())
	err = s.walk(tree.Root())
	if err != nil {
		return err
	}
	return nil
}

// Method load attempts to load and parse the given template.
func (env *Env) load(name string) (*parse.Tree, error) {
	tpl, err := env.Loader.Load(name)
	if err != nil {
		return nil, err
	}
	tree := parse.NewNamedTree(name, tpl.Contents())
	tree.Visitors = append(tree.Visitors, env.Visitors...)
	err = tree.Parse()
	if err != nil {
		return nil, err
	}
	return tree, nil
}
