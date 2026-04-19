package twig

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tyler-sommer/stick"
)

// Integration tests that exercise the built-in Twig tests (registered by
// twig.New via twig/tests.TwigTests) end-to-end. These complement the
// per-function unit tests in twig/tests/tests_test.go.

func render(t *testing.T, tpl string, ctx map[string]stick.Value) (string, error) {
	t.Helper()
	env := New(nil) // StringLoader — template text IS the template name
	buf := &bytes.Buffer{}
	err := env.Execute(tpl, buf, ctx)
	return buf.String(), err
}

func mustRender(t *testing.T, tpl string, ctx map[string]stick.Value) string {
	t.Helper()
	out, err := render(t, tpl, ctx)
	if err != nil {
		t.Fatalf("render(%q): %v", tpl, err)
	}
	return out
}

// TestIsDefined covers the special-cased interceptor: `is defined` must not
// evaluate its left operand, so bindings to nil still report as defined
// (matching PHP Twig 3.x semantics).
func TestIsDefined(t *testing.T) {
	cases := []struct {
		name string
		tpl  string
		ctx  map[string]stick.Value
		want string
	}{
		{"unbound name is undefined",
			`{% if foo is defined %}y{% else %}n{% endif %}`,
			nil, "n"},
		{"bound to nil is still defined",
			`{% if foo is defined %}y{% else %}n{% endif %}`,
			map[string]stick.Value{"foo": nil}, "y"},
		{"bound to non-nil is defined",
			`{% if foo is defined %}y{% else %}n{% endif %}`,
			map[string]stick.Value{"foo": 1}, "y"},
		{"is not defined negates",
			`{% if foo is not defined %}y{% else %}n{% endif %}`,
			nil, "y"},
		{"attribute on missing container is undefined",
			`{% if obj.x is defined %}y{% else %}n{% endif %}`,
			nil, "n"},
		{"attribute that exists is defined",
			`{% if obj.a is defined %}y{% else %}n{% endif %}`,
			map[string]stick.Value{"obj": map[string]stick.Value{"a": 1}}, "y"},
		{"missing attribute on existing container is undefined",
			`{% if obj.missing is defined %}y{% else %}n{% endif %}`,
			map[string]stick.Value{"obj": map[string]stick.Value{"a": 1}}, "n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := mustRender(t, c.tpl, c.ctx)
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

// TestBuiltInTestsEndToEnd exercises the non-defined built-in tests through
// the twig env so we know registration + lookup + evaluation all wire up.
func TestBuiltInTestsEndToEnd(t *testing.T) {
	cases := []struct {
		name string
		tpl  string
		ctx  map[string]stick.Value
		want string
	}{
		{"empty: nil", `{{ x is empty ? 'y' : 'n' }}`, nil, "y"},
		{"empty: non-empty string", `{{ x is empty ? 'y' : 'n' }}`, map[string]stick.Value{"x": "a"}, "n"},
		{"not empty: string", `{{ x is not empty ? 'y' : 'n' }}`, map[string]stick.Value{"x": "a"}, "y"},
		{"even: 4", `{{ 4 is even ? 'y' : 'n' }}`, nil, "y"},
		{"odd: 4", `{{ 4 is odd ? 'y' : 'n' }}`, nil, "n"},
		{"null: nil", `{{ x is null ? 'y' : 'n' }}`, nil, "y"},
		{"none alias: nil", `{{ x is none ? 'y' : 'n' }}`, nil, "y"},
		{"null: non-nil", `{{ x is null ? 'y' : 'n' }}`, map[string]stick.Value{"x": 0}, "n"},
		{"iterable: slice", `{{ x is iterable ? 'y' : 'n' }}`, map[string]stick.Value{"x": []int{}}, "y"},
		{"iterable: int", `{{ x is iterable ? 'y' : 'n' }}`, map[string]stick.Value{"x": 3}, "n"},
		{"divisible by: 9/3", `{{ 9 is divisible by(3) ? 'y' : 'n' }}`, nil, "y"},
		{"divisible by: 10/3", `{{ 10 is divisible by(3) ? 'y' : 'n' }}`, nil, "n"},
		{"same as: int", `{{ 1 is same as(1) ? 'y' : 'n' }}`, nil, "y"},
		{"same as: int vs string", `{{ 1 is same as('1') ? 'y' : 'n' }}`, nil, "n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := mustRender(t, c.tpl, c.ctx)
			got = strings.TrimSpace(got)
			if got != c.want {
				t.Errorf("tpl=%q got %q, want %q", c.tpl, got, c.want)
			}
		})
	}
}
