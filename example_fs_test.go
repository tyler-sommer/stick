package stick_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tyler-sommer/stick"
)

// An example showing the use of the provided FilesystemLoader.
//
// This example makes use of templates in the testdata folder. In
// particular, this example shows vertical (via extends) and horizontal
// reuse (via use).
func ExampleEnv_Execute_filesystemLoader() {
	d, _ := os.Getwd()
	env := stick.New(stick.NewFilesystemLoader(filepath.Join(d, "testdata")))

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
