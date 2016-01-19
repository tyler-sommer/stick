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

// NumberExpr represents a number literal.
type NumberExpr struct {
	pos
	value string
}

func newNumberExpr(val string, pos pos) *NumberExpr {
	return &NumberExpr{pos, val}
}

func (exp *NumberExpr) Value() string {
	return exp.value
}

func (exp *NumberExpr) String() string {
	return fmt.Sprintf("NumberExpr(%s)", exp.value)
}

// StringExpr represents a string literal.
type StringExpr struct {
	pos
	text string
}

func newStringExpr(text string, pos pos) *StringExpr {
	return &StringExpr{pos, text}
}

func (exp *StringExpr) Value() string {
	return exp.text
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

func (exp *FuncExpr) Name() string {
	return exp.name.Name()
}

func (exp *FuncExpr) Args() []Expr {
	return exp.args
}

type BinaryExpr struct {
	pos
	left  Expr
	op    string
	right Expr
}

func newBinaryExpr(left Expr, op string, right Expr, pos pos) *BinaryExpr {
	return &BinaryExpr{pos, left, op, right}
}

func (exp *BinaryExpr) Left() Expr {
	return exp.left
}

func (exp *BinaryExpr) Right() Expr {
	return exp.right
}

func (exp *BinaryExpr) Op() string {
	return exp.op
}

func (exp *BinaryExpr) String() string {
	return fmt.Sprintf("BinaryExpr(%s %s %s)", exp.left, exp.op, exp.right)
}

type UnaryExpr struct {
	pos
	op   string
	expr Expr
}

func newUnaryExpr(op string, expr Expr, pos pos) *UnaryExpr {
	return &UnaryExpr{pos, op, expr}
}

func (exp *UnaryExpr) Expr() Expr {
	return exp.expr
}

func (exp *UnaryExpr) Op() string {
	return exp.op
}

func (exp *UnaryExpr) String() string {
	return fmt.Sprintf("UnaryExpr(%s %s)", exp.op, exp.expr)
}

type GroupExpr struct {
	pos
	inner Expr
}

func newGroupExpr(inner Expr, pos pos) *GroupExpr {
	return &GroupExpr{pos, inner}
}

func (exp *GroupExpr) Inner() Expr {
	return exp.inner
}

func (exp *GroupExpr) String() string {
	return fmt.Sprintf("GroupExpr(%s)", exp.inner)
}

type GetAttrExpr struct {
	pos
	cont Expr
	attr Expr
}

func newGetAttrExpr(cont Expr, attr Expr, pos pos) *GetAttrExpr {
	return &GetAttrExpr{pos, cont, attr}
}

func (exp *GetAttrExpr) String() string {
	return fmt.Sprintf("GetAttrExpr(%s -> %s)", exp.cont, exp.attr)
}
