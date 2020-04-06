// +build debug

package parse

import (
	"fmt"
	"runtime"
)

type baseError struct {
	parseError
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
	e.name = t.Name
	l := len(t.read) - 5
	if l < 0 {
		l = 0
	}
	e.lastTokens = append(e.lastTokens, t.read[l:]...)
	e.lastTokens = append(e.lastTokens, t.unread...)
}

func newBaseError(p Pos) baseError {
	return baseError{newParseError(p), getTrace(), make([]token, 0)}
}

func (e *baseError) Debug() string {
	return fmt.Sprintf("call stack:\n%s\ntokens:\n%v", e.debug, e.lastTokens)
}
