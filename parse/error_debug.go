// +build debug

package parse

import (
	"fmt"
	"runtime"
)

type baseError struct {
	p          pos
	name       string
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

func (e *baseError) SetName(n string) {
	e.name = n
}

func (e *baseError) setTree(t *Tree) {
	e.lastTokens = append(e.lastTokens, t.read[len(t.read)-5:]...)
	e.lastTokens = append(e.lastTokens, t.unread...)
}

func newBaseError(p pos) *baseError {
	return &baseError{p, "", getTrace(), make([]token, 0)}
}

func (e *baseError) sprintf(format string, a ...interface{}) string {
	res := fmt.Sprintf(format, a...)
	nameStr := ""
	if e.name != "" {
		nameStr = fmt.Sprintf(" in %s", e.name)
	}
	if e.debug != "" {
		return fmt.Sprintf("parse: %s on line %d, column %d%s\n\ncall stack:\n%s\ntokens:\n%v", res, e.p.Line, e.p.Offset, nameStr, e.debug, e.lastTokens)
	}
	return fmt.Sprintf("parse: %s on line %d, column %d%s", res, e.p.Line, e.p.Offset, nameStr)
}
