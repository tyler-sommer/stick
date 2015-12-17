package parse

import "fmt"

type expr interface {
	Node
}

const (
	exprName nodeType = iota
	exprString
)

type nameExpr struct {
	nodeType
	pos
	name string
}

func newNameExpr(name string) *nameExpr {
	return &nameExpr{exprName, 0, name}
}

func (exp *nameExpr) String() string {
	return fmt.Sprintf("NameExpr(%s)", exp.name)
}

type stringExpr struct {
	nodeType
	pos
	text string
}

func newStringExpr(text string) *stringExpr {
	return &stringExpr{exprString, 0, text}
}

func (exp *stringExpr) String() string {
	return fmt.Sprintf("StringExpr(%s)", exp.text)
}
