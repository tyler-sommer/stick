package stick_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shane-exley/stick/v2"
)

// An example of macro definition and usage.
//
// This example uses a macro to list the values, also showing two
// ways to import macros. Check the templates in the testdata folder
// for more information.
func ExampleEnv_Execute_parent() {
	d, _ := os.Getwd()
	env := stick.New(stick.NewFilesystemLoader(filepath.Join(d, "testdata")))

	err := env.Execute("parent.txt.twig", os.Stdout, nil)
	if err != nil {
		fmt.Println(err)
	}
	// Output:
	// This is a document.
	//
	// Not A title
	//
	// Testing parent()
	//
	// This is a test
	//
	// Another section
	//
	// Some extra information.
}
