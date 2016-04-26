package stick

import (
	"github.com/tyler-sommer/stick/parse"
)

func enrichError(tpl Template, err error) error {
	if t, ok := err.(parse.Error); ok {
		t.SetName(tpl.Name())
	}
	return err
}
