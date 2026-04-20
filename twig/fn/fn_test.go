package fn

import (
	"strings"
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
	for _, name := range []string{"max", "min", "range", "cycle", "random", "attribute"} {
		if _, ok := fns[name]; !ok {
			t.Errorf("TwigFunctions missing %q", name)
		}
	}
}

// equalRange compares the two slices element-by-element, normalising int
// vs int-typed-as-int to avoid spurious mismatches from %T differences.
func equalRange(a, b []stick.Value) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestRange(t *testing.T) {
	ts := []struct {
		name string
		args []stick.Value
		want []stick.Value
	}{
		{"int asc", []stick.Value{1, 5}, []stick.Value{1, 2, 3, 4, 5}},
		{"int desc default step", []stick.Value{5, 1}, []stick.Value{5, 4, 3, 2, 1}},
		{"int with step", []stick.Value{0, 6, 2}, []stick.Value{0, 2, 4, 6}},
		{"int negative range", []stick.Value{-2, 2}, []stick.Value{-2, -1, 0, 1, 2}},
		{"single", []stick.Value{3, 3}, []stick.Value{3}},
		{"misdirected step flips", []stick.Value{5, 1, 1}, []stick.Value{5, 4, 3, 2, 1}},
		{"too few args", []stick.Value{1}, []stick.Value{}},
		{"chars asc", []stick.Value{"a", "d"}, []stick.Value{"a", "b", "c", "d"}},
		{"chars desc", []stick.Value{"c", "a"}, []stick.Value{"c", "b", "a"}},
	}
	for _, tc := range ts {
		got := fnRange(nil, tc.args...).([]stick.Value)
		if !equalRange(got, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}

	// Float input — when either endpoint isn't integer, results are
	// floats throughout.
	got := fnRange(nil, 0.5, 2.5, 1.0).([]stick.Value)
	if len(got) != 3 || got[0] != 0.5 || got[2] != 2.5 {
		t.Errorf("float range: got %v", got)
	}
}

func TestCycle(t *testing.T) {
	values := []stick.Value{"a", "b", "c"}
	ts := []struct {
		name string
		args []stick.Value
		want stick.Value
	}{
		{"index 0", []stick.Value{values, 0}, "a"},
		{"index 4 wraps", []stick.Value{values, 4}, "b"},
		{"index -1 wraps", []stick.Value{values, -1}, "c"},
		{"index -4 wraps", []stick.Value{values, -4}, "c"},
		{"empty list", []stick.Value{[]stick.Value{}, 0}, nil},
		{"too few args", []stick.Value{values}, nil},
		{"non-iterable", []stick.Value{42, 0}, nil},
	}
	for _, tc := range ts {
		got := fnCycle(nil, tc.args...)
		if got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestRandom(t *testing.T) {
	// random() with no args: returns an int63. Just make sure it doesn't
	// panic and produces something representable as int64.
	for i := 0; i < 10; i++ {
		v := fnRandom(nil)
		if _, ok := v.(int64); !ok {
			t.Errorf("random() empty: got %T, want int64", v)
		}
	}

	// random(N): result must be in [0, N].
	for i := 0; i < 100; i++ {
		v := fnRandom(nil, 5).(int64)
		if v < 0 || v > 5 {
			t.Errorf("random(5): got %d, want [0..5]", v)
		}
	}

	// random("abcde"): result must be a single character from the input.
	for i := 0; i < 50; i++ {
		s := fnRandom(nil, "abcde").(string)
		if len(s) != 1 || !strings.ContainsRune("abcde", rune(s[0])) {
			t.Errorf(`random("abcde"): got %q`, s)
		}
	}

	// random(iterable): returns one of the elements.
	for i := 0; i < 50; i++ {
		v := fnRandom(nil, []stick.Value{"x", "y", "z"})
		s, _ := v.(string)
		if s != "x" && s != "y" && s != "z" {
			t.Errorf("random(slice): got %v", v)
		}
	}

	// Empty inputs.
	if v := fnRandom(nil, ""); v != "" {
		t.Errorf(`random(""): got %v`, v)
	}
	if v := fnRandom(nil, []stick.Value{}); v != nil {
		t.Errorf("random([]): got %v", v)
	}
}

func TestAttribute(t *testing.T) {
	m := map[string]stick.Value{"name": "Alice", "age": 30}
	cases := []struct {
		name string
		args []stick.Value
		want stick.Value
	}{
		{"map key", []stick.Value{m, "name"}, "Alice"},
		{"map missing key", []stick.Value{m, "nope"}, nil},
		{"too few args", []stick.Value{m}, nil},
	}
	for _, tc := range cases {
		got := fnAttribute(nil, tc.args...)
		if got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}
