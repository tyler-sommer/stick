// Package parse handles transforming Stick source code
// into AST for further processing.
package parse

// Tree represents the state of a parser.
type Tree struct {
	root      *ModuleNode
	parent    *ExtendsNode
	blocks    []string
	blockRefs map[string]*BlockNode
	input     string
	lex       *lexer
	unread    []token
	read      []token
}

// Root returns the root module node.
func (t *Tree) Root() Node {
	return t.root
}

func (t *Tree) popBlockStack(name string) {
	t.blocks = t.blocks[0 : len(t.blocks)-1]
}

func (t *Tree) pushBlockStack(name string) {
	t.blocks = append(t.blocks, name)
}

func (t *Tree) setBlock(name string, body *BlockNode) {
	t.blockRefs[name] = body
}

// peek returns the next unread token without advancing the internal cursor.
func (t *Tree) peek() (tok token) {
	tok = t.next()
	t.backup()

	return
}

// peek returns the next unread, non-space token without advancing the internal cursor.
func (t *Tree) peekNonSpace() (tok token) {
	var next token
	for {
		next = t.next()
		if next.tokenType != tokenWhitespace || next.tokenType == tokenEof {
			t.backup()
			return next
		}
	}

	return
}

// backup pushes the last read token back onto the unread stack and reduces the internal cursor by one.
func (t *Tree) backup() {
	var tok token
	tok, t.read = t.read[len(t.read)-1], t.read[:len(t.read)-1]
	t.unread = append(t.unread, tok)
}

func (t *Tree) backup2() {
	t.backup()
	t.backup()
}

func (t *Tree) backup3() {
	t.backup()
	t.backup()
	t.backup()
}

// next returns the next unread token and advances the internal cursor by one.
func (t *Tree) next() (tok token) {
	if len(t.unread) > 0 {
		tok, t.unread = t.unread[len(t.unread)-1], t.unread[:len(t.unread)-1]
	} else {
		tok = t.lex.nextToken()
	}

	t.read = append(t.read, tok)

	return
}

// nextNonSpace returns the next non-whitespace token.
func (t *Tree) nextNonSpace() token {
	var next token
	for {
		next = t.next()
		if next.tokenType != tokenWhitespace || next.tokenType == tokenEof {
			return next
		}
	}
}

// expect returns the next non-space token. Additionally, if the token is not of one of the expected types,
// an UnexpectedTokenError is returned.
func (t *Tree) expect(typs ...tokenType) (token, error) {
	tok := t.nextNonSpace()
	for _, typ := range typs {
		if tok.tokenType == typ {
			return tok, nil
		}
	}

	return tok, newUnexpectedTokenError(tok, typs...)
}

// expectValue returns the next non-space token, with additional checks on the value of the token.
// If the token is not of the expected type, an UnexpectedTokenError is returned. If the token is not the
// expected value, an UnexpectedValueError is returned.
func (t *Tree) expectValue(typ tokenType, val string) (token, error) {
	tok, err := t.expect(typ)
	if err != nil {
		return tok, err
	}

	if tok.value != val {
		return tok, newUnexpectedValueError(tok, val)
	}

	return tok, nil
}

// Parse parses the given input.
func Parse(input string) (*Tree, error) {
	lex := newLexer(input)

	go lex.tokenize()

	t := &Tree{newModuleNode(), nil, make([]string, 0), make(map[string]*BlockNode), input, lex, make([]token, 0), make([]token, 0)}

	for {
		n, err := t.parse()
		if err != nil {
			return t, err
		}
		if n == nil {
			// expected end of input
			return t, nil
		}
		t.root.append(n)
	}
}

// parse parses generic input, such as text markup, print or tag statement opening tokens.
// parse is intended to pick up at the beginning of input, such as the start of a tag's body
// or the more obvious start of a document.
func (t *Tree) parse() (Node, error) {
	tok := t.nextNonSpace()
	switch tok.tokenType {
	case tokenText:
		return newTextNode(tok.value, tok.Pos()), nil

	case tokenPrintOpen:
		name, err := t.parseExpr()
		if err != nil {
			return nil, err
		}
		_, err = t.expect(tokenPrintClose)
		if err != nil {
			return nil, err
		}
		return newPrintNode(name, tok.Pos()), nil

	case tokenTagOpen:
		return t.parseTag()

	case tokenEof:
		// expected end of input
		return nil, nil
	}
	return nil, newParseError(tok)
}

// parseTag parses the opening of a tag "{%", then delegates to a more specific parser function
// based on the tag's name.
func (t *Tree) parseTag() (Node, error) {
	name, err := t.expect(tokenName)
	if err != nil {
		return nil, err
	}
	switch name.value {
	case "extends":
		return t.parseExtends(name.Pos())
	case "block":
		return t.parseBlock(name.Pos())
	case "if":
		return t.parseIf(name.Pos())
	default:
		return nil, newParseError(name)
	}
}

// parseExtends parses an extends tag.
func (t *Tree) parseExtends(start pos) (Node, error) {
	if t.parent != nil {
		return nil, newMultipleExtendsError(start)
	}
	tplRef, err := t.parseExpr()
	if err != nil {
		return nil, err
	}
	_, err = t.expect(tokenTagClose)
	if err != nil {
		return nil, err
	}
	t.parent = newExtendsNode(tplRef, start)
	return t.parent, nil
}

// parseBlock parses a block and any body it may contain.
func (t *Tree) parseBlock(start pos) (Node, error) {
	blockName, err := t.expect(tokenName)
	if err != nil {
		return nil, err
	}
	_, err = t.expect(tokenTagClose)
	if err != nil {
		return nil, err
	}
	t.pushBlockStack(blockName.value)
	body, err := t.parseUntilEndTag("block", start)
	if err != nil {
		return nil, err
	}
	t.popBlockStack(blockName.value)
	nod := newBlockNode(blockName.value, body, start)
	t.setBlock(blockName.value, nod)
	return nod, nil
}

// parseIf parses the opening tag and conditional expression in an if-statement.
func (t *Tree) parseIf(start pos) (Node, error) {
	cond, err := t.parseExpr()
	if err != nil {
		return nil, err
	}
	_, err = t.expect(tokenTagClose)
	if err != nil {
		return nil, err
	}
	body, els, err := t.parseIfBody(start)
	if err != nil {
		return nil, err
	}
	return newIfNode(cond, body, els, start), nil
}

// parseIfBody parses the body of an if statement.
func (t *Tree) parseIfBody(start pos) (body *ModuleNode, els *ModuleNode, e error) {
	body = newModuleNode()
	els = newModuleNode()
	for {
		switch tok := t.peek(); tok.tokenType {
		case tokenEof:
			e = newUnclosedTagError("if", start)
			return

		case tokenTagOpen:
			t.next()
			tok, err := t.expect(tokenName)
			if err != nil {
				e = err
				return
			}
			if tok.value == "else" {
				n, err := t.parseElse(tok.Pos())
				if err != nil {
					e = err
					return
				}
				els.nodes = n.nodes
			} else if tok.value == "endif" {
				_, e = t.expect(tokenTagClose)
				return
			} else {
				e = newUnclosedTagError("if", start)
				return
			}

			return

		default:
			n, err := t.parse()
			if err != nil {
				e = err
				return
			}
			body.append(n)
		}
	}
}

// parseElse parses an if statement's "else" body or "else if" statement.
func (t *Tree) parseElse(start pos) (*ModuleNode, error) {
	tok := t.nextNonSpace()
	switch tok.tokenType {
	case tokenTagClose:
		return t.parseUntilEndTag("if", start)

	case tokenName:
		if tok.value != "if" {
			return nil, newParseError(tok)
		}
		t.backup()
		in, err := t.parseTag()
		if err != nil {
			return nil, err
		}
		return newModuleNode(in), nil
	}
	return nil, newParseError(tok)
}

// parseUntilEndTag parses until it reaches the specified tag's "end", returning a specific error otherwise.
func (t *Tree) parseUntilEndTag(name string, start pos) (*ModuleNode, error) {
	tok := t.peek()
	if tok.tokenType == tokenEof {
		return nil, newUnclosedTagError(name, start)
	}

	return t.parseUntilTag("end"+name, start)
}

// parseUntilTag parses until it reaches the specified tag node, returning a parse error otherwise.
func (t *Tree) parseUntilTag(name string, start pos) (*ModuleNode, error) {
	n := newModuleNode()
	for {
		switch tok := t.peek(); tok.tokenType {
		case tokenEof:
			return n, newUnexpectedEofError(tok)

		case tokenTagOpen:
			t.next()
			tok, err := t.expect(tokenName)
			if err != nil {
				return n, err
			}
			if tok.value == name {
				_, err = t.expect(tokenTagClose)
				return n, err
			}
			t.backup3()
			o, err := t.parse()
			if err != nil {
				return n, err
			}
			n.append(o)

		default:
			o, err := t.parse()
			if err != nil {
				return n, err
			}
			n.append(o)
		}
	}
}

// parseExpr parses an expression.
func (t *Tree) parseExpr() (Expr, error) {
	expr, err := t.parseInnerExpr()
	if err != nil {
		return nil, err
	}

	return t.parseOuterExpr(expr)
}

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
