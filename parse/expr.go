package parse

import "fmt"

// Expr represents a special type of Node that represents an expression.
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

type NullExpr struct {
	pos
}

func newNullExpr(pos pos) *NullExpr {
	return &NullExpr{pos}
}

func (exp *NullExpr) String() string {
	return "NULL"
}

type BoolExpr struct {
	pos
	value bool
}

func newBoolExpr(value bool, pos pos) *BoolExpr {
	return &BoolExpr{pos, value}
}

func (exp *BoolExpr) Value() bool {
	return exp.value
}

func (exp *BoolExpr) String() string {
	if exp.value {
		return "TRUE"
	}
	return "FALSE"
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

// FilterExpr represents a filter application.
type FilterExpr struct {
	*FuncExpr
}

func newFilterExpr(name *NameExpr, args []Expr, pos pos) *FilterExpr {
	return &FilterExpr{newFuncExpr(name, args, pos)}
}

// BinaryExpr represents a binary operation, such as "x + y"
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

// UnaryExpr represents a unary operation, such as "not x"
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

// GroupExpr represents an arbitrary wrapper around an inner expression.
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

// GetAttrExpr represents an attempt to retrieve an attribute from a value.
type GetAttrExpr struct {
	pos
	cont Expr
	attr Expr
	args []Expr
}

func newGetAttrExpr(cont Expr, attr Expr, args []Expr, pos pos) *GetAttrExpr {
	return &GetAttrExpr{pos, cont, attr, args}
}

func (exp *GetAttrExpr) String() string {
	if len(exp.args) > 0 {
		return fmt.Sprintf("GetAttrExpr(%s -> %s %v)", exp.cont, exp.attr, exp.args)
	}
	return fmt.Sprintf("GetAttrExpr(%s -> %s)", exp.cont, exp.attr)
}

func (exp *GetAttrExpr) Cont() Expr {
	return exp.cont
}

func (exp *GetAttrExpr) Attr() Expr {
	return exp.attr
}

func (exp *GetAttrExpr) Args() []Expr {
	return exp.args
}
