// Package escape provides Twig-compatible escape functions.
package escape

import (
	"bytes"
	"fmt"
)

// HTML provides a Twig-compatible HTML escape function.
func HTML(in string) string {
	var out = &bytes.Buffer{}
	for _, c := range in {
		if c == 34 {
			// "
			out.WriteString("&quot;")
		} else if c == 38 {
			// &
			out.WriteString("&amp;")
		} else if c == 39 {
			// '
			out.WriteString("&#39;")
		} else if c == 60 {
			// <
			out.WriteString("&lt;")
		} else if c == 62 {
			// >
			out.WriteString("&gt;")
		} else {
			// UTF-8
			out.WriteRune(c)
		}
	}
	return out.String()
}

// HTMLAttribute provides a Twig-compatible escaper for HTML attributes.
func HTMLAttribute(in string) string {
	var out = &bytes.Buffer{}
	for _, c := range in {
		if (c >= 65 && c <= 90) || (c >= 97 && c <= 122) || (c >= 48 && c <= 57) || (c >= 44 && c <= 46) || c == 95 {
			// a-zA-Z0-9,.-_
			out.WriteRune(c)
		} else if c == 34 {
			// "
			out.WriteString("&quot;")
		} else if c == 38 {
			// &
			out.WriteString("&amp;")
		} else if c == 60 {
			// <
			out.WriteString("&lt;")
		} else if c == 62 {
			// >
			out.WriteString("&gt;")
		} else if c <= 31 && c != 9 && c != 10 && c != 13 {
			// Non-whitespace
			out.WriteString("&#xFFFD;")
		} else {
			// UTF-8
			fmt.Fprintf(out, "&#%d;", c)
		}
	}
	return out.String()
}

// JS provides a Twig-compatible javascript escaper.
func JS(in string) string {
	var out = &bytes.Buffer{}
	for _, c := range in {
		if (c >= 65 && c <= 90) || (c >= 97 && c <= 122) || (c >= 48 && c <= 57) || c == 44 || c == 46 || c == 95 {
			// a-zA-Z0-9,._
			out.WriteRune(c)
		} else {
			// UTF-8
			fmt.Fprintf(out, "\\u%04X", c)
		}
	}
	return out.String()
}

// CSS provides a Twig-compatible CSS escaper.
func CSS(in string) string {
	var out = &bytes.Buffer{}
	for _, c := range in {
		if (c >= 65 && c <= 90) || (c >= 97 && c <= 122) || (c >= 48 && c <= 57) {
			// a-zA-Z0-9
			out.WriteRune(c)
		} else {
			// UTF-8
			fmt.Fprintf(out, "\\%04X", c)
		}
	}
	return out.String()
}

// URLQueryParam provides Twig-compatible query string escaper.
func URLQueryParam(in string) string {
	var out = &bytes.Buffer{}
	var c byte
	for i := 0; i < len(in); i++ {
		c = in[i]
		if (c >= 65 && c <= 90) || (c >= 97 && c <= 122) || (c >= 48 && c <= 57) || c == 45 || c == 46 || c == 126 || c == 95 {
			// a-zA-Z0-9-._~
			out.WriteByte(c)
		} else {
			// UTF-8
			fmt.Fprintf(out, "%%%02X", c)
		}
	}
	return out.String()
}
