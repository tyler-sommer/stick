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
			return nil, newUnexpectedTokenError(nt)
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
					return nil, newUnexpectedTokenError(nt)
				}

				_, err := t.expect(tokenArrayClose)
				if err != nil {
					return nil, err
				}
			} else {
				switch exp := attr.(type) {
				case *NameExpr:
					// valid, but we want to treat the name as a string
					attr = NewStringExpr(exp.Name, exp.Pos)
				case *NumberExpr:
					// Compatibility with Twig: {{ val.0 }}
					attr = NewStringExpr(exp.Value, exp.Pos)
				case *FuncExpr:
					// method call
					for _, v := range exp.Args {
						args = append(args, v)
					}
					attr = NewStringExpr(exp.Name, exp.Pos)
				default:
					return nil, newUnexpectedTokenError(nt)
				}
			}
			return t.parseOuterExpr(NewGetAttrExpr(expr, attr, args, nt.Pos))

		case "|": // Filter application
			nx, err := t.parseExpr()
			if err != nil {
				return nil, err
			}
			switch n := nx.(type) {
			case *NameExpr:
				return NewFilterExpr(n.Name, []Expr{expr}, nt.Pos), nil

			case *FuncExpr:
				n.Args = append([]Expr{expr}, n.Args...)
				return NewFilterExpr(n.Name, n.Args, n.Pos), nil

			default:
				return nil, newUnexpectedTokenError(nt)
			}

		case "?": // Ternary if
			tx, err := t.parseExpr()
			if err != nil {
				return nil, err
			}
			_, err = t.expectValue(tokenPunctuation, ":")
			if err != nil {
				return nil, err
			}
			fx, err := t.parseExpr()
			if err != nil {
				return nil, err
			}
			return NewTernaryIfExpr(expr, tx, fx, expr.Start()), nil

		default:
			t.backup()
			return expr, nil
		}

	case tokenOperator:
		op, ok := binaryOperators[nt.value]
		if !ok {
			return nil, newUnexpectedTokenError(nt)
		}

		var right Node
		var err error
		if op.op == OpBinaryIs || op.op == OpBinaryIsNot {
			right, err = t.parseRightTestOperand(nil)
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
				return nil, newUnexpectedTokenError(ntt)
			}
			if nxop.precedence < op.precedence || (nxop.precedence == op.precedence && op.leftAssoc()) {
				t.backup()
				return t.parseOuterExpr(NewBinaryExpr(expr, op.Operator(), right, expr.Start()))
			}
			t.backup()
			right, err = t.parseOuterExpr(right)
			if err != nil {
				return nil, err
			}
		} else {
			t.backup()
		}
		return NewBinaryExpr(expr, op.Operator(), right, expr.Start()), nil

	default:
		t.backup()
		return expr, nil
	}
}

// parseIsRightOperand handles "is" and "is not" tests, which can
// themselves be two words, such as "divisible by":
//	{% if 10 is divisible by(3) %}
func (t *Tree) parseRightTestOperand(prev *NameExpr) (*TestExpr, error) {
	right, err := t.parseInnerExpr()
	if err != nil {
		return nil, err
	}
	if prev == nil {
		if r, ok := right.(*NameExpr); ok {
			if nxt := t.peekNonSpace(); nxt.tokenType == tokenName {
				return t.parseRightTestOperand(r)
			}
		}
	}
	switch r := right.(type) {
	case *NameExpr:
		if prev != nil {
			r.Name = prev.Name + " " + r.Name
		}
		return NewTestExpr(r.Name, []Expr{}, r.Pos), nil

	case *FuncExpr:
		if prev != nil {
			r.Name = prev.Name + " " + r.Name
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
			return nil, newUnexpectedTokenError(tok)
		}
		expr, err := t.parseExpr()
		if err != nil {
			return nil, err
		}
		return NewUnaryExpr(op.Operator(), expr, tok.Pos), nil

	case tokenParensOpen:
		inner, err := t.parseExpr()
		if err != nil {
			return nil, err
		}
		_, err = t.expect(tokenParensClose)
		if err != nil {
			return nil, err
		}
		return NewGroupExpr(inner, tok.Pos), nil

	case tokenHashOpen:
		els := []*KeyValueExpr{}
		for {
			nxt := t.peek()
			if nxt.tokenType == tokenHashClose {
				t.next()
				break
			}
			keyExpr, err := t.parseExpr()
			if err != nil {
				return nil, err
			}
			_, err = t.expectValue(tokenPunctuation, delimHashKeyValue)
			if err != nil {
				return nil, err
			}
			valExpr, err := t.parseExpr()
			if err != nil {
				return nil, err
			}
			els = append(els, NewKeyValueExpr(keyExpr, valExpr, nxt.Pos))
			nxt = t.peek()
			if nxt.tokenType == tokenPunctuation {
				_, err := t.expectValue(tokenPunctuation, ",")
				if err != nil {
					return nil, err
				}
			}
		}
		return NewHashExpr(tok.Pos, els...), nil

	case tokenArrayOpen:
		els := []Expr{}
		for {
			nxt := t.peek()
			if nxt.tokenType == tokenArrayClose {
				t.next()
				break
			}
			expr, err := t.parseExpr()
			if err != nil {
				return nil, err
			}
			els = append(els, expr)
			nxt = t.peek()
			if nxt.tokenType == tokenPunctuation {
				_, err := t.expectValue(tokenPunctuation, ",")
				if err != nil {
					return nil, err
				}
			}
		}
		return NewArrayExpr(tok.Pos, els...), nil

	case tokenNumber:
		nxt := t.peek()
		val := tok.value
		if nxt.tokenType == tokenPunctuation && nxt.value == "." {
			val = val + "."
			t.next()
			nxt, err := t.expect(tokenNumber)
			if err != nil {
				return nil, err
			}
			val = val + nxt.value
		}
		return NewNumberExpr(val, tok.Pos), nil

	case tokenName:
		switch tok.value {
		case "null", "NULL", "none", "NONE":
			return NewNullExpr(tok.Pos), nil
		case "true", "TRUE":
			return NewBoolExpr(true, tok.Pos), nil
		case "false", "FALSE":
			return NewBoolExpr(false, tok.Pos), nil
		}
		name := NewNameExpr(tok.value, tok.Pos)
		nt := t.nextNonSpace()
		if nt.tokenType == tokenParensOpen {
			// TODO: This duplicates some code in parseOuterExpr, are both necessary?
			return t.parseFunc(name)
		}
		t.backup()
		return name, nil

	case tokenStringOpen:
		var exprs []Expr
		for {
			nxt, err := t.expect(tokenText, tokenInterpolateOpen, tokenStringClose)
			if err != nil {
				return nil, err
			}
			switch nxt.tokenType {
			case tokenText:
				exprs = append(exprs, NewStringExpr(nxt.value, nxt.Pos))
			case tokenInterpolateOpen:
				exp, err := t.parseExpr()
				if err != nil {
					return nil, err
				}
				_, err = t.expect(tokenInterpolateClose)
				if err != nil {
					return nil, err
				}
				exprs = append(exprs, exp)
			case tokenStringClose:
				ln := len(exprs)
				if ln > 1 {
					var res *BinaryExpr
					for i := 1; i < ln; i++ {
						if res == nil {
							res = NewBinaryExpr(exprs[i-1], OpBinaryConcat, exprs[i], exprs[i-1].Start())
							continue
						}
						res = NewBinaryExpr(res, OpBinaryConcat, exprs[i], res.Pos)
					}
					return res, nil
				}
				return exprs[0], nil
			}
		}

	default:
		return nil, newUnexpectedTokenError(tok)
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
			return NewFuncExpr(name.Name, args, name.Pos), nil

		default:
			return nil, newUnexpectedTokenError(tok, tokenPunctuation, tokenParensClose)
		}
	}
}
