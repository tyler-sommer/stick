package stick_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tyler-sommer/stick"
)

// An example of macro definition and usage.
//
// This example uses a macro to list the values, also showing two
// ways to import macros. Check the templates in the testdata folder
// for more information.
func ExampleEnv_Execute_macro() {
	d, _ := os.Getwd()
	env := stick.New(stick.NewFilesystemLoader(filepath.Join(d, "testdata")))

	params := map[string]stick.Value{
		"title_first": "Hello",
		"value_first": []struct{ Key, Value string }{
			{"item1", "something about item1"},
			{"item2", "something about item2"},
		},
		"title_second": "Responses",
		"value_second": []struct{ Key, Value string }{
			{"please", "no, thank you"},
			{"why not", "cause"},
		},
	}
	err := env.Execute("other.txt.twig", os.Stdout, params)
	if err != nil {
		fmt.Println(err)
	}
	// Output:
	// Hello
	//
	// * item1: something about item1 (0)
	//
	// * item2: something about item2 (1)
	//
	// Responses
	//
	// * please: no, thank you (0)
	//
	// * why not: cause (1)
}
