// Package stick is a go-language port of the Twig templating engine.
package stick

import (
	"io"
)

type Function func(e *Env, args ...Value) Value

type Filter Function

type Env struct {
	loader    Loader
	functions map[string]Function
	filters   map[string]Filter
}

type Extension interface {
	Functions() map[string]Function
	Filters() map[string]Filter
}

func NewEnv(loader Loader) *Env {
	if loader == nil {
		loader = &StringLoader{}
	}

	return &Env{loader, make(map[string]Function), make(map[string]Filter)}
}

func (env *Env) RegisterExtension(ext Extension) {
	for name, fn := range ext.Functions() {
		env.SetFunction(name, fn)
	}

	for name, fn := range ext.Filters() {
		env.SetFilter(name, fn)
	}
}

func (env *Env) SetFunction(name string, fn Function) {
	env.functions[name] = fn
}

func (env *Env) SetFilter(name string, ft Filter) {
	env.filters[name] = ft
}

func (env *Env) Execute(tmpl string, out io.Writer, ctx map[string]Value) error {
	return execute(tmpl, out, ctx, env)
}
