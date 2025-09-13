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
		{"urlencode", func() stick.Value { return filterURLEncode(nil, "http://test.com/dude?sweet=33&1=2") }, "http%3A%2F%2Ftest.com%2Fdude%3Fsweet%3D33%261%3D2"},
		{"raw", func() stick.Value {
			safeVal, ok := filterRaw(nil, "<p>test</p>").(stick.SafeValue)
			if !ok {
				t.Errorf("Expected filterRaw to return a SafeValue")
			}
			return safeVal.Value()
		}, "<p>test</p>"},
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
