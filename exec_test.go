package stick

import (
	"bytes"
	"testing"
)

type execTest struct {
	name     string
	tmpl     string
	ctx      map[string]Value
	expected string
}

var emptyCtx = map[string]Value{}

var tests = []execTest{
	{"Hello, World", "Hello, World!", emptyCtx, "Hello, World!"},
	{"Hello, Tyler!", "Hello, {{ name }}!", map[string]Value{"name": "Tyler"}, "Hello, Tyler!"},
	{"Simple if", `<li class="{% if active %}active{% endif %}">`, map[string]Value{"active": true}, `<li class="active">`},
	{"Simple inheritance", `{% extends 'Hello, {% block test %}universe{% endblock %}!' %}{% block test %}world{% endblock %}`, emptyCtx, `Hello, world!`},
	{"Simple include", `This is a test. {% include 'Hello, {{ name }}!' %} This concludes the test.`, map[string]Value{"name": "John"}, `This is a test. Hello, John! This concludes the test.`},
	{"Include with", `{% include 'Hello, {{ name }}{{ value }}' with vars %}`, map[string]Value{"value": "!", "vars": map[string]Value{"name": "Adam"}}, `Hello, Adam!`},
	{"Embed", `Well. {% embed 'Hello, {% block name %}World{% endblock %}!' %}{% block name %}Tyler{% endblock %}{% endembed %}`, emptyCtx, `Well. Hello, Tyler!`},
	{"Constant null", `{% if test == null %}Yes{% else %}no{% endif %}`, map[string]Value{"test": nil}, `Yes`},
	{"Constant bool", `{% if test == true %}Yes{% else %}no{% endif %}`, map[string]Value{"test": false}, `no`},
}

func evaluateTest(t *testing.T, test execTest) {
	w := &bytes.Buffer{}
	env := NewEnv(&StringLoader{})
	err := execute(test.tmpl, w, test.ctx, env)
	if err != nil {
		t.Errorf("%s: unexpected error: %s", test.name, err.Error())
		return
	}
	out := w.String()
	if out != test.expected {
		t.Errorf("%s: got:\n%s\n\texpected:\n%s", test.name, out, test.expected)
	}
}

func TestExec(t *testing.T) {
	for _, test := range tests {
		evaluateTest(t, test)
	}
}
