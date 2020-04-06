package parse

import "fmt"

// Expr represents a special type of Node that represents an expression.
type Expr interface {
	Node
}

// NameExpr represents an identifier, such as a variable.
type NameExpr struct {
	Pos
	Name string // Name of the identifier.
}

// NewNameExpr returns a NameExpr.
func NewNameExpr(name string, pos Pos) *NameExpr {
	return &NameExpr{pos, name}
}

// All returns all the child Nodes in a NameExpr.
func (exp *NameExpr) All() []Node {
	return []Node{}
}

// String returns a string representation of the NameExpr.
func (exp *NameExpr) String() string {
	return fmt.Sprintf("NameExpr(%s)", exp.Name)
}

// NullExpr represents a null literal.
type NullExpr struct {
	Pos
}

// All returns all the child Nodes in a NullExpr.
func (exp *NullExpr) All() []Node {
	return []Node{}
}

// NewNullExpr returns a NullExpr.
func NewNullExpr(pos Pos) *NullExpr {
	return &NullExpr{pos}
}

// String returnsa string representation of the NullExpr.
func (exp *NullExpr) String() string {
	return "NULL"
}

// BoolExpr represents a boolean literal.
type BoolExpr struct {
	Pos
	Value bool // The raw boolean value.
}

// NewBoolExpr returns a BoolExpr.
func NewBoolExpr(value bool, pos Pos) *BoolExpr {
	return &BoolExpr{pos, value}
}

// All returns all the child Nodes in a UseNode.
func (exp *BoolExpr) All() []Node {
	return []Node{}
}

// String returns a string representation of the BoolExpr.
func (exp *BoolExpr) String() string {
	if exp.Value {
		return "TRUE"
	}
	return "FALSE"
}

// NumberExpr represents a number literal.
type NumberExpr struct {
	Pos
	Value string // The string representation of the number.
}

// NewNumberExpr returns a NumberExpr.
func NewNumberExpr(val string, pos Pos) *NumberExpr {
	return &NumberExpr{pos, val}
}

// All returns all the child Nodes in a NumberExpr.
func (exp *NumberExpr) All() []Node {
	return []Node{}
}

// String returns a string representation of the NumberExpr.
func (exp *NumberExpr) String() string {
	return fmt.Sprintf("NumberExpr(%s)", exp.Value)
}

// StringExpr represents a string literal.
type StringExpr struct {
	Pos
	Text string // The text contained within the literal.
}

// NewStringExpr returns a StringExpr.
func NewStringExpr(text string, pos Pos) *StringExpr {
	return &StringExpr{pos, text}
}

// All returns all the child Nodes in a StringExpr.
func (exp *StringExpr) All() []Node {
	return []Node{}
}

// String returns a string representation of the StringExpr.
func (exp *StringExpr) String() string {
	return fmt.Sprintf("StringExpr(%s)", exp.Text)
}

// FuncExpr represents a function call.
type FuncExpr struct {
	Pos
	Name string // The name of the function.
	Args []Expr // Arguments to be passed to the function.
}

// All returns all the child Nodes in a FuncExpr.
func (exp *FuncExpr) All() []Node {
	res := make([]Node, len(exp.Args))
	for i, n := range exp.Args {
		res[i] = n
	}
	return res
}

// NewFuncExpr returns a FuncExpr.
func NewFuncExpr(name string, args []Expr, pos Pos) *FuncExpr {
	return &FuncExpr{pos, name, args}
}

// String returns a string representation of a FuncExpr.
func (exp *FuncExpr) String() string {
	return fmt.Sprintf("FuncExpr(%s, %s)", exp.Name, exp.Args)
}

// FilterExpr represents a filter application.
type FilterExpr struct {
	*FuncExpr
}

// String returns a string representation of the FilterExpr.
func (exp *FilterExpr) String() string {
	return fmt.Sprintf("FilterExpr(%s, %s)", exp.Name, exp.Args)
}

// NewFilterExpr returns a FilterExpr.
func NewFilterExpr(name string, args []Expr, pos Pos) *FilterExpr {
	return &FilterExpr{NewFuncExpr(name, args, pos)}
}

// TestExpr represents a boolean test expression.
type TestExpr struct {
	*FuncExpr
}

// String returns a string representation of the TestExpr.
func (exp *TestExpr) String() string {
	return fmt.Sprintf("TestExpr(%s, %s)", exp.Name, exp.Args)
}

// NewTestExpr returns a TestExpr.
func NewTestExpr(name string, args []Expr, pos Pos) *TestExpr {
	return &TestExpr{NewFuncExpr(name, args, pos)}
}

// BinaryExpr represents a binary operation, such as "x + y"
type BinaryExpr struct {
	Pos
	Left  Expr   // Left side expression.
	Op    string // Binary operation in string form.
	Right Expr   // Right side expression.
}

// NewBinaryExpr returns a BinaryExpr.
func NewBinaryExpr(left Expr, op string, right Expr, pos Pos) *BinaryExpr {
	return &BinaryExpr{pos, left, op, right}
}

// All returns all the child Nodes in a BinaryExpr.
func (exp *BinaryExpr) All() []Node {
	return []Node{exp.Left, exp.Right}
}

// String returns a string representation of the BinaryExpr.
func (exp *BinaryExpr) String() string {
	return fmt.Sprintf("BinaryExpr(%s %s %s)", exp.Left, exp.Op, exp.Right)
}

// UnaryExpr represents a unary operation, such as "not x"
type UnaryExpr struct {
	Pos
	Op string // The operation, in string form.
	X  Expr   // Expression to be evaluated.
}

// NewUnaryExpr returns a new UnaryExpr.
func NewUnaryExpr(op string, expr Expr, pos Pos) *UnaryExpr {
	return &UnaryExpr{pos, op, expr}
}

// All returns all the child Nodes in a UnaryExpr.
func (exp *UnaryExpr) All() []Node {
	return []Node{exp.X}
}

// String returns a string representation of a UnaryExpr.
func (exp *UnaryExpr) String() string {
	return fmt.Sprintf("UnaryExpr(%s %s)", exp.Op, exp.X)
}

// GroupExpr represents an arbitrary wrapper around an inner expression.
type GroupExpr struct {
	Pos
	X Expr // Expression to be evaluated.
}

// NewGroupExpr returns a GroupExpr.
func NewGroupExpr(inner Expr, pos Pos) *GroupExpr {
	return &GroupExpr{pos, inner}
}

// All returns all the child Nodes in a GroupExpr.
func (exp *GroupExpr) All() []Node {
	return []Node{exp.X}
}

// String returns a string representation of a GroupExpr.
func (exp *GroupExpr) String() string {
	return fmt.Sprintf("GroupExpr(%s)", exp.X)
}

// GetAttrExpr represents an attempt to retrieve an attribute from a value.
type GetAttrExpr struct {
	Pos
	Cont Expr   // Container to get attribute from.
	Attr Expr   // Attribute to get.
	Args []Expr // Args to pass to attribute, if its a method.
}

// NewGetAttrExpr returns a GetAttrExpr.
func NewGetAttrExpr(cont Expr, attr Expr, args []Expr, pos Pos) *GetAttrExpr {
	return &GetAttrExpr{pos, cont, attr, args}
}

// All returns all the child Nodes in a GetAttrExpr.
func (exp *GetAttrExpr) All() []Node {
	res := []Node{exp.Cont, exp.Attr}
	for _, v := range exp.Args {
		res = append(res, v)
	}
	return res
}

// String returns a string representation of a GetAttrExpr.
func (exp *GetAttrExpr) String() string {
	if len(exp.Args) > 0 {
		return fmt.Sprintf("GetAttrExpr(%s -> %s %v)", exp.Cont, exp.Attr, exp.Args)
	}
	return fmt.Sprintf("GetAttrExpr(%s -> %s)", exp.Cont, exp.Attr)
}

// TernaryIfExpr represents an attempt to retrieve an attribute from a value.
type TernaryIfExpr struct {
	Pos
	Cond   Expr // Condition to test.
	TrueX  Expr // Expression if Cond is true.
	FalseX Expr // Expression if Cond is false.
}

// NewTernaryIfExpr returns a TernaryIfExpr.
func NewTernaryIfExpr(cond, tx, fx Expr, pos Pos) *TernaryIfExpr {
	return &TernaryIfExpr{pos, cond, tx, fx}
}

// All returns all the child Nodes in a TernaryIfExpr.
func (exp *TernaryIfExpr) All() []Node {
	return []Node{exp.Cond, exp.TrueX, exp.FalseX}
}

// String returns a string representation of a TernaryIfExpr.
func (exp *TernaryIfExpr) String() string {
	return fmt.Sprintf("%s ? %s : %v", exp.Cond, exp.TrueX, exp.FalseX)
}

type KeyValueExpr struct {
	Pos
	Key   Expr
	Value Expr
}

// NewKeyValueExpr returns a KeyValueExpr.
func NewKeyValueExpr(k, v Expr, pos Pos) *KeyValueExpr {
	return &KeyValueExpr{pos, k, v}
}

// All returns all the child Nodes in a KeyValueExpr.
func (exp *KeyValueExpr) All() []Node {
	return []Node{exp.Key, exp.Value}
}

// String returns a string representation of a KeyValueExpr.
func (exp *KeyValueExpr) String() string {
	return fmt.Sprintf("%s: %s", exp.Key, exp.Value)
}

type HashExpr struct {
	Pos
	Elements []*KeyValueExpr
}

// NewHashExpr returns a HashExpr.
func NewHashExpr(pos Pos, elements ...*KeyValueExpr) *HashExpr {
	return &HashExpr{pos, elements}
}

// All returns all the child Nodes in a HashExpr.
func (exp *HashExpr) All() []Node {
	all := make([]Node, len(exp.Elements))
	for i, v := range exp.Elements {
		all[i] = v
	}
	return all
}

// String returns a string representation of a HashExpr.
func (exp *HashExpr) String() string {
	return fmt.Sprintf("HashExpr{%v}", exp.Elements)
}

type ArrayExpr struct {
	Pos
	Elements []Expr
}

// NewArrayExpr returns a ArrayExpr.
func NewArrayExpr(pos Pos, els ...Expr) *ArrayExpr {
	return &ArrayExpr{pos, els}
}

// All returns all the child Nodes in a ArrayExpr.
func (exp *ArrayExpr) All() []Node {
	all := make([]Node, len(exp.Elements))
	for i, v := range exp.Elements {
		all[i] = v
	}
	return all
}

// String returns a string representation of a ArrayExpr.
func (exp *ArrayExpr) String() string {
	return fmt.Sprintf("ArrayExpr%v", exp.Elements)
}
