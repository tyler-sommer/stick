package main

import (
	"fmt"
	"strings"
	"unicode"
)

type tokenType int

const (
	tokenText tokenType = iota
	tokenName
	tokenTagOpen
	tokenTagName
	tokenTagClose
	tokenSpace
	tokenEof
)

var names = map[tokenType]string{
	tokenText:     "TEXT",
	tokenName:     "NAME",
	tokenTagOpen:  "TAG_OPEN",
	tokenTagName:  "TAG_NAME",
	tokenTagClose: "TAG_CLOSE",
	tokenEof:      "EOF",
}

const (
	delimOpenTag      = "{%"
	delimCloseTag     = "%}"
	delimOpenPrint    = "{{"
	delimClosePrint   = "}}"
	delimOpenComment  = "{#"
	delimCloseComment = "#}"
)

type lexerState int

const (
	stateData lexerState = iota
	stateBlock
	stateVar
	stateString
	stateInterpolation
)

type token struct {
	value     string
	pos       int
	tokenType tokenType
}

func (tok token) String() string {
	return fmt.Sprintf("{%s '%s' %d}\n", names[tok.tokenType], tok.value, tok.pos)
}

type tokenStream []token

type stateFn func(*lexer) stateFn

type lexer struct {
	pos    int // The position of the last emission
	cursor int // The position of the cursor
	input  string
	tokens tokenStream
	state  stateFn
}

func (lex *lexer) tokenize(code string) tokenStream {
	lex.pos = 0
	lex.cursor = 0
	lex.input = code
	lex.tokens = tokenStream{}

	for lex.state = lexData; lex.state != nil; {
		lex.state = lex.state(lex)
	}

	return lex.tokens
}

func (lex *lexer) next() string {
	fmt.Println(lex.cursor, len(lex.input))
	if lex.cursor+1 >= len(lex.input) {
		return ""
	}

	lex.cursor += 1

	return string(lex.input[lex.cursor])
}

func (lex *lexer) backup() {
	if lex.cursor <= lex.pos {
		return
	}

	fmt.Println("Backing up")
	lex.cursor -= 1
}

func (lex *lexer) peek() string {
	return lex.input[lex.cursor+1 : lex.cursor+2]
}

func (lex *lexer) current() string {
	return lex.input[lex.cursor : lex.cursor+1]
}

func (lex *lexer) ignore() {
	lex.pos = lex.cursor
}

func (lex *lexer) emit(t tokenType) {
	fmt.Println(lex.pos, lex.cursor, len(lex.input))
	val := lex.input[lex.pos:lex.cursor]
	tok := token{val, lex.pos, t}
	fmt.Println(tok)
	lex.tokens = append(lex.tokens, tok)
	lex.pos = lex.cursor
}

func (lex *lexer) consumeWhitespace() {
	if lex.pos != lex.cursor {
		panic("Whitespace may only be consumed directly after emission")
	}
	for {
		str := lex.input[lex.cursor : lex.cursor+1]
		if !isSpace(str) {
			break
		}
		lex.next()
	}

	lex.ignore()
}

func lexData(lex *lexer) stateFn {
	for {
		switch {
		case strings.HasPrefix(lex.input[lex.cursor:], delimOpenTag):
			if lex.cursor > lex.pos {
				lex.emit(tokenText)
			}
			return lexTagOpen
		}

		if lex.next() == "" {
			break
		}
	}

	if lex.cursor > lex.pos {
		lex.emit(tokenText)
	}

	lex.emit(tokenEof)

	return nil
}

func lexExpression(lex *lexer) stateFn {
	lex.consumeWhitespace()

	switch str := lex.current(); {
	case str[0] == delimCloseTag[0]:
		return lexTagClose

	case isAlphaNumeric(str):
		return lexName
	}

	panic("Unknown expression")
}

func lexName(lex *lexer) stateFn {
	for {
		str := lex.current()
		if isAlphaNumeric(str) {
			lex.next()
		} else {
			break
		}
	}

	lex.emit(tokenName)

	return lexExpression
}

func lexTagOpen(lex *lexer) stateFn {
	lex.cursor += len(delimOpenTag)
	lex.emit(tokenTagOpen)

	return lexTagName
}

func lexTagName(lex *lexer) stateFn {
	lex.consumeWhitespace()
	for {
		str := lex.next()
		if !isAlphaNumeric(str) {
			break
		}
	}

	lex.emit(tokenTagName)

	return lexExpression
}

func lexTagClose(lex *lexer) stateFn {
	lex.cursor += len(delimCloseTag)
	lex.emit(tokenTagClose)

	return lexData
}

func isSpace(str string) bool {
	return str == " " || str == "\t"
}

func isAlphaNumeric(str string) bool {
	for _, s := range str {
		if string(s) != "_" && !unicode.IsLetter(s) && !unicode.IsDigit(s) {
			return false
		}
	}

	return true
}

func main() {
	data := "<html><head><title>{% block title %}{% endblock %}"

	lex := lexer{}

	fmt.Println(lex.tokenize(data))
}
