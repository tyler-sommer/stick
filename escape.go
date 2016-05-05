package stick

import (
	"html"

	"net/url"

	"github.com/tyler-sommer/stick/parse"
)

type Escaper func(string) string

func escapeHTML(in string) string {
	return html.EscapeString(in)
}

func escapeHTMLAttribute(in string) string {
	// TODO: Implementation
	return in
}

func escapeJS(in string) string {
	// TODO: Implementation
	return in
}

func escapeCSS(in string) string {
	// TODO: Implementation
	return in
}

func escapeURL(in string) string {
	u, err := url.Parse(in)
	if err != nil {
		return ""
	}
	return u.String()
}

type AutoEscapeExtension struct {
	Escapers map[string]Escaper
}

func (e *AutoEscapeExtension) Init(env *Env) error {
	env.Visitors = append(env.Visitors, &AutoEscapeVisitor{})
	env.Filters["escape"] = EscapeFilter
	return nil
}

func NewAutoEscapeExtension() *AutoEscapeExtension {
	return &AutoEscapeExtension{
		Escapers: map[string]Escaper{
			"html":      escapeHTML,
			"html_attr": escapeHTMLAttribute,
			"js":        escapeJS,
			"css":       escapeCSS,
			"url":       escapeURL,
		},
	}
}

// AutoEscapeVisitor can be used to automatically apply the "escape" filter
// to any PrintNode.
type AutoEscapeVisitor struct {
	stack []string
}

// push adds a scope on top of the stack.
func (v *AutoEscapeVisitor) push(name string) {
	v.stack = append(v.stack, name)
}

// pop removes the top-most scope.
func (v *AutoEscapeVisitor) pop() {
	v.stack = v.stack[0 : len(v.stack)-1]
}

func (v *AutoEscapeVisitor) Enter(n parse.Node) {
	switch node := n.(type) {
	case *parse.ModuleNode:
		v.stack = []string{node.Origin}
	case *parse.BlockNode:
		v.push(node.Origin)
	case *parse.PrintNode:
		v := node.X
		r := &parse.FilterExpr{&parse.FuncExpr{v.Start(), "escape", []parse.Expr{v}}}
		node.X = r
	}
}

func (v *AutoEscapeVisitor) Leave(n parse.Node) {
	switch n.(type) {
	case *parse.ModuleNode:
		v.pop()
	case *parse.BlockNode:
		v.pop()
	}
}

func EscapeFilter(ctx Context, val Value, args ...Value) Value {
	if _, ok := val.(SafeValue); ok {
		return val
	}
	// TODO: Context-sensitive escaping is a must, currently only escapes HTML.
	return NewSafeValue(html.EscapeString(CoerceString(val)))
}
