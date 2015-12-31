package parse

import (
	"regexp"
	"strings"
)

func init() {
	var ops = make([]string, 0)
	for op, _ := range binaryOperators {
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
	OpUnaryNot:      operator{OpUnaryNot, 50, opNonAssoc, true},
	OpUnaryPositive: operator{OpUnaryPositive, 500, opNonAssoc, true},
	OpUnaryNegative: operator{OpUnaryNegative, 500, opNonAssoc, true},
}

var binaryOperators = map[string]operator{
	OpBinaryOr:           operator{OpBinaryOr, 10, opLeftAssoc, false},
	OpBinaryAnd:          operator{OpBinaryAnd, 15, opLeftAssoc, false},
	OpBinaryBitwiseOr:    operator{OpBinaryBitwiseOr, 16, opLeftAssoc, false},
	OpBinaryBitwiseXor:   operator{OpBinaryBitwiseXor, 17, opLeftAssoc, false},
	OpBinaryBitwiseAnd:   operator{OpBinaryBitwiseAnd, 18, opLeftAssoc, false},
	OpBinaryEqual:        operator{OpBinaryEqual, 20, opLeftAssoc, false},
	OpBinaryNotEqual:     operator{OpBinaryNotEqual, 20, opLeftAssoc, false},
	OpBinaryLessThan:     operator{OpBinaryLessThan, 20, opLeftAssoc, false},
	OpBinaryLessEqual:    operator{OpBinaryLessEqual, 20, opLeftAssoc, false},
	OpBinaryGreaterThan:  operator{OpBinaryGreaterThan, 20, opLeftAssoc, false},
	OpBinaryGreaterEqual: operator{OpBinaryGreaterEqual, 20, opLeftAssoc, false},
	OpBinaryNotIn:        operator{OpBinaryNotIn, 20, opLeftAssoc, false},
	OpBinaryIn:           operator{OpBinaryIn, 20, opLeftAssoc, false},
	OpBinaryMatches:      operator{OpBinaryMatches, 20, opLeftAssoc, false},
	OpBinaryStartsWith:   operator{OpBinaryStartsWith, 20, opLeftAssoc, false},
	OpBinaryEndsWith:     operator{OpBinaryEndsWith, 20, opLeftAssoc, false},
	OpBinaryRange:        operator{OpBinaryRange, 20, opLeftAssoc, false},
	OpBinaryAdd:          operator{OpBinaryAdd, 30, opLeftAssoc, false},
	OpBinarySubtract:     operator{OpBinarySubtract, 30, opLeftAssoc, false},
	OpBinaryConcat:       operator{OpBinaryConcat, 40, opLeftAssoc, false},
	OpBinaryMultiply:     operator{OpBinaryMultiply, 60, opLeftAssoc, false},
	OpBinaryDivide:       operator{OpBinaryDivide, 60, opLeftAssoc, false},
	OpBinaryFloorDiv:     operator{OpBinaryFloorDiv, 60, opLeftAssoc, false},
	OpBinaryModulo:       operator{OpBinaryModulo, 60, opLeftAssoc, false},
	OpBinaryIs:           operator{OpBinaryIs, 100, opLeftAssoc, false},
	OpBinaryIsNot:        operator{OpBinaryIsNot, 100, opLeftAssoc, false},
	OpBinaryPower:        operator{OpBinaryPower, 200, opRightAssoc, false},
}
