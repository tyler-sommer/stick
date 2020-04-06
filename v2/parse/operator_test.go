package parse

import (
	"testing"
)

var tests = []string{"not"}

func init() {
	for _, op := range binaryOperators {
		tests = append(tests, op.op)
	}
}

func TestOperator(t *testing.T) {
	for _, test := range tests {
		o := operatorMatcher.FindString(test)
		if o != test {
			t.Errorf("got \"%+v\" expected \"%v\"", o, test)
		}
	}
}
