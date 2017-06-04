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
	l := NewModuleNode("", nodes...)

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
	newParseTest("text", "some text", mkModule(NewTextNode("some text", noPos))),
	newParseTest("hello", "Hello {{ name }}", mkModule(NewTextNode("Hello ", noPos), NewPrintNode(NewNameExpr("name", noPos), noPos))),
	newParseTest("string expr", "Hello {{ 'Tyler' }}", mkModule(NewTextNode("Hello ", noPos), NewPrintNode(NewStringExpr("Tyler", noPos), noPos))),
	newParseTest(
		"string interpolation",
		`{{ "Hello, #{greeting} #{name|titlecase}." }}`,
		mkModule(NewPrintNode(
			NewBinaryExpr(
				NewBinaryExpr(
					NewBinaryExpr(
						NewBinaryExpr(NewStringExpr("Hello, ", noPos), OpBinaryConcat, NewNameExpr("greeting", noPos), noPos),
						OpBinaryConcat,
						NewStringExpr(" ", noPos), noPos),
					OpBinaryConcat,
					NewFilterExpr("titlecase", []Expr{NewNameExpr("name", noPos)}, noPos), noPos),
				OpBinaryConcat,
				NewStringExpr(".", noPos), noPos),
			noPos)),
	),
	newParseTest(
		"simple tag",
		"{% block something %}Body{% endblock %}",
		mkModule(NewBlockNode("something", NewBodyNode(noPos, NewTextNode("Body", noPos)), noPos)),
	),
	newParseTest(
		"if",
		"{% if something %}Do Something{% endif %}",
		mkModule(NewIfNode(NewNameExpr("something", noPos), NewBodyNode(noPos, NewTextNode("Do Something", noPos)), NewBodyNode(noPos), noPos)),
	),
	newParseTest(
		"if else",
		"{% if something %}Do Something{% else %}Another thing{% endif %}",
		mkModule(NewIfNode(NewNameExpr("something", noPos), NewBodyNode(noPos, NewTextNode("Do Something", noPos)), NewBodyNode(noPos, NewTextNode("Another thing", noPos)), noPos)),
	),
	newParseTest(
		"if elseif else",
		"{% if something %}Do Something{% elseif another %}Another thing{% else %}Final thing{% endif %}",
		mkModule(NewIfNode(
			NewNameExpr("something", noPos),
			NewBodyNode(noPos, NewTextNode("Do Something", noPos)),
			NewBodyNode(noPos, NewIfNode(
				NewNameExpr("another", noPos),
				NewBodyNode(noPos, NewTextNode("Another thing", noPos)),
				NewBodyNode(noPos, NewTextNode("Final thing", noPos)),
				noPos)),
			noPos)),
	),
	newParseTest(
		"nested if",
		"{% if something %}Do {% if another %}something {% endif %} finally{% endif %}",
		mkModule(NewIfNode(
			NewNameExpr("something", noPos),
			NewBodyNode(noPos,
				NewTextNode("Do ", noPos),
				NewIfNode(
					NewNameExpr("another", noPos),
					NewBodyNode(noPos, NewTextNode("something ", noPos)),
					NewBodyNode(noPos),
					noPos),
				NewTextNode(" finally", noPos)),
			NewBodyNode(noPos),
			noPos)),
	),
	newParseTest(
		"function expr",
		"{{ func('arg1', arg2) }}",
		mkModule(NewPrintNode(NewFuncExpr("func", []Expr{NewStringExpr("arg1", noPos), NewNameExpr("arg2", noPos)}, noPos), noPos)),
	),
	newParseTest(
		"extends statement",
		"{% extends '::base.html.twig' %}",
		mkModule(NewExtendsNode(NewStringExpr("::base.html.twig", noPos), noPos)),
	),
	newParseTest(
		"basic binary operation",
		"{{ something + else }}",
		mkModule(NewPrintNode(NewBinaryExpr(NewNameExpr("something", noPos), OpBinaryAdd, NewNameExpr("else", noPos), noPos), noPos)),
	),
	newParseTest(
		"number literal binary operation",
		"{{ 4.123 + else - test }}",
		mkModule(NewPrintNode(NewBinaryExpr(NewBinaryExpr(NewNumberExpr("4.123", noPos), OpBinaryAdd, NewNameExpr("else", noPos), noPos), OpBinarySubtract, NewNameExpr("test", noPos), noPos), noPos)),
	),
	newParseTest(
		"parenthesis grouping expression",
		"{{ (4 + else) / 10 }}",
		mkModule(NewPrintNode(NewBinaryExpr(NewGroupExpr(NewBinaryExpr(NewNumberExpr("4", noPos), OpBinaryAdd, NewNameExpr("else", noPos), noPos), noPos), OpBinaryDivide, NewNumberExpr("10", noPos), noPos), noPos)),
	),
	newParseTest(
		"correct order of operations",
		"{{ 10 + 5 / 5 }}",
		mkModule(NewPrintNode(NewBinaryExpr(NewNumberExpr("10", noPos), OpBinaryAdd, NewBinaryExpr(NewNumberExpr("5", noPos), OpBinaryDivide, NewNumberExpr("5", noPos), noPos), noPos), noPos)),
	),
	newParseTest(
		"correct ** associativity",
		"{{ 10 ** 2 ** 5 }}",
		mkModule(NewPrintNode(NewBinaryExpr(NewNumberExpr("10", noPos), OpBinaryPower, NewBinaryExpr(NewNumberExpr("2", noPos), OpBinaryPower, NewNumberExpr("5", noPos), noPos), noPos), noPos)),
	),
	newParseTest(
		"extended binary expression",
		"{{ 5 + 10 + 15 * 12 / 4 }}",
		mkModule(NewPrintNode(NewBinaryExpr(NewBinaryExpr(NewNumberExpr("5", noPos), OpBinaryAdd, NewNumberExpr("10", noPos), noPos), OpBinaryAdd, NewBinaryExpr(NewBinaryExpr(NewNumberExpr("15", noPos), OpBinaryMultiply, NewNumberExpr("12", noPos), noPos), OpBinaryDivide, NewNumberExpr("4", noPos), noPos), noPos), noPos)),
	),
	newParseTest(
		"unary not expression",
		"{{ not something }}",
		mkModule(NewPrintNode(NewUnaryExpr(OpUnaryNot, NewNameExpr("something", noPos), noPos), noPos)),
	),
	newParseTest(
		"dot notation accessor",
		"{{ something.another.else }}",
		mkModule(NewPrintNode(NewGetAttrExpr(NewGetAttrExpr(NewNameExpr("something", noPos), NewStringExpr("another", noPos), []Expr{}, noPos), NewStringExpr("else", noPos), []Expr{}, noPos), noPos)),
	),
	newParseTest(
		"explicit method call",
		"{{ something.doThing('arg1', arg2) }}",
		mkModule(NewPrintNode(NewGetAttrExpr(NewNameExpr("something", noPos), NewStringExpr("doThing", noPos), []Expr{NewStringExpr("arg1", noPos), NewNameExpr("arg2", noPos)}, noPos), noPos)),
	),
	newParseTest(
		"dot notation array access mix",
		"{{ something().another['test'].further }}",
		mkModule(NewPrintNode(NewGetAttrExpr(NewGetAttrExpr(NewGetAttrExpr(NewFuncExpr("something", []Expr{}, noPos), NewStringExpr("another", noPos), []Expr{}, noPos), NewStringExpr("test", noPos), []Expr{}, noPos), NewStringExpr("further", noPos), []Expr{}, noPos), noPos)),
	),
	newParseTest(
		"basic filter",
		"{{ something|default('another') }}",
		mkModule(NewPrintNode(NewFilterExpr("default", []Expr{NewNameExpr("something", noPos), NewStringExpr("another", noPos)}, noPos), noPos)),
	),
	newParseTest(
		"filter with no args",
		"{{ something|default }}",
		mkModule(NewPrintNode(NewFilterExpr("default", []Expr{NewNameExpr("something", noPos)}, noPos), noPos)),
	),
	newParseTest(
		"basic for loop",
		"{% for val in 1..10 %}body{% endfor %}",
		mkModule(NewForNode("", "val", NewBinaryExpr(NewNumberExpr("1", noPos), OpBinaryRange, NewNumberExpr("10", noPos), noPos), NewBodyNode(noPos, NewTextNode("body", noPos)), NewBodyNode(noPos), noPos)),
	),
	newParseTest(
		"for loop",
		"{% for k, val in something if val %}body{% else %}No results.{% endfor %}",
		mkModule(NewForNode("k", "val", NewNameExpr("something", noPos), NewIfNode(NewNameExpr("val", noPos), NewBodyNode(noPos, NewTextNode("body", noPos)), nil, noPos), NewBodyNode(noPos, NewTextNode("No results.", noPos)), noPos)),
	),
	newParseTest(
		"include",
		"{% include '::_subnav.html.twig' %}",
		mkModule(NewIncludeNode(NewStringExpr("::_subnav.html.twig", noPos), nil, false, noPos)),
	),
	newParseTest(
		"include with",
		"{% include '::_subnav.html.twig' with var %}",
		mkModule(NewIncludeNode(NewStringExpr("::_subnav.html.twig", noPos), NewNameExpr("var", noPos), false, noPos)),
	),
	newParseTest(
		"include with only",
		"{% include '::_subnav.html.twig' with var only %}",
		mkModule(NewIncludeNode(NewStringExpr("::_subnav.html.twig", noPos), NewNameExpr("var", noPos), true, noPos)),
	),
	newParseTest(
		"include only",
		"{% include '::_subnav.html.twig' only %}",
		mkModule(NewIncludeNode(NewStringExpr("::_subnav.html.twig", noPos), nil, true, noPos)),
	),
	newParseTest(
		"embed",
		"{% embed '::_modal.html.twig' %}{% block title %}Hello{% endblock %}{% endembed  %}",
		mkModule(NewEmbedNode(NewStringExpr("::_modal.html.twig", noPos), nil, false, map[string]*BlockNode{"title": NewBlockNode("title", NewBodyNode(noPos, NewTextNode("Hello", noPos)), noPos)}, noPos)),
	),
	newParseTest(
		"null",
		"{{ null }}{{ NONE }}{{ NULL }}{{ none }}",
		mkModule(NewPrintNode(NewNullExpr(noPos), noPos), NewPrintNode(NewNullExpr(noPos), noPos), NewPrintNode(NewNullExpr(noPos), noPos), NewPrintNode(NewNullExpr(noPos), noPos)),
	),
	newParseTest(
		"boolean",
		"{{ true }}{{ TRUE }}{{ false }}{{ FALSE }}",
		mkModule(NewPrintNode(NewBoolExpr(true, noPos), noPos), NewPrintNode(NewBoolExpr(true, noPos), noPos), NewPrintNode(NewBoolExpr(false, noPos), noPos), NewPrintNode(NewBoolExpr(false, noPos), noPos)),
	),
	newParseTest(
		"unary then getattr",
		"{% if not loop.last %}{% endif %}",
		mkModule(NewIfNode(NewUnaryExpr(OpUnaryNot, NewGetAttrExpr(NewNameExpr("loop", noPos), NewStringExpr("last", noPos), []Expr{}, noPos), noPos), NewBodyNode(noPos), NewBodyNode(noPos), noPos)),
	),
	newParseTest(
		"binary expr if",
		"<option{% if script.Type == 'Javascript' %} selected{% endif %}>Javascript</option>",
		mkModule(NewTextNode("<option", noPos), NewIfNode(NewBinaryExpr(NewGetAttrExpr(NewNameExpr("script", noPos), NewStringExpr("Type", noPos), []Expr{}, noPos), OpBinaryEqual, NewStringExpr("Javascript", noPos), noPos), NewBodyNode(noPos, NewTextNode(" selected", noPos)), NewBodyNode(noPos), noPos), NewTextNode(">Javascript</option>", noPos)),
	),
	newParseTest(
		"test parsing",
		"{{ animal is mammal }}{{ 10 is not divisible by(3) }}",
		mkModule(NewPrintNode(NewBinaryExpr(NewNameExpr("animal", noPos), OpBinaryIs, NewTestExpr("mammal", []Expr{}, noPos), noPos), noPos), NewPrintNode(NewBinaryExpr(NewNumberExpr("10", noPos), OpBinaryIsNot, NewTestExpr("divisible by", []Expr{NewNumberExpr("3", noPos)}, noPos), noPos), noPos)),
	),
	newParseTest(
		"comment",
		"But{# This is a test #} not this.",
		mkModule(NewTextNode("But", noPos), NewCommentNode(" This is a test ", noPos), NewTextNode(" not this.", noPos)),
	),
	newParseTest(
		"use statement",
		"{% use '::blocks.html.twig' %}",
		mkModule(NewUseNode(NewStringExpr("::blocks.html.twig", noPos), map[string]string{}, noPos)),
	),
	newParseTest(
		"use statement with one alias",
		"{% use '::blocks.html.twig' with form_input as base_input %}",
		mkModule(NewUseNode(NewStringExpr("::blocks.html.twig", noPos), map[string]string{"form_input": "base_input"}, noPos)),
	),
	newParseTest(
		"use statement with muiltiple aliases",
		"{% use '::blocks.html.twig' with form_input as base_input, form_radio as base_radio %}",
		mkModule(NewUseNode(NewStringExpr("::blocks.html.twig", noPos), map[string]string{"form_input": "base_input", "form_radio": "base_radio"}, noPos)),
	),
	newParseTest(
		"set statement",
		"{% set varn = 'test' %}",
		mkModule(NewSetNode("varn", NewStringExpr("test", noPos), noPos)),
	),
	newParseTest(
		"do statement",
		"{% do somefunc() %}",
		mkModule(NewDoNode(NewFuncExpr("somefunc", []Expr{}, noPos), noPos)),
	),
	newParseTest(
		"filter statement",
		"{% filter upper|escape %}Some text{% endfilter %}",
		mkModule(NewFilterNode([]string{"upper", "escape"}, NewBodyNode(noPos, NewTextNode("Some text", noPos)), noPos)),
	),
	newParseTest(
		"simple macro",
		"{% macro thing(var1, var2) %}Hello{% endmacro %}",
		mkModule(NewMacroNode("thing", []string{"var1", "var2"}, NewBodyNode(noPos, NewTextNode("Hello", noPos)), noPos)),
	),
	newParseTest(
		"simple macro2",
		"{% macro thing(var2) %}Hello{% endmacro %}",
		mkModule(NewMacroNode("thing", []string{"var2"}, NewBodyNode(noPos, NewTextNode("Hello", noPos)), noPos)),
	),
	newParseTest(
		"import statement",
		"{% import '::macros.html.twig' as mac %}",
		mkModule(NewImportNode(NewStringExpr("::macros.html.twig", noPos), "mac", noPos)),
	),
	newParseTest(
		"from statement",
		"{% from '::macros.html.twig' import input as field, textarea %}",
		mkModule(NewFromNode(NewStringExpr("::macros.html.twig", noPos), map[string]string{"input": "field", "textarea": "textarea"}, noPos)),
	),
	newParseTest(
		"ternary if expression",
		"{{ test ? 'Hello' : 'World' }}",
		mkModule(NewPrintNode(NewTernaryIfExpr(NewNameExpr("test", noPos), NewStringExpr("Hello", noPos), NewStringExpr("World", noPos), noPos), noPos)),
	),
	newParseTest(
		"for loop filter application (#3)",
		"{% for row in items|batch(3, 'No Item') %}{% endfor %}",
		mkModule(NewForNode("", "row", NewFilterExpr("batch", []Expr{NewNameExpr("items", noPos), NewNumberExpr("3", noPos), NewStringExpr("No Item", noPos)}, noPos), NewBodyNode(noPos), NewBodyNode(noPos), noPos)),
	),
	newParseTest(
		"hash literal",
		`{% set v = {"test": 1, "bar": 10,} %}`,
		mkModule(NewSetNode("v", NewHashExpr(noPos, NewKeyValueExpr(NewStringExpr("test", noPos), NewNumberExpr("1", noPos), noPos), NewKeyValueExpr(NewStringExpr("bar", noPos), NewNumberExpr("10", noPos), noPos)), noPos)),
	),
	newParseTest(
		"array literal",
		`{% set v = [1, "bar", 10,] %}`,
		mkModule(NewSetNode("v", NewArrayExpr(noPos, NewNumberExpr("1", noPos), NewStringExpr("bar", noPos), NewNumberExpr("10", noPos)), noPos)),
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
		if e, ok := err.(DebugError); ok {
			t.Errorf("%s: trace:\n%s", test.name, e.Debug())
		}
	} else if !nodeEqual(tree.root, test.expected) {
		t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, tree.root, test.expected)
		if err != nil {
			t.Errorf("%s: got error\n\t%v", test.name, err.Error())
			if e, ok := err.(DebugError); ok {
				t.Errorf("%s: trace:\n%s", test.name, e.Debug())
			}
		}
	}
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		evaluateTest(t, test)
	}
}
