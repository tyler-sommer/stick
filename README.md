Stick
=====

[![Build Status](https://travis-ci.org/tyler-sommer/stick.svg?branch=master)](https://travis-ci.org/tyler-sommer/stick)
[![GoDoc](https://godoc.org/github.com/tyler-sommer/stick?status.svg)](https://godoc.org/github.com/tyler-sommer/stick)

A Go language port of the [Twig](http://twig.sensiolabs.org/) templating engine. 


Introduction
------------

Twig is a powerful templating language that supports macros, vertical and 
horizontal reuse, and an easy-to-learn syntax that promotes separation of 
logic and markup. Twig is also extremely extensible, and by default will
autoescape content based on content type.


**Stick brings these great features to Go projects.**


### Current status

##### In development

Stick is currently fairly usable except for a few important missing features: macros,
autoescaping, whitespace control, and proper error handling, unfortunately.

Stick is made up of three main parts: a lexer, a parser, and a template executer. Stick's lexer
is complete. Parsing and template execution is under development, but core functionality is complete.

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
    env := stick.NewEnv(nil)
	env.Execute("Hello, {{ name }}!", os.Stdout, map[string]stick.Value{"name": "Tyler"})
}
```

See [godoc for more information](https://godoc.org/github.com/tyler-sommer/stick).


To do
-----

#### Lexer
- [x] Text
- [x] Tags
- [x] Print statements
- [x] Comments
- [x] Expressions
- [x] String interpolation

#### Parser
- [x] Raw output
- [x] Comments
- [ ] Basic tag support
    - [x] if/else if/else/endif
    - [x] extends
    - [x] block
    - [x] for loop
    - [x] include
    - [ ] import
    - [ ] from
    - [ ] set
    - [ ] filter
    - [x] embed
    - [x] use
    - [ ] macro
    - [ ] do
- [ ] Expressions
    - [x] Unary expressions
    - [x] Binary expressions
    - [ ] Ternary "if"
    - [x] String interpolation
    - [x] Array and dot access
    - [x] Function calls
    - [x] Inline filter "expr|filter()"
    - [x] Method calls

#### Executer
- [x] Basic execution
- [x] Template loading
- [x] Inheritance
    - [x] extends
    - [x] embed
    - [x] use
    - [x] include
- [x] Expressions
    - [x] literals
    - [x] binary operators
    - [x] unary operators
    - [x] get attribute
    - [x] function call
    - [x] filter application
- [ ] Other basic tags
- [ ] Macros
- [x] User defined functions
- [x] User defined filters
- [ ] Autoescaping
- [ ] Whitespace control

##### Further
- [ ] Improve error reporting
- [ ] Improve test coverage (especially error cases)
- [ ] Custom operators and tags
- [ ] Sandbox
- [ ] Generate native Go code from a given parser tree
