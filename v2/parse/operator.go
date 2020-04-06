package parse

import (
	"regexp"
	"strings"
)

func init() {
	var ops = make([]string, 0)
	for op := range binaryOperators {
		// Because there is overlap between operators (like "*" and "**") we have to
		// ensure that some ordering is forced.
		if op != "**" && op != "is not" && op != "//" && op != "not in" && op != ">=" && op != "<=" {
			ops = append(ops, regexp.QuoteMeta(op))
		}
	}
	// Additionally, we add the unary "not" operator since it has no binary counterpart.
	operatorMatcher = regexp.MustCompile(`^(not in|not|\*\*|is not|//|>=|<=|` + strings.Join(ops, "|") + ")")
}

var operatorMatcher *regexp.Regexp

type associativity int

const (
	opLeftAssoc associativity = iota
	opRightAssoc
	opNonAssoc
)

type operator struct {
	op         string
	precedence int
	assoc      associativity
	unary      bool
}

// Built-in operators.
const (
	OpUnaryNot      = "not"
	OpUnaryPositive = "+"
	OpUnaryNegative = "-"

	OpBinaryOr           = "or"
	OpBinaryAnd          = "and"
	OpBinaryBitwiseOr    = "b-or"
	OpBinaryBitwiseXor   = "b-xor"
	OpBinaryBitwiseAnd   = "b-and"
	OpBinaryEqual        = "=="
	OpBinaryNotEqual     = "!="
	OpBinaryLessThan     = "<"
	OpBinaryLessEqual    = "<="
	OpBinaryGreaterThan  = ">"
	OpBinaryGreaterEqual = ">="
	OpBinaryNotIn        = "not in"
	OpBinaryIn           = "in"
	OpBinaryMatches      = "matches"
	OpBinaryStartsWith   = "starts with"
	OpBinaryEndsWith     = "ends with"
	OpBinaryRange        = ".."
	OpBinaryAdd          = "+"
	OpBinarySubtract     = "-"
	OpBinaryConcat       = "~"
	OpBinaryMultiply     = "*"
	OpBinaryDivide       = "/"
	OpBinaryFloorDiv     = "//"
	OpBinaryModulo       = "%"
	OpBinaryIs           = "is"
	OpBinaryIsNot        = "is not"
	OpBinaryPower        = "**"
)

func (o operator) Operator() string {
	return o.op
}

func (o operator) leftAssoc() bool {
	return o.assoc == opLeftAssoc
}

func (o operator) String() string {
	return o.op
}

var unaryOperators = map[string]operator{
	OpUnaryNot:      {OpUnaryNot, 50, opNonAssoc, true},
	OpUnaryPositive: {OpUnaryPositive, 500, opNonAssoc, true},
	OpUnaryNegative: {OpUnaryNegative, 500, opNonAssoc, true},
}

var binaryOperators = map[string]operator{
	OpBinaryOr:           {OpBinaryOr, 10, opLeftAssoc, false},
	OpBinaryAnd:          {OpBinaryAnd, 15, opLeftAssoc, false},
	OpBinaryBitwiseOr:    {OpBinaryBitwiseOr, 16, opLeftAssoc, false},
	OpBinaryBitwiseXor:   {OpBinaryBitwiseXor, 17, opLeftAssoc, false},
	OpBinaryBitwiseAnd:   {OpBinaryBitwiseAnd, 18, opLeftAssoc, false},
	OpBinaryEqual:        {OpBinaryEqual, 20, opLeftAssoc, false},
	OpBinaryNotEqual:     {OpBinaryNotEqual, 20, opLeftAssoc, false},
	OpBinaryLessThan:     {OpBinaryLessThan, 20, opLeftAssoc, false},
	OpBinaryLessEqual:    {OpBinaryLessEqual, 20, opLeftAssoc, false},
	OpBinaryGreaterThan:  {OpBinaryGreaterThan, 20, opLeftAssoc, false},
	OpBinaryGreaterEqual: {OpBinaryGreaterEqual, 20, opLeftAssoc, false},
	OpBinaryNotIn:        {OpBinaryNotIn, 20, opLeftAssoc, false},
	OpBinaryIn:           {OpBinaryIn, 20, opLeftAssoc, false},
	OpBinaryMatches:      {OpBinaryMatches, 20, opLeftAssoc, false},
	OpBinaryStartsWith:   {OpBinaryStartsWith, 20, opLeftAssoc, false},
	OpBinaryEndsWith:     {OpBinaryEndsWith, 20, opLeftAssoc, false},
	OpBinaryRange:        {OpBinaryRange, 20, opLeftAssoc, false},
	OpBinaryAdd:          {OpBinaryAdd, 30, opLeftAssoc, false},
	OpBinarySubtract:     {OpBinarySubtract, 30, opLeftAssoc, false},
	OpBinaryConcat:       {OpBinaryConcat, 40, opLeftAssoc, false},
	OpBinaryMultiply:     {OpBinaryMultiply, 60, opLeftAssoc, false},
	OpBinaryDivide:       {OpBinaryDivide, 60, opLeftAssoc, false},
	OpBinaryFloorDiv:     {OpBinaryFloorDiv, 60, opLeftAssoc, false},
	OpBinaryModulo:       {OpBinaryModulo, 60, opLeftAssoc, false},
	OpBinaryIs:           {OpBinaryIs, 100, opLeftAssoc, false},
	OpBinaryIsNot:        {OpBinaryIsNot, 100, opLeftAssoc, false},
	OpBinaryPower:        {OpBinaryPower, 200, opRightAssoc, false},
}
