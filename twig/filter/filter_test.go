package filter

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tyler-sommer/stick"
)

func TestFilters(t *testing.T) {
	newBatchFunc := func(in stick.Value, args ...stick.Value) func() stick.Value {
		return func() stick.Value {
			batched := filterBatch(nil, in, args...)
			res := ""
			stick.Iterate(batched, func(k, v stick.Value, l stick.Loop) (bool, error) {
				stick.Iterate(v, func(k, v stick.Value, l stick.Loop) (bool, error) {
					res += stick.CoerceString(v) + "."
					return false, nil
				})
				res += "."
				return false, nil
			})
			return res
		}
	}

	tz, err := time.LoadLocation("Australia/Perth")
	if nil != err {
		t.Error(err)
	}
	testDate := time.Date(1980, 5, 31, 22, 01, 0, 0, tz)
	testDate2 := time.Date(2018, 2, 3, 2, 1, 44, 123456000, tz)

	tests := []struct {
		name     string
		actual   func() stick.Value
		expected stick.Value
	}{
		{"default nil", func() stick.Value { return filterDefault(nil, nil, "person") }, "person"},
		{"default empty string", func() stick.Value { return filterDefault(nil, "", "person") }, "person"},
		{"default not empty", func() stick.Value { return filterDefault(nil, "user", "person") }, "user"},
		{"abs positive", func() stick.Value { return filterAbs(nil, 5.1) }, 5.1},
		{"abs negative", func() stick.Value { return filterAbs(nil, -42) }, 42.0 /* note: coerced to float */},
		{"abs invalid", func() stick.Value { return filterAbs(nil, "invalid") }, 0.0},
		{"len string", func() stick.Value { return filterLength(nil, "hello") }, 5},
		{"len nil", func() stick.Value { return filterLength(nil, nil) }, 0},
		{"len slice", func() stick.Value { return filterLength(nil, []string{"h", "e"}) }, 2},
		{"capitalize", func() stick.Value { return filterCapitalize(nil, "word") }, "Word"},
		{"lower", func() stick.Value { return filterLower(nil, "HELLO, WORLD!") }, "hello, world!"},
		{"title", func() stick.Value { return filterTitle(nil, "hello, world!") }, "Hello, World!"},
		{"trim", func() stick.Value { return filterTrim(nil, " Hello   ") }, "Hello"},
		{"upper", func() stick.Value { return filterUpper(nil, "hello, world!") }, "HELLO, WORLD!"},
		{"batch underfull with fill", newBatchFunc([]int{1, 2, 3, 4, 5, 6, 7, 8}, 3, "No Item"), "1.2.3..4.5.6..7.8.No Item.."},
		{"batch underfull without fill", newBatchFunc([]int{1, 2, 3, 4, 5}, 3), "1.2.3..4.5.."},
		{"batch full", newBatchFunc([]int{1, 2, 3, 4}, 2), "1.2..3.4.."},
		{"batch empty", newBatchFunc([]int{}, 10), ""},
		{"batch nil", newBatchFunc(nil, 10), ""},
		{"first array", func() stick.Value { return filterFirst(nil, []string{"1", "2", "3", "4"}) }, "1"},
		{"first string", func() stick.Value { return filterFirst(nil, "1234") }, "1"},
		{"first string utf8", func() stick.Value { return filterFirst(nil, "東京") }, "東"},
		{"last array", func() stick.Value { return filterLast(nil, []string{"1", "2", "3", "4"}) }, "4"},
		{"last string", func() stick.Value { return filterLast(nil, "1234") }, "4"},
		{"last string utf8", func() stick.Value { return filterLast(nil, "東京") }, "京"},
		{"date c", func() stick.Value { return filterDate(nil, testDate, "c") }, "1980-05-31T22:01:00+08:00"},
		{"date r", func() stick.Value { return filterDate(nil, testDate, "r") }, "Sat, 31 May 1980 22:01:00 +0800"},
		{"date test", func() stick.Value { return filterDate(nil, testDate2, "d D j l F m M n Y y a A g G h H i s O P T") }, "03 Sat 3 Saturday February 02 Feb 2 2018 18 am AM 2 02 02 02 01 44 +0800 +08:00 AWST"},
		{"date u", func() stick.Value { return filterDate(nil, testDate2, "s.u") }, "44.123456"},
		{"date S", func() stick.Value { return filterDate(nil, testDate, "S") }, "st"},
		{"date S 2", func() stick.Value { return filterDate(nil, testDate2, "S") }, "rd"},
		{"date now", func() stick.Value { return filterDate(nil, "now", "Y-m-d") }, time.Now().Format("2006-01-02")},
		{"join", func() stick.Value { return filterJoin(nil, []string{"a", "b", "c"}, "-") }, "a-b-c"},
		{"join not a slice", func() stick.Value { return filterJoin(nil, "a", "-") }, "a"},
		{"round common down", func() stick.Value { return filterRound(nil, 3.4) }, 3.0},
		{"round common up", func() stick.Value { return filterRound(nil, 3.6) }, 4.0},
		{"round common half", func() stick.Value { return filterRound(nil, 3.5) }, 4.0},
		{"round common down 2 digits", func() stick.Value { return filterRound(nil, 3.114, 2) }, 3.11},
		{"round common up 2 digits", func() stick.Value { return filterRound(nil, 3.116, 2) }, 3.12},
		{"round common half 2 digits", func() stick.Value { return filterRound(nil, 3.115, 2) }, 3.12},
		{"round ceil", func() stick.Value { return filterRound(nil, 3.123, 0, "ceil") }, 4.0},
		{"round ceil 2 digits", func() stick.Value { return filterRound(nil, 3.123, 2, "ceil") }, 3.13},
		{"round floor", func() stick.Value { return filterRound(nil, 3.123, 0, "floor") }, 3.0},
		{"round floor 2 digits", func() stick.Value { return filterRound(nil, 3.123, 2, "floor") }, 3.12},
		{"reverse array", func() stick.Value { return stickSliceToString(filterReverse(nil, []string{"1", "2", "3", "4"})) }, "4.3.2.1"},
		{"reverse string", func() stick.Value { return filterReverse(nil, "1234") }, "4321"},
		{"reverse string utf8", func() stick.Value { return filterReverse(nil, "東京") }, "京東"},
		{"keys array", func() stick.Value { return stickSliceToString(filterKeys(nil, []string{"a", "b", "c"})) }, `0.1.2`},
		{"keys map", func() stick.Value {
			return stickSliceToString(filterKeys(nil, map[string]string{"a": "1", "b": "2", "c": "3"}))
		}, `a.b.c`},
		{"merge", func() stick.Value {
			return stickSliceToString(filterMerge(nil, []string{"a", "b"}, []string{"c", "d"}))
		}, "a.b.c.d"},
		{
			"replace",
			func() stick.Value {
				return filterReplace(nil, "I like %this% and %that%.", map[string]string{"%this%": "foo", "%that%": "bar"})
			},
			"I like foo and bar.",
		},
		{
			"json encode",
			func() stick.Value {
				return filterJSONEncode(nil, map[string]interface{}{"a": 1, "b": true, "c": 3.14, "d": "a string", "e": []string{"one", "two"}, "f": map[string]interface{}{"alpha": "foo", "beta": nil}})
			},
			`{"a":1,"b":true,"c":3.14,"d":"a string","e":["one","two"],"f":{"alpha":"foo","beta":null}}`,
		},
		{
			"merge array",
			func() stick.Value {
				return filterMerge(nil, []string{"test", "foo"}, []string{"baz"})
			},
			`[test foo baz]`,
		},
		{
			"merge object",
			func() stick.Value {
				return filterMerge(nil, map[string]stick.Value{"test": "wot"}, map[string]stick.Value{"foo": "bar"})
			},
			func(actual stick.Value) (ex string, ok bool) {
				ex = "map[foo:bar test:wot]"
				ok = false
				if v, ok := actual.(map[string]stick.Value); ok {
					// elaborate check is needed here because map order is not guaranteed; a simple string
					// comparison will not reliably pass.
					if len(v) == 2 && v["test"] == "wot" && v["foo"] == "bar" {
						return ex, true
					}
				}
				return
			},
		},
		// merge: hash + list — used to silently drop the list, returning the
		// hash unchanged. See PR fixing this.
		{"merge empty hash with list", func() stick.Value {
			return stickSliceToString(filterMerge(nil, map[string]stick.Value{}, []string{"a", "b"}))
		}, "a.b"},
		{"merge accumulator pattern (hash + list, repeated)", func() stick.Value {
			out := stick.Value(map[string]stick.Value{})
			out = filterMerge(nil, out, []string{"x"})
			out = filterMerge(nil, out, []string{"y"})
			out = filterMerge(nil, out, []string{"z"})
			return stickSliceToString(out)
		}, "x.y.z"},
		{
			"merge non-empty hash with list",
			func() stick.Value {
				return filterMerge(nil, map[string]stick.Value{"a": "1", "b": "2"}, []string{"3"})
			},
			func(actual stick.Value) (ex string, ok bool) {
				ex = "[1 2 3]"
				if v, isSlice := actual.([]stick.Value); isSlice && len(v) == 3 {
					seen := map[string]bool{}
					for _, e := range v {
						seen[stick.CoerceString(e)] = true
					}
					if seen["1"] && seen["2"] && seen["3"] {
						return ex, true
					}
				}
				return ex, false
			},
		},
		{"urlencode", func() stick.Value { return filterURLEncode(nil, "http://test.com/dude?sweet=33&1=2") }, "http%3A%2F%2Ftest.com%2Fdude%3Fsweet%3D33%261%3D2"},
		{"raw", func() stick.Value {
			safeVal, ok := filterRaw(nil, "<p>test</p>").(stick.SafeValue)
			if !ok {
				t.Errorf("Expected filterRaw to return a SafeValue")
			}
			return safeVal.Value()
		}, "<p>test</p>"},

		// split
		{"split no separator → per character", func() stick.Value {
			return stickSliceToString(filterSplit(nil, "a,c"))
		}, "a.,.c"},
		{"split with separator", func() stick.Value {
			return stickSliceToString(filterSplit(nil, "a,b,c", ","))
		}, "a.b.c"},
		{"split with limit", func() stick.Value {
			return stickSliceToString(filterSplit(nil, "a,b,c,d", ",", 2))
		}, "a.b,c,d"},
		{"split per character", func() stick.Value {
			return stickSliceToString(filterSplit(nil, "abc", ""))
		}, "a.b.c"},
		{"split per character utf8", func() stick.Value {
			return stickSliceToString(filterSplit(nil, "東京", ""))
		}, "東.京"},
		{"split negative limit", func() stick.Value {
			return stickSliceToString(filterSplit(nil, "a,b,c,d", ",", -1))
		}, "a.b.c"},

		// striptags
		{"striptags simple", func() stick.Value {
			return filterStripTags(nil, "<p>hello <b>world</b></p>")
		}, "hello world"},
		{"striptags multiline", func() stick.Value {
			return filterStripTags(nil, "<a\nhref='x'>link</a>")
		}, "link"},
		{"striptags none", func() stick.Value {
			return filterStripTags(nil, "no tags here")
		}, "no tags here"},

		// nl2br
		{"nl2br LF", func() stick.Value { return filterNL2BR(nil, "a\nb") }, "a<br />\nb"},
		{"nl2br CRLF", func() stick.Value { return filterNL2BR(nil, "a\r\nb") }, "a<br />\r\nb"},
		{"nl2br no newline", func() stick.Value { return filterNL2BR(nil, "abc") }, "abc"},

		// number_format
		{"number_format default", func() stick.Value { return filterNumberFormat(nil, 1234567) }, "1,234,567"},
		{"number_format two decimals", func() stick.Value { return filterNumberFormat(nil, 1234.5, 2) }, "1,234.50"},
		{"number_format custom separators", func() stick.Value {
			return filterNumberFormat(nil, 1234567.89, 2, ",", " ")
		}, "1 234 567,89"},
		{"number_format negative", func() stick.Value { return filterNumberFormat(nil, -9876.5, 1) }, "-9,876.5"},
		{"number_format small", func() stick.Value { return filterNumberFormat(nil, 5) }, "5"},

		// slice — array form
		{"slice array start", func() stick.Value {
			return stickSliceToString(filterSlice(nil, []string{"a", "b", "c", "d"}, 1))
		}, "b.c.d"},
		{"slice array start+length", func() stick.Value {
			return stickSliceToString(filterSlice(nil, []string{"a", "b", "c", "d"}, 1, 2))
		}, "b.c"},
		{"slice array negative start", func() stick.Value {
			return stickSliceToString(filterSlice(nil, []string{"a", "b", "c", "d"}, -2))
		}, "c.d"},
		{"slice array negative length", func() stick.Value {
			return stickSliceToString(filterSlice(nil, []string{"a", "b", "c", "d"}, 1, -1))
		}, "b.c"},
		{"slice array out of range", func() stick.Value {
			return stickSliceToString(filterSlice(nil, []string{"a", "b"}, 10, 5))
		}, ""},

		// slice — string form
		{"slice string", func() stick.Value { return filterSlice(nil, "abcdef", 1, 3) }, "bcd"},
		{"slice string utf8", func() stick.Value { return filterSlice(nil, "héllo", 1, 3) }, "éll"},
		{"slice string negative", func() stick.Value { return filterSlice(nil, "abcdef", -3) }, "def"},

		// sort
		{"sort strings", func() stick.Value {
			return stickSliceToString(filterSort(nil, []string{"b", "a", "c"}))
		}, "a.b.c"},
		{"sort numbers", func() stick.Value {
			return stickSliceToString(filterSort(nil, []int{3, 1, 2}))
		}, "1.2.3"},
		{"sort empty", func() stick.Value {
			return stickSliceToString(filterSort(nil, []string{}))
		}, ""},
		{"sort nil", func() stick.Value {
			return stickSliceToString(filterSort(nil, nil))
		}, ""},
	}
	for _, test := range tests {
		matches := false
		res := test.actual()
		expected := test.expected
		if fn, ok := expected.(func(actual stick.Value) (string, bool)); ok {
			if expected, ok = fn(res); ok {
				matches = true
			}
		} else {
			res = test.actual()
			if res != expected {
				if v := fmt.Sprintf("%v", res); v == expected {
					// the Go representation of the value matches expected
					matches = true
				}
			} else {
				matches = true
			}
		}
		if !matches {
			t.Errorf("%s:\n\texpected: %v\n\tgot: %v", test.name, expected, res)
		}
	}
}

func stickSliceToString(value stick.Value) (output string) {
	var slice []string
	stick.Iterate(value, func(k, v stick.Value, l stick.Loop) (bool, error) {
		slice = append(slice, stick.CoerceString(v))
		return false, nil
	})

	return strings.Join(slice, ".")
}
