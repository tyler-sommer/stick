/*
Package stick is a go-language port of the Twig templating engine.

Stick executes Twig templates and allows users to define custom Functions,
Filters, and Tests. The parser allows parse-time node inspection with
NodeVisitors, and a template Loader to load named templates from any source.

Twig compatibility

Stick itself is a parser and template executor. If you're looking for Twig
compatibility, check out package https://godoc.org/github.com/tyler-sommer/stick/twig

For additional information on Twig, check http://twig.sensiolabs.org/

Basic usage

Obligatory "Hello, World!" example:

	env := stick.New(nil);    // A nil loader means stick will simply execute
	                          // the string passed into env.Execute.

	// Templates receive a map of string to any value.
	p := map[string]stick.Value{"name": "World"}

	// Substitute os.Stdout with any io.Writer.
	env.Execute("Hello, {{ name }}!", os.Stdout, p)

Another example, using a FilesystemLoader and responding to an HTTP request:

	import "net/http"

	// ...

	fsRoot := os.Getwd() // Templates are loaded relative to this directory.
	env := stick.New(stick.NewFilesystemLoader(fsRoot))
	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		env.Execute("bar.html.twig", w, nil) // Loads "bar.html.twig" relative to fsRoot.
	})
	http.ListenAndServe(":80", nil)


Types and values

Any user value in Stick is represented by a stick.Value. There are three main types
in Stick when it comes to built-in operations: strings, numbers, and booleans. Of note,
numbers are represented by float64 as this matches regular Twig behavior most closely.

Stick makes no restriction on what is stored in a stick.Value, but some built-in
operators will try to coerce a value into a boolean, string, or number depending
on the operation.

Additionally, custom types that implement specific interfaces can be coerced. Stick
defines three interfaces: Stringer, Number, and Boolean. Each interface defines a single
method that should convert a custom type into the specified type.

	type myType struct {
		// ...
	}

	func (t *myType) String() string {
		return fmt.Sprintf("%v", t.someField)
	}

	func (t *myType) Number() float64 {
		return t.someFloatField
	}

	func (t *myType) Boolean() bool {
		return t.someValue != nil
	}

On a final note, there exists three functions to coerce any type into a string,
number, or boolean, respectively.

	// Coerce any value to a string
	v := stick.CoerceString(anything)

	// Coerce any value to a float64
	f := stick.CoerceNumber(anything)

	// Coerce any vale to a boolean
	b := stick.CoerceBool(anything)


User defined helpers

It is possible to define custom Filters, Functions, and boolean Tests available to
your Stick templates. Each user-defined type is simply a function with a specific
signature.

A Func represents a user-defined function.

	type Func func(e *Env, args ...Value) Value

Functions can be called anywhere expressions are allowed. Functions may take any number
of arguments.

A Filter is a user-defined filter.

	type Filter func(e *Env, val Value, args ...Value) Value

Filters receive a value and modify it in some way. Filters also accept zero or more arguments
beyond the value to be filtered.

A Test represents a user-defined boolean test.

	type Test func(e *Env, val Value, args ...Value) bool

Tests are used to make some comparisons more expressive. Tests also accept zero to any
number of arguments, and Test names can contain up to one space.

User-defined types are added to an Env after it is created. For example:

	env := stick.New(nil)
	env.Functions["form_valid"] = func(e *stick.Env, args ...stick.Value) stick.Value {
		// Do something useful..
		return true
	}
	env.Filters["number_format"] = func(e *stick.Env, val stick.Value, args ...stick.Value) stick.Value {
		v := stick.CoerceNumber(val)
		// Do some formatting.
		return fmt.Sprintf("%.2d", v)
	}
	env.Tests["empty"] = func(e *stick.Env, val stick.Value, args ...stick.Value) bool {
		// Probably not that useful.
		return stick.CoerceBool(val) == false
	}

*/
package stick
