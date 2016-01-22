package parse

import "fmt"

// parseExpr parses an expression.
func (t *Tree) parseExpr() (Expr, error) {
	expr, err := t.parseInnerExpr()
	if err != nil {
		return nil, err
	}

	return t.parseOuterExpr(expr)
}

// parseOuterExpr attempts to parse an expression outside of an inner
// expression.
// An outer expression is defined as a modification to an inner expression.
// Examples include attribute accessing, filter application, or binary operations.
func (t *Tree) parseOuterExpr(expr Expr) (Expr, error) {
	switch nt := t.nextNonSpace(); nt.tokenType {
	case tokenParensOpen:
		switch name := expr.(type) {
		case *NameExpr:
			// TODO: This duplicates some code in parseInnerExpr, are both necessary?
			return t.parseFunc(name)
		default:
			return nil, newError(nt)
		}

	case tokenArrayOpen, tokenPunctuation:
		switch nt.value {
		case ".", "[": // Dot or array access
			var args = make([]Expr, 0)
			attr, err := t.parseInnerExpr()
			if err != nil {
				return nil, err
			}

			if nt.value == "[" {
				switch attr.(type) {
				case *NameExpr, *StringExpr, *NumberExpr, *GroupExpr:
					// valid
				default:
					return nil, newError(nt)
				}

				_, err := t.expect(tokenArrayClose)
				if err != nil {
					return nil, err
				}
			} else {
				switch exp := attr.(type) {
				case *NameExpr:
					// valid, but we want to treat the name as a string
					attr = newStringExpr(exp.Name(), exp.Pos())
				case *FuncExpr:
					// method call
					for _, v := range exp.Args() {
						args = append(args, v)
					}
					attr = newStringExpr(exp.Name(), exp.Pos())
				default:
					return nil, newError(nt)
				}
			}
			return t.parseOuterExpr(newGetAttrExpr(expr, attr, args, nt.Pos()))

		case "|": // Filter application
			nx, err := t.parseExpr()
			if err != nil {
				return nil, err
			}
			switch n := nx.(type) {
			case *NameExpr:
				return newFilterExpr(n, []Expr{expr}, nt.Pos()), nil

			case *FuncExpr:
				n.args = append([]Expr{expr}, n.args...)
				return newFilterExpr(n.name, n.args, n.pos), nil

			default:
				return nil, newError(nt)
			}

		default:
			t.backup()
			return expr, nil
		}

	case tokenOperator:
		op, ok := binaryOperators[nt.value]
		if !ok {
			return nil, newError(nt)
		}

		var right Node
		var err error
		if op.op == OpBinaryIs || op.op == OpBinaryIsNot {
			right, err = t.parseIsRightOperand(nil)
			if err != nil {
				return nil, err
			}
		} else {
			right, err = t.parseInnerExpr()
			if err != nil {
				return nil, err
			}
		}

		ntt := t.nextNonSpace()
		if ntt.tokenType == tokenOperator {
			nxop, ok := binaryOperators[ntt.value]
			if !ok {
				return nil, newError(ntt)
			}
			if nxop.precedence < op.precedence || (nxop.precedence == op.precedence && op.leftAssoc()) {
				t.backup()
				return t.parseOuterExpr(newBinaryExpr(expr, op.Operator(), right, expr.Pos()))
			}
			t.backup()
			right, err = t.parseOuterExpr(right)
			if err != nil {
				return nil, err
			}
		} else {
			t.backup()
		}
		return newBinaryExpr(expr, op.Operator(), right, expr.Pos()), nil

	default:
		t.backup()
		return expr, nil
	}
}

// parseIsRightOperand handles "is" and "is not" tests, which can
// themselves be two words, such as "divisible by":
//	{% if 10 is divisible by(3) %}
func (t *Tree) parseIsRightOperand(prev *NameExpr) (*TestExpr, error) {
	right, err := t.parseInnerExpr()
	if err != nil {
		return nil, err
	}
	if prev == nil {
		if r, ok := right.(*NameExpr); ok {
			if nxt := t.peekNonSpace(); nxt.tokenType == tokenName {
				return t.parseIsRightOperand(r)
			}
		}
	}
	switch r := right.(type) {
	case *NameExpr:
		if prev != nil {
			r.name = prev.name + " " + r.name
		}
		return newTestExpr(r, []Expr{}, r.Pos()), nil

	case *FuncExpr:
		if prev != nil {
			r.name.name = prev.name + " " + r.name.name
		}
		return &TestExpr{r}, nil
	default:
		return nil, fmt.Errorf(`Expected name or function, got "%v"`, right)
	}
}

// parseInnerExpr attempts to parse an inner expression.
// An inner expression is defined as a cohesive expression, such as a literal.
func (t *Tree) parseInnerExpr() (Expr, error) {
	switch tok := t.nextNonSpace(); tok.tokenType {
	case tokenEOF:
		return nil, newUnexpectedEOFError(tok)

	case tokenOperator:
		op, ok := unaryOperators[tok.value]
		if !ok {
			return nil, newError(tok)
		}
		expr, err := t.parseExpr()
		if err != nil {
			return nil, err
		}
		return newUnaryExpr(op.Operator(), expr, tok.Pos()), nil

	case tokenParensOpen:
		inner, err := t.parseExpr()
		if err != nil {
			return nil, err
		}
		_, err = t.expect(tokenParensClose)
		if err != nil {
			return nil, err
		}
		return newGroupExpr(inner, tok.Pos()), nil

	case tokenNumber:
		nxt := t.peek()
		val := tok.value
		if nxt.tokenType == tokenPunctuation || nxt.value == "." {
			val = val + "."
			t.next()
			nxt, err := t.expect(tokenNumber)
			if err != nil {
				return nil, err
			}
			val = val + nxt.value
		}
		return newNumberExpr(val, tok.Pos()), nil

	case tokenName:
		switch tok.value {
		case "null", "NULL", "none", "NONE":
			return newNullExpr(tok.Pos()), nil
		case "true", "TRUE":
			return newBoolExpr(true, tok.Pos()), nil
		case "false", "FALSE":
			return newBoolExpr(false, tok.Pos()), nil
		}
		name := newNameExpr(tok.value, tok.Pos())
		nt := t.nextNonSpace()
		if nt.tokenType == tokenParensOpen {
			// TODO: This duplicates some code in parseOuterExpr, are both necessary?
			return t.parseFunc(name)
		}
		t.backup()
		return name, nil

	case tokenStringOpen:
		txt, err := t.expect(tokenText)
		if err != nil {
			return nil, err
		}
		_, err = t.expectValue(tokenStringClose, tok.value)
		if err != nil {
			return nil, err
		}
		return newStringExpr(txt.value, txt.Pos()), nil

	default:
		return nil, newError(tok)
	}
}

// parseFunc parses a function call expression from the first argument expression until the closing parenthesis.
func (t *Tree) parseFunc(name *NameExpr) (Expr, error) {
	var args []Expr
	for {
		switch tok := t.peek(); tok.tokenType {
		case tokenEOF:
			return nil, newUnexpectedEOFError(tok)

		case tokenParensClose:
		// do nothing

		default:
			argexp, err := t.parseExpr()
			if err != nil {
				return nil, err
			}

			args = append(args, argexp)
		}

		switch tok := t.nextNonSpace(); tok.tokenType {
		case tokenEOF:
			return nil, newUnexpectedEOFError(tok)

		case tokenPunctuation:
			if tok.value != "," {
				return nil, newUnexpectedValueError(tok, ",")
			}

		case tokenParensClose:
			return newFuncExpr(name, args, name.Pos()), nil

		default:
			return nil, newUnexpectedTokenError(tok, tokenPunctuation, tokenParensClose)
		}
	}
}
