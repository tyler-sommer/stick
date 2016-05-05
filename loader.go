package stick

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

// Loader defines a type that can load Stick templates using the given name.
type Loader interface {
	// Load attempts to load the specified template, returning a Template or an error.
	Load(name string) (Template, error)
}

type stringTemplate struct {
	name     string
	contents string
}

func (t *stringTemplate) Name() string {
	return t.name
}

func (t *stringTemplate) Contents() io.Reader {
	return bytes.NewBufferString(t.contents)
}

// StringLoader is intended to be used to load Stick templates directly from a string.
type StringLoader struct{}

// Load on a StringLoader simply returns the name that is passed in.
func (l *StringLoader) Load(name string) (Template, error) {
	return &stringTemplate{name, name}, nil
}

// MemoryLoader loads templates from an in-memory map.
type MemoryLoader struct {
	Templates map[string]string
}

// Load tries to load the template from the in-memory map.
func (l *MemoryLoader) Load(name string) (Template, error) {
	v, ok := l.Templates[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return &stringTemplate{name, v}, nil
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
