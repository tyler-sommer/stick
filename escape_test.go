package stick_test

import (
	"testing"

	"os"

	"github.com/tyler-sommer/stick"
	"github.com/tyler-sommer/stick/parse"
)

// This example shows how the AutoEscapeVisitor can be used to automatically
// sanitize input. It does this by wrapping printed expressions with a filter
// application, which resolves to stick.EscapeFilter.
func ExampleAutoEscapeVisitor() {
	env := stick.NewEnv(nil)
	env.Visitors = append(env.Visitors, &stick.AutoEscapeVisitor{})
	env.Filters["escape"] = stick.EscapeFilter

	env.Execute("<html>{{ '<script>bad stuff</script>' }}", os.Stdout, map[string]stick.Value{})
	// Output:
	// <html>&lt;script&gt;bad stuff&lt;/script&gt;
}

// This example displays the EscapeFilter in action.
//
// Note the "already_safe" value wrapped in a NewSafeValue; it is not
// escaped.
func ExampleEscapeFilter() {
	env := stick.NewEnv(nil)
	env.Filters["escape"] = stick.EscapeFilter

	env.Execute("<html>{{ dangerous|escape }} {{ already_safe|escape }}", os.Stdout, map[string]stick.Value{
		"already_safe": stick.NewSafeValue("<script>good script</script>"),
		"dangerous":    "<script>bad script</script>",
	})
	// Output:
	// <html>&lt;script&gt;bad script&lt;/script&gt; <script>good script</script>
}

func newVisitorEnv() *stick.Env {
	env := stick.NewEnv(nil)
	env.Visitors = append(env.Visitors, &stick.AutoEscapeVisitor{})
	env.Filters["escape"] = stick.EscapeFilter
	return env
}

func TestAutoEscapeVisitor(t *testing.T) {
	env := newVisitorEnv()
	tree, err := env.Parse("Some {{ 'text' }}")
	if err != nil {
		t.Error(err)
		return
	}
	root := tree.Root()
	a := root.All()
	if l := len(a); l != 2 {
		t.Errorf("expected two children, got %d", l)
		return
	}
	ti, ok := a[0].(*parse.TextNode)
	if !ok {
		t.Errorf("expected TextNode, got %s", ti)
		return
	}
	if tfi := ti.Data; tfi != "Some " {
		t.Errorf("expected 'Some ', got %s", tfi)
		return
	}
	fb, ok := a[1].(*parse.PrintNode)
	if !ok {
		t.Errorf("expected PrintNode, got %s", fb)
		return
	}
	fi, ok := fb.X.(*parse.FilterExpr)
	if !ok {
		t.Errorf("expected FilterNode, got %s", fi)
		return
	}
	fa, ok := fi.Args[0].(*parse.StringExpr)
	if !ok {
		t.Errorf("expected StringExpr, got %s", fa)
		return
	}
	if fv := fa.Text; fv != "text" {
		t.Errorf("expected 'text'", fv)
	}
}
