package parse

import (
	"fmt"
	"testing"
)

type parseTest struct {
	name     string
	input    string
	expected Node
}

func mkModule(nodes ...Node) Node {
	l := newModuleNode()
	for _, n := range nodes {
		l.append(n)
	}

	return l
}

var parseTests = []parseTest{
	{"text", "some text", mkModule(newTextNode("some text", 0))},
	{"hello", "Hello {{ name }}", mkModule(newTextNode("Hello ", 0), newPrintNode(newNameExpr("name"), 6))},
	{"string expr", "Hello {{ 'Tyler' }}", mkModule(newTextNode("Hello ", 0), newPrintNode(newStringExpr("Tyler"), 6))},
	{
		"simple tag",
		"{% block something %}Body{% endblock %}",
		mkModule(newBlockNode(newNameExpr("something"), mkModule(newTextNode("Body", 0)), 0)),
	},
	{
		"if else",
		"{% if something %}Do Something{% else %}Another thing{% endif %}",
		mkModule(newIfNode(newNameExpr("something"), mkModule(newTextNode("Do Something", 0)), mkModule(newTextNode("Another thing", 0)), 0)),
	},
}

func nodeEqual(a, b Node) bool {
	if a.String() != b.String() {
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
