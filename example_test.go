package stick_test

import (
	"fmt"
	"os"

	"strconv"

	"bytes"

	"io/ioutil"

	"github.com/tyler-sommer/stick"
)

// An example of executing a template in the simplest possible manner.
func ExampleEnv_Execute() {
	env := stick.New(nil)

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
	env := stick.New(nil)
	env.Tests["empty"] = func(ctx stick.Context, val stick.Value, args ...stick.Value) bool {
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
	env := stick.New(nil)
	env.Tests["divisible by"] = func(ctx stick.Context, val stick.Value, args ...stick.Value) bool {
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
	env := stick.New(nil)
	env.Functions["get_post"] = func(ctx stick.Context, args ...stick.Value) stick.Value {
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
	env := stick.New(nil)
	env.Filters["raw"] = func(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
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
	env := stick.New(nil)
	env.Filters["number_format"] = func(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
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

func ExampleFunc_usingContext() {
	env := stick.New(&stick.MemoryLoader{
		Templates: map[string]string{
			"base.html.twig": `<!doctype html>
<html>
<head>
	<title>{% block title %}{% endblock %}</title>
</head>
<body>
{% block nav %}{% endblock %}
#3 base: {{ current_template() }}
{% block content %}{% endblock %}
</body>
</html>
`,
			"side.html.twig": `{% block nav %}#2 side: {{ current_template() }}{% endblock %}`,
			"child.html.twig": `{% extends 'base.html.twig' %}
{% use 'side.html.twig' %}
{% block title %}#1 child: {{ current_template() }}{% endblock %}
{% block content %}#4 child: {{ current_template() }}{% endblock %}`,
		},
	})
	buf := &bytes.Buffer{}
	env.Functions["current_template"] = func(ctx stick.Context, args ...stick.Value) stick.Value {
		// Reading persistent metadata
		v, _ := ctx.Meta().Get("current_template_calls")
		nc := stick.CoerceNumber(v)
		nc++

		fmt.Fprintf(buf, "#%.0f Current Template: %s\n", nc, ctx.Name())

		// Writing persistent metadata
		ctx.Meta().Set("current_template_calls", stick.CoerceString(nc))

		return nil
	}

	// Notice that we discard the actual output. We only care about what the
	// current_template function writes to buf.
	err := env.Execute(
		`child.html.twig`,
		ioutil.Discard,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(buf.String())
	// Output:
	// #1 Current Template: child.html.twig
	// #2 Current Template: side.html.twig
	// #3 Current Template: base.html.twig
	// #4 Current Template: child.html.twig
}
