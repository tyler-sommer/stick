package parse

import "fmt"

type tree struct {
	root   *moduleNode
	input  string
	lex    *lexer
	unread []token
	read   []token
}

func (t *tree) peek() (tok token) {
	tok = t.next()
	t.backup()

	return
}

func (t *tree) backup() {
	var tok token
	tok, t.read = t.read[len(t.read)-1], t.read[:len(t.read)-1]
	t.unread = append(t.unread, tok)
}

func (t *tree) next() (tok token) {
	if len(t.unread) > 0 {
		tok, t.unread = t.unread[len(t.unread)-1], t.unread[:len(t.unread)-1]
	} else {
		tok = t.lex.nextToken()
	}

	t.read = append(t.read, tok)

	return
}

func (t *tree) nextNonSpace() token {
	var next token
	for {
		next = t.next()
		if next.tokenType != tokenWhitespace {
			return next
		}
	}
}

func (t *tree) expect(typ tokenType) token {
	tok := t.nextNonSpace()
	if tok.tokenType != typ {
		panic(fmt.Sprintf("expected %s got %s", typ, tok.tokenType))
	}

	return tok
}

func Parse(input string) (t *tree) {
	lex := newLexer(input)

	go lex.tokenize()

	t = &tree{newModuleNode(), input, lex, make([]token, 0), make([]token, 0)}

	for {
		n := t.parse()
		if n == nil {
			return
		}
		t.root.append(n)
	}

	return
}

func (t *tree) parse() node {
	tok := t.nextNonSpace()
	switch {
	case tok.tokenType == tokenText:
		return newTextNode(tok.value, pos(tok.pos))

	case tok.tokenType == tokenPrintOpen:
		name := t.parseExpr()
		t.expect(tokenPrintClose)
		return newPrintNode(name, pos(tok.pos))

	case tok.tokenType == tokenTagOpen:
		name := t.expect(tokenName)
		attr := t.parseExpr()
		t.expect(tokenTagClose)
		body := t.parseUntilEndTag(name.value)
		return newTagNode(name.value, body, map[string]expr{"name": attr}, pos(tok.pos))

	case tok.tokenType == tokenEof:
		return nil

	default:
		panic(fmt.Sprintf("parse error near %s", tok.value))
	}

	return nil
}

func (t *tree) parseUntilEndTag(name string) (n *moduleNode) {
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

func (t *tree) parseExpr() expr {
	for {
		tok := t.nextNonSpace()
		switch typ := tok.tokenType; {
		case typ == tokenName:
			return newNameExpr(tok.value)

		default:
			panic("unknown expression")
		}
	}
}
