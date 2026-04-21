package twig

import (
	"testing"

	"github.com/tyler-sommer/stick"
)

// TestBuiltinRange covers range() across the int, char, and step
// variants the Twig docs call out.
func TestBuiltinRange(t *testing.T) {
	cases := []struct {
		name string
		tpl  string
		want string
	}{
		{"int asc", `{% for i in range(1, 3) %}{{ i }}-{% endfor %}`, "1-2-3-"},
		{"int desc", `{% for i in range(3, 1) %}{{ i }}-{% endfor %}`, "3-2-1-"},
		{"int with step", `{% for i in range(0, 6, 2) %}{{ i }};{% endfor %}`, "0;2;4;6;"},
		{"chars", `{% for c in range('a', 'c') %}{{ c }}{% endfor %}`, "abc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := mustRender(t, tc.tpl, nil); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// TestBuiltinCycle covers the wrap-around behavior of cycle().
func TestBuiltinCycle(t *testing.T) {
	tpl := `{% for i in range(0, 5) %}{{ cycle(['x','y','z'], i) }}{% endfor %}`
	if got := mustRender(t, tpl, nil); got != "xyzxyz" {
		t.Errorf("cycle wrap: got %q, want %q", got, "xyzxyz")
	}
}

// TestBuiltinAttribute covers dynamic attribute access via attribute().
func TestBuiltinAttribute(t *testing.T) {
	tpl := `{{ attribute(p, 'name') }}`
	ctx := map[string]stick.Value{"p": map[string]stick.Value{"name": "Alice"}}
	if got := mustRender(t, tpl, ctx); got != "Alice" {
		t.Errorf("attribute: got %q, want Alice", got)
	}
}

// TestBuiltinRandom asserts random() returns a value of the expected
// shape, not a specific element (since output is unpredictable).
func TestBuiltinRandom(t *testing.T) {
	tpl := `{{ random([1, 2, 3]) }}`
	got := mustRender(t, tpl, nil)
	if got != "1" && got != "2" && got != "3" {
		t.Errorf("random: got %q, want one of 1/2/3", got)
	}
}

// TestBuiltinMaxMin exercises twig.New's Functions wiring end-to-end, so a
// future change that accidentally drops fn.TwigFunctions fails here.
func TestBuiltinMaxMin(t *testing.T) {
	cases := []struct {
		name string
		tpl  string
		ctx  map[string]stick.Value
		want string
	}{
		{"max scalars", `{{ max(1, 2, 3) }}`, nil, "3"},
		{"min scalars", `{{ min(1, 2, 3) }}`, nil, "1"},
		{"max from slice", `{{ max([4, 1, 9, 2]) }}`, nil, "9"},
		{"min from slice", `{{ min([4, 1, 9, 2]) }}`, nil, "1"},
		{"max from hash values",
			`{{ max(counts) }}`, map[string]stick.Value{
				"counts": map[string]stick.Value{"a": 2, "b": 7, "c": 5},
			}, "7"},
		{"max strings lexicographic", `{{ max('apple', 'banana') }}`, nil, "banana"},
		{"min strings lexicographic", `{{ min('banana', 'apple') }}`, nil, "apple"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mustRender(t, tc.tpl, tc.ctx)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
