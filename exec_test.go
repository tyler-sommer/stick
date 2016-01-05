package stick

import (
	"bytes"
	"testing"
)

type execTest struct {
	name string
	tmpl string
	ctx map[string]Value
	expected string
}

var emptyCtx = map[string]Value{}

var tests = []execTest{
	{"Hello, World", "Hello, World!", emptyCtx, "Hello, World!"},
	{"Hello, Tyler!", "Hello, {{ name }}!", map[string]Value{"name": "Tyler"}, "Hello, Tyler!"},
	{"Simple if", `<li class="{% if active %}active{% endif %}">`, map[string]Value{"active": true}, `<li class="active">`},
}

func evaluateTest(t *testing.T, test execTest) {
	w := &bytes.Buffer{}
	err := Execute(test.tmpl, w, test.ctx)
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
