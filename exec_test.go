package stick

import (
	"testing"
	"github.com/tyler-sommer/stick/parse"
	"bytes"
)

func TestExec(t *testing.T) {
	tree := parse.Parse("Hello")

	w := &bytes.Buffer{}

	s := newState(w)

	s.walk(tree.Root())

	if (w.String() != "Hello") {
		t.Errorf("Unexpected output: '%s'", w.String())
	}
}
