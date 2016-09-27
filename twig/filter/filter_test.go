package filter

import (
	"testing"

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
	}
	for _, test := range tests {
		res := test.actual()
		if res != test.expected {
			t.Errorf("%s:\n\texpected: %v\n\tgot: %v", test.name, test.expected, res)
		}
	}
}
