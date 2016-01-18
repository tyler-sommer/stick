package stick

import (
	"errors"
	"fmt"
	"github.com/tyler-sommer/stick/parse"
	"io"
	"strconv"
)

type state struct {
	out     io.Writer
	node    parse.Node
	context map[string]Value
	blocks  []map[string]*parse.BlockNode

	loader Loader
	cache  map[string]*parse.Tree
}

func newState(out io.Writer, ctx map[string]Value, loader Loader) *state {
	return &state{out, nil, ctx, make([]map[string]*parse.BlockNode, 0), loader, make(map[string]*parse.Tree)}
}

func (s *state) load(name string) (*parse.Tree, error) {
	if v, ok := s.cache[name]; ok {
		return v, nil
	}
	tmpl, err := s.loader.Load(name)
	if err != nil {
		return nil, err
	}
	tree, err := parse.Parse(tmpl)
	if err != nil {
		return nil, err
	}
	s.cache[name] = tree
	return tree, nil
}

func (s *state) getBlock(name string) *parse.BlockNode {
	for _, blocks := range s.blocks {
		if block, ok := blocks[name]; ok {
			return block
		}
	}

	return nil
}

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
		err = execute(tpl, s.out, ctx, s.loader)
		if err != nil {
			return err
		}
	case *parse.EmbedNode:
		tpl, ctx, err := s.walkInclude(node.IncludeNode)
		if err != nil {
			return err
		}
		s := newState(s.out, ctx, s.loader)
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
		// TODO: Assert with is a hash?
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

func (s *state) walkExpr(exp parse.Expr) (v Value, e error) {
	switch exp := exp.(type) {
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
	}
	return
}

func execute(name string, out io.Writer, ctx map[string]Value, loader Loader) error {
	s := newState(out, ctx, loader)
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
