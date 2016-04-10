/*
Package stick is a go-language port of the Twig templating engine.

Twig is a powerful templating language that promotes separation of logic
from the view.

Stick executes Twig templates using an instance of Env. An Env contains all
the configured Functions, Filters, and Tests as well as a Loader to load
named templates from any source.

Obligatory "Hello, World!" example:

	env := stick.NewEnv(nil); // A nil loader means stick will simply execute
	                          // the string passed into env.Execute.
	// Templates receive a map of string to any value.
	p := map[string]stick.Value{"name": "World"}
	err := env.Execute("Hello, {{ name }}!", os.Stdout, )
	if err != nil { panic(err) }

In the previous example, notice that we passed in os.Stdout. Any io.Writer can be used.

Another example, using a FilesystemLoader and responding to an HTTP request:

	import "net/http"

	// ...

	fsRoot := os.Getwd() // Templates are loaded relative to this directory.
	env := stick.NewEnv(stick.NewFilesystemLoader(fsRoot))
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


Functions, filters, and tests

It is possible to define custom Filters, Functions, and boolean Tests available to
your Stick templates. Each user-defined type is simply a function with a specific
signature.

A Func represents a user-defined function.

	type Func func(e *Env, args ...Value) Value

Functions can be called anywhere expressions are allowed. Functions may take any number
of arguments.

	{% if form_valid(form) %}

A Filter is a user-defined filter.

	type Filter func(e *Env, val Value, args ...Value) Value

Filters receive a value and modify it in some way. Example of using a filter:

	{{ post|raw }}

Filters also accept zero or more arguments beyond the value to be filtered:

	{{ balance|number_format(2) }}

A Test represents a user-defined boolean test.

	type Test func(e *Env, val Value, args ...Value) bool

Tests are used to make some comparisons more expressive, for example:

	{% if users is empty %}

Tests also accept zero to any number of arguments, and Test names can contain
up to one space. Here, "divisible by" is an example of a two-word test that takes
a parameter:

	{% if loop.index is divisible by(3) %}

User-defined types are added to an Env after it is created. For example:

	env := stick.NewEnv(nil)
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

For additional information on Twig, check http://twig.sensiolabs.org/
*/
package stick

// BUG(ts): Missing documentation on operators, tests, functions, filters, and much more.
