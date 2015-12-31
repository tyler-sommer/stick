package parse

func (t *Tree) parseOuterExpr(expr Expr) (Expr, error) {
	switch nt := t.nextNonSpace(); nt.tokenType {
	case tokenParensOpen:
		switch name := expr.(type) {
		case *NameExpr:
			return t.parseFunc(name)
		default:
			return nil, newParseError(nt)
		}

	case tokenArrayOpen:
		fallthrough
	case tokenPunctuation:
		switch nt.value {
		case ".", "[":
			attr, err := t.parseInnerExpr()
			if err != nil {
				return nil, err
			}

			if nt.value == "[" {
				switch attr.(type) {
				case *NameExpr, *StringExpr, *NumberExpr:
				// valid
				default:
					return nil, newParseError(nt)
				}

				_, err := t.expect(tokenArrayClose)
				if err != nil {
					return nil, err
				}
			} else {
				switch attr.(type) {
				case *NameExpr:
				// valid
				default:
					return nil, newParseError(nt)
				}
			}

			getattr := newGetAttrExpr(expr, attr, nt.Pos())

			ntt := t.peek()
			if (ntt.tokenType == tokenPunctuation && ntt.value == ".") || ntt.tokenType == tokenArrayOpen {
				return t.parseOuterExpr(getattr)
			}

			return getattr, nil

		case "|":
			nx, err := t.parseExpr()
			if err != nil {
				return nil, err
			}
			switch n := nx.(type) {
			case *NameExpr:
				return newFuncExpr(n, []Expr{expr}, nt.Pos()), nil

			case *FuncExpr:
				n.args = append([]Expr{expr}, n.args...)
				return n, nil

			default:
				return nil, newParseError(nt)
			}

		default:
			t.backup()
			return expr, nil
		}

	case tokenOperator:
		op, ok := binaryOperators[nt.value]
		if !ok {
			return nil, newParseError(nt)
		}

		right, err := t.parseInnerExpr()
		if err != nil {
			return nil, err
		}

		ntt := t.nextNonSpace()
		if ntt.tokenType == tokenOperator {
			nxop, ok := binaryOperators[ntt.value]
			if !ok {
				return nil, newParseError(ntt)
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

func (t *Tree) parseInnerExpr() (Expr, error) {
	switch tok := t.nextNonSpace(); tok.tokenType {
	case tokenEof:
		return nil, newUnexpectedEofError(tok)

	case tokenOperator:
		op, ok := unaryOperators[tok.value]
		if !ok {
			return nil, newParseError(tok)
		}
		expr, err := t.parseInnerExpr()
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
		name := newNameExpr(tok.value, tok.Pos())
		nt := t.nextNonSpace()
		if nt.tokenType == tokenParensOpen {
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
		return nil, newParseError(tok)
	}
}

// parseFunc parses a function call expression from the first argument expression until the closing parenthesis.
func (t *Tree) parseFunc(name *NameExpr) (Expr, error) {
	var args []Expr
	for {
		switch tok := t.peek(); tok.tokenType {
		case tokenEof:
			return nil, newUnexpectedEofError(tok)

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
		case tokenEof:
			return nil, newUnexpectedEofError(tok)

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
