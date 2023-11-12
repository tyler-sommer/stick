package twig_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/tyler-sommer/stick"
	"github.com/tyler-sommer/stick/parse"
	"github.com/tyler-sommer/stick/twig"
)

// This example shows how the AutoEscapeVisitor can be used to automatically
// sanitize input. It does this by wrapping printed expressions with a filter
// application, which resolves to stick.EscapeFilter.
func ExampleAutoEscapeExtension() {
	env := twig.New(nil)
	env.Execute("<html>{{ '<script>bad stuff</script>' }}", os.Stdout, map[string]stick.Value{})
	// Output:
	// <html>&lt;script&gt;bad stuff&lt;/script&gt;
}

// This example displays the EscapeFilter in action.
//
// Note the "already_safe" value wrapped in a NewSafeValue; it is not
// escaped.
func ExampleAutoEscapeExtension_alreadySafe() {
	env := twig.New(nil)
	env.Execute("<html>{{ dangerous|escape }} {{ already_safe|escape }}", os.Stdout, map[string]stick.Value{
		"already_safe": stick.NewSafeValue("<script>good script</script>", "html"),
		"dangerous":    "<script>bad script</script>",
	})
	// Output:
	// <html>&lt;script&gt;bad script&lt;/script&gt; <script>good script</script>
}

func TestAutoEscapeVisitor(t *testing.T) {
	env := twig.New(nil)
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
		t.Errorf("expected 'text', got %s", fv)
	}
}

func TestAutoEscapeExtension(t *testing.T) {
	env := twig.New(&stick.MemoryLoader{Templates: map[string]string{
		"index.html.twig": "<html><script>{% include 'utils.js.twig' %}</script><body>Hello, {{ user }}!</body></html>",
		"utils.js.twig":   `console.log("{{ message }}");`,
	}})
	buf := bytes.Buffer{}
	err := env.Execute("index.html.twig", &buf, map[string]stick.Value{
		"user":    "<a href='bad'>tyler-sommer</a>",
		"message": "bad '\" message",
	})
	if err != nil {
		t.Errorf("unexpected error executing template: %s", err)
	}
	expected := "<html><script>console.log(\"bad\\u0020\\u0027\\u0022\\u0020message\");</script><body>Hello, &lt;a href=&#39;bad&#39;&gt;tyler-sommer&lt;/a&gt;!</body></html>"
	actual := buf.String()
	if actual != expected {
		t.Errorf("expected output to be escaped, but got: %s", actual)
	}
}
