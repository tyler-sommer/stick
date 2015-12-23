package stick

import (
	"errors"
	"github.com/tyler-sommer/stick/parse"
	"io"
	"reflect"
)

type state struct {
	out     io.Writer
	node    parse.Node
	context map[string]*variable
}

func newState(out io.Writer, ctx map[string]*variable) *state {
	return &state{out, nil, ctx}
}

type variable struct {
	reflect.Value
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

		io.WriteString(s.out, v.String())
	default:
		return errors.New("Unknown node " + node.String())
	}

	return nil
}

func (s *state) walkExpr(exp parse.Expr) (v *variable, e error) {
	switch exp := exp.(type) {
	case *parse.NameExpr:
		if val, ok := s.context[exp.Name()]; ok {
			v = val
		} else {
			e = errors.New("Undefined variable \"" + exp.Name() + "\"")
		}
	}
	return
}

func Execute(tmpl string, out io.Writer, ctx map[string]interface{}) error {
	tree, err := parse.Parse(tmpl)
	if err != nil {
		return err
	}

	sctx := make(map[string]*variable)
	for k, v := range ctx {
		sctx[k] = &variable{reflect.ValueOf(v)}
	}

	s := newState(out, sctx)
	err = s.walk(tree.Root())
	if err != nil {
		return err
	}
	return nil
}
