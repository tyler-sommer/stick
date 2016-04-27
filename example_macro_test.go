package stick_test

import (
	"os"

	"fmt"

	"path/filepath"

	"github.com/tyler-sommer/stick"
)

// ExampleEnv_Execute_macro shows an example of macro definition and
// usage.
func ExampleEnv_Execute_macro() {
	d, _ := os.Getwd()
	env := stick.NewEnv(stick.NewFilesystemLoader(filepath.Join(d, "testdata")))

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
