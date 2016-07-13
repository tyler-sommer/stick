package stick

import (
	"testing"
	"time"
)

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

	tz, err := time.LoadLocation("Australia/Perth")
	if nil != err {
		t.Error(err)
	}
	testDate := time.Date(1980, 5, 31, 22, 01, 0, 0, tz)
	testDate2 := time.Date(2018, 2, 3, 2, 1, 44, 123456000, tz)

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
		{"first array", func() Value {return filterFirst(nil, []string{"1","2","3","4"})}, "1"},
		{"first string", func() Value {return filterFirst(nil, "1234")}, "1"},

		{"date c", func() Value {return filterDate(nil, testDate, "c")}, "1980-05-31T22:01:00+08:00"},
		{"date r", func() Value {return filterDate(nil, testDate, "r")}, "Sat, 31 May 1980 22:01:00 +0800"},
		{"date test", func() Value {return filterDate(nil, testDate2, "d D j l F m M n Y y a A g G h H i s O P T")}, "03 Sat 3 Saturday February 02 Feb 2 2018 18 am AM 2 02 02 02 01 44 +0800 +08:00 AWST"},
		{"date u", func() Value {return filterDate(nil, testDate2, "s.u")}, "44.123456"},

	}
	for _, test := range tests {
		res := test.actual()
		if res != test.expected {
			t.Errorf("%s:\n\texpected: '%v'\n\tgot     : '%v' (%T)", test.name, test.expected, res, res)
		}
	}
}
