package parse

import (
	"regexp"
	"strings"
)

func init() {
	var ops = make([]string, 0)
	for op, _ := range binaryOperators {
		// Because there is some overlap between operators (like "*" and "**") we have to
		// ensure that some ordering is forced.
		if op != "**" && op != "is not" && op != "//" && op != "not in" {
			ops = append(ops, regexp.QuoteMeta(op))
		}
	}
	// Additionally, we add the unary "not" operator since it has no binary counterpart.
	operatorTest = regexp.MustCompile(`^(not in|not|\*\*|is not|//|` + strings.Join(ops, "|") + ")")
}

var operatorTest *regexp.Regexp

type associativity int

const (
	operatorLeftAssoc associativity = iota
	operatorRightAssoc
	operatorNonAssoc
)

type operator struct {
	op         string
	precedence int
	assoc      associativity
	unary      bool
}

func (o operator) Operation() string {
	return o.op
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
	return o.op
}

var (
	UnaryNot = operator{"not", 50, operatorNonAssoc, true}
	UnaryPos = operator{"+", 500, operatorNonAssoc, true}
	UnaryNeg = operator{"-", 500, operatorNonAssoc, true}

	BinaryOr           = operator{"or", 10, operatorLeftAssoc, false}
	BinaryAnd          = operator{"and", 15, operatorLeftAssoc, false}
	BinaryBitwiseOr    = operator{"b-or", 16, operatorLeftAssoc, false}
	BinaryBitwiseXor   = operator{"b-xor", 17, operatorLeftAssoc, false}
	BinaryBitwiseAnd   = operator{"b-and", 18, operatorLeftAssoc, false}
	BinaryEqual        = operator{"==", 20, operatorLeftAssoc, false}
	BinaryNotEqual     = operator{"!=", 20, operatorLeftAssoc, false}
	BinaryLessThan     = operator{"<", 20, operatorLeftAssoc, false}
	BinaryLessEqual    = operator{"<=", 20, operatorLeftAssoc, false}
	BinaryGreaterThan  = operator{">", 20, operatorLeftAssoc, false}
	BinaryGreaterEqual = operator{">=", 20, operatorLeftAssoc, false}
	BinaryNotIn        = operator{"not in", 20, operatorLeftAssoc, false}
	BinaryIn           = operator{"in", 20, operatorLeftAssoc, false}
	BinaryMatches      = operator{"matches", 20, operatorLeftAssoc, false}
	BinaryStartsWith   = operator{"starts with", 20, operatorLeftAssoc, false}
	BinaryEndsWith     = operator{"ends with", 20, operatorLeftAssoc, false}
	BinaryRange        = operator{"..", 20, operatorLeftAssoc, false}
	BinaryAdd          = operator{"+", 30, operatorLeftAssoc, false}
	BinarySubtract     = operator{"-", 30, operatorLeftAssoc, false}
	BinaryConcat       = operator{"~", 40, operatorLeftAssoc, false}
	BinaryMultiply     = operator{"*", 60, operatorLeftAssoc, false}
	BinaryDivide       = operator{"/", 60, operatorLeftAssoc, false}
	BinaryFloorDiv     = operator{"//", 60, operatorLeftAssoc, false}
	BinaryModulo       = operator{"%", 60, operatorLeftAssoc, false}
	BinaryIs           = operator{"is", 100, operatorLeftAssoc, false}
	BinaryIsNot        = operator{"is not", 100, operatorLeftAssoc, false}
	BinaryPower        = operator{"**", 200, operatorRightAssoc, false}
)

var unaryOperators = map[string]operator{
	"not": UnaryNot,
	"+":   UnaryPos,
	"-":   UnaryNeg,
}

var binaryOperators = map[string]operator{
	"or":          BinaryOr,
	"and":         BinaryAnd,
	"b-or":        BinaryBitwiseOr,
	"b-xor":       BinaryBitwiseXor,
	"b-and":       BinaryBitwiseAnd,
	"==":          BinaryEqual,
	"!=":          BinaryNotEqual,
	"<":           BinaryLessThan,
	"<=":          BinaryLessEqual,
	">":           BinaryGreaterThan,
	">=":          BinaryGreaterEqual,
	"not in":      BinaryNotIn,
	"in":          BinaryIn,
	"matches":     BinaryMatches,
	"starts with": BinaryStartsWith,
	"ends with":   BinaryEndsWith,
	"..":          BinaryRange,
	"+":           BinaryAdd,
	"-":           BinarySubtract,
	"~":           BinaryConcat,
	"*":           BinaryMultiply,
	"/":           BinaryDivide,
	"//":          BinaryFloorDiv,
	"%":           BinaryModulo,
	"is":          BinaryIs,
	"is not":      BinaryIsNot,
	"**":          BinaryPower,
}
