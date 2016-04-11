package parse

import "errors"

// A tagParser can parse the body of a tag, returning the resulting Node or an error.
// TODO: This will be used to implement user-defined tags.
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
	case "embed":
		return parseEmbed(t, name.Pos())
	case "use":
		return parseUse(t, name.Pos())
	case "set":
		return parseSet(t, name.Pos())
	case "do":
		return parseDo(t, name.Pos())
	case "filter":
		return parseFilter(t, name.Pos())
	default:
		return nil, newError(name)
	}
}

// parseUntilEndTag parses until it reaches the specified tag's "end", returning a specific error otherwise.
func (t *Tree) parseUntilEndTag(name string, start pos) (*BodyNode, error) {
	tok := t.peek()
	if tok.tokenType == tokenEOF {
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
		case tokenEOF:
			return n, newUnexpectedEOFError(tok)

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
	body, err := t.parseUntilEndTag("block", start)
	if err != nil {
		return nil, err
	}
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
func parseIfBody(t *Tree, start pos) (body *BodyNode, els *BodyNode, err error) {
	body = newBodyNode(start)
	for {
		switch tok := t.peek(); tok.tokenType {
		case tokenEOF:
			return nil, nil, newUnclosedTagError("if", start)
		case tokenTagOpen:
			t.next()
			tok, err := t.expect(tokenName)
			if err != nil {
				return nil, nil, err
			}
			switch tok.value {
			case "else":
				_, err := t.expect(tokenTagClose)
				if err != nil {
					return nil, nil, err
				}
				els, err = t.parseUntilEndTag("if", start)
				if err != nil {
					return nil, nil, err
				}
			case "elseif":
				t.backup()
				in, err := t.parseTag()
				if err != nil {
					return nil, nil, err
				}
				els = newBodyNode(tok.Pos(), in)
			case "endif":
				_, err := t.expect(tokenTagClose)
				if err != nil {
					return nil, nil, err
				}
			default:
				return nil, nil, newUnclosedTagError("if", start)
			}
			if els == nil {
				els = newBodyNode(start)
			}
			return body, els, nil
		default:
			n, err := t.parse()
			if err != nil {
				return nil, nil, err
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
	var kn, vn string
	nam, err := t.parseInnerExpr()
	if err != nil {
		return nil, err
	}
	if nam, ok := nam.(*NameExpr); ok {
		vn = nam.Name()
	} else {
		return nil, errors.New("parse error: a parse error occured, expected name")
	}
	nxt := t.peekNonSpace()
	if nxt.tokenType == tokenPunctuation && nxt.value == "," {
		t.next()
		kn = vn
		nam, err = t.parseInnerExpr()
		if err != nil {
			return nil, err
		}
		if nam, ok := nam.(*NameExpr); ok {
			vn = nam.Name()
		} else {
			return nil, errors.New("parse error: a parse error occured, expected name")
		}
	}
	tok := t.nextNonSpace()
	if tok.tokenType != tokenName && tok.value != "in" {
		return nil, newError(tok)
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
	var elseBody Node = newBodyNode(tok.Pos())
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
	return newForNode(kn, vn, expr, body, elseBody, start), nil
}

// parseInclude parses an include statement.
func parseInclude(t *Tree, start pos) (Node, error) {
	expr, with, only, err := parseIncludeOrEmbed(t)
	if err != nil {
		return nil, err
	}
	return newIncludeNode(expr, with, only, start), nil
}

// parseEmbed parses an embed statement and body.
func parseEmbed(t *Tree, start pos) (Node, error) {
	expr, with, only, err := parseIncludeOrEmbed(t)
	if err != nil {
		return nil, err
	}
	t.pushBlockStack()
	for {
		tok := t.nextNonSpace()
		if tok.tokenType == tokenEOF {
			return nil, newUnclosedTagError("embed", start)
		} else if tok.tokenType == tokenTagOpen {
			tok, err := t.expect(tokenName)
			if err != nil {
				return nil, err
			}
			if tok.value == "endembed" {
				t.next()
				_, err := t.expect(tokenTagClose)
				if err != nil {
					return nil, err
				}
				break
			} else if tok.value == "block" {
				n, err := parseBlock(t, start)
				if err != nil {
					return nil, err
				}
				if _, ok := n.(*BlockNode); !ok {
					return nil, newError(tok)
				}
			} else {
				return nil, newUnexpectedValueError(tok, "endembed or block")
			}
		}
	}
	blockRefs := t.popBlockStack()
	return newEmbedNode(expr, with, only, blockRefs, start), nil
}

// parseIncludeOrEmbed parses an include or embed tag's parameters.
// TODO: Implement "ignore missing" support
//
//   {% include <expr> %}
//   {% include <expr> with <expr> %}
//   {% include <expr> with <expr> only %}
//   {% include <expr> only %}
func parseIncludeOrEmbed(t *Tree) (expr Expr, with Expr, only bool, err error) {
	expr, err = t.parseExpr()
	if err != nil {
		return
	}
	only = false
	switch tok := t.peekNonSpace(); tok.tokenType {
	case tokenEOF:
		err = newUnexpectedEOFError(tok)
		return
	case tokenName:
		if tok.value == "only" { // {% include <expr> only %}
			t.next()
			_, err = t.expect(tokenTagClose)
			if err != nil {
				return
			}
			only = true
			return expr, with, only, nil
		} else if tok.value != "with" {
			err = newError(tok)
			return
		}
		t.next()
		with, err = t.parseExpr()
		if err != nil {
			return
		}
	case tokenTagClose:
	// no op
	default:
		err = newError(tok)
		return
	}
	switch tok := t.nextNonSpace(); tok.tokenType {
	case tokenEOF:
		err = newUnexpectedEOFError(tok)
		return
	case tokenName:
		if tok.value != "only" {
			err = newError(tok)
			return
		}
		_, err = t.expect(tokenTagClose)
		if err != nil {
			return
		}
		only = true
	case tokenTagClose:
	// no op
	default:
		err = newError(tok)
		return
	}
	return
}

func parseUse(t *Tree, start pos) (Node, error) {
	tmpl, err := t.parseExpr()
	if err != nil {
		return nil, err
	}
	tok, err := t.expect(tokenName, tokenTagClose)
	if err != nil {
		return nil, err
	}
	aliases := make(map[string]string)
	if tok.tokenType == tokenName {
		if tok.value != "with" {
			return nil, newUnexpectedValueError(tok, "with")
		}
		for {
			orig, err := t.expect(tokenName)
			if err != nil {
				return nil, err
			}
			tok, err = t.expectValue(tokenName, "as")
			if err != nil {
				return nil, err
			}
			alias, err := t.expect(tokenName)
			if err != nil {
				return nil, err
			}
			aliases[orig.value] = alias.value
			tok, err = t.expect(tokenTagClose, tokenPunctuation)
			if err != nil {
				return nil, err
			}
			if tok.tokenType == tokenTagClose {
				break
			} else if tok.value != "," {
				return nil, newUnexpectedValueError(tok, ",")
			}
		}
	}
	return newUseNode(tmpl, aliases, start), nil
}

// parseSet parses a set statement.
//
//   {% set <var> = <expr> %}
func parseSet(t *Tree, start pos) (Node, error) {
	tok, err := t.expect(tokenName)
	if err != nil {
		return nil, err
	}
	_, err = t.expectValue(tokenPunctuation, "=")
	if err != nil {
		return nil, err
	}
	expr, err := t.parseExpr()
	if err != nil {
		return nil, err
	}
	_, err = t.expect(tokenTagClose)
	if err != nil {
		return nil, err
	}
	return newSetNode(tok.value, expr, start), nil
}

// parseDo parses a do statement.
//
//   {% do <expr> %}
func parseDo(t *Tree, start pos) (Node, error) {
	expr, err := t.parseExpr()
	if err != nil {
		return nil, err
	}
	_, err = t.expect(tokenTagClose)
	if err != nil {
		return nil, err
	}
	return newDoNode(expr, start), nil
}

// parseFilter parses a filter statement.
//
// 	{% filter <name> %}
//
// Multiple filters can be applied to a block:
//
// 	{% filter <name>|<name>|<name> %}
func parseFilter(t *Tree, start pos) (Node, error) {
	var filters []string
	for {
		tok, err := t.expect(tokenName)
		if err != nil {
			return nil, err
		}
		filters = append(filters, tok.value)
		tok = t.peekNonSpace()
		switch tok.tokenType {
		case tokenEOF:
			return nil, newUnexpectedEOFError(tok)
		case tokenPunctuation:
			if tok.value != "|" {
				return nil, newUnexpectedValueError(tok, "|")
			}
			t.nextNonSpace()
		case tokenTagClose:
			t.nextNonSpace()
			goto body
		}
	}
body:
	body, err := t.parseUntilEndTag("filter", start)
	if err != nil {
		return nil, err
	}
	return newFilterNode(filters, body, start), nil
}
