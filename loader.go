package stick

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

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

// StringLoader is intended to be used to load Stick templates directly from a string.
type StringLoader struct{}

func (l *StringLoader) Load(name string) (string, error) {
	return name, nil
}

type FilesystemLoader struct {
	rootDir string
}

func NewFilesystemLoader(rootDir string) *FilesystemLoader {
	return &FilesystemLoader{rootDir}
}

func (l *FilesystemLoader) Load(name string) (string, error) {
	sep := string(os.PathSeparator)
	path := l.rootDir + sep + strings.TrimLeft(name, sep)
	res, err := ioutil.ReadFile(path)
	return string(res), err
}
