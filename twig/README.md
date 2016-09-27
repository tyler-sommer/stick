Twig
====

[![Build Status](https://travis-ci.org/tyler-sommer/stick.svg?branch=master)](https://travis-ci.org/tyler-sommer/stick)
[![GoDoc](https://godoc.org/github.com/tyler-sommer/stick/twig?status.svg)](https://godoc.org/github.com/tyler-sommer/stick/twig)

Provides [Twig-compatibility](http://twig.sensiolabs.org/) for the stick
templating engine.


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


Installation
------------

The `twig` package is intended to be used as a library. The recommended
way to install the library is using `go get`.

```bash
go get -u github.com/tyler-sommer/stick/twig
```


Usage
-----

Execute a simple Twig template.

```go
package main

import (
	"github.com/tyler-sommer/stick/twig"
	"github.com/tyler-sommer/stick"
	"os"
)

func main() {
    env := twig.New(nil)
	env.Execute("Hello, {{ name }}!", os.Stdout, map[string]stick.Value{"name": "Tyler"})
}
```

See [godoc for more information](https://godoc.org/github.com/tyler-sommer/stick/twig).

