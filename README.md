Stick
=====

A Go language port of the [Twig](http://twig.sensiolabs.org/) templating engine. 


Introduction
------------

Twig is a powerful templating language that supports macros, vertical and 
horizontal reuse including multiple inheritance, and an easy-to-learn syntax that
promotes separation of logic and markup.


### Current status

##### In development

Stick is made up of three main parts: a lexer, a parser, and an executer. Stick's lexer and parser are 
nearly complete. Basic template execution is implemented as well.

See the [to do list](#to-do) for additional information.


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
    - [ ] include
    - [ ] import
    - [ ] from
    - [ ] set
    - [ ] filter
    - [ ] embed
    - [ ] use
    - [ ] macros
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
    - [ ] embed
    - [ ] use
    - [ ] include
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
