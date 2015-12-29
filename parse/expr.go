package parse

import "fmt"

type Expr interface {
	Node
}

// NameExpr represents an identifier, such as a variable.
type NameExpr struct {
	pos
	name string
}

func newNameExpr(name string, pos pos) *NameExpr {
	return &NameExpr{pos, name}
}

func (exp *NameExpr) Name() string {
	return exp.name
}

func (exp *NameExpr) String() string {
	return fmt.Sprintf("NameExpr(%s)", exp.name)
}

// StringExpr represents a string literal.
type StringExpr struct {
	pos
	text string
}

func newStringExpr(text string, pos pos) *StringExpr {
	return &StringExpr{pos, text}
}

func (exp *StringExpr) String() string {
	return fmt.Sprintf("StringExpr(%s)", exp.text)
}

// FuncExpr represents a function call.
type FuncExpr struct {
	pos
	name *NameExpr
	args []Expr
}

func newFuncExpr(name *NameExpr, args []Expr, pos pos) *FuncExpr {
	return &FuncExpr{pos, name, args}
}

func (exp *FuncExpr) String() string {
	return fmt.Sprintf("FuncExpr(%s, %s)", exp.name, exp.args)
}
