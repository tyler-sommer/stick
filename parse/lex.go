package parse

import (
	"fmt"
	"strings"
	"unicode"
)

type tokenType int

const (
	tokenText tokenType = iota
	tokenName
	tokenNumber
	tokenTagOpen
	tokenTagName
	tokenTagClose
	tokenPrintOpen
	tokenPrintClose
	tokenParensOpen
	tokenParensClose
	tokenArrayOpen
	tokenArrayClose
	tokenHashOpen
	tokenHashClose
	tokenStringOpen
	tokenStringClose
	tokenPunctuation
	tokenOperator
	tokenWhitespace
	tokenError
	tokenEof
)

var names = map[tokenType]string{
	tokenText:        "TEXT",
	tokenName:        "NAME",
	tokenNumber:      "NUMBER",
	tokenTagOpen:     "TAG_OPEN",
	tokenTagName:     "TAG_NAME",
	tokenTagClose:    "TAG_CLOSE",
	tokenPrintOpen:   "PRINT_OPEN",
	tokenPrintClose:  "PRINT_CLOSE",
	tokenParensOpen:  "PARENS_OPEN",
	tokenParensClose: "PARENS_CLOSE",
	tokenArrayOpen:   "ARRAY_OPEN",
	tokenArrayClose:  "ARRAY_CLOSE",
	tokenHashOpen:    "HASH_OPEN",
	tokenHashClose:   "HASH_CLOSE",
	tokenStringOpen:  "STRING_OPEN",
	tokenStringClose: "STRING_CLOSE",
	tokenPunctuation: "PUNCTUATION",
	tokenOperator:    "OPERATOR",
	tokenWhitespace:  "WHITESPACE",
	tokenError:       "ERROR",
	tokenEof:         "EOF",
}

const (
	delimEof          = ""
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
	return fmt.Sprintf("{%s '%s' %d}", names[tok.tokenType], tok.value, tok.pos)
}

type stateFn func(*lexer) stateFn

type lexer struct {
	start  int // The position of the last emission
	pos    int // The position of the cursor
	parens int // Number of still-open parenthesis in the current expression
	input  string
	tokens chan token
	state  stateFn
}

func newLexer(input string) *lexer {
	return &lexer{0, 0, 0, input, make(chan token), nil}
}

func (l *lexer) next() (val string) {
	if l.pos >= len(l.input) {
		val = delimEof

	} else {
		val = l.input[l.pos : l.pos+1]

		l.pos++
	}

	return
}

func (l *lexer) backup() {
	l.pos--
}

func (l *lexer) peek() (val string) {
	val = l.next()

	l.backup()

	return
}

func (l *lexer) emit(t tokenType) {
	val := ""
	if l.pos <= len(l.input) {
		val = l.input[l.start:l.pos]
	}

	tok := token{val, l.start, t}
	l.tokens <- tok
	l.start = l.pos
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	tok := token{fmt.Sprintf(format, args...), l.start, tokenError}
	l.tokens <- tok

	return nil
}

func (l *lexer) tokenize() {
	for l.state = lexData; l.state != nil; {
		l.state = l.state(l)
	}
}

func lexData(l *lexer) stateFn {
	for {
		switch {
		case strings.HasPrefix(l.input[l.pos:], delimOpenTag):
			if l.pos > l.start {
				l.emit(tokenText)
			}
			return lexTagOpen

		case strings.HasPrefix(l.input[l.pos:], delimOpenPrint):
			if l.pos > l.start {
				l.emit(tokenText)
			}
			return lexPrintOpen
		}

		if l.next() == delimEof {
			break
		}
	}

	if l.pos > l.start {
		l.emit(tokenText)
	}

	l.emit(tokenEof)

	return nil
}

func lexExpression(l *lexer) stateFn {
	switch str := l.peek(); {
	case strings.HasPrefix(l.input[l.pos:], delimCloseTag):
		if l.pos > l.start {
			panic("Incomplete token?")
		}
		return lexTagClose

	case strings.HasPrefix(l.input[l.pos:], delimClosePrint):
		if l.pos > l.start {
			panic("Incomplete token?")
		}
		return lexPrintClose

	case isOperator(str):
		return lexOperator

	case isPunctuation(str):
		return lexPunctuation

	case strings.ContainsAny(str, "([{"):
		return lexOpenParens

	case strings.ContainsAny(str, "}])"):
		return lexCloseParens

	case strings.ContainsAny(str, "\"'"):
		return lexString

	case isNumeric(str):
		return lexNumber

	case isName(str):
		return lexName

	case isSpace(str):
		return lexSpace

	default:
		panic("Unknown expression")
	}
}

func lexSpace(l *lexer) stateFn {
	for {
		str := l.next()
		if !isSpace(str) {
			l.backup()
			break
		}
	}

	l.emit(tokenWhitespace)

	return lexExpression
}

func lexNumber(l *lexer) stateFn {
	for {
		str := l.next()
		if !isNumeric(str) {
			l.backup()
			break
		}
	}

	l.emit(tokenNumber)

	return lexExpression
}

func lexOperator(l *lexer) stateFn {
	for {
		str := l.next()
		if !isOperator(str) {
			l.backup()
			break
		}
	}

	l.emit(tokenOperator)

	return lexExpression
}

func lexPunctuation(l *lexer) stateFn {
	for {
		str := l.next()
		if !isPunctuation(str) {
			l.backup()
			break
		}
	}
	
	l.emit(tokenPunctuation)

	return lexExpression
}

func lexString(l *lexer) stateFn {
	open := l.next()
	l.emit(tokenStringOpen)
	closePos := strings.Index(l.input[l.pos:], open)
	if closePos < 0 {
		return l.errorf("unclosed string")
	}

	l.pos += closePos
	l.emit(tokenText)

	l.next()
	l.emit(tokenStringClose)

	return lexExpression
}

func lexOpenParens(l *lexer) stateFn {
	switch str := l.next(); {
	case str == "(":
		l.emit(tokenParensOpen)

	case str == "[":
		l.emit(tokenArrayOpen)

	case str == "{":
		l.emit(tokenHashOpen)

	default:
		panic("Unknown parens")
	}

	l.parens += 1

	return lexExpression
}

func lexCloseParens(l *lexer) stateFn {
	switch str := l.next(); {
	case str == ")":
		l.emit(tokenParensClose)

	case str == "]":
		l.emit(tokenArrayClose)

	case str == "}":
		l.emit(tokenHashClose)

	default:
		panic("Unknown parens")
	}

	l.parens -= 1

	return lexExpression
}

func lexName(l *lexer) stateFn {
	for {
		str := l.next()
		if !isAlphaNumeric(str) {
			l.backup()
			break
		}
	}

	l.emit(tokenName)

	return lexExpression
}

func lexTagOpen(l *lexer) stateFn {
	l.pos += len(delimOpenTag)
	l.emit(tokenTagOpen)

	return lexExpression
}

func lexTagClose(l *lexer) stateFn {
	if l.parens > 0 {
		return l.errorf("unclosed parenthesis")
	}
	l.pos += len(delimCloseTag)
	l.emit(tokenTagClose)

	return lexData
}

func lexPrintOpen(l *lexer) stateFn {
	l.pos += len(delimOpenPrint)
	l.emit(tokenPrintOpen)

	return lexExpression
}

func lexPrintClose(l *lexer) stateFn {
	if l.parens > 0 {
		return l.errorf("unclosed parenthesis")
	}
	l.pos += len(delimClosePrint)
	l.emit(tokenPrintClose)

	return lexData
}

func isSpace(str string) bool {
	return str == " " || str == "\t"
}

func isName(str string) bool {
	for _, s := range str {
		if string(s) != "_" && !unicode.IsLetter(s) && !unicode.IsDigit(s) {
			return false
		}
	}

	return true
}

func isAlphaNumeric(str string) bool {
	for _, s := range str {
		if !unicode.IsLetter(s) && !unicode.IsDigit(s) {
			return false
		}
	}

	return true
}

func isNumeric(str string) bool {
	for _, s := range str {
		if !unicode.IsDigit(s) {
			return false
		}
	}

	return true
}

func isOperator(str string) bool {
	for _, s := range str {
		if !strings.ContainsAny(string(s), "+-/*%~=") {
			return false
		}
	}
	
	return true
}

func isPunctuation(str string) bool {
	for _, s := range str {
		if !strings.ContainsAny(string(s), ",?:|") {
			return false
		}
	}
	
	return true
}


