package parse

import "fmt"

type expr interface {
	Node
}

const (
	exprName nodeType = iota
	exprString
	exprFunc
)

type NameExpr struct {
	nodeType
	pos
	name string
}

func newNameExpr(name string) *NameExpr {
	return &NameExpr{exprName, 0, name}
}

func (exp *NameExpr) Name() string {
	return exp.name
}

func (exp *NameExpr) String() string {
	return fmt.Sprintf("NameExpr(%s)", exp.name)
}

type StringExpr struct {
	nodeType
	pos
	text string
}

func newStringExpr(text string) *StringExpr {
	return &StringExpr{exprString, 0, text}
}

func (exp *StringExpr) String() string {
	return fmt.Sprintf("StringExpr(%s)", exp.text)
}

type FuncExpr struct {
	nodeType
	pos
	name string
	args []expr
}

func newFuncExpr(name string, args []expr) *FuncExpr {
	return &FuncExpr{exprFunc, 0, name, args}
}

func (exp *FuncExpr) String() string {
	return fmt.Sprintf("FuncExpr(%s, [%s])", exp.name, exp.args)
}
