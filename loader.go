package stick

import "fmt"

// Loader defines a type that can load stick templates using the given name.
type Loader interface {
	Load(name string) (string, error)
}

// UnableToLoadTemplateErr describes a template that was not able to be loaded.
type UnableToLoadTemplateErr struct {
	name string
}

func (e *UnableToLoadTemplateErr) Error() string {
	return fmt.Sprintf("Unable to load template: %s", e.name)
}

type StringLoader struct {}

func (l *StringLoader) Load(name string) (string, error) {
	return name, nil
}
