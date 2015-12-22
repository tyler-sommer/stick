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
	return fmt.Sprintf("parse error: unexpected token \"%s\" on line %d, offset %d", e.tok, e.tok.line, e.tok.offset)
}

// newParseError returns a new ParseError.
func newParseError(tok token) error {
	return &ParseError{tok}
}

// UnexpectedTokenError is generated when the current token
// is not of the expected type.
type UnexpectedTokenError struct {
	actual token
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

	return fmt.Sprintf("parse error: expected one of %s, got \"%s\" on line %d, offset %d", s, e.actual.tokenType, e.actual.line, e.actual.offset)
}

// newUnexpectedTokenError returns a new UnexpectedTokenError
func newUnexpectedTokenError(actual token, expected ...tokenType) error {
	return &UnexpectedTokenError{actual, expected}
}

// UnclosedTagError is generated when a tag is not properly closed.
type UnclosedTagError struct {
	name string
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
	return fmt.Sprintf("parse error: unexpected end of input on line %d, offset %d", e.tok.line, e.tok.offset)
}

// newUnexpectedEofError returns a new UnexpectedEofError
func newUnexpectedEofError(tok token) error {
	return &UnexpectedEofError{tok}
}

// UnexpectedPunctuationError describes an invalid or unexpected punctuation token.
type UnexpectedPunctuationError struct {
	tok token
	expected string
}

func (e *UnexpectedPunctuationError) Error() string {
	return fmt.Sprintf("parse error: unexpected punctuation \"%s\", expected \"%s\" on line %d, offset %d", e.tok.value, e.expected, e.tok.line, e.tok.offset)
}

// newUnexpectedPunctuationError returns a new UnexpectedPunctuationError
func newUnexpectedPunctuationError(tok token, expected string) error {
	return &UnexpectedPunctuationError{tok, expected}
}
