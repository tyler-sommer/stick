package stick

import (
	"html"

	"github.com/tyler-sommer/stick/parse"
)

// AutoEscapeVisitor can be used to automatically apply the "escape" filter
// to any PrintNode.
type AutoEscapeVisitor struct {
}

func (v *AutoEscapeVisitor) Enter(n parse.Node) {
	if node, ok := n.(*parse.PrintNode); ok {
		v := node.X
		r := &parse.FilterExpr{&parse.FuncExpr{v.Start(), "escape", []parse.Expr{v}}}
		node.X = r
	}
}

func (v *AutoEscapeVisitor) Leave(n parse.Node) {}

// EscapeFilter can be added to an Env and used to escape content.
//
// At present, this filter only handles HTML input.
var EscapeFilter = func(e *Env, val Value, args ...Value) Value {
	if _, ok := val.(SafeValue); ok {
		return val
	}
	// TODO: Context-sensitive escaping is a must, currently only escapes HTML.
	return NewSafeValue(html.EscapeString(CoerceString(val)))
}
