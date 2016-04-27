package stick_test

import (
	"os"

	"fmt"

	"path/filepath"

	"github.com/tyler-sommer/stick"
)

// ExampleEnv_Execute_fs shows an example of using the provided
// FilesystemLoader.
//
// This example makes use of templates in the testdata folder. In
// particular, this example shows vertical (via extends) and horizontal
// reuse (via use).
func ExampleEnv_Execute_fs() {
	d, _ := os.Getwd()
	env := stick.NewEnv(stick.NewFilesystemLoader(filepath.Join(d, "testdata")))

	params := map[string]stick.Value{"name": "World"}
	err := env.Execute("main.txt.twig", os.Stdout, params)
	if err != nil {
		fmt.Println(err)
	}
	// Output:
	// This is a document.
	//
	// Hello
	//
	// An introduction to the topic.
	//
	// The body of this topic.
	//
	// Another section
	//
	// Some extra information.
	//
	// Still nobody knows.
	//
	// Some kind of footer.
}
