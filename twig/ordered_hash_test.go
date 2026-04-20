package twig

import (
	"testing"

	"github.com/tyler-sommer/stick"
)

// TestOrderedHashMergeAccumulator mirrors the real-world Twig 1.x pattern
// where a theme iterates sorted keys and accumulates into a hash via merge,
// then iterates the accumulator for display. Under stick's previous Go-map
// hash literals this lost the sort order; with *stick.Hash it is preserved.
func TestOrderedHashMergeAccumulator(t *testing.T) {
	tpl := `{% set out = {} %}` +
		`{% for k in ["c","a","b"]|sort %}{% set out = out|merge({(k): k~k}) %}{% endfor %}` +
		`{% for k,v in out %}{{k}}={{v}};{% endfor %}`
	got := mustRender(t, tpl, nil)
	if got != "a=aa;b=bb;c=cc;" {
		t.Errorf("got %q, want %q", got, "a=aa;b=bb;c=cc;")
	}
}

func TestOrderedHashReverse(t *testing.T) {
	tpl := `{% set m = {"a": 1, "b": 2, "c": 3}|reverse %}{% for k,v in m %}{{k}}={{v}};{% endfor %}`
	got := mustRender(t, tpl, nil)
	if got != "c=3;b=2;a=1;" {
		t.Errorf("got %q, want %q", got, "c=3;b=2;a=1;")
	}
}

func TestOrderedHashKeys(t *testing.T) {
	tpl := `{% set m = {"c": 1, "a": 2, "b": 3} %}{% for k in m|keys %}{{k}};{% endfor %}`
	got := mustRender(t, tpl, nil)
	if got != "c;a;b;" {
		t.Errorf("got %q, want %q (keys should be insertion order, not sorted)", got, "c;a;b;")
	}
}

// TestUserMapUnchanged covers the out-of-scope guarantee: maps passed by
// the caller as context keep plain Go-map semantics (unordered) and merge
// results between two plain maps also stay plain maps.
func TestUserContextMapUnchanged(t *testing.T) {
	ctx := map[string]stick.Value{
		"m": map[string]stick.Value{"x": 1},
		"n": map[string]stick.Value{"y": 2},
	}
	// The merged value iterated through stick sees both keys regardless of
	// iteration order; we just assert that no panic / no nil-drop happens.
	tpl := `{% set r = m|merge(n) %}{{ r.x }}+{{ r.y }}`
	got := mustRender(t, tpl, ctx)
	if got != "1+2" {
		t.Errorf("got %q, want %q", got, "1+2")
	}
}
