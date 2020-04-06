package twig

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/shane-exley/stick/v2"
)

func Test_Escaping(t *testing.T) {
	var buf bytes.Buffer
	var env = New(nil)

	var tests = []struct {
		htmlString, expString string
		values                map[string]stick.Value
	}{
		{ // escape a value directly injected into the content
			htmlString: "<html>{{ '<script>bad stuff</script>' }}",
			expString:  "<html>&lt;script&gt;bad stuff&lt;/script&gt;",
			values:     map[string]stick.Value{},
		},
		{ // escape a value indirectly injected into the content
			htmlString: "<html><p>{{ bad }}</p>",
			expString:  "<html><p>&lt;script&gt;bad stuff&lt;/script&gt;</p>",
			values: map[string]stick.Value{
				"bad": "<script>bad stuff</script>",
			},
		},
		{ // ensure that we we dont provide a value it doesnt break the engine
			htmlString: "<html>{{ bad }}",
			expString:  "<html>",
			values:     map[string]stick.Value{},
		},
		{ // unescape a value indirectly injected into the content
			htmlString: "<html>{{ bad|raw }}",
			expString:  "<html><script>bad stuff</script>",
			values: map[string]stick.Value{
				"bad": "<script>bad stuff</script>",
			},
		},
		{ // same, but not everything
			htmlString: "<html>{{ bad|raw }}{{ bad }}",
			expString:  "<html><script>bad stuff</script>&lt;script&gt;bad stuff&lt;/script&gt;",
			values: map[string]stick.Value{
				"bad": "<script>bad stuff</script>",
			},
		},
		{ // nested vars
			htmlString: "{{ test.bad }}",
			expString:  "&lt;script&gt;test bad stuff&lt;/script&gt;",
			values: map[string]stick.Value{
				"test": map[string]stick.Value{
					"bad": "<script>test bad stuff</script>",
				},
			},
		},
		{ // nested vars
			htmlString: "{{ test.bad }}{{ test1.bad }}",
			expString:  "&lt;script&gt;test bad stuff&lt;/script&gt;",
			values: map[string]stick.Value{
				"test": map[string]stick.Value{
					"bad": "<script>test bad stuff</script>",
				},
			},
		},
	}

	for k, test := range tests {
		t.Run(fmt.Sprintf("#%d", k), func(t *testing.T) {
			env.Execute(test.htmlString, io.Writer(&buf), test.values)

			if buf.String() != test.expString {
				t.Errorf("Escaping Error: got %s which is not correctly escaped", buf.String())
			}

			buf.Reset()
		})
	}
}
