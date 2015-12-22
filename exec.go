package stick

import (
	"github.com/tyler-sommer/stick/parse"
	"io"
	"reflect"
)

type state struct {
	out     io.Writer
	node    parse.Node
	context map[string]variable
}

func newState(out io.Writer) (s *state) {
	s = &state{}
	s.out = out
	return s
}

type variable struct {
	val reflect.Value
}

func (s *state) walk(node parse.Node) {
	s.node = node
	switch node := node.(type) {
	case *parse.ModuleNode:
		for _, c := range node.Children() {
			s.walk(c)
		}
	case *parse.TextNode:
		io.WriteString(s.out, node.Text())
	default:
		io.WriteString(s.out, "Unknown node "+node.String())
	}
}

func Execute(tmpl string, out io.Writer) {
	tree, err := parse.Parse(tmpl)
	if err != nil {
		panic(err)
	}

	s := newState(out)

	s.walk(tree.Root())
}
