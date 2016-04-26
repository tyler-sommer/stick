// +build !debug

package parse

import (
	"fmt"
)

type baseError struct {
	p    pos
	name string
}

func (e *baseError) SetName(n string) {
	e.name = n
}

func (e *baseError) setTree(t *Tree) {}

func newBaseError(p pos) *baseError {
	return &baseError{p, ""}
}

func (e *baseError) sprintf(format string, a ...interface{}) string {
	res := fmt.Sprintf(format, a...)
	nameStr := ""
	if e.name != "" {
		nameStr = fmt.Sprintf(" in %s", e.name)
	}
	return fmt.Sprintf("parse: %s on line %d, column %d%s", res, e.p.Line, e.p.Offset, nameStr)
}
