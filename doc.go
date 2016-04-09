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
	err := env.Execute("Hello, {{ name }}!", os.Stdout, map[string]stick.Value{"name": "World"})
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

Stick makes no restriction on what is stored in a stick.Value, but some built-in
operators will try to coerce a value into a boolean, string, or number depending
on the operation.

For additional information on Twig, check http://twig.sensiolabs.org/
*/
package stick

// BUG(ts): Missing documentation on operators, tests, functions, filters, and much more.
