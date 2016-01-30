package stick

import (
	"bytes"
	"strings"
	"testing"
)

type execTest struct {
	name     string
	tmpl     string
	ctx      map[string]Value
	expected expectedChecker
}

var emptyCtx = map[string]Value{}

type testPerson struct {
	name string
}

func (p testPerson) Name(prefix string) string {
	return prefix + p.name
}

var tests = []execTest{
	{"Hello, World", "Hello, World!", emptyCtx, expect("Hello, World!")},
	{"Hello, Tyler!", "Hello, {{ name }}!", map[string]Value{"name": "Tyler"}, expect("Hello, Tyler!")},
	{"Simple if", `<li class="{% if active %}active{% endif %}">`, map[string]Value{"active": true}, expect(`<li class="active">`)},
	{"Simple inheritance", `{% extends 'Hello, {% block test %}universe{% endblock %}!' %}{% block test %}world{% endblock %}`, emptyCtx, expect(`Hello, world!`)},
	{"Simple include", `This is a test. {% include 'Hello, {{ name }}!' %} This concludes the test.`, map[string]Value{"name": "John"}, expect(`This is a test. Hello, John! This concludes the test.`)},
	{"Include with", `{% include 'Hello, {{ name }}{{ value }}' with vars %}`, map[string]Value{"value": "!", "vars": map[string]Value{"name": "Adam"}}, expect(`Hello, Adam!`)},
	{"Embed", `Well. {% embed 'Hello, {% block name %}World{% endblock %}!' %}{% block name %}Tyler{% endblock %}{% endembed %}`, emptyCtx, expect(`Well. Hello, Tyler!`)},
	{"Constant null", `{% if test == null %}Yes{% else %}no{% endif %}`, map[string]Value{"test": nil}, expect(`Yes`)},
	{"Constant bool", `{% if test == true %}Yes{% else %}no{% endif %}`, map[string]Value{"test": false}, expect(`no`)},
	{"Chained attributes", `{{ entity.attr.Name }}`, map[string]Value{"entity": map[string]Value{"attr": struct{ Name string }{"Tyler"}}}, expect(`Tyler`)},
	{"Attribute method call", `{{ entity.Name('lower') }}`, map[string]Value{"entity": testPerson{"Johnny"}}, expect(`lowerJohnny`)},
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
	err := execute(test.tmpl, w, test.ctx, env)
	if err != nil {
		t.Errorf("%s: unexpected error: %s", test.name, err.Error())
		return
	}
	out := w.String()
	if expected, ok := test.expected(out); !ok {
		t.Errorf("%s: got:\n%s\n\texpected:\n%s", test.name, out, expected)
	}
}

func TestExec(t *testing.T) {
	env := NewEnv(&StringLoader{})
	env.Functions["multiply"] = func(env *Env, args ...Value) Value {
		if len(args) != 2 {
			return 0
		}
		return CoerceNumber(args[0]) * CoerceNumber(args[1])
	}
	env.Filters["default"] = func(env *Env, val Value, args ...Value) Value {
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
