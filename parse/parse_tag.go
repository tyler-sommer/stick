package parse

import "errors"

// A tagParser can parse the body of a tag, returning the resulting Node or an error.
type tagParser func(t *Tree, start pos) (Node, error)

// parseTag parses the opening of a tag "{%", then delegates to a more specific parser function
// based on the tag's name.
func (t *Tree) parseTag() (Node, error) {
	name, err := t.expect(tokenName)
	if err != nil {
		return nil, err
	}
	switch name.value {
	case "extends":
		return parseExtends(t, name.Pos())
	case "block":
		return parseBlock(t, name.Pos())
	case "if", "elseif":
		return parseIf(t, name.Pos())
	case "for":
		return parseFor(t, name.Pos())
	case "include":
		return parseInclude(t, name.Pos())
	default:
		return nil, newParseError(name)
	}
}

// parseUntilEndTag parses until it reaches the specified tag's "end", returning a specific error otherwise.
func (t *Tree) parseUntilEndTag(name string, start pos) (*BodyNode, error) {
	tok := t.peek()
	if tok.tokenType == tokenEof {
		return nil, newUnclosedTagError(name, start)
	}

	n, err := t.parseUntilTag(start, "end"+name)
	if err != nil {
		return nil, err
	}
	_, err = t.expect(tokenTagClose)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func contains(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

// parseUntilTag parses until it reaches the specified tag node, returning a parse error otherwise.
func (t *Tree) parseUntilTag(start pos, names ...string) (*BodyNode, error) {
	n := newBodyNode(start)
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
			if contains(names, tok.value) {
				return n, nil
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

// parseExtends parses an extends tag.
//
//   {% extends <expr> %}
func parseExtends(t *Tree, start pos) (Node, error) {
	if t.Root().parent != nil {
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
	n := newExtendsNode(tplRef, start)
	t.Root().parent = n
	return n, nil
}

// parseBlock parses a block and any body it may contain.
// TODO: {% endblock <name> %} support
//
//   {% block <name> %}
//   {% endblock %}
func parseBlock(t *Tree, start pos) (Node, error) {
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
//
//   {% if <expr> %}
//   {% elseif <expr> %}
func parseIf(t *Tree, start pos) (Node, error) {
	cond, err := t.parseExpr()
	if err != nil {
		return nil, err
	}
	_, err = t.expect(tokenTagClose)
	if err != nil {
		return nil, err
	}
	body, els, err := parseIfBody(t, start)
	if err != nil {
		return nil, err
	}
	return newIfNode(cond, body, els, start), nil
}

// parseIfBody parses the body of an if statement.
//
//   {% else %}
//   {% endif %}
func parseIfBody(t *Tree, start pos) (body *BodyNode, els *BodyNode, e error) {
	body = newBodyNode(start)
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
			switch tok.value {
			case "else":
				_, err := t.expect(tokenTagClose)
				if err != nil {
					e = err
					return
				}
				els, err = t.parseUntilEndTag("if", start)
				if err != nil {
					e = err
					return
				}
			case "elseif":
				t.backup()
				in, err := t.parseTag()
				if err != nil {
					return nil, nil, err
				}
				els = newBodyNode(start, in)
			case "endif":
				_, e = t.expect(tokenTagClose)
				return
			default:
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

// parseFor parses a for loop construct.
// TODO: This needs proper error reporting.
//
//   {% for <name, [name]> in <expr> %}
//   {% for <name, [name]> in <expr> if <expr> %}
//   {% else %}
//   {% endfor %}
func parseFor(t *Tree, start pos) (*ForNode, error) {
	var kName, vName *NameExpr
	name, err := t.parseInnerExpr()
	if err != nil {
		return nil, err
	}
	if _, ok := name.(*NameExpr); !ok {
		return nil, errors.New("parse error: a parse error occured, expected name")
	}
	nxt := t.peekNonSpace()
	if nxt.tokenType == tokenPunctuation && nxt.value == "," {
		t.next()
		kName = name.(*NameExpr)
		name, err = t.parseInnerExpr()
		if err != nil {
			return nil, err
		}
		if _, ok := name.(*NameExpr); !ok {
			return nil, errors.New("parse error: a parse error occured, expected name")
		}
		vName = name.(*NameExpr)
	} else {
		vName = name.(*NameExpr)
	}
	tok := t.nextNonSpace()
	if tok.tokenType != tokenName && tok.value != "in" {
		return nil, newParseError(tok)
	}
	expr, err := t.parseExpr()
	if err != nil {
		return nil, err
	}
	tok, err = t.expect(tokenTagClose, tokenName)
	if err != nil {
		return nil, err
	}
	var ifCond Expr
	if tok.tokenType == tokenName {
		if tok.value != "if" {
			return nil, errors.New("parse error: a parse error occured")
		}
		ifCond, err = t.parseExpr()
		if err != nil {
			return nil, err
		}
		tok, err = t.expect(tokenTagClose)
		if err != nil {
			return nil, err
		}
	}
	var body Node
	body, err = t.parseUntilTag(tok.Pos(), "endfor", "else")
	if err != nil {
		return nil, err
	}
	if ifCond != nil {
		body = newIfNode(ifCond, body, nil, tok.Pos())
	}
	t.backup()
	tok = t.next()
	var elseBody Node
	if tok.value == "else" {
		_, err = t.expect(tokenTagClose)
		if err != nil {
			return nil, err
		}
		elseBody, err = t.parseUntilTag(tok.Pos(), "endfor")
		if err != nil {
			return nil, err
		}
	}
	_, err = t.expect(tokenTagClose)
	if err != nil {
		return nil, err
	}
	return newForNode(kName, vName, expr, body, elseBody, start), nil
}

// parseInclude parses an include statement.
// TODO: Implement "ignore missing" support
//
//   {% include <expr> %}
//   {% include <expr> with <expr> %}
//   {% include <expr> with <expr> only %}
//   {% include <expr> only %}
func parseInclude(t *Tree, start pos) (Node, error) {
	expr, err := t.parseExpr()
	if err != nil {
		return nil, err
	}
	var with Expr
	var only bool = false
	switch tok := t.peekNonSpace(); tok.tokenType {
	case tokenEof:
		return nil, newUnexpectedEofError(tok)
	case tokenName:
		if tok.value == "only" { // {% include <expr> only %}
			t.next()
			_, err = t.expect(tokenTagClose)
			if err != nil {
				return nil, err
			}
			only = true
			return newIncludeNode(expr, with, only, start), nil
		} else if tok.value != "with" {
			return nil, newParseError(tok)
		}
		t.next()
		with, err = t.parseExpr()
		if err != nil {
			return nil, err
		}
	case tokenTagClose:
		// no op
	default:
		return nil, newParseError(tok)
	}
	switch tok := t.nextNonSpace(); tok.tokenType {
	case tokenEof:
		return nil, newUnexpectedEofError(tok)
	case tokenName:
		if tok.value != "only" {
			return nil, newParseError(tok)
		}
		_, err = t.expect(tokenTagClose)
		if err != nil {
			return nil, err
		}
		only = true
	case tokenTagClose:
		// no op
	default:
		return nil, newParseError(tok)
	}
	return newIncludeNode(expr, with, only, start), nil
}
