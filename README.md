stick
=====

A Go language port of the [Twig](http://twig.sensiolabs.org/) templating engine. 


Introduction
------------

Twig is a powerful templating language that supports macros, vertical and 
horizontal reuse including multiple inheritance, and an easy-to-learn syntax that
promotes separation of logic and markup.


Current status
--------------

Stick is made up of three main parts: a lexer, a parser, and an executer.

Stick's lexer and parser are nearly complete.

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
    - [ ] for loop
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
- [ ] Template loading
- [ ] Inheritance
    - [ ] extends
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
