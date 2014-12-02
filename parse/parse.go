package parse

type tree struct {
	root *moduleNode
	input string
	lex *lexer
}

func (t *tree) next() token {
	return t.lex.nextToken()
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

func Parse(input string) (t *tree) {
	lex := newLexer(input)

	go lex.tokenize()

	t = &tree{newModuleNode(), input, lex}

	for {
		tok := t.nextNonSpace()
		switch {
		case tok.tokenType == tokenText:
			t.root.append(newTextNode([]byte(tok.value), pos(tok.pos)))

		case tok.tokenType == tokenPrintOpen:
			name := t.nextNonSpace()
			t.nextNonSpace() // Consume the closing bracket
			t.root.append(newPrintNode(newNameExpr(name.value), pos(name.pos)))

		case tok.tokenType == tokenTagOpen:
			name := t.nextNonSpace()
			attr := t.nextNonSpace()
			t.nextNonSpace() // Consume the closing bracket
			body := t.nextNonSpace()
			t.nextNonSpace() // Next open bracket
			t.nextNonSpace() // Next tag name
			t.nextNonSpace() // Next closing bracket
			t.root.append(newTagNode(name.value, newTextNode([]byte(body.value), pos(body.pos)), map[string]expr{"name":newNameExpr(attr.value)}, pos(name.pos)))

		case tok.tokenType == tokenEof:
			return

		default:
			panic("parse error!")
		}
	}

	return
}
