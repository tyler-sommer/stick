package stick

import (
	"bytes"
	"github.com/tyler-sommer/stick/parse"
	"testing"
)

func TestExec(t *testing.T) {
	tree, err := parse.Parse("Hello")
	if err != nil {
		t.Errorf("Parse error: %s", err)
	}

	w := &bytes.Buffer{}

	s := newState(w, map[string]*variable{})

	s.walk(tree.Root())

	if w.String() != "Hello" {
		t.Errorf("Unexpected output: '%s'", w.String())
	}
}
