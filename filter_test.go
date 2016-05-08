package stick

import (
	"bytes"
	"testing"
)

func TestFilters(t *testing.T) {

	type test struct {
		test, expected string
		data           map[string]Value
	}

	values := []test{
		{test: "Hi {{ name|default('person') }}", expected: "Hi person", data: nil},
		{test: "{{ num|abs }}", expected: "5", data: map[string]Value{"num": -5}},
		{test: "{{ num|abs }}", expected: "6", data: map[string]Value{"num": 6}},
		{test: "{{ pi|abs }}", expected: "3.14", data: map[string]Value{"pi": 3.14}},
		{test: "{{ pi|abs }}", expected: "3.14", data: map[string]Value{"pi": -3.14}},
		{test: "{{ name|capitalize }}", expected: "Mr ed", data: map[string]Value{"name": "MR ED"}},
		{test: "{{ name|lower }}", expected: "mr ed", data: map[string]Value{"name": "MR ED"}},
		{test: "{{ name|title }}", expected: "Mr Ed", data: map[string]Value{"name": "mr ed"}},
		{test: "{{ name|trim }}", expected: "mr ed", data: map[string]Value{"name": " mr ed "}},
		{test: "{{ name|upper }}", expected: "MR ED", data: map[string]Value{"name": "mr ed"}},
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

func TestFilterBatch(t *testing.T) {
	data := map[string]Value{"items": []int{1, 2, 3, 4, 5, 6, 7, 8}}
	var ctx Context
	batched := filterBatch(ctx, data, 3, "No Item")

	bLen := len(batched)
	if 3 != bLen {
		t.Errorf("Expected the batched array to be 3 items long, got %d", bLen)
	}

	for _, batchedVals := range batched {
		bLen := len(batchedVals)
		if 3 != bLen {
			t.Errorf("Expected batched value length to be 3, got %d", bLen)
		}
	}
}

func TestFilterOnCmdBlock(t *testing.T) {
	expected := "1.2.3..4.5.6..7.8.No Item.."

	env := NewEnv(nil)
	template := "{% for row in items|batch(3, 'No Item') %}{% for item in row %}{{ item }}.{% endfor %}.{% endfor %}"
	data := map[string]Value{"items": []int{1, 2, 3, 4, 5, 6, 7, 8}}

	var actual string
	buf := bytes.NewBufferString(actual)
	err := env.Execute(template, buf, data)

	if nil != err {
		t.Error(err.Error())
		return
	}

	if expected != actual {
		t.Errorf("Failed to parse template. expected: %s != actual:%s", expected, actual)
	}
}
