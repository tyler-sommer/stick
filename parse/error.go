package parse

import (
	"fmt"
)

// ParseError represents a generic error during parsing when no additional
// context is available other than the current token.
type ParseError struct {
	tok token
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error: unexpected token \"%s\" on line %d, offset %d", e.tok, e.tok.Line, e.tok.Offset)
}

// newParseError returns a new ParseError.
func newParseError(tok token) error {
	return &ParseError{tok}
}

// UnexpectedTokenError is generated when the current token
// is not of the expected type.
type UnexpectedTokenError struct {
	actual   token
	expected []tokenType
}

func (e *UnexpectedTokenError) Error() string {
	var s = "["
	for i, e := range e.expected {
		if i > 0 {
			s = s + ", "
		}

		s = s + e.String()
	}

	s = s + "]"

	return fmt.Sprintf("parse error: expected one of %s, got \"%s\" on line %d, offset %d", s, e.actual.tokenType, e.actual.Line, e.actual.Offset)
}

// newUnexpectedTokenError returns a new UnexpectedTokenError
func newUnexpectedTokenError(actual token, expected ...tokenType) error {
	return &UnexpectedTokenError{actual, expected}
}

// UnclosedTagError is generated when a tag is not properly closed.
type UnclosedTagError struct {
	name  string
	start pos
}

func (e *UnclosedTagError) Error() string {
	return fmt.Sprintf("parse error: unclosed tag \"%s\" starting on line %d, offset %d", e.name, e.start.Line, e.start.Offset)
}

// newUnclosedTagError returns a new UnclosedTagError.
func newUnclosedTagError(name string, start pos) error {
	return &UnclosedTagError{name, start}
}

// UnexpectedEofError describes an unexpected end of input.
type UnexpectedEofError struct {
	tok token
}

func (e *UnexpectedEofError) Error() string {
	return fmt.Sprintf("parse error: unexpected end of input on line %d, offset %d", e.tok.Line, e.tok.Offset)
}

// newUnexpectedEofError returns a new UnexpectedEofError
func newUnexpectedEofError(tok token) error {
	return &UnexpectedEofError{tok}
}

// UnexpectedValueError describes an invalid or unexpected value inside a token.
type UnexpectedValueError struct {
	tok token  // The actual token
	val string // The expected value
}

func (e *UnexpectedValueError) Error() string {
	return fmt.Sprintf("parse error: unexpected \"%s\", expected \"%s\" on line %d, offset %d", e.tok.value, e.val, e.tok.Line, e.tok.Offset)
}

// newUnexpectedValueError returns a new UnexpectedPunctuationError
func newUnexpectedValueError(tok token, expected string) error {
	return &UnexpectedValueError{tok, expected}
}

// MultipleExtendsError describes an attempt to extend from multiple parent templates.
type MultipleExtendsError struct {
	start pos
}

func (e *MultipleExtendsError) Error() string {
	return fmt.Sprintf("parse error: a template may have only one \"extends\" statement, near line %d, offset %d", e.start.Line, e.start.Offset)
}

// newMultipleExtendsError returns a new MultipleExtendsError
func newMultipleExtendsError(start pos) error {
	return &MultipleExtendsError{start}
}
