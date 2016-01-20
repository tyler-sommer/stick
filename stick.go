// Package stick is a go-language port of the Twig templating engine.
package stick

import (
	"io"
)

// A Function represents a user-defined function.
type Function func(e *Env, args ...Value) Value

// A Filter represents a user-defined filter.
type Filter Function

// Env represents the configuration of a Stick environment.
type Env struct {
	loader    Loader
	functions map[string]Function
	filters   map[string]Filter
}

// Extension defines the methods a Stick extension must implement.
type Extension interface {
	Functions() map[string]Function
	Filters() map[string]Filter
}

// NewEnv creates a new Env and returns it, ready to use.
func NewEnv(loader Loader) *Env {
	if loader == nil {
		loader = &StringLoader{}
	}

	return &Env{loader, make(map[string]Function), make(map[string]Filter)}
}

// RegisterExtension registers the given extension, adding any defined functions
// or filters to the Env.
func (env *Env) RegisterExtension(ext Extension) {
	for name, fn := range ext.Functions() {
		env.SetFunction(name, fn)
	}

	for name, fn := range ext.Filters() {
		env.SetFilter(name, fn)
	}
}

// SetFunction registers a user-defined function with the Env.
func (env *Env) SetFunction(name string, fn Function) {
	env.functions[name] = fn
}

// SetFilter registers a user-defined filter with the Env.
func (env *Env) SetFilter(name string, ft Filter) {
	env.filters[name] = ft
}

// Execute parses and executes the given template.
func (env *Env) Execute(tmpl string, out io.Writer, ctx map[string]Value) error {
	return execute(tmpl, out, ctx, env)
}
