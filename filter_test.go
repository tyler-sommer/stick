package stick

import "testing"

func TestFilters(t *testing.T) {
	newBatchFunc := func(in Value, args ...Value) func() Value {
		return func() Value {
			batched := filterBatch(nil, in, args...)
			res := ""
			Iterate(batched, func(k, v Value, l Loop) (bool, error) {
				Iterate(v, func(k, v Value, l Loop) (bool, error) {
					res += CoerceString(v) + "."
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
		actual   func() Value
		expected Value
	}{
		{"default nil", func() Value { return filterDefault(nil, nil, "person") }, "person"},
		{"default empty string", func() Value { return filterDefault(nil, "", "person") }, "person"},
		{"default not empty", func() Value { return filterDefault(nil, "user", "person") }, "user"},
		{"abs positive", func() Value { return filterAbs(nil, 5.1) }, 5.1},
		{"abs negative", func() Value { return filterAbs(nil, -42) }, 42.0 /* note: coerced to float */},
		{"abs invalid", func() Value { return filterAbs(nil, "invalid") }, 0.0},
		{"len string", func() Value { return filterLength(nil, "hello") }, 5},
		{"len nil", func() Value { return filterLength(nil, nil) }, 0},
		{"len slice", func() Value { return filterLength(nil, []string{"h", "e"}) }, 2},
		{"capitalize", func() Value { return filterCapitalize(nil, "word") }, "Word"},
		{"lower", func() Value { return filterLower(nil, "HELLO, WORLD!") }, "hello, world!"},
		{"title", func() Value { return filterTitle(nil, "hello, world!") }, "Hello, World!"},
		{"trim", func() Value { return filterTrim(nil, " Hello   ") }, "Hello"},
		{"upper", func() Value { return filterUpper(nil, "hello, world!") }, "HELLO, WORLD!"},
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
