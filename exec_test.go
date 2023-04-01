package stick

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/tyler-sommer/stick/parse"
)

// A testValidator checks that the result of a test execution is as expected.
type testValidator func(actual string, err error) error

// execTest is a configurable template execution test.
type execTest struct {
	name string
	tpl  string
	ctx  map[string]Value

	checkResult testValidator

	visitNode func(parse.Node) // When visitNode is set, it will be called after each node is parsed.
}

type testOption func(t *execTest)

func newExecTest(name string, tpl string, v testValidator, opts ...testOption) execTest {
	t := execTest{name: name, tpl: tpl, checkResult: v}
	for _, o := range opts {
		o(&t)
	}
	return t
}

// withContext sets the context variables for the template.
func withContext(ctx map[string]Value) testOption {
	return func(t *execTest) {
		t.ctx = ctx
	}
}

// withNodeVisitor enhances a test with the ability to inspect parsed nodes.
//
// For the purposes of this testing, this provides the ability to muck around with the internal
// structure of a template.  This can be used to create an invalid tree that would not normally
// be parsed, but may be necessary to test a certain error condition.
func withNodeVisitor(visitFunc func(parse.Node)) testOption {
	return func(t *execTest) {
		t.visitNode = visitFunc
	}
}

var tests = []execTest{
	newExecTest("Hello, World", "Hello, World!", expect("Hello, World!")),
	newExecTest("Hello, Tyler!", "Hello, {{ name }}!", expect("Hello, Tyler!"), withContext(map[string]Value{"name": "Tyler"})),
	newExecTest("Simple if", `<li class="{% if active %}active{% endif %}">`, expect(`<li class="active">`), withContext(map[string]Value{"active": true})),
	newExecTest("Simple inheritance", `{% extends 'Hello, {% block test %}universe{% endblock %}!' %}{% block test %}world{% endblock %}`, expect(`Hello, world!`)),
	newExecTest("Simple include", `This is a test. {% include 'Hello, {{ name }}!' %} This concludes the test.`, expect(`This is a test. Hello, John! This concludes the test.`), withContext(map[string]Value{"name": "John"})),
	newExecTest("Include with", `{% include 'Hello, {{ name }}{{ value }}' with vars %}`, expect(`Hello, Adam!`), withContext(map[string]Value{"value": "!", "vars": map[string]Value{"name": "Adam"}})),
	newExecTest("Include with literal", `{% include 'Hello, {{ name }}{{ value }}' with {"name": "world", "value": "!"} only %}`, expect(`Hello, world!`)),
	newExecTest("Embed", `Well. {% embed 'Hello, {% block name %}World{% endblock %}!' %}{% block name %}Tyler{% endblock %}{% endembed %}`, expect(`Well. Hello, Tyler!`)),
	newExecTest("Constant null", `{% if test == null %}Yes{% else %}no{% endif %}`, expect(`Yes`), withContext(map[string]Value{"test": nil})),
	newExecTest("Constant bool", `{% if test == true %}Yes{% else %}no{% endif %}`, expect(`no`), withContext(map[string]Value{"test": false})),
	newExecTest("Chained attributes", `{{ entity.attr.Name }}`, expect(`Tyler`), withContext(map[string]Value{"entity": map[string]Value{"attr": struct{ Name string }{"Tyler"}}})),
	newExecTest("Attribute method call", `{{ entity.Name('lower') }}`, expect(`lowerJohnny`), withContext(map[string]Value{"entity": &fakePerson{"Johnny"}})),
	newExecTest("For loop", `{% for i in 1..3 %}{{ i }}{% endfor %}`, expect(`123`)),
	newExecTest(
		"For loop with inner loop",
		`{% for i in test %}{% for j in i %}{{ j }}{{ loop.index }}{{ loop.parent.index }}{% if loop.first %},{% endif %}{% if loop.last %};{% endif %}{% endfor %}{% if loop.first %}f{% endif %}{% if loop.last %}l{% endif %}:{% endfor %}`,
		expect(`111,221331;f:412,522632;:713,823933;l:`),
		withContext(map[string]Value{
			"test": [][]int{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9}},
		}),
	),
	newExecTest(
		"For loop variables",
		`{% for i in 1..3 %}{{ i }}{{ loop.index }}{{ loop.index0 }}{{ loop.revindex }}{{ loop.revindex0 }}{{ loop.length }}{% if loop.first %}f{% endif %}{% if loop.last %}l{% endif %}{% endfor %}`,
		expect(`110323f221213332103l`),
	),
	newExecTest("For else", `{% for i in emptySet %}{{ i }}{% else %}No results.{% endfor %}`, expect(`No results.`), withContext(map[string]Value{"emptySet": []int{}})),
	newExecTest(
		"For map",
		`{% for k, v in data %}Record {{ loop.index }}: {{ k }}: {{ v }}{% if not loop.last %} - {% endif %}{% endfor %}`,
		expect(`Record 1: Group A: 5.12 - Record 2: Group B: 5.09`, `Record 1: Group B: 5.09 - Record 2: Group A: 5.12`),
		withContext(map[string]Value{"data": map[string]float64{"Group A": 5.12, "Group B": 5.09}}),
	),
	newExecTest(
		"Some operators",
		`{{ 4.5 * 10 }} - {{ 3 + true }} - {{ 3 + 4 == 7.0 }} - {{ 10 % 2 == 0 }} - {{ 10 ** 2 > 99.9 and 10 ** 2 <= 100 }}`,
		expect(`45 - 4 - 1 - 1 - 1`),
	),
	newExecTest("In and not in", `{{ 5 in set and 4 not in set }}`, expect(`1`), withContext(map[string]Value{"set": []int{5, 10}})),
	newExecTest("Function call", `{{ multiply(num, 5) }}`, expect(`50`), withContext(map[string]Value{"num": 10})),
	newExecTest("Filter call", `Welcome, {{ name }}`, expect(`Welcome, `)),
	newExecTest("Filter call", `Welcome, {{ name|default('User') }}`, expect(`Welcome, User`), withContext(map[string]Value{"name": nil})),
	newExecTest("Filter call", `Welcome, {{ surname|default('User') }}`, expect(`Welcome, User`), withContext(map[string]Value{"name": nil})),
	newExecTest(
		"Basic use statement",
		`{% extends '{% block message %}{% endblock %}' %}{% use '{% block message %}Hello{% endblock %}' %}`,
		expect("Hello"),
	),
	newExecTest(
		"Extended use statement",
		`{% extends '{% block message %}{% endblock %}' %}{% use '{% block message %}Hello{% endblock %}' with message as base_message %}{% block message %}{{ block('base_message') }}, World!{% endblock %}`,
		expect("Hello, World!"),
	),
	newExecTest(
		"Set statement",
		`{% set val = 'a value' %}{{ val }}`,
		expect("a value"),
	),
	newExecTest(
		"Set statement with body",
		`{% set var1 = 'Hello,' %}{% set var2 %}{{ var1 }} World!{% endset %}{{ var2 }}`,
		expect("Hello, World!"),
	),
	newExecTest(
		"Set statement evaluates once",
		`{% set var0 = 0 %}{% set var1 = 'Hello,' %}
{% macro tester(var0, var1) %}
{% set var0 = var0 + 1 %}
{% endmacro %}
{% set var2 %}{{ _self.tester(var0, var1) }}{{ var1 }} World!{% endset %}
{{ var2 }}
{{ var0 }}
{{ var2 }}
{{ var0 }}`,
		expectContains("Hello, World!\n1\n\n\nHello, World!\n1"),
	),
	newExecTest(
		"Set statement invalid expr type",
		`{% set v = 10 %}`,
		expectErrorContains("unable to evaluate unsupported Expr type: *parse.TextNode"),
		withNodeVisitor(func(n parse.Node) {
			if sn, ok := n.(*parse.SetNode); ok {
				sn.X = parse.NewTextNode("Hello", sn.X.Start())
			}
		}),
	),
	newExecTest(
		"Do statement",
		`{% do p.Name('Mister ') %}{{ p.Name('') }}`,
		expect("Mister Meeseeks"),
		withContext(map[string]Value{"p": &fakePerson{"Meeseeks"}}),
	),
	newExecTest(
		"Filter statement",
		`{% filter upper %}hello, world!{% endfilter %}`,
		expect("HELLO, WORLD!"),
	),
	newExecTest(
		"Import statement",
		`{% import 'macros.twig' as mac %}{{ mac.test("hi") }}`,
		expect("test: hi"),
	),
	newExecTest(
		"From statement",
		`{% from 'macros.twig' import test, def as other %}{{ other("", "HI!") }}`,
		expect("HI!"),
	),
	newExecTest(
		"Ternary if",
		`{{ false ? (true ? "Hello" : "World") : "Words" }}`,
		expect("Words"),
	),
	newExecTest(
		"Hash literal",
		`{{ {"test": 1}["test"] }}`,
		expect("1"),
	),
	newExecTest(
		"Another hash literal",
		`{% set v = {quadruple: "to the power of four!", 0: "ew", "0": "it's not that bad"} %}ew? {{ v.0 }} {{ v.quadruple }}`,
		expect("ew? it's not that bad to the power of four!"),
	),
	newExecTest(
		"Array literal",
		`{{ ["test", 1, "bar"][2] }}`,
		expect("bar"),
	),
	newExecTest(
		"Another Array literal",
		`{{ ["test", 1, "bar"].1 }}`,
		expect("1"),
	),
	newExecTest(
		"Comparison with or",
		`{% if item1 == "banana" or item2 == "apple" %}At least one item is correct{% else %}neither item is correct{% endif %}`,
		expect("At least one item is correct"),
		withContext(map[string]Value{"item1": "orange", "item2": "apple"}),
	),
	newExecTest(
		"Non-existent map element without default",
		`{{ data.A }} {{ data.NotThere }} {{ data.B }}`,
		expect("Foo  Bar"),
		withContext(map[string]Value{"data": map[string]string{"A": "Foo", "B": "Bar"}}),
	),
	newExecTest(
		"Non-existent map element with default",
		`{{ data.A }} {{ data.NotThere|default("default value") }} {{ data.B }}`,
		expect("Foo default value Bar"),
		withContext(map[string]Value{"data": map[string]string{"A": "Foo", "B": "Bar"}}),
	),
	newExecTest(
		"Accessing templateName on _self",
		`Template: {{ _self.templateName }}`,
		expect("Template: Template: {{ _self.templateName }}"),
	),
	newExecTest(
		"Unsupported binary operator",
		`{{ 1 + 2 }}`,
		expectErrorContains("unsupported binary operator: _"),
		withNodeVisitor(func(n parse.Node) {
			if bn, ok := n.(*parse.BinaryExpr); ok {
				bn.Op = "_"
			}
		}),
	),
	newExecTest(
		"Unsupported binary operator",
		`{{ 1 + 2 }}`,
		expectErrorContains("unable to evaluate unsupported Expr type: *parse.TextNode"),
		withNodeVisitor(func(n parse.Node) {
			if bn, ok := n.(*parse.BinaryExpr); ok {
				bn.Right = parse.NewTextNode("foo", bn.Right.Start())
			}
		}),
	),
}

func joinExpected(expected []string) string {
	res := ""
	for i, e := range expected {
		if i != 0 {
			res = res + " or "
		}
		res = res + fmt.Sprintf("%#v", e)
	}
	return res
}

// expectMismatchError is an error that describes a test output that does not match what is expected.
type expectMismatchError struct {
	actual   string
	expected []string
	loose    bool
}

func (err *expectMismatchError) Error() string {
	if err.loose {
		return fmt.Sprintf("%#v does not contain %s", err.actual, joinExpected(err.expected))
	}
	return fmt.Sprintf("%#v does not equal %s", err.actual, joinExpected(err.expected))
}

func newExpectFailedError(loose bool, actual string, expected ...string) error {
	return &expectMismatchError{actual, expected, loose}
}

// expectErrorMismatchError is an error that describes a test that does not result in the expected error.
type expectErrorMismatchError struct {
	actual   error
	expected []string
}

func (err *expectErrorMismatchError) Error() string {
	if len(err.expected) == 0 {
		if err.actual == nil {
			// shouldn't happen in practice, but technically possible
			return fmt.Sprint("error mismatch expected but there was no actual error and no error was expected! (bug?)")
		}
		return fmt.Sprintf("unexpected error %#v", err.actual.Error())
	}
	ex := "<nil>"
	if err.actual != nil {
		ex = fmt.Sprintf("%#v", err.actual.Error())
	}
	return fmt.Sprintf("%s is not the expected error %s", ex, joinExpected(err.expected))
}

func newExpectErrorMismatchError(actual error, expected ...string) error {
	return &expectErrorMismatchError{actual, expected}
}

func expectChained(expects ...testValidator) testValidator {
	return func(actual string, err error) error {
		for _, expect := range expects {
			if e := expect(actual, err); e != nil {
				return e
			}
		}
		return nil
	}
}

func expect(expected ...string) testValidator {
	return expectChained(expectNoError(), func(actual string, err error) error {
		for _, exp := range expected {
			if actual == exp {
				return nil
			}
		}
		return newExpectFailedError(false, actual, expected...)
	})
}

func expectContains(matches string) testValidator {
	return expectChained(expectNoError(), func(actual string, err error) error {
		if !strings.Contains(actual, matches) {
			return newExpectFailedError(true, actual, matches)
		}
		return nil
	})
}

func expectNoError() testValidator {
	return func(actual string, err error) error {
		if err != nil {
			return newExpectErrorMismatchError(err)
		}
		return nil
	}
}

func expectErrorContains(expected ...string) testValidator {
	return func(_ string, err error) error {
		if err == nil && len(expected) == 0 {
			return nil
		}
		actual := "<nil>"
		if err != nil {
			actual = err.Error()
		}
		for _, e := range expected {
			if strings.Contains(actual, e) {
				// actual error matches one of the expected values
				return nil
			}
		}
		// no match was found for actual error
		return newExpectErrorMismatchError(err, expected...)
	}
}

type testLoader struct {
	templates map[string]Template
}

func newTestLoader(templates []Template) *testLoader {
	tpls := make(map[string]Template)
	for i := 0; i < len(templates); i++ {
		tpl := templates[i]
		tpls[tpl.Name()] = tpl
	}
	return &testLoader{tpls}
}

type testTemplate struct {
	name     string
	contents string
}

func (t *testTemplate) Name() string {
	return t.name
}

func (t *testTemplate) Contents() io.Reader {
	return bytes.NewReader([]byte(t.contents))
}

func tpl(name, content string) *testTemplate {
	return &testTemplate{name, content}
}

func (t *testLoader) Load(name string) (Template, error) {
	if b, ok := t.templates[name]; ok {
		return b, nil
	}
	return tpl(name, name), nil
}

// testVisitor is a parse.NodeVisitor to enable tests to arbitrarily modify a parsed tree.
type testVisitor struct {
	// visit is called when the visitor leaves a node
	visit func(parse.Node)
}

func (v *testVisitor) Enter(parse.Node) {
	// only Leave is used for simplicity's sake
}

func (v *testVisitor) Leave(n parse.Node) {
	if v.visit != nil {
		v.visit(n)
	}
}

func evaluateTest(t *testing.T, env *Env, test execTest) {
	w := &bytes.Buffer{}
	err := execute(test.tpl, w, test.ctx, env)

	out := w.String()
	if err := test.checkResult(out, err); err != nil {
		t.Errorf("%s: %s", test.name, err)
	}
}

func TestExec(t *testing.T) {
	env := New(newTestLoader(
		[]Template{
			tpl("macros.twig", `
{% macro test(arg) %}test: {{ arg }}{% endmacro %}

{% macro def(val, default) %}{% if not val %}{{ default }}{% else %}{{ val }}{% endif %}{% endmacro %}
`),
		},
	))
	env.Functions["multiply"] = func(ctx Context, args ...Value) Value {
		if len(args) != 2 {
			return 0
		}
		return CoerceNumber(args[0]) * CoerceNumber(args[1])
	}
	env.Filters["upper"] = func(ctx Context, val Value, args ...Value) Value {
		return strings.ToUpper(CoerceString(val))
	}
	env.Filters["default"] = func(ctx Context, val Value, args ...Value) Value {
		var d Value
		if len(args) == 0 {
			d = nil
		} else {
			d = args[0]
		}
		if CoerceString(val) == "" {
			return d
		}
		return val
	}
	tv := &testVisitor{}
	env.Visitors = append(env.Visitors, tv)
	for _, test := range tests {
		tv.visit = test.visitNode
		evaluateTest(t, env, test)
	}
}

type fakePerson struct {
	name string
}

func (p *fakePerson) Name(prefix string) string {
	p.name = prefix + p.name
	return p.name
}
