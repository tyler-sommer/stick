package stick_test

import (
	"os"

	"fmt"

	"github.com/tyler-sommer/stick"
)

// ExampleEnv_Execute shows a simple example of executing a template.
func ExampleEnv_Execute() {
	env := stick.NewEnv(nil)

	params := map[string]stick.Value{"name": "World"}
	env.Execute(`Hello, {{ name }}!`, os.Stdout, params)
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

// ExampleCoerceBool demonstrates how a type can be coerced to a boolean.
// The struct in this example has the Boolean method implemented.
//
// 	func (e exampleType) Boolean() bool {
//    	return true
//    }
func ExampleCoerceBool() {
	v := exampleType{}
	fmt.Printf("%t", stick.CoerceBool(v))
	// Output: true
}

// ExampleCoerceBool2 demonstrates how various values are coerced to boolean.
func ExampleCoerceBool2() {
	v0 := ""
	v1 := "some string"
	v2 := 0
	v3 := 3.14
	fmt.Printf("%t %t %t %t", stick.CoerceBool(v0), stick.CoerceBool(v1), stick.CoerceBool(v2), stick.CoerceBool(v3))
	// Output: false true false true
}

// ExampleCoerceNumber demonstrates how a type can be coerced to a number.
// The struct in this example has the Number method implemented.
//
// 	func (e exampleType) Number() float64 {
//    	return 3.14
//    }
func ExampleCoerceNumber() {
	v := exampleType{}
	fmt.Printf("%.2f", stick.CoerceNumber(v))
	// Output: 3.14
}

// ExampleCoerceNumber2 demonstrates how various values are coerced to number.
func ExampleCoerceNumber2() {
	v0 := true
	v1 := ""
	v2 := "54"
	v3 := "1.33"
	fmt.Printf("%.f %.f %.f %.2f", stick.CoerceNumber(v0), stick.CoerceNumber(v1), stick.CoerceNumber(v2), stick.CoerceNumber(v3))
	// Output: 1 0 54 1.33
}

// ExampleCoerceString demonstrates how a type can be coerced to a string.
// The struct in this example has the String method implemented.
//
// 	func (e exampleType) String() string {
//    	return "some kinda string"
//    }
func ExampleCoerceString() {
	v := exampleType{}
	fmt.Printf("%s", v)
	// Output: some kinda string
}

// ExampleCoerceString2 demonstrates how various values are coerced to string.
func ExampleCoerceString2() {
	v0 := true
	v1 := false // Coerces into ""
	v2 := 54
	v3 := 1.33
	v4 := 0
	fmt.Printf("%s '%s' %s %s %s", stick.CoerceString(v0), stick.CoerceString(v1), stick.CoerceString(v2), stick.CoerceString(v3), stick.CoerceString(v4))
	// Output: 1 '' 54 1.33 0
}
