package parse

import (
	"fmt"
)

// newParseError generates a generic error when no additional context
// is available other than the current token.
func newParseError(tok token) error {
	return fmt.Errorf("parse error: unexpected token \"%s\" on line %d, offset %d", tok, tok.line, tok.offset)
}

// newUnexpectedTokenError generates an error describing when one or more known
// token types are expected and not received.
func newUnexpectedTokenError(actual token, expected ...tokenType) error {
	var s = "["
	for i, e := range expected {
		if i > 0 {
			s = s + ", "
		}

		s = s + e.String()
	}

	s = s + "]"

	return fmt.Errorf("parse error: expected one of %s, got \"%s\" on line %d, offset %d", s, actual.tokenType, actual.line, actual.offset)
}

// newUnclosedTagError generates an error describing an unclosed tag.
func newUnclosedTagError(name string, start pos) error {
	return fmt.Errorf("parse error: unclosed tag \"%s\" starting on line %d, offset %d", name, start.Line, start.Offset)
}

// newUnexpectedEofError generates an error describing an unexpected end of input.
func newUnexpectedEofError(tok token) error {
	return fmt.Errorf("parse error: unexpected end of input on line %d, offset %d", tok.line, tok.offset)
}

// newUnexpectedPunctuationError generates an error describing
// an incorrect punctuation token.
func newUnexpectedPunctuationError(tok token, expected string) error {
	return fmt.Errorf("parse error: unexpected punctuation \"%s\", expected \"%s\" on line %d, offset %d", tok.value, expected, tok.line, tok.offset)
}
