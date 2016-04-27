package stick

import (
	"io"

	"github.com/tyler-sommer/stick/parse"
)

// A Template represents a named template and its contents.
type Template interface {
	// Name returns the name of this Template.
	Name() string

	// Contents returns an io.Reader for reading the Template contents.
	Contents() io.Reader
}

// A Func represents a user-defined function.
// Functions can be called anywhere expressions are allowed and
// take any number of arguments.
type Func func(e *Env, args ...Value) Value

// A Filter is a user-defined filter.
// Filters receive a value and modify it in some way. Filters
// also accept parameters.
type Filter func(e *Env, val Value, args ...Value) Value

// A Test represents a user-defined test.
// Tests are used to make some comparisons more expressive. Tests
// also accept arguments and can consist of two words.
type Test func(e *Env, val Value, args ...Value) bool

// Env represents a configured Stick environment.
type Env struct {
	Loader    Loader              // Template loader.
	Functions map[string]Func     // User-defined functions.
	Filters   map[string]Filter   // User-defined filters.
	Tests     map[string]Test     // User-defined tests.
	Visitors  []parse.NodeVisitor // User-defined node visitors.
}

// NewEnv creates a new Env and returns it, ready to use.
// Passing nil in as the loader will make the Env use a
// StringLoader.
func NewEnv(loader Loader) *Env {
	if loader == nil {
		loader = &StringLoader{}
	}
	return &Env{loader, make(map[string]Func), make(map[string]Filter), make(map[string]Test), make([]parse.NodeVisitor, 0)}
}

// Execute parses and executes the given template.
func (env *Env) Execute(tmpl string, out io.Writer, ctx map[string]Value) error {
	return execute(tmpl, out, ctx, env)
}
