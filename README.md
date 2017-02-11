Stick
=====

[![Build Status](https://travis-ci.org/tyler-sommer/stick.svg?branch=master)](https://travis-ci.org/tyler-sommer/stick)
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

##### In development

Stick itself is mostly feature-complete, with the exception of
whitespace control, and better error handling in places.

Stick is made up of three main parts: a lexer, a parser, and a template
executor. Stick's lexer and parser are complete. Template execution is
under development, but essentially complete.

See the [to do list](#to-do) for additional information.


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
	"github.com/tyler-sommer/stick"
	"os"
)

func main() {
	env := stick.New(nil)
	env.Execute("Hello, {{ name }}!", os.Stdout, map[string]stick.Value{"name": "Tyler"})
}
```

See [godoc for more information](https://godoc.org/github.com/tyler-sommer/stick).


To do
-----

- [x] Autoescaping (see: https://github.com/tyler-sommer/stick/twig)
- [ ] Whitespace control
- [ ] Improve error reporting

##### Further
- [ ] Improve test coverage (especially error cases)
- [ ] Custom operators and tags
- [ ] Sandbox
- [ ] Generate [native Go code from a given parser tree](https://github.com/tyler-sommer/go-stickgen)
