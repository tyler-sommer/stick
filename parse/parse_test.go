package parse

import (
	"testing"
)

type parseTest struct {
	name     string
	input    string
	expected *ModuleNode
	err      string
}

const noError = ""

func newParseTest(name, input string, expected *ModuleNode) parseTest {
	return parseTest{name, input, expected, noError}
}

func newErrorTest(name, input string, err string) parseTest {
	return parseTest{name, input, mkModule(), err}
}

func mkModule(nodes ...Node) *ModuleNode {
	l := newModuleNode()
	for _, n := range nodes {
		l.append(n)
	}

	return l
}

var parseTests = []parseTest{
	newErrorTest("parse error", "{% block test %}", "Unclosed tag block"),
	newParseTest("text", "some text", mkModule(newTextNode("some text", 0))),
	newParseTest("hello", "Hello {{ name }}", mkModule(newTextNode("Hello ", 0), newPrintNode(newNameExpr("name"), 6))),
	newParseTest("string expr", "Hello {{ 'Tyler' }}", mkModule(newTextNode("Hello ", 0), newPrintNode(newStringExpr("Tyler"), 6))),
	newParseTest(
		"simple tag",
		"{% block something %}Body{% endblock %}",
		mkModule(newBlockNode("something", mkModule(newTextNode("Body", 0)), 0)),
	),
	newParseTest(
		"if else",
		"{% if something %}Do Something{% else %}Another thing{% endif %}",
		mkModule(newIfNode(newNameExpr("something"), mkModule(newTextNode("Do Something", 0)), mkModule(newTextNode("Another thing", 0)), 0)),
	),
	newParseTest(
		"function expr",
		"{{ func('arg1', arg2) }}",
		mkModule(newPrintNode(newFuncExpr("func", []expr{newStringExpr("arg1"), newNameExpr("arg2")}), 0)),
	),
}

func nodeEqual(a, b Node) bool {
	if a.String() != b.String() {
		return false
	}

	return true
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		tree, err := Parse(test.input)
		if test.err != noError && err != nil && test.err != err.Error() {
			t.Errorf("%s: got error\n\t%+v\nexpected error\n\t%v", test.name, err, test.err)
		} else if !nodeEqual(tree.root, test.expected) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, tree.root, test.expected)
			if err != nil {
				t.Errorf("%s: got error\n\t%v", test.name, err.Error())
			}
		}
	}
}
