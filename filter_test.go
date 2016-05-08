package stick

import (
	"testing"
	"bytes"
)

func TestFilterDefault(t *testing.T) {
	env := NewEnv(nil)

	if _, ok := env.Filters["default"]; !ok {
		t.Errorf("There is no filter named default!")
	}

	var out string
	var expected string = "Hi person"
	buf := bytes.NewBufferString(out)

	m := make(map[string]Value)

	err := env.Execute("Hi {{ name|default('person') }}", buf, m)

	if nil != err {
		t.Error(err)
	}

	if expected != buf.String() {
		t.Errorf("Failed to fill in default value for a variable. Expected '%s', got '%s'", expected, out)
	}

}