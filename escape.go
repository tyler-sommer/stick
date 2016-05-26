package stick

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tyler-sommer/stick/parse"
)

type Escaper func(string) string

func escapeHTML(in string) string {
	var out = &bytes.Buffer{}
	for _, c := range in {
		if c == 34 {
			// "
			out.WriteString("&quot;")
		} else if c == 38 {
			// &
			out.WriteString("&amp;")
		} else if c == 39 {
			// '
			out.WriteString("&#39;")
		} else if c == 60 {
			// <
			out.WriteString("&lt;")
		} else if c == 62 {
			// >
			out.WriteString("&gt;")
		} else {
			// UTF-8
			out.WriteRune(c)
		}
	}
	return out.String()
}

func escapeHTMLAttribute(in string) string {
	var out = &bytes.Buffer{}
	for _, c := range in {
		if (c >= 65 && c <= 90) || (c >= 97 && c <= 122) || (c >= 48 && c <= 57) || (c >= 44 && c <= 46) || c == 95 {
			// a-zA-Z0-9,.-_
			out.WriteRune(c)
		} else if c == 34 {
			// "
			out.WriteString("&quot;")
		} else if c == 38 {
			// &
			out.WriteString("&amp;")
		} else if c == 60 {
			// <
			out.WriteString("&lt;")
		} else if c == 62 {
			// >
			out.WriteString("&gt;")
		} else if c <= 31 && c != 9 && c != 10 && c != 13 {
			// Non-whitespace
			out.WriteString("&#xFFFD;")
		} else {
			// UTF-8
			fmt.Fprintf(out, "&#%d;", c)
		}
	}
	return out.String()
}

func escapeJS(in string) string {
	var out = &bytes.Buffer{}
	for _, c := range in {
		if (c >= 65 && c <= 90) || (c >= 97 && c <= 122) || (c >= 48 && c <= 57) || c == 44 || c == 46 || c == 95 {
			// a-zA-Z0-9,._
			out.WriteRune(c)
		} else {
			// UTF-8
			fmt.Fprintf(out, "\\u%04X", c)
		}
	}
	return out.String()
}

func escapeCSS(in string) string {
	var out = &bytes.Buffer{}
	for _, c := range in {
		if (c >= 65 && c <= 90) || (c >= 97 && c <= 122) || (c >= 48 && c <= 57) {
			// a-zA-Z0-9
			out.WriteRune(c)
		} else {
			// UTF-8
			fmt.Fprintf(out, "\\%04X", c)
		}
	}
	return out.String()
}

func escapeURL(in string) string {
	var out = &bytes.Buffer{}
	var c byte
	for i := 0; i < len(in); i++ {
		c = in[i]
		if (c >= 65 && c <= 90) || (c >= 97 && c <= 122) || (c >= 48 && c <= 57) || c == 45 || c == 46 || c == 126 || c == 95 {
			// a-zA-Z0-9-._~
			out.WriteByte(c)
		} else {
			// UTF-8
			fmt.Fprintf(out, "%%%02X", c)
		}
	}
	return out.String()
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
	case *parse.ModuleNode, *parse.BlockNode:
		v.pop()
	}
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
