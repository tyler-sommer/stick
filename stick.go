package stick

import (
	"io"

	"github.com/tyler-sommer/stick/parse"
)

// A Func represents a user-defined function.
// Functions can be called anywhere expressions are allowed.
//	{% if form_valid(form) %}
// Functions may take any number of arguments.
type Func func(e *Env, args ...Value) Value

// A Filter is a user-defined filter.
// Filters receive a value and modify it in some way.
//	{{ post|raw }}
// Filters also accept parameters.
//	{{ balance|number_format(2) }}
type Filter func(e *Env, val Value, args ...Value) Value

// A Test represents a user-defined test.
// Tests are used to make some comparisons more expressive, for example:
//	{% if users is empty %}
// Tests also accept arguments.
//	{% if loop.index is divisible by(3) %}
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
