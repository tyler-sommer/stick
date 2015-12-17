package parse

import "fmt"

type Tree struct {
	root   *ModuleNode
	input  string
	lex    *lexer
	unread []token
	read   []token
}

func (t *Tree) Root() Node {
	return t.root
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
		if next.tokenType != tokenWhitespace {
			return next
		}
	}
}

func (t *Tree) expect(typ tokenType) token {
	tok := t.nextNonSpace()
	if tok.tokenType != typ {
		panic(fmt.Sprintf("expected %s got %s", typ, tok.tokenType))
	}

	return tok
}

func Parse(input string) (t *Tree) {
	lex := newLexer(input)

	go lex.tokenize()

	t = &Tree{newModuleNode(), input, lex, make([]token, 0), make([]token, 0)}

	for {
		n := t.parse()
		if n == nil {
			return
		}
		t.root.append(n)
	}

	return
}

func (t *Tree) parse() Node {
	tok := t.nextNonSpace()
	switch {
	case tok.tokenType == tokenText:
		return newTextNode(tok.value, pos(tok.pos))

	case tok.tokenType == tokenPrintOpen:
		name := t.parseExpr()
		t.expect(tokenPrintClose)
		return newPrintNode(name, pos(tok.pos))

	case tok.tokenType == tokenTagOpen:
		return t.parseTag()

	case tok.tokenType == tokenEof:
		return nil

	default:
		panic(fmt.Sprintf("parse error near %s", tok.value))
	}

	return nil
}

func (t *Tree) parseTag() expr {
	name := t.expect(tokenName)
	switch name.value {
	case "block":
		blockName := t.parseExpr()
		t.expect(tokenTagClose)
		body := t.parseUntilEndTag("block")
		return newBlockNode(blockName, body, pos(name.pos))
	case "if":
		cond := t.parseExpr()
		t.expect(tokenTagClose)
		body, els := t.parseEndifOrElse()
		return newIfNode(cond, body, els, pos(name.pos))
	}

	return nil
}

func (t *Tree) parseEndifOrElse() (body *ModuleNode, els *ModuleNode) {
	body = newModuleNode()
	for {
		switch tok := t.peek(); tok.tokenType {
		case tokenTagOpen:
			t.next()
			tok := t.expect(tokenName)
			if tok.value == "else" {
				t.expect(tokenTagClose)
				els = t.parseUntilEndTag("if")
				return

			} else if tok.value == "endif" {
				t.expect(tokenTagClose)
				return
			}
			t.backup()
			t.backup()
			t.backup()
			body.append(t.parse())

		default:
			body.append(t.parse())
		}
	}
}

func (t *Tree) parseUntilEndTag(name string) (n *ModuleNode) {
	n = newModuleNode()
	for {
		switch tok := t.peek(); tok.tokenType {
		case tokenTagOpen:
			t.next()
			tok := t.expect(tokenName)
			if tok.value == "end"+name {
				t.expect(tokenTagClose)
				return
			}
			t.backup()
			t.backup()
			t.backup()
			n.append(t.parse())

		default:
			n.append(t.parse())
		}
	}
}

func (t *Tree) parseExpr() expr {
	for {
		tok := t.nextNonSpace()
		switch typ := tok.tokenType; {
		case typ == tokenName:
			return newNameExpr(tok.value)

		case typ == tokenStringOpen:
			tok := t.expect(tokenText)
			t.expect(tokenStringClose)
			return newStringExpr(tok.value)

		default:
			panic("unknown expression")
		}
	}
}
