package stick_test

import (
	"fmt"
	"os"

	"strconv"

	"github.com/tyler-sommer/stick"
)

// An example of executing a template in the simplest possible manner.
func ExampleEnv_Execute() {
	env := stick.NewEnv(nil)

	params := map[string]stick.Value{"name": "World"}
	err := env.Execute(`Hello, {{ name }}!`, os.Stdout, params)
	if err != nil {
		fmt.Println(err)
	}
	// Output: Hello, World!
}

type exampleType struct{}

func (e exampleType) Boolean() bool {
	return true
}

func (e exampleType) Number() float64 {
	return 3.14
}

func (e exampleType) String() string {
	return "some kinda string"
}

// This demonstrates how a type can be coerced to a boolean.
// The struct in this example has the Boolean method implemented.
//
//	func (e exampleType) Boolean() bool {
//		return true
//	}
func ExampleBoolean() {
	v := exampleType{}
	fmt.Printf("%t", stick.CoerceBool(v))
	// Output: true
}

// This example demonstrates how various values are coerced to boolean.
func ExampleCoerceBool() {
	v0 := ""
	v1 := "some string"
	v2 := 0
	v3 := 3.14
	fmt.Printf("%t %t %t %t", stick.CoerceBool(v0), stick.CoerceBool(v1), stick.CoerceBool(v2), stick.CoerceBool(v3))
	// Output: false true false true
}

// This demonstrates how a type can be coerced to a number.
// The struct in this example has the Number method implemented.
//
// 	func (e exampleType) Number() float64 {
//		return 3.14
//	}
func ExampleNumber() {
	v := exampleType{}
	fmt.Printf("%.2f", stick.CoerceNumber(v))
	// Output: 3.14
}

// This example demonstrates how various values are coerced to number.
func ExampleCoerceNumber() {
	v0 := true
	v1 := ""
	v2 := "54"
	v3 := "1.33"
	fmt.Printf("%.f %.f %.f %.2f", stick.CoerceNumber(v0), stick.CoerceNumber(v1), stick.CoerceNumber(v2), stick.CoerceNumber(v3))
	// Output: 1 0 54 1.33
}

// This example demonstrates how a type can be coerced to a string.
// The struct in this example has the String method implemented.
//
//	func (e exampleType) String() string {
//		return "some kinda string"
//	}
func ExampleStringer() {
	v := exampleType{}
	fmt.Printf("%s", v)
	// Output: some kinda string
}

// This demonstrates how various values are coerced to string.
func ExampleCoerceString() {
	v0 := true
	v1 := false // Coerces into ""
	v2 := 54
	v3 := 1.33
	v4 := 0
	fmt.Printf("%s '%s' %s %s %s", stick.CoerceString(v0), stick.CoerceString(v1), stick.CoerceString(v2), stick.CoerceString(v3), stick.CoerceString(v4))
	// Output: 1 '' 54 1.33 0
}

// A simple test to check if a value is empty
func ExampleTest() {
	env := stick.NewEnv(nil)
	env.Tests["empty"] = func(env *stick.Env, val stick.Value, args ...stick.Value) bool {
		return stick.CoerceBool(val) == false
	}

	err := env.Execute(
		`{{ (false is empty) ? 'empty' : 'not empty' }} - {{ ("a string" is empty) ? 'empty' : 'not empty' }}`,
		os.Stdout,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	// Output: empty - not empty
}

// A test made up of two words that takes an argument.
func ExampleTest_twoWordsWithArgs() {
	env := stick.NewEnv(nil)
	env.Tests["divisible by"] = func(env *stick.Env, val stick.Value, args ...stick.Value) bool {
		if len(args) != 1 {
			return false
		}
		i := stick.CoerceNumber(args[0])
		if i == 0 {
			return false
		}
		v := stick.CoerceNumber(val)
		return int(v)%int(i) == 0
	}

	err := env.Execute(
		`{{ ('something' is divisible by(3)) ? "yep, 'something' evals to 0" : 'nope'  }} - {{ (9 is divisible by(3)) ? 'sure' : 'nope' }} - {{ (4 is divisible by(3)) ? 'sure' : 'nope' }}`,
		os.Stdout,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	// Output: yep, 'something' evals to 0 - sure - nope
}

// A contrived example of a user-defined function.
func ExampleFunc() {
	env := stick.NewEnv(nil)
	env.Functions["get_post"] = func(e *stick.Env, args ...stick.Value) stick.Value {
		if len(args) == 0 {
			return nil
		}
		return struct {
			Title string
			ID    float64
		}{"A post", stick.CoerceNumber(args[0])}
	}

	err := env.Execute(
		`{% set post = get_post(123) %}{{ post.Title }} (# {{ post.ID }})`,
		os.Stdout,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	// Output: A post (# 123)
}

// A simple user-defined filter.
func ExampleFilter() {
	env := stick.NewEnv(nil)
	env.Filters["raw"] = func(e *stick.Env, val stick.Value, args ...stick.Value) stick.Value {
		return stick.NewSafeValue(val)
	}

	err := env.Execute(
		`{{ name|raw }}`,
		os.Stdout,
		map[string]stick.Value{"name": "<name>"},
	)
	if err != nil {
		fmt.Println(err)
	}
	// Output: <name>
}

// A simple user-defined filter that accepts a parameter.
func ExampleFilter_withParam() {
	env := stick.NewEnv(nil)
	env.Filters["number_format"] = func(e *stick.Env, val stick.Value, args ...stick.Value) stick.Value {
		var d float64
		if len(args) > 0 {
			d = stick.CoerceNumber(args[0])
		}
		return strconv.FormatFloat(stick.CoerceNumber(val), 'f', int(d), 64)
	}

	err := env.Execute(
		`${{ price|number_format(2) }}`,
		os.Stdout,
		map[string]stick.Value{"price": 4.99},
	)
	if err != nil {
		fmt.Println(err)
	}
	// Output: $4.99
}
