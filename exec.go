package stick

import (
	"errors"
	"github.com/tyler-sommer/stick/parse"
	"io"
	"strconv"
	"fmt"
)

type state struct {
	out     io.Writer
	node    parse.Node
	context map[string]Value
}

func newState(out io.Writer, ctx map[string]Value) *state {
	return &state{out, nil, ctx}
}

func (s *state) walk(node parse.Node) error {
	s.node = node
	switch node := node.(type) {
	case *parse.ModuleNode:
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
	default:
		return errors.New("Unknown node " + node.String())
	}

	return nil
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

func Execute(tmpl string, out io.Writer, ctx map[string]Value) error {
	tree, err := parse.Parse(tmpl)
	if err != nil {
		return err
	}

	s := newState(out, ctx)
	err = s.walk(tree.Root())
	if err != nil {
		return err
	}
	return nil
}
