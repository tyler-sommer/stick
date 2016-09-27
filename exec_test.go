package stick

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

type execTest struct {
	name     string
	tpl      string
	ctx      map[string]Value
	expected expectedChecker
}

func tpl(name, content string) *testTemplate {
	return &testTemplate{name, content}
}

var emptyCtx = map[string]Value{}

type testPerson struct {
	name string
}

func (p *testPerson) Name(prefix string) string {
	p.name = prefix + p.name
	return p.name
}

var tests = []execTest{
	{"Hello, World", "Hello, World!", emptyCtx, expect("Hello, World!")},
	{"Hello, Tyler!", "Hello, {{ name }}!", map[string]Value{"name": "Tyler"}, expect("Hello, Tyler!")},
	{"Simple if", `<li class="{% if active %}active{% endif %}">`, map[string]Value{"active": true}, expect(`<li class="active">`)},
	{"Simple inheritance", `{% extends 'Hello, {% block test %}universe{% endblock %}!' %}{% block test %}world{% endblock %}`, emptyCtx, expect(`Hello, world!`)},
	{"Simple include", `This is a test. {% include 'Hello, {{ name }}!' %} This concludes the test.`, map[string]Value{"name": "John"}, expect(`This is a test. Hello, John! This concludes the test.`)},
	{"Include with", `{% include 'Hello, {{ name }}{{ value }}' with vars %}`, map[string]Value{"value": "!", "vars": map[string]Value{"name": "Adam"}}, expect(`Hello, Adam!`)},
	{"Include with literal", `{% include 'Hello, {{ name }}{{ value }}' with {"name": "world", "value": "!"} only %}`, emptyCtx, expect(`Hello, world!`)},
	{"Embed", `Well. {% embed 'Hello, {% block name %}World{% endblock %}!' %}{% block name %}Tyler{% endblock %}{% endembed %}`, emptyCtx, expect(`Well. Hello, Tyler!`)},
	{"Constant null", `{% if test == null %}Yes{% else %}no{% endif %}`, map[string]Value{"test": nil}, expect(`Yes`)},
	{"Constant bool", `{% if test == true %}Yes{% else %}no{% endif %}`, map[string]Value{"test": false}, expect(`no`)},
	{"Chained attributes", `{{ entity.attr.Name }}`, map[string]Value{"entity": map[string]Value{"attr": struct{ Name string }{"Tyler"}}}, expect(`Tyler`)},
	{"Attribute method call", `{{ entity.Name('lower') }}`, map[string]Value{"entity": &testPerson{"Johnny"}}, expect(`lowerJohnny`)},
	{"For loop", `{% for i in 1..3 %}{{ i }}{% endfor %}`, emptyCtx, expect(`123`)},
	{"For else", `{% for i in emptySet %}{{ i }}{% else %}No results.{% endfor %}`, map[string]Value{"emptySet": []int{}}, expect(`No results.`)},
	{
		"For map",
		`{% for k, v in data %}Record {{ loop.Index }}: {{ k }}: {{ v }}{% if not loop.Last %} - {% endif %}{% endfor %}`,
		map[string]Value{"data": map[string]float64{"Group A": 5.12, "Group B": 5.09}},
		optionExpect(`Record 1: Group A: 5.12 - Record 2: Group B: 5.09`, `Record 1: Group B: 5.09 - Record 2: Group A: 5.12`),
	},
	{
		"Some operators",
		`{{ 4.5 * 10 }} - {{ 3 + true }} - {{ 3 + 4 == 7.0 }} - {{ 10 % 2 == 0 }} - {{ 10 ** 2 > 99.9 and 10 ** 2 <= 100 }}`,
		emptyCtx,
		expect(`45 - 4 - 1 - 1 - 1`),
	},
	{"In and not in", `{{ 5 in set and 4 not in set }}`, map[string]Value{"set": []int{5, 10}}, expect(`1`)},
	{"Function call", `{{ multiply(num, 5) }}`, map[string]Value{"num": 10}, expect(`50`)},
	{"Filter call", `Welcome, {{ name|default('User') }}`, map[string]Value{"name": nil}, expect(`Welcome, User`)},
	{
		"Basic use statement",
		`{% extends '{% block message %}{% endblock %}' %}{% use '{% block message %}Hello{% endblock %}' %}`,
		emptyCtx,
		expect("Hello"),
	},
	{
		"Extended use statement",
		`{% extends '{% block message %}{% endblock %}' %}{% use '{% block message %}Hello{% endblock %}' with message as base_message %}{% block message %}{{ block('base_message') }}, World!{% endblock %}`,
		emptyCtx,
		expect("Hello, World!"),
	},
	{
		"Set statement",
		`{% set val = 'a value' %}{{ val }}`,
		emptyCtx,
		expect("a value"),
	},
	{
		"Do statement",
		`{% do p.Name('Mister ') %}{{ p.Name('') }}`,
		map[string]Value{"p": &testPerson{"Meeseeks"}},
		expect("Mister Meeseeks"),
	},
	{
		"Filter statement",
		`{% filter upper %}hello, world!{% endfilter %}`,
		emptyCtx,
		expect("HELLO, WORLD!"),
	},
	{
		"Import statement",
		`{% import 'macros.twig' as mac %}{{ mac.test("hi") }}`,
		emptyCtx,
		expect("test: hi"),
	},
	{
		"From statement",
		`{% from 'macros.twig' import test, def as other %}{{ other("", "HI!") }}`,
		emptyCtx,
		expect("HI!"),
	},
	{
		"Ternary if",
		`{{ false ? (true ? "Hello" : "World") : "Words" }}`,
		emptyCtx,
		expect("Words"),
	},
	{
		"Hash literal",
		`{{ {"test": 1}["test"] }}`,
		emptyCtx,
		expect("1"),
	},
	{
		"Another hash literal",
		`{% set v = {quadruple: "to the power of four!", 0: "ew", "0": "it's not that bad"} %}ew? {{ v.0 }} {{ v.quadruple }}`,
		emptyCtx,
		expect("ew? it's not that bad to the power of four!"),
	},
	{
		"Array literal",
		`{{ ["test", 1, "bar"][2] }}`,
		emptyCtx,
		expect("bar"),
	},
	{
		"Another Array literal",
		`{{ ["test", 1, "bar"].1 }}`,
		emptyCtx,
		expect("1"),
	},
}

type expectedChecker func(actual string) (string, bool)

func expect(expected string) expectedChecker {
	return func(actual string) (string, bool) {
		return expected, expected == actual
	}
}

func optionExpect(expected ...string) expectedChecker {
	return func(actual string) (string, bool) {
		for _, exp := range expected {
			if actual == exp {
				return exp, true
			}
		}
		return strings.Join(expected, "\n\tor:\n"), false
	}
}

func evaluateTest(t *testing.T, env *Env, test execTest) {
	w := &bytes.Buffer{}
	err := execute(test.tpl, w, test.ctx, env)
	if err != nil {
		t.Errorf("%s: unexpected error: %s", test.name, err.Error())
		return
	}
	out := w.String()
	if expected, ok := test.expected(out); !ok {
		t.Errorf("%s: got:\n%s\n\texpected:\n%s", test.name, out, expected)
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

func (t *testLoader) Load(name string) (Template, error) {
	if b, ok := t.templates[name]; ok {
		return b, nil
	}
	return tpl(name, name), nil
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
	for _, test := range tests {
		evaluateTest(t, env, test)
	}
}
