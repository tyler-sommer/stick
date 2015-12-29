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

type associativity int

const (
	operatorLeftAssoc associativity = iota
	operatorRightAssoc
)

type operator struct {
	operand string
	precedence int
	assoc associativity
	unary bool
}

func (o operator) Operand() string {
	return o.operand
}

func (o operator) Precedence() int {
	return o.precedence
}

func (o operator) IsLeftAssociative() bool {
	return o.assoc == operatorLeftAssoc
}

func (o operator) IsUnary() bool {
	return o.unary
}

func (o operator) IsBinary() bool {
	return !o.unary
}

func (o operator) String() string {
	return o.operand
}

var Operators = map[string]operator{
	"+": operator{"+", 30, operatorLeftAssoc, false},
	"-": operator{"-", 30, operatorLeftAssoc, false},
	"*": operator{"*", 60, operatorLeftAssoc, false},
	"/": operator{"/", 60, operatorLeftAssoc, false},
	"%": operator{"%", 60, operatorLeftAssoc, false},
	"**": operator{"**", 200, operatorRightAssoc, false},
}

type BinaryExpr struct {
	pos
	left    Expr
	operand operator
	right   Expr
}

func newBinaryExpr(left Expr, operand operator, right Expr, pos pos) *BinaryExpr {
	return &BinaryExpr{pos, left, operand, right}
}

func (exp *BinaryExpr) String() string {
	return fmt.Sprintf("BinaryExpr(%s %s %s)", exp.left, exp.operand, exp.right)
}

type GroupExpr struct {
	pos
	inner Expr
}

func newGroupExpr(inner Expr, pos pos) *GroupExpr {
	return &GroupExpr{pos, inner}
}

func (exp *GroupExpr) String() string {
	return fmt.Sprintf("GroupExpr(%s)", exp.inner)
}
