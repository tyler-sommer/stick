// Package stick is a go-language port of the Twig templating engine.
package stick

import "io"

type Env struct {
	loader Loader
}

func NewEnv(loader Loader) *Env {
	if loader == nil {
		loader = &StringLoader{}
	}

	return &Env{loader}
}

func (env *Env) Execute(tmpl string, out io.Writer, ctx map[string]Value) error {
	in, err := env.loader.Load(tmpl)
	if err != nil {
		return err
	}

	return execute(in, out, ctx, env.loader)
}
