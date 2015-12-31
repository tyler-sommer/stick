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

// parseExpr parses an expression.
func (t *Tree) parseExpr() (Expr, error) {
	expr, err := t.parseInnerExpr()
	if err != nil {
		return nil, err
	}

	return t.parseOuterExpr(expr)
}
