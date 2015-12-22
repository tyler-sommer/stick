// Package parse handles transforming Stick source code
// into AST for further processing.
package parse

import (
	"fmt"
)

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

func (t *Tree) expect(typ tokenType) (token, error) {
	tok := t.nextNonSpace()
	if tok.tokenType != typ {
		return tok, fmt.Errorf("expected %s got %s", typ, tok.tokenType)
	}

	return tok, nil
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
	tok := t.nextNonSpace()
	switch {
	case tok.tokenType == tokenText:
		n = newTextNode(tok.value, pos(tok.pos))

	case tok.tokenType == tokenPrintOpen:
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
		n = newPrintNode(name, pos(tok.pos))

	case tok.tokenType == tokenTagOpen:
		n, e = t.parseTag()

	case tok.tokenType == tokenEof:
		return

	default:
		e = fmt.Errorf("parse error near %s", tok.value)
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
		body, err := t.parseUntilEndTag("block")
		if err != nil {
			e = err
			return
		}
		t.popBlockStack(blockName.value)
		nod := newBlockNode(blockName.value, body, pos(name.pos))
		t.setBlock(blockName.value, nod)
		return nod, nil
	case "if":
		cond, err := t.parseExpr()
		if err != nil {
			e = err
			return
		}
		t.expect(tokenTagClose)
		body, els, err := t.parseEndifOrElse()
		if err != nil {
			e = err
			return
		}
		n = newIfNode(cond, body, els, pos(name.pos))
	}
	return
}

func (t *Tree) parseEndifOrElse() (body *ModuleNode, els *ModuleNode, e error) {
	body = newModuleNode()
	for {
		switch tok := t.peek(); tok.tokenType {
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
				els, err = t.parseUntilEndTag("if")
				if err != nil {
					e = err
					return
				}
				return

			} else if tok.value == "endif" {
				_, e = t.expect(tokenTagClose)
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

func (t *Tree) parseUntilEndTag(name string) (n *ModuleNode, e error) {
	n = newModuleNode()
	for {
		switch tok := t.peek(); tok.tokenType {
		case tokenEof:
			e = fmt.Errorf("Unclosed tag %s", name)
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
		exp = newNameExpr(tok.value)

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
		exp = newStringExpr(tok.value)

	default:
		return nil, fmt.Errorf("unknown expression: %s", tok)
	}
	return
}
