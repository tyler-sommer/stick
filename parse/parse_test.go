package parse

import (
	"strings"
	"testing"
)

type parseTest struct {
	name     string
	input    string
	expected *ModuleNode
	err      string
}

const noError = ""

// Position testing isnt implemented
var noPos = Pos{0, 0}

func newParseTest(name, input string, expected *ModuleNode) parseTest {
	return parseTest{name, input, expected, noError}
}

func newErrorTest(name, input string, err string) parseTest {
	return parseTest{name, input, mkModule(), err}
}

func mkModule(nodes ...Node) *ModuleNode {
	l := newModuleNode("", nodes...)

	return l
}

var parseTests = []parseTest{
	// Errors
	newErrorTest("unclosed block", "{% block test %}", `unclosed tag "block" starting on line 1, column 3`),
	newErrorTest("unclosed if", "{% if test %}", `unclosed tag "if" starting on line 1, column 3`),
	newErrorTest("unexpected end (function call)", "{{ func('arg1'", `unexpected end of input on line 1, column 14`),
	newErrorTest("unclosed parenthesis", "{{ func(arg1 }}", `expected one of [PUNCTUATION, PARENS_CLOSE], got "ERROR" on line 1, column 13`),
	newErrorTest("unexpected punctuation", "{{ func(arg1? arg2) }}", `expected "PUNCTUATION", got "PARENS_CLOSE"`),

	// Valid
	newParseTest("text", "some text", mkModule(newTextNode("some text", noPos))),
	newParseTest("hello", "Hello {{ name }}", mkModule(newTextNode("Hello ", noPos), newPrintNode(newNameExpr("name", noPos), noPos))),
	newParseTest("string expr", "Hello {{ 'Tyler' }}", mkModule(newTextNode("Hello ", noPos), newPrintNode(newStringExpr("Tyler", noPos), noPos))),
	newParseTest(
		"string interpolation",
		`{{ "Hello, #{greeting} #{name|titlecase}." }}`,
		mkModule(newPrintNode(
			newBinaryExpr(
				newBinaryExpr(
					newBinaryExpr(
						newBinaryExpr(newStringExpr("Hello, ", noPos), OpBinaryConcat, newNameExpr("greeting", noPos), noPos),
						OpBinaryConcat,
						newStringExpr(" ", noPos), noPos),
					OpBinaryConcat,
					newFilterExpr("titlecase", []Expr{newNameExpr("name", noPos)}, noPos), noPos),
				OpBinaryConcat,
				newStringExpr(".", noPos), noPos),
			noPos)),
	),
	newParseTest(
		"simple tag",
		"{% block something %}Body{% endblock %}",
		mkModule(newBlockNode("something", newBodyNode(noPos, newTextNode("Body", noPos)), noPos)),
	),
	newParseTest(
		"if",
		"{% if something %}Do Something{% endif %}",
		mkModule(newIfNode(newNameExpr("something", noPos), newBodyNode(noPos, newTextNode("Do Something", noPos)), newBodyNode(noPos), noPos)),
	),
	newParseTest(
		"if else",
		"{% if something %}Do Something{% else %}Another thing{% endif %}",
		mkModule(newIfNode(newNameExpr("something", noPos), newBodyNode(noPos, newTextNode("Do Something", noPos)), newBodyNode(noPos, newTextNode("Another thing", noPos)), noPos)),
	),
	newParseTest(
		"if elseif else",
		"{% if something %}Do Something{% elseif another %}Another thing{% else %}Final thing{% endif %}",
		mkModule(newIfNode(
			newNameExpr("something", noPos),
			newBodyNode(noPos, newTextNode("Do Something", noPos)),
			newBodyNode(noPos, newIfNode(
				newNameExpr("another", noPos),
				newBodyNode(noPos, newTextNode("Another thing", noPos)),
				newBodyNode(noPos, newTextNode("Final thing", noPos)),
				noPos)),
			noPos)),
	),
	newParseTest(
		"function expr",
		"{{ func('arg1', arg2) }}",
		mkModule(newPrintNode(newFuncExpr("func", []Expr{newStringExpr("arg1", noPos), newNameExpr("arg2", noPos)}, noPos), noPos)),
	),
	newParseTest(
		"extends statement",
		"{% extends '::base.html.twig' %}",
		mkModule(newExtendsNode(newStringExpr("::base.html.twig", noPos), noPos)),
	),
	newParseTest(
		"basic binary operation",
		"{{ something + else }}",
		mkModule(newPrintNode(newBinaryExpr(newNameExpr("something", noPos), OpBinaryAdd, newNameExpr("else", noPos), noPos), noPos)),
	),
	newParseTest(
		"number literal binary operation",
		"{{ 4.123 + else - test }}",
		mkModule(newPrintNode(newBinaryExpr(newBinaryExpr(newNumberExpr("4.123", noPos), OpBinaryAdd, newNameExpr("else", noPos), noPos), OpBinarySubtract, newNameExpr("test", noPos), noPos), noPos)),
	),
	newParseTest(
		"parenthesis grouping expression",
		"{{ (4 + else) / 10 }}",
		mkModule(newPrintNode(newBinaryExpr(newGroupExpr(newBinaryExpr(newNumberExpr("4", noPos), OpBinaryAdd, newNameExpr("else", noPos), noPos), noPos), OpBinaryDivide, newNumberExpr("10", noPos), noPos), noPos)),
	),
	newParseTest(
		"correct order of operations",
		"{{ 10 + 5 / 5 }}",
		mkModule(newPrintNode(newBinaryExpr(newNumberExpr("10", noPos), OpBinaryAdd, newBinaryExpr(newNumberExpr("5", noPos), OpBinaryDivide, newNumberExpr("5", noPos), noPos), noPos), noPos)),
	),
	newParseTest(
		"correct ** associativity",
		"{{ 10 ** 2 ** 5 }}",
		mkModule(newPrintNode(newBinaryExpr(newNumberExpr("10", noPos), OpBinaryPower, newBinaryExpr(newNumberExpr("2", noPos), OpBinaryPower, newNumberExpr("5", noPos), noPos), noPos), noPos)),
	),
	newParseTest(
		"extended binary expression",
		"{{ 5 + 10 + 15 * 12 / 4 }}",
		mkModule(newPrintNode(newBinaryExpr(newBinaryExpr(newNumberExpr("5", noPos), OpBinaryAdd, newNumberExpr("10", noPos), noPos), OpBinaryAdd, newBinaryExpr(newBinaryExpr(newNumberExpr("15", noPos), OpBinaryMultiply, newNumberExpr("12", noPos), noPos), OpBinaryDivide, newNumberExpr("4", noPos), noPos), noPos), noPos)),
	),
	newParseTest(
		"unary not expression",
		"{{ not something }}",
		mkModule(newPrintNode(newUnaryExpr(OpUnaryNot, newNameExpr("something", noPos), noPos), noPos)),
	),
	newParseTest(
		"dot notation accessor",
		"{{ something.another.else }}",
		mkModule(newPrintNode(newGetAttrExpr(newGetAttrExpr(newNameExpr("something", noPos), newStringExpr("another", noPos), []Expr{}, noPos), newStringExpr("else", noPos), []Expr{}, noPos), noPos)),
	),
	newParseTest(
		"explicit method call",
		"{{ something.doThing('arg1', arg2) }}",
		mkModule(newPrintNode(newGetAttrExpr(newNameExpr("something", noPos), newStringExpr("doThing", noPos), []Expr{newStringExpr("arg1", noPos), newNameExpr("arg2", noPos)}, noPos), noPos)),
	),
	newParseTest(
		"dot notation array access mix",
		"{{ something().another['test'].further }}",
		mkModule(newPrintNode(newGetAttrExpr(newGetAttrExpr(newGetAttrExpr(newFuncExpr("something", []Expr{}, noPos), newStringExpr("another", noPos), []Expr{}, noPos), newStringExpr("test", noPos), []Expr{}, noPos), newStringExpr("further", noPos), []Expr{}, noPos), noPos)),
	),
	newParseTest(
		"basic filter",
		"{{ something|default('another') }}",
		mkModule(newPrintNode(newFilterExpr("default", []Expr{newNameExpr("something", noPos), newStringExpr("another", noPos)}, noPos), noPos)),
	),
	newParseTest(
		"filter with no args",
		"{{ something|default }}",
		mkModule(newPrintNode(newFilterExpr("default", []Expr{newNameExpr("something", noPos)}, noPos), noPos)),
	),
	newParseTest(
		"basic for loop",
		"{% for val in 1..10 %}body{% endfor %}",
		mkModule(newForNode("", "val", newBinaryExpr(newNumberExpr("1", noPos), OpBinaryRange, newNumberExpr("10", noPos), noPos), newBodyNode(noPos, newTextNode("body", noPos)), newBodyNode(noPos), noPos)),
	),
	newParseTest(
		"for loop",
		"{% for k, val in something if val %}body{% else %}No results.{% endfor %}",
		mkModule(newForNode("k", "val", newNameExpr("something", noPos), newIfNode(newNameExpr("val", noPos), newBodyNode(noPos, newTextNode("body", noPos)), nil, noPos), newBodyNode(noPos, newTextNode("No results.", noPos)), noPos)),
	),
	newParseTest(
		"include",
		"{% include '::_subnav.html.twig' %}",
		mkModule(newIncludeNode(newStringExpr("::_subnav.html.twig", noPos), nil, false, noPos)),
	),
	newParseTest(
		"include with",
		"{% include '::_subnav.html.twig' with var %}",
		mkModule(newIncludeNode(newStringExpr("::_subnav.html.twig", noPos), newNameExpr("var", noPos), false, noPos)),
	),
	newParseTest(
		"include with only",
		"{% include '::_subnav.html.twig' with var only %}",
		mkModule(newIncludeNode(newStringExpr("::_subnav.html.twig", noPos), newNameExpr("var", noPos), true, noPos)),
	),
	newParseTest(
		"include only",
		"{% include '::_subnav.html.twig' only %}",
		mkModule(newIncludeNode(newStringExpr("::_subnav.html.twig", noPos), nil, true, noPos)),
	),
	newParseTest(
		"embed",
		"{% embed '::_modal.html.twig' %}{% block title %}Hello{% endblock %}{% endembed  %}",
		mkModule(newEmbedNode(newStringExpr("::_modal.html.twig", noPos), nil, false, map[string]*BlockNode{"title": newBlockNode("title", newBodyNode(noPos, newTextNode("Hello", noPos)), noPos)}, noPos)),
	),
	newParseTest(
		"null",
		"{{ null }}{{ NONE }}{{ NULL }}{{ none }}",
		mkModule(newPrintNode(newNullExpr(noPos), noPos), newPrintNode(newNullExpr(noPos), noPos), newPrintNode(newNullExpr(noPos), noPos), newPrintNode(newNullExpr(noPos), noPos)),
	),
	newParseTest(
		"boolean",
		"{{ true }}{{ TRUE }}{{ false }}{{ FALSE }}",
		mkModule(newPrintNode(newBoolExpr(true, noPos), noPos), newPrintNode(newBoolExpr(true, noPos), noPos), newPrintNode(newBoolExpr(false, noPos), noPos), newPrintNode(newBoolExpr(false, noPos), noPos)),
	),
	newParseTest(
		"unary then getattr",
		"{% if not loop.last %}{% endif %}",
		mkModule(newIfNode(newUnaryExpr(OpUnaryNot, newGetAttrExpr(newNameExpr("loop", noPos), newStringExpr("last", noPos), []Expr{}, noPos), noPos), newBodyNode(noPos), newBodyNode(noPos), noPos)),
	),
	newParseTest(
		"binary expr if",
		"<option{% if script.Type == 'Javascript' %} selected{% endif %}>Javascript</option>",
		mkModule(newTextNode("<option", noPos), newIfNode(newBinaryExpr(newGetAttrExpr(newNameExpr("script", noPos), newStringExpr("Type", noPos), []Expr{}, noPos), OpBinaryEqual, newStringExpr("Javascript", noPos), noPos), newBodyNode(noPos, newTextNode(" selected", noPos)), newBodyNode(noPos), noPos), newTextNode(">Javascript</option>", noPos)),
	),
	newParseTest(
		"test parsing",
		"{{ animal is mammal }}{{ 10 is not divisible by(3) }}",
		mkModule(newPrintNode(newBinaryExpr(newNameExpr("animal", noPos), OpBinaryIs, newTestExpr("mammal", []Expr{}, noPos), noPos), noPos), newPrintNode(newBinaryExpr(newNumberExpr("10", noPos), OpBinaryIsNot, newTestExpr("divisible by", []Expr{newNumberExpr("3", noPos)}, noPos), noPos), noPos)),
	),
	newParseTest(
		"comment",
		"But{# This is a test #} not this.",
		mkModule(newTextNode("But", noPos), newCommentNode(" This is a test ", noPos), newTextNode(" not this.", noPos)),
	),
	newParseTest(
		"use statement",
		"{% use '::blocks.html.twig' %}",
		mkModule(newUseNode(newStringExpr("::blocks.html.twig", noPos), map[string]string{}, noPos)),
	),
	newParseTest(
		"use statement with one alias",
		"{% use '::blocks.html.twig' with form_input as base_input %}",
		mkModule(newUseNode(newStringExpr("::blocks.html.twig", noPos), map[string]string{"form_input": "base_input"}, noPos)),
	),
	newParseTest(
		"use statement with muiltiple aliases",
		"{% use '::blocks.html.twig' with form_input as base_input, form_radio as base_radio %}",
		mkModule(newUseNode(newStringExpr("::blocks.html.twig", noPos), map[string]string{"form_input": "base_input", "form_radio": "base_radio"}, noPos)),
	),
	newParseTest(
		"set statement",
		"{% set varn = 'test' %}",
		mkModule(newSetNode("varn", newStringExpr("test", noPos), noPos)),
	),
	newParseTest(
		"do statement",
		"{% do somefunc() %}",
		mkModule(newDoNode(newFuncExpr("somefunc", []Expr{}, noPos), noPos)),
	),
	newParseTest(
		"filter statement",
		"{% filter upper|escape %}Some text{% endfilter %}",
		mkModule(NewFilterNode([]string{"upper", "escape"}, newBodyNode(noPos, newTextNode("Some text", noPos)), noPos)),
	),
	newParseTest(
		"simple macro",
		"{% macro thing(var1, var2) %}Hello{% endmacro %}",
		mkModule(newMacroNode("thing", []string{"var1", "var2"}, newBodyNode(noPos, newTextNode("Hello", noPos)), noPos)),
	),
	newParseTest(
		"simple macro2",
		"{% macro thing(var2) %}Hello{% endmacro %}",
		mkModule(newMacroNode("thing", []string{"var2"}, newBodyNode(noPos, newTextNode("Hello", noPos)), noPos)),
	),
	newParseTest(
		"import statement",
		"{% import '::macros.html.twig' as mac %}",
		mkModule(newImportNode(newStringExpr("::macros.html.twig", noPos), "mac", noPos)),
	),
	newParseTest(
		"from statement",
		"{% from '::macros.html.twig' import input as field, textarea %}",
		mkModule(newFromNode(newStringExpr("::macros.html.twig", noPos), map[string]string{"input": "field", "textarea": "textarea"}, noPos)),
	),
	newParseTest(
		"ternary if expression",
		"{{ test ? 'Hello' : 'World' }}",
		mkModule(newPrintNode(newTernaryIfExpr(newNameExpr("test", noPos), newStringExpr("Hello", noPos), newStringExpr("World", noPos), noPos), noPos)),
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
	if test.err != noError && err != nil && !strings.Contains(err.Error(), test.err) {
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
