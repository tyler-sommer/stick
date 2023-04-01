package stick

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/tyler-sommer/stick/parse"
)

// execTest is an extensible template execution test.
type execTest interface {
	name() string
	tpl() string
	ctx() map[string]Value
	expected() expectedChecker
}

// _execTest is the standard implementation of execTest.
type _execTest struct {
	_name     string
	_tpl      string
	_ctx      map[string]Value
	_expected expectedChecker
}

func (t _execTest) name() string              { return t._name }
func (t _execTest) tpl() string               { return t._tpl }
func (t _execTest) ctx() map[string]Value     { return t._ctx }
func (t _execTest) expected() expectedChecker { return t._expected }

func newExecTest(name string, tpl string, ctx map[string]Value, expected expectedChecker) execTest {
	return _execTest{name, tpl, ctx, expected}
}

// execTestWithPatch is an execTest with a monkey patch function defined.
type execTestWithPatch struct {
	execTest

	patchNode func(parse.Node) // When patchNode is set, it will be called during parsing after each node is parsed.
}

// withPatch enhances a test with the ability to monkey patch the parsed nodes.
//
// Patching is used when a test needs to create alter the internal structure of a template.
// This can be used to create an invalid tree that would not normally be parsed, but may be
// necessary to test a certain error condition.
func withPatch(t execTest, patch func(parse.Node)) execTestWithPatch {
	return execTestWithPatch{t, patch}
}

var tests = []execTest{
	newExecTest("Hello, World", "Hello, World!", nil, expect("Hello, World!")),
	newExecTest("Hello, Tyler!", "Hello, {{ name }}!", map[string]Value{"name": "Tyler"}, expect("Hello, Tyler!")),
	newExecTest("Simple if", `<li class="{% if active %}active{% endif %}">`, map[string]Value{"active": true}, expect(`<li class="active">`)),
	newExecTest("Simple inheritance", `{% extends 'Hello, {% block test %}universe{% endblock %}!' %}{% block test %}world{% endblock %}`, nil, expect(`Hello, world!`)),
	newExecTest("Simple include", `This is a test. {% include 'Hello, {{ name }}!' %} This concludes the test.`, map[string]Value{"name": "John"}, expect(`This is a test. Hello, John! This concludes the test.`)),
	newExecTest("Include with", `{% include 'Hello, {{ name }}{{ value }}' with vars %}`, map[string]Value{"value": "!", "vars": map[string]Value{"name": "Adam"}}, expect(`Hello, Adam!`)),
	newExecTest("Include with literal", `{% include 'Hello, {{ name }}{{ value }}' with {"name": "world", "value": "!"} only %}`, nil, expect(`Hello, world!`)),
	newExecTest("Embed", `Well. {% embed 'Hello, {% block name %}World{% endblock %}!' %}{% block name %}Tyler{% endblock %}{% endembed %}`, nil, expect(`Well. Hello, Tyler!`)),
	newExecTest("Constant null", `{% if test == null %}Yes{% else %}no{% endif %}`, map[string]Value{"test": nil}, expect(`Yes`)),
	newExecTest("Constant bool", `{% if test == true %}Yes{% else %}no{% endif %}`, map[string]Value{"test": false}, expect(`no`)),
	newExecTest("Chained attributes", `{{ entity.attr.Name }}`, map[string]Value{"entity": map[string]Value{"attr": struct{ Name string }{"Tyler"}}}, expect(`Tyler`)),
	newExecTest("Attribute method call", `{{ entity.Name('lower') }}`, map[string]Value{"entity": &fakePerson{"Johnny"}}, expect(`lowerJohnny`)),
	newExecTest("For loop", `{% for i in 1..3 %}{{ i }}{% endfor %}`, nil, expect(`123`)),
	newExecTest(
		"For loop with inner loop",
		`{% for i in test %}{% for j in i %}{{ j }}{{ loop.index }}{{ loop.parent.index }}{% if loop.first %},{% endif %}{% if loop.last %};{% endif %}{% endfor %}{% if loop.first %}f{% endif %}{% if loop.last %}l{% endif %}:{% endfor %}`,
		map[string]Value{
			"test": [][]int{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9}},
		},
		expect(`111,221331;f:412,522632;:713,823933;l:`),
	),
	newExecTest(
		"For loop variables",
		`{% for i in 1..3 %}{{ i }}{{ loop.index }}{{ loop.index0 }}{{ loop.revindex }}{{ loop.revindex0 }}{{ loop.length }}{% if loop.first %}f{% endif %}{% if loop.last %}l{% endif %}{% endfor %}`,
		nil,
		expect(`110323f221213332103l`),
	),
	newExecTest("For else", `{% for i in emptySet %}{{ i }}{% else %}No results.{% endfor %}`, map[string]Value{"emptySet": []int{}}, expect(`No results.`)),
	newExecTest(
		"For map",
		`{% for k, v in data %}Record {{ loop.index }}: {{ k }}: {{ v }}{% if not loop.last %} - {% endif %}{% endfor %}`,
		map[string]Value{"data": map[string]float64{"Group A": 5.12, "Group B": 5.09}},
		expect(`Record 1: Group A: 5.12 - Record 2: Group B: 5.09`, `Record 1: Group B: 5.09 - Record 2: Group A: 5.12`),
	),
	newExecTest(
		"Some operators",
		`{{ 4.5 * 10 }} - {{ 3 + true }} - {{ 3 + 4 == 7.0 }} - {{ 10 % 2 == 0 }} - {{ 10 ** 2 > 99.9 and 10 ** 2 <= 100 }}`,
		nil,
		expect(`45 - 4 - 1 - 1 - 1`),
	),
	newExecTest("In and not in", `{{ 5 in set and 4 not in set }}`, map[string]Value{"set": []int{5, 10}}, expect(`1`)),
	newExecTest("Function call", `{{ multiply(num, 5) }}`, map[string]Value{"num": 10}, expect(`50`)),
	newExecTest("Filter call", `Welcome, {{ name }}`, nil, expect(`Welcome, `)),
	newExecTest("Filter call", `Welcome, {{ name|default('User') }}`, map[string]Value{"name": nil}, expect(`Welcome, User`)),
	newExecTest("Filter call", `Welcome, {{ surname|default('User') }}`, map[string]Value{"name": nil}, expect(`Welcome, User`)),
	newExecTest(
		"Basic use statement",
		`{% extends '{% block message %}{% endblock %}' %}{% use '{% block message %}Hello{% endblock %}' %}`,
		nil,
		expect("Hello"),
	),
	newExecTest(
		"Extended use statement",
		`{% extends '{% block message %}{% endblock %}' %}{% use '{% block message %}Hello{% endblock %}' with message as base_message %}{% block message %}{{ block('base_message') }}, World!{% endblock %}`,
		nil,
		expect("Hello, World!"),
	),
	newExecTest(
		"Set statement",
		`{% set val = 'a value' %}{{ val }}`,
		nil,
		expect("a value"),
	),
	newExecTest(
		"Do statement",
		`{% do p.Name('Mister ') %}{{ p.Name('') }}`,
		map[string]Value{"p": &fakePerson{"Meeseeks"}},
		expect("Mister Meeseeks"),
	),
	newExecTest(
		"Filter statement",
		`{% filter upper %}hello, world!{% endfilter %}`,
		nil,
		expect("HELLO, WORLD!"),
	),
	newExecTest(
		"Import statement",
		`{% import 'macros.twig' as mac %}{{ mac.test("hi") }}`,
		nil,
		expect("test: hi"),
	),
	newExecTest(
		"From statement",
		`{% from 'macros.twig' import test, def as other %}{{ other("", "HI!") }}`,
		nil,
		expect("HI!"),
	),
	newExecTest(
		"Ternary if",
		`{{ false ? (true ? "Hello" : "World") : "Words" }}`,
		nil,
		expect("Words"),
	),
	newExecTest(
		"Hash literal",
		`{{ {"test": 1}["test"] }}`,
		nil,
		expect("1"),
	),
	newExecTest(
		"Another hash literal",
		`{% set v = {quadruple: "to the power of four!", 0: "ew", "0": "it's not that bad"} %}ew? {{ v.0 }} {{ v.quadruple }}`,
		nil,
		expect("ew? it's not that bad to the power of four!"),
	),
	newExecTest(
		"Array literal",
		`{{ ["test", 1, "bar"][2] }}`,
		nil,
		expect("bar"),
	),
	newExecTest(
		"Another Array literal",
		`{{ ["test", 1, "bar"].1 }}`,
		nil,
		expect("1"),
	),
	newExecTest(
		"Comparison with or",
		`{% if item1 == "banana" or item2 == "apple" %}At least one item is correct{% else %}neither item is correct{% endif %}`,
		map[string]Value{"item1": "orange", "item2": "apple"},
		expect("At least one item is correct"),
	),
	newExecTest(
		"Non-existent map element without default",
		`{{ data.A }} {{ data.NotThere }} {{ data.B }}`,
		map[string]Value{"data": map[string]string{"A": "Foo", "B": "Bar"}},
		expect("Foo  Bar"),
	),
	newExecTest(
		"Non-existent map element with default",
		`{{ data.A }} {{ data.NotThere|default("default value") }} {{ data.B }}`,
		map[string]Value{"data": map[string]string{"A": "Foo", "B": "Bar"}},
		expect("Foo default value Bar"),
	),
	newExecTest(
		"Accessing templateName on _self",
		`Template: {{ _self.templateName }}`,
		nil,
		expect("Template: Template: {{ _self.templateName }}"),
	),
	withPatch(_execTest{
		"Unsupported binary operator",
		`{{ 1 + 2 }}`,
		nil,
		expectErrorContains("unsupported binary operator: _"),
	}, func(n parse.Node) {
		if bn, ok := n.(*parse.BinaryExpr); ok {
			bn.Op = "_"
		}
	}),
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
			return fmt.Sprint("expected error mismatch but there was no actual error and no error was expected! (bug?)")
		}
		return fmt.Sprintf("unexpected error %#v", err.actual.Error())
	}
	return fmt.Sprintf("%#v is not the expected error %s", err.actual, joinExpected(err.expected))
}

func newExpectErrorMismatchError(actual error, expected ...string) error {
	return &expectErrorMismatchError{actual, expected}
}

type expectedChecker func(actual string, err error) error

func expectChained(expects ...expectedChecker) expectedChecker {
	return func(actual string, err error) error {
		for _, expect := range expects {
			if e := expect(actual, err); e != nil {
				return e
			}
		}
		return nil
	}
}

func expect(expected ...string) expectedChecker {
	return expectChained(expectNoError(), func(actual string, err error) error {
		for _, exp := range expected {
			if actual == exp {
				return nil
			}
		}
		return newExpectFailedError(false, actual, expected...)
	})
}

func expectContains(matches string) expectedChecker {
	return expectChained(expectNoError(), func(actual string, err error) error {
		if !strings.Contains(actual, matches) {
			return newExpectFailedError(true, actual, matches)
		}
		return nil
	})
}

func expectNoError() expectedChecker {
	return func(actual string, err error) error {
		if err != nil {
			return newExpectErrorMismatchError(err)
		}
		return nil
	}
}

func expectErrorContains(expected ...string) expectedChecker {
	return func(_ string, err error) error {
		if err == nil {
			return nil
		}
		actual := err.Error()
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

func evaluateTest(t *testing.T, env *Env, test execTest) {
	w := &bytes.Buffer{}
	err := execute(test.tpl(), w, test.ctx(), env)
	check := test.expected()

	out := w.String()
	if err := check(out, err); err != nil {
		t.Errorf("%s: %s", test.name(), err)
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

// nodeMonkeyPatcher is a parse.NodeVisitor to enable tests to arbitrarily modify a parsed tree.
type nodeMonkeyPatcher struct {
	patch func(parse.Node)
}

func (v *nodeMonkeyPatcher) Enter(parse.Node) {}

func (v *nodeMonkeyPatcher) Leave(n parse.Node) {
	if v.patch != nil {
		v.patch(n)
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
	patcher := &nodeMonkeyPatcher{}
	env.Visitors = append(env.Visitors, patcher)
	for _, test := range tests {
		patcher.patch = nil
		if t, ok := test.(execTestWithPatch); ok {
			patcher.patch = t.patchNode
		}
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
