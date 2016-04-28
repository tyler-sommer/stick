// +build debug

package parse

import (
	"fmt"
	"runtime"
)

type baseError struct {
	Pos
	debug      string
	lastTokens []token
}

func getTrace() string {
	res := ""
	for i := 3; true; i++ { // Starting at caller 3 leaves out the junk error handling trace.
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		res = res + fmt.Sprintf("%s\n\t%s:%d", fn.Name(), file, line) + "\n"
	}
	return res
}

func (e *baseError) setTree(t *Tree) {
	l := len(t.read) - 5
	if l < 0 {
		l = 0
	}
	e.lastTokens = append(e.lastTokens, t.read[l:]...)
	e.lastTokens = append(e.lastTokens, t.unread...)
}

func newBaseError(p Pos) *baseError {
	return &baseError{p, getTrace(), make([]token, 0)}
}

func (e *baseError) sprintf(format string, a ...interface{}) string {
	res := fmt.Sprintf(format, a...)
	return fmt.Sprintf("parse: %s on line %d, column %d", res, e.Line, e.Offset)
}

func (e *baseError) Debug() string {
	return fmt.Sprintf("call stack:\n%s\ntokens:\n%v")
}
