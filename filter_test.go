package stick

import (
	"testing"
	"bytes"
)

func TestFilters(t *testing.T) {

	type test struct {
		test, expected string
		data           map[string]Value
	}

	values := []test{
		{test:"Hi {{ name|default('person') }}", expected: "Hi person", data:nil},
		{test:"{{ num|abs }}", expected:"5", data:map[string]Value{"num":-5}},
		{test:"{{ num|abs }}", expected:"6", data:map[string]Value{"num":6}},
		{test:"{{ pi|abs }}", expected:"3.14", data:map[string]Value{"pi":3.14}},
		{test:"{{ pi|abs }}", expected:"3.14", data:map[string]Value{"pi":-3.14}},
		{test:"{{ name|capitalize }}", expected:"Mr ed", data:map[string]Value{"name":"MR ED"}},
		{test:"{{ name|lower }}", expected:"mr ed", data:map[string]Value{"name":"MR ED"}},
		{test:"{{ name|title }}", expected:"Mr Ed", data:map[string]Value{"name":"mr ed"}},
		{test:"{{ name|trim }}", expected:"mr ed", data:map[string]Value{"name":" mr ed "}},
		{test:"{{ name|upper }}", expected:"MR ED", data:map[string]Value{"name":"mr ed"}},
	}

	var out string
	buf := bytes.NewBufferString(out)

	env := NewEnv(nil)
	for _, test := range values {
		buf.Reset()
		t.Logf("Testing '%s' and expecting to get '%s'", test.test, test.expected)
		out = ""
		if err := env.Execute(test.test, buf, test.data); err != nil {
			t.Error("Failed to Execute Template", err)
			continue
		}

		if test.expected != buf.String() {
			t.Errorf("Failed. Expected '%s', got '%s'", test.expected, buf.String())
			continue
		}

	}
}