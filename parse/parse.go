package parse

type tree struct {
	root *listNode
	input string
	lex *lexer
}

func Parse(input string) (t *tree) {
	lex := newLexer(input)

	go lex.tokenize()

	t = &tree{newListNode(), input, lex}

	for {
		tok := t.lex.nextToken()
		switch {
		case tok.tokenType == tokenText:
			t.root.append(newTextNode([]byte(tok.value), pos(tok.pos)))

		case tok.tokenType == tokenEof:
			return

		default:

		}
	}

	return
}
