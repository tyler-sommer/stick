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

// NameExpr represents an identifier, such as a variable.
type NameExpr struct {
	nodeType
	pos
	name string
}

func newNameExpr(name string, pos pos) *NameExpr {
	return &NameExpr{exprName, pos, name}
}

func (exp *NameExpr) Name() string {
	return exp.name
}

func (exp *NameExpr) String() string {
	return fmt.Sprintf("NameExpr(%s)", exp.name)
}

// StringExpr represents a string literal.
type StringExpr struct {
	nodeType
	pos
	text string
}

func newStringExpr(text string, pos pos) *StringExpr {
	return &StringExpr{exprString, pos, text}
}

func (exp *StringExpr) String() string {
	return fmt.Sprintf("StringExpr(%s)", exp.text)
}

// FuncExpr represents a function call.
type FuncExpr struct {
	nodeType
	pos
	name *NameExpr
	args []expr
}

func newFuncExpr(name *NameExpr, args []expr, pos pos) *FuncExpr {
	return &FuncExpr{exprFunc, pos, name, args}
}

func (exp *FuncExpr) String() string {
	return fmt.Sprintf("FuncExpr(%s, %s)", exp.name, exp.args)
}
