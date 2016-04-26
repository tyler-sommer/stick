package stick

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Loader defines a type that can load Stick templates using the given name.
type Loader interface {
	// Load attempts to load the specified template, returning a Template or an error.
	Load(name string) (Template, error)
}

// UnableToLoadTemplateErr describes a template that was not able to be loaded.
type UnableToLoadTemplateErr struct {
	name string
}

func (e *UnableToLoadTemplateErr) Error() string {
	return fmt.Sprintf("Unable to load template: %s", e.name)
}

type stringTemplate struct {
	contents string
}

func (t *stringTemplate) Name() string {
	return t.contents
}

func (t *stringTemplate) Contents() io.Reader {
	return bytes.NewReader([]byte(t.contents))
}

// StringLoader is intended to be used to load Stick templates directly from a string.
type StringLoader struct{}

// Load on a StringLoader simply returns the name that is passed in.
func (l *StringLoader) Load(name string) (Template, error) {
	return &stringTemplate{name}, nil
}

type fileTemplate struct {
	name   string
	reader io.Reader
}

func (t *fileTemplate) Name() string {
	return t.name
}

func (t *fileTemplate) Contents() io.Reader {
	return t.reader
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
func (l *FilesystemLoader) Load(name string) (Template, error) {
	path := filepath.Join(l.rootDir, name)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &fileTemplate{name, f}, nil
}
