package stick

import (
	"errors"
	"io"
	"math"
	"strconv"

	"strings"

	"regexp"

	"fmt"

	"github.com/tyler-sommer/stick/parse"
)

// Type state represents the internal state of a template execution.
type state struct {
	out    io.Writer
	node   parse.Node
	blocks []map[string]*parse.BlockNode

	env   *Env
	scope *scopeStack
}

type scopeStack struct {
	scopes []map[string]Value
}

func (s *scopeStack) push() {
	s.scopes = append(s.scopes, make(map[string]Value))
}

func (s *scopeStack) pop() {
	s.scopes = s.scopes[0 : len(s.scopes)-1]
}

func (s *scopeStack) all() map[string]Value {
	res := make(map[string]Value)
	for _, scope := range s.scopes {
		for k, v := range scope {
			res[k] = v
		}
	}
	return res
}

func (s *scopeStack) get(name string) (Value, bool) {
	for i := len(s.scopes); i > 0; i-- {
		scope := s.scopes[i-1]
		if v, ok := scope[name]; ok {
			return v, true
		}
	}
	return nil, false
}

func (s *scopeStack) set(name string, val Value) {
	for _, scope := range s.scopes {
		if _, ok := scope[name]; ok {
			scope[name] = val
			return
		}
	}
	s.scopes[len(s.scopes)-1][name] = val
}

// Function newState creates a new template execution state, ready for use.
func newState(out io.Writer, ctx map[string]Value, env *Env) *state {
	return &state{out, nil, make([]map[string]*parse.BlockNode, 0), env, &scopeStack{[]map[string]Value{ctx}}}
}

// Method load attempts to load and parse the given template.
func (s *state) load(name string) (*parse.Tree, error) {
	cnt, err := s.env.Loader.Load(name)
	if err != nil {
		return nil, err
	}
	tree := parse.NewTree(name, cnt)
	err = tree.Parse()
	if err != nil {
		return nil, err
	}
	return tree, nil
}

// Method getBlock iterates through each set of blocks, returning the first
// block it encounters.
func (s *state) getBlock(name string) *parse.BlockNode {
	for _, blocks := range s.blocks {
		if block, ok := blocks[name]; ok {
			return block
		}
	}

	return nil
}

// Enter is called when the given Node is entered.
func (s *state) enter(node parse.Node) {
	s.node = node
	for _, v := range s.env.Visitors {
		v.Enter(node)
	}
}

// Leave is called just before the state exits the given Node.
func (s *state) leave(node parse.Node) {
	for _, v := range s.env.Visitors {
		v.Leave(node)
	}
}

// Method walk is the main entry-point into template execution.
func (s *state) walk(node parse.Node) error {
	s.enter(node)
	defer s.leave(node)
	switch node := node.(type) {
	case *parse.ModuleNode:
		if p := node.Parent(); p != nil {
			tplName, err := s.walkExpr(p.TemplateRef())
			if err != nil {
				return err
			}
			tree, err := s.load(CoerceString(tplName))
			if err != nil {
				return err
			}
			s.blocks = append(s.blocks, tree.Blocks())
			return s.walk(tree.Root())
		}
		return s.walk(node.BodyNode)
	case *parse.BodyNode:
		for _, c := range node.Children() {
			err := s.walk(c)
			if err != nil {
				return err
			}
		}
	case *parse.TextNode:
		io.WriteString(s.out, node.Text())
	case *parse.PrintNode:
		v, err := s.walkExpr(node.Expr())
		if err != nil {
			return err
		}
		io.WriteString(s.out, CoerceString(v))
	case *parse.BlockNode:
		name := node.Name()
		if block := s.getBlock(name); block != nil {
			return s.walk(block.Body())
		}
		// TODO: It seems this should never occur.
		return errors.New("Unable to locate block " + name)
	case *parse.IfNode:
		v, err := s.walkExpr(node.Cond())
		if err != nil {
			return err
		}
		if CoerceBool(v) {
			s.walk(node.Body())
		} else {
			s.walk(node.Else())
		}
	case *parse.IncludeNode:
		tpl, ctx, err := s.walkInclude(node)
		if err != nil {
			return err
		}
		err = execute(tpl, s.out, ctx, s.env)
		if err != nil {
			return err
		}
	case *parse.EmbedNode:
		tpl, ctx, err := s.walkInclude(node.IncludeNode)
		if err != nil {
			return err
		}
		// TODO: We duplicate most of the "execute" function here.
		s := newState(s.out, ctx, s.env)
		tree, err := s.load(tpl)
		if err != nil {
			return err
		}
		s.blocks = append(s.blocks, node.Blocks(), tree.Blocks())
		err = s.walk(tree.Root())
		if err != nil {
			return err
		}
	case *parse.ForNode:
		return s.walkForNode(node)
	default:
		return errors.New("Unknown node " + node.String())
	}
	return nil
}

func (s *state) walkForNode(node *parse.ForNode) error {
	res, err := s.walkExpr(node.Expr())
	if err != nil {
		return err
	}
	kn := node.Key()
	vn := node.Val()
	ct, err := iterate(res, func(k Value, v Value, l loop) (bool, error) {
		s.scope.push()
		defer s.scope.pop()

		if kn != "" {
			s.scope.set(kn, k)
		}
		s.scope.set(vn, v)
		s.scope.set("loop", l)

		err := s.walk(node.Body())
		if err != nil {
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	if ct == 0 {
		return s.walk(node.Else())
	}
	return nil
}

// Method walkInclude determines the necessary parameters for including or embedding a template.
func (s *state) walkInclude(node *parse.IncludeNode) (tpl string, ctx map[string]Value, err error) {
	ctx = make(map[string]Value)
	v, err := s.walkExpr(node.Tpl())
	if err != nil {
		return "", nil, err
	}
	tpl = CoerceString(v)
	var with Value
	if n := node.With(); n != nil {
		with, err = s.walkExpr(n)
		// TODO: Assert "with" is a hash?
		if err != nil {
			return "", nil, err
		}
	}
	if !node.Only() {
		ctx = s.scope.all()
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

// Method walkExpr executes the given expression, returning a Value or error.
func (s *state) walkExpr(exp parse.Expr) (v Value, e error) {
	switch exp := exp.(type) {
	case *parse.NullExpr:
		return nil, nil
	case *parse.BoolExpr:
		return exp.Value(), nil
	case *parse.NameExpr:
		if val, ok := s.scope.get(exp.Name()); ok {
			v = val
		} else {
			e = errors.New("Undefined variable \"" + exp.Name() + "\"")
		}
	case *parse.NumberExpr:
		num, err := strconv.ParseFloat(exp.Value(), 64)
		if err != nil {
			return nil, err
		}
		return num, nil
	case *parse.StringExpr:
		return exp.Value(), nil
	case *parse.GroupExpr:
		return s.walkExpr(exp.Inner())
	case *parse.UnaryExpr:
		in, err := s.walkExpr(exp.Expr())
		if err != nil {
			return nil, err
		}
		switch exp.Op() {
		case parse.OpUnaryNot:
			return !CoerceBool(in), nil
		case parse.OpUnaryPositive:
			// no-op, +1 = 1, +(-1) = -1, +(false) = 0
			return CoerceNumber(in), nil
		case parse.OpUnaryNegative:
			return -CoerceNumber(in), nil
		}
	case *parse.BinaryExpr:
		left, err := s.walkExpr(exp.Left())
		if err != nil {
			return nil, err
		}
		right, err := s.walkExpr(exp.Right())
		if err != nil {
			return nil, err
		}
		switch exp.Op() {
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
			return contains(right, left)
		case parse.OpBinaryNotIn:
			res, err := contains(right, left)
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
			return equal(left, right), nil
		case parse.OpBinaryNotEqual:
			return !equal(left, right), nil
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
		fnName := exp.Name()
		if fn, ok := s.env.Functions[fnName]; ok {
			eargs := exp.Args()
			args := make([]Value, len(eargs))
			for i, e := range eargs {
				v, err := s.walkExpr(e)
				if err != nil {
					return nil, err
				}
				args[i] = v
			}

			return fn(s.env, args...), nil
		}
		return nil, errors.New("Undeclared function \"" + fnName + "\"")
	case *parse.GetAttrExpr:
		c, err := s.walkExpr(exp.Cont())
		if err != nil {
			return nil, err
		}
		k, err := s.walkExpr(exp.Attr())
		if err != nil {
			return nil, err
		}
		exargs := exp.Args()
		args := make([]Value, len(exargs))
		for k, e := range exargs {
			v, err := s.walkExpr(e)
			if err != nil {
				return nil, err
			}
			args[k] = v
		}
		v, err = GetAttr(c, CoerceString(k), args...)
		if err != nil {
			return nil, err
		}
	case *parse.TestExpr:
		if tfn, ok := s.env.Tests[exp.Name()]; ok {
			eargs := exp.Args()
			args := make([]Value, len(eargs))
			for i, e := range eargs {
				v, err := s.walkExpr(e)
				if err != nil {
					return nil, err
				}
				args[i] = v
			}
			return func(v Value) bool {
				return tfn(s.env, v, args...)
			}, nil
		}
		return nil, fmt.Errorf(`unknown test "%v"`, exp.Name())
	}

	return v, nil
}

// execute kicks off execution of the given template.
func execute(name string, out io.Writer, ctx map[string]Value, env *Env) error {
	s := newState(out, ctx, env)
	tree, err := s.load(name)
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
