package fn

import (
	"testing"

	"github.com/tyler-sommer/stick"
)

func TestMaxMin(t *testing.T) {
	ts := []struct {
		name string
		fn   stick.Func
		args []stick.Value
		want stick.Value
	}{
		{"max scalars", fnMax, []stick.Value{1, 2, 3}, 3},
		{"min scalars", fnMin, []stick.Value{1, 2, 3}, 1},
		{"max slice", fnMax, []stick.Value{[]stick.Value{1, 3, 2}}, 3},
		{"min slice", fnMin, []stick.Value{[]stick.Value{3, 1, 2}}, 1},
		{"max hash values", fnMax, []stick.Value{map[string]stick.Value{"a": 1, "b": 3, "c": 2}}, 3},
		{"max single arg", fnMax, []stick.Value{42}, 42},
		{"max no args", fnMax, []stick.Value{}, nil},
		{"min no args", fnMin, []stick.Value{}, nil},
		{"max float beats int", fnMax, []stick.Value{2, 2.5}, 2.5},
		{"min negative", fnMin, []stick.Value{-5, -3, -10}, -10},
		// Lexicographic compare when every operand is a string.
		{"max strings", fnMax, []stick.Value{"apple", "banana", "cherry"}, "cherry"},
		{"min strings", fnMin, []stick.Value{"banana", "apple", "cherry"}, "apple"},
		{"max string slice", fnMax, []stick.Value{[]stick.Value{"z", "a", "m"}}, "z"},
		{"min single string", fnMin, []stick.Value{"only"}, "only"},
		// Mixed type: any non-string flips to numeric comparison.
		{"max mixed string and int", fnMax, []stick.Value{2, "10"}, "10"},
		{"min mixed", fnMin, []stick.Value{"5", 1}, 1},
	}
	for _, tc := range ts {
		got := tc.fn(nil, tc.args...)
		if got != tc.want {
			t.Errorf("%s: got %v (%T), want %v (%T)", tc.name, got, got, tc.want, tc.want)
		}
	}
}

func TestTwigFunctionsRegistration(t *testing.T) {
	fns := TwigFunctions()
	for _, name := range []string{"max", "min"} {
		if _, ok := fns[name]; !ok {
			t.Errorf("TwigFunctions missing %q", name)
		}
	}
}
