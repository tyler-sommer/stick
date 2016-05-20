package stick

import (
	"html"

	"net/url"

	"strings"

	"regexp"

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
	env.Filters["escape"] = func(ctx Context, val Value, args ...Value) Value {
		ct := "html"
		if len(args) > 0 {
			ct = CoerceString(args[0])
		}

		if sval, ok := val.(SafeValue); ok {
			if sval.IsSafe(ct) {
				return val
			}
		}

		escfn, ok := e.Escapers[ct]
		if !ok {
			// TODO: Communicate error
			return NewSafeValue("", ct)
		}

		return NewSafeValue(escfn(CoerceString(val)), ct)
	}
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
func (v *AutoEscapeVisitor) pop() string {
	ret := v.current()
	v.stack = v.stack[0 : len(v.stack)-1]
	return ret
}

func (v *AutoEscapeVisitor) current() string {
	if len(v.stack) == 0 {
		// TODO: This is an invalid state.
		return ""
	}
	return v.stack[len(v.stack)-1]
}

func (v *AutoEscapeVisitor) Enter(n parse.Node) {
	switch node := n.(type) {
	case *parse.ModuleNode:
		v.push(v.guessTypeFromName(node.Origin))
	case *parse.BlockNode:
		v.push(v.guessTypeFromName(node.Origin))
	case *parse.TextNode:
		v.push(v.guessTypeFromData(node.Data))
	case *parse.PrintNode:
		ct := v.current()
		v := node.X
		r := parse.NewFilterExpr(
			"escape",
			[]parse.Expr{v, parse.NewStringExpr(ct, v.Start())},
			v.Start(),
		)
		node.X = r
	}
}

func (v *AutoEscapeVisitor) Leave(n parse.Node) {
	switch n.(type) {
	case *parse.ModuleNode, *parse.BlockNode, *parse.TextNode:
		v.pop()
	}
}

var matcher = regexp.MustCompile("([a-zA-Z]+?)=\"$")

func (v *AutoEscapeVisitor) guessTypeFromData(data string) string {
	if v.current() != "html" {
		return v.current()
	}
	m := matcher.FindStringSubmatch(data)
	if len(m) == 2 {
		// TODO: This is extremely naive
		if m[2] == "href" || m[2] == "src" {
			return "url"
		}
		return "html_attr"
	}
	return v.current()
}

func (v *AutoEscapeVisitor) guessTypeFromName(name string) string {
	name = strings.TrimSuffix(name, ".twig")
	p := strings.LastIndex(name, ".")
	if p < 0 {
		// Default to html
		return "html"
	}
	return name[p:]
}
