package stick

import (
	"fmt"

	"github.com/tyler-sommer/stick/parse"
)

type enrichedParseError struct {
	parse.ParsingError // The original error.

	name string // The template in which this error occurred.
}

func (e *enrichedParseError) Error() string {
	return fmt.Sprintf("%s in %s", e.ParsingError.Error(), e.name)
}

func enrichError(tpl Template, err error) error {
	if t, ok := err.(parse.ParsingError); ok {
		return &enrichedParseError{t, tpl.Name()}
	}
	return err
}
