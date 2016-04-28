package parse

// ParsingError is followed by all errors emitted from the parse package.
type ParsingError interface {
	error

	// Pos returns the position where the error originated.
	Pos() Pos

	// setTree is used internally to enrich the error with extra information.
	setTree(t *Tree)
}

// UnexpectedTokenError is generated when the current token
// is not of the expected type.
type UnexpectedTokenError struct {
	*baseError
	actual   token
	expected []tokenType
}

func (e *UnexpectedTokenError) Error() string {
	if len(e.expected) == 0 {
		return e.sprintf(`unexpected token "%s"`, e.actual.tokenType)
	}
	if len(e.expected) == 1 {
		return e.sprintf(`expected "%s", got "%s"`, e.expected[0], e.actual.tokenType)
	}

	s := "["
	for i, e := range e.expected {
		if i > 0 {
			s = s + ", "
		}
		s = s + e.String()
	}
	s = s + "]"
	return e.sprintf(`expected one of %s, got "%s"`, s, e.actual.tokenType)
}

// newUnexpectedTokenError returns a new UnexpectedTokenError
func newUnexpectedTokenError(actual token, expected ...tokenType) error {
	return &UnexpectedTokenError{newBaseError(actual.Pos), actual, expected}
}

// UnclosedTagError is generated when a tag is not properly closed.
type UnclosedTagError struct {
	*baseError
	name string
}

func (e *UnclosedTagError) Error() string {
	return e.sprintf(`unclosed tag "%s" starting`, e.name)
}

// newUnclosedTagError returns a new UnclosedTagError.
func newUnclosedTagError(name string, start Pos) error {
	return &UnclosedTagError{newBaseError(start), name}
}

// UnexpectedEOFError describes an unexpected end of input.
type UnexpectedEOFError struct {
	*baseError
}

func (e *UnexpectedEOFError) Error() string {
	return e.sprintf(`unexpected end of input`)
}

// newUnexpectedEOFError returns a new UnexpectedEOFError
func newUnexpectedEOFError(tok token) error {
	return &UnexpectedEOFError{newBaseError(tok.Pos)}
}

// UnexpectedValueError describes an invalid or unexpected value inside a token.
type UnexpectedValueError struct {
	*baseError
	tok token  // The actual token
	val string // The expected value
}

func (e *UnexpectedValueError) Error() string {
	return e.sprintf(`unexpected "%s", expected "%s"`, e.tok.value, e.val)
}

// newUnexpectedValueError returns a new UnexpectedPunctuationError
func newUnexpectedValueError(tok token, expected string) error {
	return &UnexpectedValueError{newBaseError(tok.Pos), tok, expected}
}

// MultipleExtendsError describes an attempt to extend from multiple parent templates.
type MultipleExtendsError struct {
	*baseError
}

func (e *MultipleExtendsError) Error() string {
	return e.sprintf(`a template may have only one "extends" statement`)
}

// newMultipleExtendsError returns a new MultipleExtendsError
func newMultipleExtendsError(start Pos) error {
	return &MultipleExtendsError{newBaseError(start)}
}
