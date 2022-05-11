Stick
=====

[![CircleCI](https://circleci.com/gh/tyler-sommer/stick/tree/master.svg?style=shield)](https://circleci.com/gh/tyler-sommer/stick/tree/master)
[![GoDoc](https://godoc.org/github.com/tyler-sommer/stick?status.svg)](https://godoc.org/github.com/tyler-sommer/stick)

A Go language port of the [Twig](http://twig.sensiolabs.org/) templating engine. 


Overview
--------

This project is split over two main parts.

Package
[`github.com/tyler-sommer/stick`](https://github.com/tyler-sommer/stick)
is a Twig template parser and executor. It provides the core
functionality and offers many of the same extension points as Twig like
functions, filters, node visitors, etc.

Package
[`github.com/tyler-sommer/stick/twig`](https://github.com/tyler-sommer/stick/tree/master/twig)
contains extensions to provide the most Twig-like experience for
template writers. It aims to feature the same functions, filters, etc.
to be closely Twig-compatible.

### Current status

##### Stable, mostly feature complete

Stick itself is mostly feature-complete, with the exception of
whitespace control, and better error handling in places.

Stick is made up of three main parts: a lexer, a parser, and a template
executor. Stick's lexer and parser are complete. Template execution is
under development, but essentially complete.

See the [to do list](#to-do) for additional information.

### Alternatives

These alternatives are worth checking out if you're considering using Stick.

- [`text/template`](https://pkg.go.dev/text/template) and [`html/template`](https://pkg.go.dev/html/template) from the Go standard library.
- [`pongo2`](https://pkg.go.dev/github.com/flosch/pongo2/v5) is a full-featured Go language port of Django's templating language.


Installation
------------

Stick is intended to be used as a library. The recommended way to install the library is using `go get`.

```bash
go get -u github.com/tyler-sommer/stick
```


Usage
-----

Execute a simple Stick template.

```go
package main

import (
	"log"
	"os"
    
	"github.com/tyler-sommer/stick"
)

func main() {
	env := stick.New(nil)
	if err := env.Execute("Hello, {{ name }}!", os.Stdout, map[string]stick.Value{"name": "Tyler"}); err != nil {
		log.Fatal(err)
	}
}
```

See [godoc for more information](https://pkg.go.dev/github.com/tyler-sommer/stick).


To do
-----

- [x] Autoescaping (see: [Twig compatibility](https://github.com/tyler-sommer/stick/blob/master/twig))
- [ ] Whitespace control
- [ ] Improve error reporting

##### Further
- [ ] Improve test coverage (especially error cases)
- [ ] Custom operators and tags
- [ ] Sandbox
- [ ] Generate [native Go code from a given parser tree](https://github.com/tyler-sommer/go-stickgen)
