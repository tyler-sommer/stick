package stick

import (
	"errors"
	"fmt"
	"github.com/tyler-sommer/stick/parse"
	"io"
	"strconv"
)

// Type state represents the internal state of a template execution.
type state struct {
	out     io.Writer
	node    parse.Node
	context map[string]Value
	blocks  []map[string]*parse.BlockNode

	env *Env
}

// Function newState creates a new template execution state, ready for use.
func newState(out io.Writer, ctx map[string]Value, env *Env) *state {
	return &state{out, nil, ctx, make([]map[string]*parse.BlockNode, 0), env}
}

// Method load attempts to load and parse the given template.
func (s *state) load(name string) (*parse.Tree, error) {
	tmpl, err := s.env.loader.Load(name)
	if err != nil {
		return nil, err
	}
	tree, err := parse.Parse(tmpl)
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

// Method walk is the main entry-point into template execution.
func (s *state) walk(node parse.Node) error {
	s.node = node
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
		io.WriteString(s.out, fmt.Sprintf("%v", v))
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
	default:
		return errors.New("Unknown node " + node.String())
	}
	return nil
}

// Method walkInclude determines the necessary parameters for including or embedding a template.
func (s *state) walkInclude(node *parse.IncludeNode) (tpl string, ctx map[string]Value, err error) {
	ctx = make(map[string]Value)
	v, err := s.walkExpr(node.Tpl())
	if err != nil {
		return
	}
	tpl = CoerceString(v)
	var with Value
	if n := node.With(); n != nil {
		with, err = s.walkExpr(n)
		// TODO: Assert "with" is a hash?
		if err != nil {
			return
		}
	}
	if !node.Only() {
		for k, v := range s.context {
			ctx[k] = v
		}
	}
	if with != nil {
		if with, ok := with.(map[string]Value); ok {
			for k, v := range with {
				ctx[k] = v
			}
		}
	}
	return
}

// Method walkExpr executes the given expression, returning a Value or error.
func (s *state) walkExpr(exp parse.Expr) (v Value, e error) {
	switch exp := exp.(type) {
	case *parse.NullExpr:
		return nil, nil
	case *parse.BoolExpr:
		return exp.Value(), nil
	case *parse.NameExpr:
		if val, ok := s.context[exp.Name()]; ok {
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
		case parse.OpBinaryConcat:
			return CoerceString(left) + CoerceString(right), nil
		case parse.OpBinaryEqual:
			// TODO: Stop-gap for now, this will need to be much more sophisticated.
			return CoerceString(left) == CoerceString(right), nil
		}
	case *parse.FuncExpr:
		fnName := exp.Name()
		if fn, ok := s.env.functions[fnName]; ok {
			args := make([]Value, 0)
			for _, e := range exp.Args() {
				v, err := s.walkExpr(e)
				if err != nil {
					return nil, err
				}
				args = append(args, v)
			}

			return fn(s.env, args...), nil
		} else {
			return nil, errors.New("Undeclared function \"" + fnName + "\"")
		}
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
		v, err := GetAttr(c, CoerceString(k), args...)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return
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
