package stick

import "testing"

type lexTest struct {
	name   string
	input  string
	tokens []token
}

func mkTok(t tokenType, val string) token {
	return token{val, 0, t}
}

var (
	tEof            = mkTok(tokenEof, delimEof)
	tTagOpen        = mkTok(tokenTagOpen, delimOpenTag)
	tTagClose       = mkTok(tokenTagClose, delimCloseTag)
	tPrintOpen      = mkTok(tokenPrintOpen, delimOpenPrint)
	tPrintClose     = mkTok(tokenPrintClose, delimClosePrint)
	tDblStringOpen  = mkTok(tokenStringOpen, "\"")
	tDblStringClose = mkTok(tokenStringClose, "\"")
)

var lexTests = []lexTest{
	{"empty", "", []token{tEof}},

	{"text", "<html><head></head></html>", []token{
		mkTok(tokenText, "<html><head></head></html>"),
		tEof,
	}},

	{"simple block", "{% block test %}Some text{% endblock %}", []token{
		tTagOpen,
		mkTok(tokenTagName, "block"),
		mkTok(tokenName, "test"),
		tTagClose,
		mkTok(tokenText, "Some text"),
		tTagOpen,
		mkTok(tokenTagName, "endblock"),
		tTagClose,
		tEof,
	}},

	{"print string", "{{ \"this is a test\" }}", []token{
		tPrintOpen,
		tDblStringOpen,
		mkTok(tokenText, "this is a test"),
		tDblStringClose,
		tPrintClose,
		tEof,
	}},

	{"unclosed string", "{{ \"this is a test }}", []token{
		tPrintOpen,
		tDblStringOpen,
		mkTok(tokenError, "unclosed string"),
	}},
}

func collect(t *lexTest) tokenStream {
	return lex(t.input)
}

func equal(stream1, stream2 tokenStream) bool {
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
