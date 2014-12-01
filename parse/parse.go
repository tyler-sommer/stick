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

		case tok.tokenType == tokenEof:
			return

		default:
			panic("parse error!")
		}
	}

	return
}
