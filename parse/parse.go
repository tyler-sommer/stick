// Package parse handles transforming Stick source code
// into AST for further processing.
package parse

// Tree represents the state of a parser.
type Tree struct {
	root      *ModuleNode
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

func (t *Tree) peek() (tok token) {
	tok = t.next()
	t.backup()

	return
}

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

func (t *Tree) next() (tok token) {
	if len(t.unread) > 0 {
		tok, t.unread = t.unread[len(t.unread)-1], t.unread[:len(t.unread)-1]
	} else {
		tok = t.lex.nextToken()
	}

	t.read = append(t.read, tok)

	return
}

func (t *Tree) nextNonSpace() token {
	var next token
	for {
		next = t.next()
		if next.tokenType != tokenWhitespace || next.tokenType == tokenEof {
			return next
		}
	}
}

func (t *Tree) expect(typs ...tokenType) (token, error) {
	tok := t.nextNonSpace()
	for _, typ := range typs {
		if tok.tokenType == typ {
			return tok, nil
		}
	}

	return tok, newUnexpectedTokenError(tok, typs...)
}

// Parse parses the given input.
func Parse(input string) (t *Tree, e error) {
	lex := newLexer(input)

	go lex.tokenize()

	t = &Tree{newModuleNode(), make([]string, 0), make(map[string]*BlockNode), input, lex, make([]token, 0), make([]token, 0)}

	for {
		n, err := t.parse()
		if err != nil || n == nil {
			e = err
			return
		}
		t.root.append(n)
	}

	return
}

func (t *Tree) parse() (n Node, e error) {

	switch tok := t.nextNonSpace(); tok.tokenType {
	case tokenText:
		n = newTextNode(tok.value, tok.Pos())

	case tokenPrintOpen:
		name, err := t.parseExpr()
		if err != nil {
			e = err
			return
		}
		_, err = t.expect(tokenPrintClose)
		if err != nil {
			e = err
			return
		}
		n = newPrintNode(name, tok.Pos())

	case tokenTagOpen:
		n, e = t.parseTag()

	case tokenEof:
		return

	default:
		e = newParseError(tok)
	}

	return
}

func (t *Tree) parseTag() (n Node, e error) {
	name, err := t.expect(tokenName)
	if err != nil {
		e = err
		return
	}
	switch name.value {
	case "block":
		blockName, err := t.expect(tokenName)
		if err != nil {
			e = err
			return
		}
		t.expect(tokenTagClose)
		t.pushBlockStack(blockName.value)
		body, err := t.parseUntilEndTag("block", name.Pos())
		if err != nil {
			e = err
			return
		}
		t.popBlockStack(blockName.value)
		nod := newBlockNode(blockName.value, body, name.Pos())
		t.setBlock(blockName.value, nod)
		return nod, nil
	case "if":
		cond, err := t.parseExpr()
		if err != nil {
			e = err
			return
		}
		t.expect(tokenTagClose)
		body, els, err := t.parseEndifOrElse(name.Pos())
		if err != nil {
			e = err
			return
		}
		n = newIfNode(cond, body, els, name.Pos())
	}
	return
}

func (t *Tree) parseEndifOrElse(start pos) (body *ModuleNode, els *ModuleNode, e error) {
	body = newModuleNode()
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
				_, err := t.expect(tokenTagClose)
				if err != nil {
					e = err
					return
				}
				els, err = t.parseUntilEndTag("if", start.Pos())
				if err != nil {
					e = err
					return
				}
				return

			} else if tok.value == "endif" {
				_, e = t.expect(tokenTagClose)
				return
			} else {
				e = newUnclosedTagError("if", start)
				return
			}
			t.backup3()
			n, err := t.parse()
			if err != nil {
				e = err
				return
			}
			body.append(n)

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

func (t *Tree) parseUntilEndTag(name string, start pos) (n *ModuleNode, e error) {
	n = newModuleNode()
	for {
		switch tok := t.peek(); tok.tokenType {
		case tokenEof:
			e = newUnclosedTagError(name, start)
			return

		case tokenTagOpen:
			t.next()
			tok, err := t.expect(tokenName)
			if err != nil {
				e = err
				return
			}
			if tok.value == "end"+name {
				_, err = t.expect(tokenTagClose)
				if err != nil {
					e = err
				}
				return
			}
			t.backup3()
			o, err := t.parse()
			if err != nil {
				e = err
				return
			}
			n.append(o)

		default:
			o, err := t.parse()
			if err != nil {
				e = err
				return
			}
			n.append(o)
		}
	}
}

func (t *Tree) parseExpr() (exp expr, e error) {
	tok := t.nextNonSpace()
	switch tok.tokenType {
	case tokenEof:
		return

	case tokenName:
		n := newNameExpr(tok.value, tok.Pos())
		tok = t.nextNonSpace()
		if tok.tokenType == tokenParensOpen {
			f, err := t.parseFunc(n)
			if err != nil {
				e = err
				return
			}
			exp = f
		} else {
			t.backup()
			exp = n
		}

	case tokenStringOpen:
		tok, err := t.expect(tokenText)
		if err != nil {
			e = err
			return
		}
		_, err = t.expect(tokenStringClose)
		if err != nil {
			e = err
			return
		}
		exp = newStringExpr(tok.value, tok.Pos())

	default:
		return nil, newParseError(tok)
	}
	return
}

func (t *Tree) parseFunc(name *NameExpr) (exp expr, e error) {
	var args []expr
	for {
		switch tok := t.peek(); tok.tokenType {
		case tokenEof:
			e = newUnexpectedEofError(tok)
			return

		default:
			argexp, err := t.parseExpr()
			if err != nil {
				e = err
				return
			}

			args = append(args, argexp)
		}

		switch tok := t.nextNonSpace(); tok.tokenType {
		case tokenEof:
			e = newUnexpectedEofError(tok)
			return

		case tokenPunctuation:
			if tok.value != "," {
				e = newUnexpectedPunctuationError(tok, ",")
				return
			}

		case tokenParensClose:
			exp = newFuncExpr(name, args, name.Pos())
			return

		default:
			e = newUnexpectedTokenError(tok, tokenPunctuation, tokenParensClose)
			return
		}
	}
}