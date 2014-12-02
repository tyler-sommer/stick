package parse

import "fmt"

type expr interface {
	node
}

const (
	exprName nodeType = iota
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
