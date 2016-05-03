package parse

import (
	"bytes"
	"testing"
)

type lexTest struct {
	name   string
	input  string
	tokens []token
}

func mkTok(t tokenType, val string) token {
	return token{val, t, Pos{0, 0}}
}

var (
	tEOF              = mkTok(tokenEOF, delimEOF)
	tSpace            = mkTok(tokenWhitespace, " ")
	tNewLine          = mkTok(tokenWhitespace, "\n")
	tCommentOpen      = mkTok(tokenCommentOpen, delimOpenComment)
	tCommentClose     = mkTok(tokenCommentClose, delimCloseComment)
	tCommentTrimOpen  = mkTok(tokenCommentOpen, delimOpenComment+delimTrimWhitespace)
	tCommentTrimClose = mkTok(tokenCommentClose, delimTrimWhitespace+delimCloseComment)
	tTagOpen          = mkTok(tokenTagOpen, delimOpenTag)
	tTagClose         = mkTok(tokenTagClose, delimCloseTag)
	tTagTrimOpen      = mkTok(tokenTagOpen, delimOpenTag+delimTrimWhitespace)
	tTagTrimClose     = mkTok(tokenTagClose, delimTrimWhitespace+delimCloseTag)
	tPrintOpen        = mkTok(tokenPrintOpen, delimOpenPrint)
	tPrintClose       = mkTok(tokenPrintClose, delimClosePrint)
	tPrintTrimOpen    = mkTok(tokenPrintOpen, delimOpenPrint+delimTrimWhitespace)
	tPrintTrimClose   = mkTok(tokenPrintClose, delimTrimWhitespace+delimClosePrint)
	tDblStringOpen    = mkTok(tokenStringOpen, "\"")
	tDblStringClose   = mkTok(tokenStringClose, "\"")
	tStringOpen       = mkTok(tokenStringOpen, "'")
	tStringClose      = mkTok(tokenStringClose, "'")
	tInterpolateOpen  = mkTok(tokenInterpolateOpen, delimOpenInterpolate)
	tInterpolateClose = mkTok(tokenInterpolateClose, delimCloseInterpolate)
	tParensOpen       = mkTok(tokenParensOpen, "(")
	tParensClose      = mkTok(tokenParensClose, ")")
)

var lexTests = []lexTest{
	{"empty", "", []token{tEOF}},

	{"comment", "Some text{# Hello there #}", []token{
		mkTok(tokenText, "Some text"),
		tCommentOpen,
		mkTok(tokenText, " Hello there "),
		tCommentClose,
		tEOF,
	}},

	{"unclosed comment", "{# Hello there", []token{
		tCommentOpen,
		mkTok(tokenText, " Hello there"),
		mkTok(tokenError, "expected comment close"),
	}},

	{"number", "{{ 5 }}", []token{
		tPrintOpen,
		tSpace,
		mkTok(tokenNumber, "5"),
		tSpace,
		tPrintClose,
		tEOF,
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
		tEOF,
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
		tEOF,
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
		tEOF,
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
		tEOF,
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
		tEOF,
	}},

	{"unary not operator", "{{ not 100 }}", []token{
		tPrintOpen,
		tSpace,
		mkTok(tokenOperator, "not"),
		tSpace,
		mkTok(tokenNumber, "100"),
		tSpace,
		tPrintClose,
		tEOF,
	}},

	{"unary negation operator", "{{ -100 }}", []token{
		tPrintOpen,
		tSpace,
		mkTok(tokenOperator, "-"),
		mkTok(tokenNumber, "100"),
		tSpace,
		tPrintClose,
		tEOF,
	}},

	{"text", "<html><head></head></html>", []token{
		mkTok(tokenText, "<html><head></head></html>"),
		tEOF,
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
		tEOF,
	}},

	{"print string", "{{ \"this is a test\" }}", []token{
		tPrintOpen,
		tSpace,
		tDblStringOpen,
		mkTok(tokenText, "this is a test"),
		tDblStringClose,
		tSpace,
		tPrintClose,
		tEOF,
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
		mkTok(tokenError, "unclosed parenthesis"),
	}},

	{"unclosed tag (block)", "{% block test %}", []token{
		tTagOpen,
		tSpace,
		mkTok(tokenName, "block"),
		tSpace,
		mkTok(tokenName, "test"),
		tSpace,
		tTagClose,
		tEOF,
	}},

	{"name with underscore", "{% block additional_javascripts %}", []token{
		tTagOpen,
		tSpace,
		mkTok(tokenName, "block"),
		tSpace,
		mkTok(tokenName, "additional_javascripts"),
		tSpace,
		tTagClose,
		tEOF,
	}},

	{"string interpolation", `{{ "Hello, #{name}" }}`, []token{
		tPrintOpen,
		tSpace,
		tDblStringOpen,
		mkTok(tokenText, "Hello, "),
		tInterpolateOpen,
		mkTok(tokenName, "name"),
		tInterpolateClose,
		tDblStringClose,
		tSpace,
		tPrintClose,
		tEOF,
	}},

	{"string interpolation", `{{ "Item #: #{item.id}<br>" }}`, []token{
		tPrintOpen,
		tSpace,
		tDblStringOpen,
		mkTok(tokenText, "Item #: "),
		tInterpolateOpen,
		mkTok(tokenName, "item"),
		mkTok(tokenPunctuation, "."),
		mkTok(tokenName, "id"),
		tInterpolateClose,
		mkTok(tokenText, "<br>"),
		tDblStringClose,
		tSpace,
		tPrintClose,
		tEOF,
	}},

	{"whitespace control print", `{{- test -}}`, []token{
		tPrintTrimOpen,
		tSpace,
		mkTok(tokenName, "test"),
		tSpace,
		tPrintTrimClose,
		tEOF,
	}},

	{"whitespace control tag", `{%- test -%}`, []token{
		tTagTrimOpen,
		tSpace,
		mkTok(tokenName, "test"),
		tSpace,
		tTagTrimClose,
		tEOF,
	}},

	{"whitespace control comment", `{#- test -#}`, []token{
		tCommentTrimOpen,
		mkTok(tokenText, " test "),
		tCommentTrimClose,
		tEOF,
	}},
}

func collect(t *lexTest) (tokens []token) {
	lex := newLexer(bytes.NewReader([]byte(t.input)))
	go lex.tokenize()
	for {
		tok := lex.nextToken()
		tokens = append(tokens, tok)
		if tok.tokenType == tokenEOF || tok.tokenType == tokenError {
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
