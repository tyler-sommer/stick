package main

import (
	"fmt"
	"strings"
	"unicode"
)

type tokenType int

const (
	tokenText tokenType = iota
	tokenTagOpen
	tokenTagName
	tokenTagArguments
	tokenTagClose
	tokenSpace
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
	if lex.current+1 >= len(lex.input) {
		return ""
	}

	lex.current += 1

	return string(lex.input[lex.current])
}

func (lex *lexer) backup() {
	if lex.current <= lex.pos {
		return
	}
	
	fmt.Println("Backing up")
	lex.current -= 1
}

func (lex *lexer) peek() string {
	str := lex.next()
	lex.backup()
	
	return str
}

func (lex *lexer) emit(t tokenType) {
	fmt.Println(lex.pos, lex.current, len(lex.input))
	val := lex.input[lex.pos:lex.current]
	tok := token{val, lex.pos, t}
	fmt.Println(tok)
	lex.tokens = append(lex.tokens, tok)
	lex.pos = lex.current
}

func (lex *lexer) consumeWhitespace() {
	fmt.Println("Consuming whitespace")
	for {
		str := lex.input[lex.current:lex.current+1]
		fmt.Println("A space?", str)
		if (!isSpace(str)) {
			break
		}
		lex.next()
	}
	
	fmt.Println("Emitting a space token")
	lex.emit(tokenSpace)
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

	return lexTagName
}

func lexTagName(lex *lexer) stateFn {
	lex.consumeWhitespace()
	for {
		str := lex.next();
		if !isAlphaNumeric(str) {
			break
		}
	}

	lex.emit(tokenTagName)

	return lexTagArguments
}

func lexTagArguments(lex *lexer) stateFn {
	lex.consumeWhitespace()
	closePos := strings.Index(lex.input[lex.current:], delimCloseTag)
	if closePos > 0 {
		lex.current += closePos
		lex.emit(tokenTagArguments)
	}
	
	return lexTagClose
}

func lexTagClose(lex *lexer) stateFn {
	lex.current += len(delimCloseTag)
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
