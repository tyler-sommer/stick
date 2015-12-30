package parse

import (
	"testing"
)

type lexTest struct {
	name   string
	input  string
	tokens []token
}

func mkTok(t tokenType, val string) token {
	return token{val, t, newPos(0, 0)}
}

var (
	tEof            = mkTok(tokenEof, delimEof)
	tSpace          = mkTok(tokenWhitespace, " ")
	tNewLine        = mkTok(tokenWhitespace, "\n")
	tTagOpen        = mkTok(tokenTagOpen, delimOpenTag)
	tTagClose       = mkTok(tokenTagClose, delimCloseTag)
	tPrintOpen      = mkTok(tokenPrintOpen, delimOpenPrint)
	tPrintClose     = mkTok(tokenPrintClose, delimClosePrint)
	tDblStringOpen  = mkTok(tokenStringOpen, "\"")
	tDblStringClose = mkTok(tokenStringClose, "\"")
	tStringOpen     = mkTok(tokenStringOpen, "'")
	tStringClose    = mkTok(tokenStringClose, "'")
	tParensOpen     = mkTok(tokenParensOpen, "(")
	tParensClose    = mkTok(tokenParensClose, ")")
)

var lexTests = []lexTest{
	{"empty", "", []token{tEof}},

	{"number", "{{ 5 }}", []token{
		tPrintOpen,
		tSpace,
		mkTok(tokenNumber, "5"),
		tSpace,
		tPrintClose,
		tEof,
	}},

	{"operator", "{{\n5 == 4 ? 'Yes' : 'No'\n}}", []token{
		tPrintOpen,
		tNewLine,
		mkTok(tokenNumber, "5"),
		tSpace,
		mkTok(tokenOperator, "=="),
		tSpace,
		mkTok(tokenNumber, "4"),
		tSpace,
		mkTok(tokenPunctuation, "?"),
		tSpace,
		tStringOpen,
		mkTok(tokenText, "Yes"),
		tStringClose,
		tSpace,
		mkTok(tokenPunctuation, ":"),
		tSpace,
		tStringOpen,
		mkTok(tokenText, "No"),
		tStringClose,
		tNewLine,
		tPrintClose,
		tEof,
	}},

	{"power and multiply", "{{ 1 ** 10 * 5 }}", []token{
		tPrintOpen,
		tSpace,
		mkTok(tokenNumber, "1"),
		tSpace,
		mkTok(tokenOperator, "**"),
		tSpace,
		mkTok(tokenNumber, "10"),
		tSpace,
		mkTok(tokenOperator, "*"),
		tSpace,
		mkTok(tokenNumber, "5"),
		tSpace,
		tPrintClose,
		tEof,
	}},

	{"div and floordiv", "{{ 10 // 4 / 2 }}", []token{
		tPrintOpen,
		tSpace,
		mkTok(tokenNumber, "10"),
		tSpace,
		mkTok(tokenOperator, "//"),
		tSpace,
		mkTok(tokenNumber, "4"),
		tSpace,
		mkTok(tokenOperator, "/"),
		tSpace,
		mkTok(tokenNumber, "2"),
		tSpace,
		tPrintClose,
		tEof,
	}},

	{"is and is not", "{{ 1 is not 10 and 5 is 5 }}", []token{
		tPrintOpen,
		tSpace,
		mkTok(tokenNumber, "1"),
		tSpace,
		mkTok(tokenOperator, "is not"),
		tSpace,
		mkTok(tokenNumber, "10"),
		tSpace,
		mkTok(tokenOperator, "and"),
		tSpace,
		mkTok(tokenNumber, "5"),
		tSpace,
		mkTok(tokenOperator, "is"),
		tSpace,
		mkTok(tokenNumber, "5"),
		tSpace,
		tPrintClose,
		tEof,
	}},

	{"word operators", "{{ name not in data }}", []token{
		tPrintOpen,
		tSpace,
		mkTok(tokenName, "name"),
		tSpace,
		mkTok(tokenOperator, "not in"),
		tSpace,
		mkTok(tokenName, "data"),
		tSpace,
		tPrintClose,
		tEof,
	}},

	{"unary not operator", "{{ not 100 }}", []token{
		tPrintOpen,
		tSpace,
		mkTok(tokenOperator, "not"),
		tSpace,
		mkTok(tokenNumber, "100"),
		tSpace,
		tPrintClose,
		tEof,
	}},

	{"unary negation operator", "{{ -100 }}", []token{
		tPrintOpen,
		tSpace,
		mkTok(tokenOperator, "-"),
		mkTok(tokenNumber, "100"),
		tSpace,
		tPrintClose,
		tEof,
	}},

	{"text", "<html><head></head></html>", []token{
		mkTok(tokenText, "<html><head></head></html>"),
		tEof,
	}},

	{"simple block", "{% block test %}Some text{% endblock %}", []token{
		tTagOpen,
		tSpace,
		mkTok(tokenName, "block"),
		tSpace,
		mkTok(tokenName, "test"),
		tSpace,
		tTagClose,
		mkTok(tokenText, "Some text"),
		tTagOpen,
		tSpace,
		mkTok(tokenName, "endblock"),
		tSpace,
		tTagClose,
		tEof,
	}},

	{"print string", "{{ \"this is a test\" }}", []token{
		tPrintOpen,
		tSpace,
		tDblStringOpen,
		mkTok(tokenText, "this is a test"),
		tDblStringClose,
		tSpace,
		tPrintClose,
		tEof,
	}},

	{"unclosed string", "{{ \"this is a test }}", []token{
		tPrintOpen,
		tSpace,
		tDblStringOpen,
		mkTok(tokenError, "unclosed string"),
	}},

	{"unclosed parens", "{{ (test + 5 }}", []token{
		tPrintOpen,
		tSpace,
		tParensOpen,
		mkTok(tokenName, "test"),
		tSpace,
		mkTok(tokenOperator, "+"),
		tSpace,
		mkTok(tokenNumber, "5"),
		tSpace,
		tPrintClose,
		tEof,
	}},

	{"unclosed tag (block)", "{% block test %}", []token{
		tTagOpen,
		tSpace,
		mkTok(tokenName, "block"),
		tSpace,
		mkTok(tokenName, "test"),
		tSpace,
		tTagClose,
		tEof,
	}},
}

func collect(t *lexTest) (tokens []token) {
	lex := newLexer(t.input)
	go lex.tokenize()
	for {
		tok := lex.nextToken()
		tokens = append(tokens, tok)
		if tok.tokenType == tokenEof || tok.tokenType == tokenError {
			break
		}
	}

	return
}

func equal(stream1, stream2 []token) bool {
	if len(stream1) != len(stream2) {
		return false
	}
	for k := range stream1 {
		switch {
		case stream1[k].tokenType != stream2[k].tokenType,
			stream1[k].value != stream2[k].value:
			return false
		}
	}

	return true
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		tokens := collect(&test)
		if !equal(tokens, test.tokens) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, tokens, test.tokens)
		}
	}
}
