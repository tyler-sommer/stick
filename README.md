Stick
=====

A Go language port of the [Twig](http://twig.sensiolabs.org/) templating engine. 


Introduction
------------

Twig is a powerful templating language that supports macros, vertical and 
horizontal reuse, and an easy-to-learn syntax that promotes separation of 
logic and markup. Twig is also extremely extensible, and by default will
autoescape content based on content type.


Stick is an attempt to bring these great features to Go projects.


### Current status

##### In development

Stick is made up of three main parts: a lexer, a parser, and an executer. Stick's lexer and parser are 
nearly complete. Basic template execution is implemented as well.

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


To do
-----

#### Lexer
- [x] Text
- [x] Tags
- [x] Print statements
- [ ] Comments
- [x] Expressions

#### Parser
- [x] Raw output
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
    - [ ] use
    - [ ] macro
    - [ ] do
- [x] Unary expressions
- [x] Binary expressions
- [ ] Ternary "if"
- [ ] Comments
- [x] Array and dot access
- [x] Function calls
- [x] Inline filter "expr|filter()"
- [ ] Method calls

#### Executer
- [x] Basic execution
- [x] Template loading
- [ ] Inheritance
    - [x] extends
    - [x] embed
    - [ ] use
    - [x] include
- [ ] Other basic tags
- [ ] Macros
- [ ] User defined functions
- [ ] User defined filters
- [ ] Autoescaping

##### Further
- [ ] Whitespace control
- [ ] Custom operators and tags
- [ ] Sandbox
- [ ] Generate native Go code from a given parser tree
