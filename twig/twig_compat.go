// Package twig provides Twig 1.x compatible template parsing and executing.
//
// Note: This package is still in development.
//
// Twig is a powerful templating language that promotes separation of logic
// from the view. This package attempts to replicate the functionality of the
// Twig engine using the Go programming language.
//
// Twig provides an extensive list of built-in functions, filters, tests, and
// even auto-escaping. This package provides this functionality out of the box,
// aiming to fully support the Twig spec.
//
// A simple example might look like:
//
// 	env := twig.New(nil);	// A nil loader means stick will execute
// 					// the string passed into env.Execute.
//
// 	// Templates receive a map of string to any value.
// 	p := map[string]stick.Value{"name": "World"}
//
// 	// Substitute os.Stdout with any io.Writer.
// 	env.Execute("Hello, {{ name }}!", os.Stdout, p)
//
// Check the main package https://godoc.org/github.com/tyler-sommer/stick for
// more information on general functionality and usage.
package twig

import (
	"github.com/tyler-sommer/stick"
	"github.com/tyler-sommer/stick/parse"
	"github.com/tyler-sommer/stick/twig/filter"
)

// New creates a new, default Env that aims to be compatible with Twig.
// If nil is passed as loader, a StringLoader is used.
func New(loader stick.Loader) *stick.Env {
	if loader == nil {
		loader = &stick.StringLoader{}
	}
	env := &stick.Env{
		Loader:    loader,
		Functions: make(map[string]stick.Func),
		Filters:   filter.TwigFilters(),
		Tests:     make(map[string]stick.Test),
		Visitors:  make([]parse.NodeVisitor, 0),
	}
	env.Register(NewAutoEscapeExtension())
	return env
}
