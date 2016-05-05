// +build !debug

package parse

type baseError struct {
	parseError
}

func newBaseError(p Pos) baseError {
	return baseError{newParseError(p)}
}
