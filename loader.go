package stick

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Loader defines a type that can load Stick templates using the given name.
type Loader interface {
	// Load attempts to load the specified template, returning its content or an error.
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

// Load on a StringLoader simply returns the name that is passed in.
func (l *StringLoader) Load(name string) (string, error) {
	return name, nil
}

// A FilesystemLoader loads templates from a filesystem.
type FilesystemLoader struct {
	rootDir string
}

// NewFilesystemLoader creates a new FilesystemLoader with the specified root directory.
func NewFilesystemLoader(rootDir string) *FilesystemLoader {
	return &FilesystemLoader{rootDir}
}

// Load on a FileSystemLoader attempts to load the given file, relative to the
// configured root directory.
func (l *FilesystemLoader) Load(name string) (string, error) {
	sep := string(os.PathSeparator)
	path := l.rootDir + sep + strings.TrimLeft(name, sep)
	res, err := ioutil.ReadFile(path)
	return string(res), err
}
