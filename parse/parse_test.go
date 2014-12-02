package parse

import (
	"testing"
	"fmt"
)

type parseTest struct {
	name string
	input string
	expected node
}

func mkModule(nodes []node) node {
	l := newModuleNode()
	for _, n := range nodes {
		l.append(n)
	}

	return node(l)
}

var parseTests = []parseTest{
	{"text", "some text", mkModule([]node{newTextNode([]byte("some text"), 0)})},
	{"hello", "Hello {{ name }}", mkModule([]node{newTextNode([]byte("Hello "), 0), newPrintNode(expr(newNameExpr("name")), 6)})},
	{"simple tag", "{% block something %}Body{% endblock %}", mkModule([]node{newTagNode("block", newTextNode([]byte("Body"), 0), map[string]expr{"name": newNameExpr("name")}, 0)})},
}

func nodeEqual(a, b node) bool {
	if (a.String() != b.String()) {
		return false
	}

	return true
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		tree := Parse(test.input)
		fmt.Println(tree.root)
		if !nodeEqual(tree.root, test.expected) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, tree.root, test.expected)
		}
	}
}
