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

// All returns all the child Nodes in a NameExpr.
func (exp *NameExpr) All() []Node {
	return []Node{}
}

// Name returns the expression's name.
func (exp *NameExpr) Name() string {
	return exp.name
}

// String returns a string representation of the NameExpr.
func (exp *NameExpr) String() string {
	return fmt.Sprintf("NameExpr(%s)", exp.name)
}

// NullExpr represents a null literal.
type NullExpr struct {
	pos
}

// All returns all the child Nodes in a NullExpr.
func (exp *NullExpr) All() []Node {
	return []Node{}
}

func newNullExpr(pos pos) *NullExpr {
	return &NullExpr{pos}
}

// String returnsa string representation of the NullExpr.
func (exp *NullExpr) String() string {
	return "NULL"
}

// BoolExpr represents a boolean literal.
type BoolExpr struct {
	pos
	value bool
}

func newBoolExpr(value bool, pos pos) *BoolExpr {
	return &BoolExpr{pos, value}
}

// All returns all the child Nodes in a UseNode.
func (exp *BoolExpr) All() []Node {
	return []Node{}
}

// Value returns the boolean value stored in the expression.
func (exp *BoolExpr) Value() bool {
	return exp.value
}

// String returns a string representation of the BoolExpr.
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

// All returns all the child Nodes in a NumberExpr.
func (exp *NumberExpr) All() []Node {
	return []Node{}
}

// Value returns the value stored in the NumberExpr.
func (exp *NumberExpr) Value() string {
	return exp.value
}

// String returns a string representation of the NumberExpr.
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

// All returns all the child Nodes in a StringExpr.
func (exp *StringExpr) All() []Node {
	return []Node{}
}

// Value returns the stored string value.
func (exp *StringExpr) Value() string {
	return exp.text
}

// String returns a string representation of the StringExpr.
func (exp *StringExpr) String() string {
	return fmt.Sprintf("StringExpr(%s)", exp.text)
}

// FuncExpr represents a function call.
type FuncExpr struct {
	pos
	name *NameExpr
	args []Expr
}

// All returns all the child Nodes in a FuncExpr.
func (exp *FuncExpr) All() []Node {
	res := []Node{exp.name}
	for _, n := range exp.args {
		res = append(res, n)
	}
	return res
}

func newFuncExpr(name *NameExpr, args []Expr, pos pos) *FuncExpr {
	return &FuncExpr{pos, name, args}
}

// String returns a string representation of a FuncExpr.
func (exp *FuncExpr) String() string {
	return fmt.Sprintf("FuncExpr(%s, %s)", exp.name, exp.args)
}

// Name returns the name of the function to be called.
func (exp *FuncExpr) Name() string {
	return exp.name.Name()
}

// Args returns any arguments that should be evaluated and passed
// into the function.
func (exp *FuncExpr) Args() []Expr {
	return exp.args
}

// FilterExpr represents a filter application.
type FilterExpr struct {
	*FuncExpr
}

// String returns a string representation of the FilterExpr.
func (exp *FilterExpr) String() string {
	return fmt.Sprintf("FilterExpr(%s, %s)", exp.name, exp.args)
}

func newFilterExpr(name *NameExpr, args []Expr, pos pos) *FilterExpr {
	return &FilterExpr{newFuncExpr(name, args, pos)}
}

// TestExpr represents a boolean test expression.
type TestExpr struct {
	*FuncExpr
}

// String returns a string representation of the TestExpr.
func (exp *TestExpr) String() string {
	return fmt.Sprintf("TestExpr(%s, %s)", exp.name, exp.args)
}

func newTestExpr(name *NameExpr, args []Expr, pos pos) *TestExpr {
	return &TestExpr{newFuncExpr(name, args, pos)}
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

// All returns all the child Nodes in a BinaryExpr.
func (exp *BinaryExpr) All() []Node {
	return []Node{exp.left, exp.right}
}

// Left returns the left operand expression.
func (exp *BinaryExpr) Left() Expr {
	return exp.left
}

// Right returns the right operand expression.
func (exp *BinaryExpr) Right() Expr {
	return exp.right
}

// Op returns the operation to be performed.
func (exp *BinaryExpr) Op() string {
	return exp.op
}

// String returns a string representation of the BinaryExpr.
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

// All returns all the child Nodes in a UnaryExpr.
func (exp *UnaryExpr) All() []Node {
	return []Node{exp.expr}
}

// Expr returns the expression to be evaluated and operated on.
func (exp *UnaryExpr) Expr() Expr {
	return exp.expr
}

// Op returns the operation to be performed.
func (exp *UnaryExpr) Op() string {
	return exp.op
}

// String returns a string representation of a UnaryExpr.
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

// All returns all the child Nodes in a GroupExpr.
func (exp *GroupExpr) All() []Node {
	return []Node{exp.inner}
}

// Inner returns the inner expression of a GroupExpr.
func (exp *GroupExpr) Inner() Expr {
	return exp.inner
}

// String returns a string representation of a GroupExpr.
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

// All returns all the child Nodes in a GetAttrExpr.
func (exp *GetAttrExpr) All() []Node {
	res := []Node{exp.cont, exp.attr}
	for _, v := range exp.args {
		res = append(res, v)
	}
	return res
}

// String returns a string representation of a GetAttrExpr.
func (exp *GetAttrExpr) String() string {
	if len(exp.args) > 0 {
		return fmt.Sprintf("GetAttrExpr(%s -> %s %v)", exp.cont, exp.attr, exp.args)
	}
	return fmt.Sprintf("GetAttrExpr(%s -> %s)", exp.cont, exp.attr)
}

// Cont returns the container expression to be evaluated.
func (exp *GetAttrExpr) Cont() Expr {
	return exp.cont
}

// Attr returns the expression to be used as the attribute name.
func (exp *GetAttrExpr) Attr() Expr {
	return exp.attr
}

// Args returns any arguments that should be used during attribute fetching.
func (exp *GetAttrExpr) Args() []Expr {
	return exp.args
}
