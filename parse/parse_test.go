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
	// Errors
	newErrorTest("unclosed block", "{% block test %}", "parse error: unclosed tag \"block\" starting on line 1, offset 3"),
	newErrorTest("unclosed if", "{% if test %}", "parse error: unclosed tag \"if\" starting on line 1, offset 3"),
	newErrorTest("unexpected end (function call)", "{{ func('arg1'", "parse error: unexpected end of input on line 1, offset 14"),
	newErrorTest("unclosed parenthesis", "{{ func(arg1 }}", "parse error: expected one of [PUNCTUATION, PARENS_CLOSE], got \"PRINT_CLOSE\" on line 1, offset 13"),
	newErrorTest("unexpected punctuation", "{{ func(arg1. arg2) }}", "parse error: unexpected punctuation \".\", expected \",\" on line 1, offset 12"),

	// Valid
	newParseTest("text", "some text", mkModule(newTextNode("some text", pos{1,6}))),
	newParseTest("hello", "Hello {{ name }}", mkModule(newTextNode("Hello ", pos{1,0}), newPrintNode(newNameExpr("name", pos{1,6}), pos{1,6}))),
	newParseTest("string expr", "Hello {{ 'Tyler' }}", mkModule(newTextNode("Hello ", pos{1,0}), newPrintNode(newStringExpr("Tyler", pos{1,6}), pos{1,6}))),
	newParseTest(
		"simple tag",
		"{% block something %}Body{% endblock %}",
		mkModule(newBlockNode("something", mkModule(newTextNode("Body", pos{1,6})), pos{1,6})),
	),
	newParseTest(
		"if else",
		"{% if something %}Do Something{% else %}Another thing{% endif %}",
		mkModule(newIfNode(newNameExpr("something", pos{1,6}), mkModule(newTextNode("Do Something", pos{1,6})), mkModule(newTextNode("Another thing", pos{1,6})), pos{1,6})),
	),
	newParseTest(
		"function expr",
		"{{ func('arg1', arg2) }}",
		mkModule(newPrintNode(newFuncExpr(newNameExpr("func", pos{1,0}), []expr{newStringExpr("arg1", pos{1,6}), newNameExpr("arg2", pos{1,6})}, pos{1,6}), pos{1,6})),
	),
}

func nodeEqual(a, b Node) bool {
	if a.String() != b.String() {
		return false
	}

	return true
}

func evaluateTest(t *testing.T, test parseTest) {
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

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		evaluateTest(t, test)
	}
}
