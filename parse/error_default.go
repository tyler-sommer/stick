// +build !debug

package parse

import (
	"fmt"
)

type baseError struct {
	p    pos
	name string
}

func (e *baseError) setTree(t *Tree) {
	e.name = t.name
}

func newBaseError(p pos) *baseError {
	return &baseError{p, ""}
}

func (e *baseError) sprintf(format string, a ...interface{}) string {
	res := fmt.Sprintf(format, a...)
	return fmt.Sprintf("parse: %s on line %d, column %d in %s", res, e.p.Line, e.p.Offset, e.name)
}
