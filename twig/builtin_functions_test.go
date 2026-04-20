package twig

import (
	"testing"

	"github.com/tyler-sommer/stick"
)

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
