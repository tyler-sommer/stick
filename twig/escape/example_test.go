package escape_test

import (
	"fmt"

	"github.com/tyler-sommer/stick/twig/escape"
)

func ExampleHTML() {
	input := "Very <unsafe> \"string & stuff'"
	fmt.Print(escape.HTML(input))
	// Output:
	// Very &lt;unsafe&gt; &quot;string &amp; stuff&#39;
}

func ExampleHTMLAttribute() {
	input := "a bad\">\battribute<נש"
	fmt.Printf("<a href=\"%s\">A link</a>", escape.HTMLAttribute(input))
	// Output:
	// <a href="a&#32;bad&quot;&gt;&#xFFFD;attribute&lt;&#1504;&#1513;">A link</a>
}

func ExampleCSS() {
	input := "some \" bad content"
	fmt.Printf("div:after { content: \"%s\"; }", escape.CSS(input))
	// Output:
	// div:after { content: "some\0020\0022\0020bad\0020content"; }
}

func ExampleJS() {
	input := "some \"' bad javascript"
	fmt.Printf("var test = \"%s\";", escape.JS(input))
	// Output:
	// var test = "some\u0020\u0022\u0027\u0020bad\u0020javascript";
}

func ExampleURLQueryParam() {
	input := "מיין מאמעם"
	fmt.Printf("?who=%s", escape.URLQueryParam(input))
	// Output:
	// ?who=%D7%9E%D7%99%D7%99%D7%9F%20%D7%9E%D7%90%D7%9E%D7%A2%D7%9D
}
