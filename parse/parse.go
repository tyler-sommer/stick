// Package parse handles transforming Stick source code
// into AST for further processing.
package parse

import (
	"bytes"
	"io"
)

// A NodeVisitor can be used to modify node contents and structure.
type NodeVisitor interface {
	Enter(Node) // Enter is called before the node is traversed.
	Leave(Node) // Exit is called before leaving the given Node.
}

// Tree represents the state of a parser.
type Tree struct {
	lex *lexer

	root   *ModuleNode
	blocks []map[string]*BlockNode // Contains each block available to this template.
	macros map[string]*MacroNode   // All macros defined on this template.

	unread []token // Any tokens received by the lexer but not yet read.
	read   []token // Tokens that have already been read.

	Name string // A name identifying this tree; the template name.

	Visitors []NodeVisitor
}

// NewTree creates a new parser Tree, ready for use.
func NewTree(input io.Reader) *Tree {
	return NewNamedTree("", input)
}

// NewNamedTree is an alternative constructor which creates a Tree with a name
func NewNamedTree(name string, input io.Reader) *Tree {
	return &Tree{
		lex: newLexer(input),

		root:   NewModuleNode(name),
		blocks: []map[string]*BlockNode{make(map[string]*BlockNode)},
		macros: make(map[string]*MacroNode),

		unread: make([]token, 0),
		read:   make([]token, 0),

		Name:     name,
		Visitors: make([]NodeVisitor, 0),
	}
}

// Root returns the root module node.
func (t *Tree) Root() *ModuleNode {
	return t.root
}

// Blocks returns a map of blocks in this tree.
func (t *Tree) Blocks() map[string]*BlockNode {
	return t.blocks[len(t.blocks)-1]
}

// Macros returns a map of macros defined in this tree.
func (t *Tree) Macros() map[string]*MacroNode {
	return t.macros
}

func (t *Tree) popBlockStack() map[string]*BlockNode {
	blocks := t.Blocks()
	t.blocks = t.blocks[0 : len(t.blocks)-1]
	return blocks
}

func (t *Tree) pushBlockStack() {
	t.blocks = append(t.blocks, make(map[string]*BlockNode))
}

func (t *Tree) setBlock(name string, body *BlockNode) {
	t.blocks[len(t.blocks)-1][name] = body
}

func (t *Tree) enrichError(err error) error {
	if err, ok := err.(ParsingError); ok {
		err.setTree(t)
	}
	return err
}

// peek returns the next unread token without advancing the internal cursor.
func (t *Tree) peek() token {
	tok := t.next()
	t.backup()

	return tok
}

// peek returns the next unread, non-space token without advancing the internal cursor.
func (t *Tree) peekNonSpace() token {
	var next token
	for {
		next = t.next()
		if next.tokenType != tokenWhitespace {
			t.backup()
			return next
		}
	}
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
func (t *Tree) next() token {
	var tok token
	if len(t.unread) > 0 {
		tok, t.unread = t.unread[len(t.unread)-1], t.unread[:len(t.unread)-1]
	} else {
		tok = t.lex.nextToken()
	}

	t.read = append(t.read, tok)

	return tok
}

// nextNonSpace returns the next non-whitespace token.
func (t *Tree) nextNonSpace() token {
	var next token
	for {
		next = t.next()
		if next.tokenType != tokenWhitespace || next.tokenType == tokenEOF {
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

// Enter is called when the given Node is entered.
func (t *Tree) enter(n Node) {
	for _, v := range t.Visitors {
		v.Enter(n)
	}
}

// Leave is called just before the state exits the given Node.
func (t *Tree) leave(n Node) {
	for _, v := range t.Visitors {
		v.Leave(n)
	}
}

func (t *Tree) traverse(n Node) {
	if n == nil {
		return
	}
	t.enter(n)
	for _, c := range n.All() {
		t.traverse(c)
	}
	t.leave(n)
}

// Parse parses the given input.
func Parse(input string) (*Tree, error) {
	t := NewTree(bytes.NewReader([]byte(input)))
	return t, t.Parse()
}

// Parse begins parsing, returning an error, if any.
func (t *Tree) Parse() error {
	go t.lex.tokenize()
	for {
		n, err := t.parse()
		if err != nil {
			return t.enrichError(err)
		}
		if n == nil {
			break
		}
		t.root.Append(n)
	}
	t.traverse(t.root)
	return nil
}

// parse parses generic input, such as text markup, print or tag statement opening tokens.
// parse is intended to pick up at the beginning of input, such as the start of a tag's body
// or the more obvious start of a document.
func (t *Tree) parse() (Node, error) {
	tok := t.nextNonSpace()
	switch tok.tokenType {
	case tokenText:
		return NewTextNode(tok.value, tok.Pos), nil

	case tokenPrintOpen:
		name, err := t.parseExpr()
		if err != nil {
			return nil, err
		}
		_, err = t.expect(tokenPrintClose)
		if err != nil {
			return nil, err
		}
		return NewPrintNode(name, tok.Pos), nil

	case tokenTagOpen:
		return t.parseTag()

	case tokenCommentOpen:
		tok, err := t.expect(tokenText)
		if err != nil {
			return nil, err
		}
		_, err = t.expect(tokenCommentClose)
		if err != nil {
			return nil, err
		}
		return NewCommentNode(tok.value, tok.Pos), nil

	case tokenEOF:
		// expected end of input
		return nil, nil
	}
	return nil, newUnexpectedTokenError(tok)
}
