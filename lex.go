package main

import (
	"fmt"
	"strings"
)

type tokenType int

const (
	tokenText tokenType = iota
	tokenTagOpen
	tokenTagDefinition
	tokenTagEnd
	tokenTagClose
	tokenEof
)

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

type tokenStream []token

type stateFn func(*lexer) stateFn

type lexer struct {
	pos     int // The position of the last emission
	current int // The position of the cursor
	input   string
	tokens  tokenStream
	state   stateFn
}

func (lex *lexer) tokenize(code string) tokenStream {
	lex.pos = 0
	lex.current = 0
	lex.input = code
	lex.tokens = tokenStream{}

	for lex.state = lexData; lex.state != nil; {
		lex.state = lex.state(lex)
	}

	return lex.tokens
}

func (lex *lexer) next() string {
	fmt.Println(lex.current, len(lex.input))
	if lex.current == len(lex.input) {
		return ""
	}

	lex.current++

	return string(lex.input[lex.current])
}

func (lex *lexer) emit(t tokenType) {
	fmt.Println(lex.pos, lex.current, len(lex.input))
	val := lex.input[lex.pos:lex.current]
	lex.tokens = append(lex.tokens, token{val, lex.pos, t})
	lex.pos = lex.current
}

func lexData(lex *lexer) stateFn {
	for {
		switch {
		case strings.HasPrefix(lex.input[lex.current:], delimOpenTag):
			if lex.current > lex.pos {
				lex.emit(tokenText)
			}
			return lexTagOpen
		}

		if lex.next() == "" {
			break
		}
	}

	if lex.current > lex.pos {
		lex.emit(tokenText)
	}

	lex.emit(tokenEof)

	return nil
}

func lexTagOpen(lex *lexer) stateFn {
	lex.current += len(delimOpenTag)
	lex.emit(tokenTagOpen)

	return lexTagDefinition
}

func lexTagDefinition(lex *lexer) stateFn {
	closePos := strings.Index(lex.input[lex.current:], delimCloseTag)
	if closePos < 0 {
		panic("unexpected end of tag")
	}

	lex.current += closePos
	lex.emit(tokenTagDefinition)

	return lexTagClose
}

func lexTagClose(lex *lexer) stateFn {
	lex.current += len(delimCloseTag)
	lex.emit(tokenTagClose)

	return lexData
}

func main() {
	data := "<html><head><title>{% block title %}{% endblock %}"

	lex := lexer{}

	fmt.Println(lex.tokenize(data))
}
