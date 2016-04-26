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
//
//	// A simple function call:
//	// {% if form_valid(form) %}
//	env.Functions["form_valid"] = func(e *stick.Env, args ...stick.Value) stick.Value {
//		if len(args) == 0 {
//			return false
//		}
//		form := args[0]
//		// Do something useful...
//	}
type Func func(e *Env, args ...Value) Value

// A Filter is a user-defined filter.
// Filters receive a value and modify it in some way. Filters
// also accept parameters.
//
//	// A simple filter example:
//	// {{ post|raw }}
//	env.Filters["raw"] = func(e *stick.Env, val stick.Value, args ...stick.Value) stick.Value {
//		return stick.NewSafeValue(val)
//	}
//
// 	// A filter that accepts parameters:
//	// {{ balance|number_format(2) }}
//	env.Filters["number_format"] = func(e *stick.Env, val stick.Value, args ...stick.Value) stick.Value {
//		var d float64
//		if len(args) > 0 {
//			d = CoerceNumber(args[0])
//		}
//		return strconv.FormatFloat(CoerceNumber(val), 'f', d, 64)
//	}
type Filter func(e *Env, val Value, args ...Value) Value

// A Test represents a user-defined test.
// Tests are used to make some comparisons more expressive. Tests
// also accept arguments and can consist of two words.
//
//	// A simple test to check if a value is empty:
//	// {% if users is empty %}
//	env.Tests["empty"] = func(env *stick.Env, val stick.Value, args ...stick.Value) bool {
//		return stick.CoerceBool(val) == false
//	}
//
//	// A test consisting of two words and taking a parameter:
//	// {% if loop.index is divisible by(3) %}
//	env.Tests["divisible by"] = func(env *stick.Env, val stick.Value, args ...stick.Value) bool {
//		if len(args) != 1 {
//			return false
//		}
//		i := stick.CoerceNumber(args[0])
//		if i == 0 {
//			return false
//		}
//		v := stick.CoerceNumber(val)
//		return v / i == 0
//	}
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
