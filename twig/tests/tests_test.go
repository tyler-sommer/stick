package tests

import (
	"testing"

	"github.com/tyler-sommer/stick"
)

type testCase struct {
	name string
	val  stick.Value
	args []stick.Value
	want bool
}

func run(t *testing.T, fn stick.Test, cases []testCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := fn(nil, c.val, c.args...)
			if got != c.want {
				t.Errorf("val=%#v args=%#v: got %v, want %v", c.val, c.args, got, c.want)
			}
		})
	}
}

func TestEmpty(t *testing.T) {
	run(t, testEmpty, []testCase{
		{"nil", nil, nil, true},
		{"empty string", "", nil, true},
		{"empty slice", []int{}, nil, true},
		{"empty map", map[string]int{}, nil, true},
		{"nonempty string", "x", nil, false},
		{"nonempty slice", []int{1}, nil, false},
		{"int zero", 0, nil, false},
		{"bool false", false, nil, false},
	})
}

func TestEven(t *testing.T) {
	run(t, testEven, []testCase{
		{"zero", 0, nil, true},
		{"two", 2, nil, true},
		{"minus four", -4, nil, true},
		{"one", 1, nil, false},
		{"three", 3, nil, false},
	})
}

func TestOdd(t *testing.T) {
	run(t, testOdd, []testCase{
		{"one", 1, nil, true},
		{"minus three", -3, nil, true},
		{"zero", 0, nil, false},
		{"two", 2, nil, false},
	})
}

func TestNull(t *testing.T) {
	run(t, testNull, []testCase{
		{"nil", nil, nil, true},
		{"zero", 0, nil, false},
		{"empty string", "", nil, false},
		{"false", false, nil, false},
	})
}

func TestIterable(t *testing.T) {
	run(t, testIterable, []testCase{
		{"slice", []int{}, nil, true},
		{"map", map[string]int{}, nil, true},
		{"array", [2]int{1, 2}, nil, true},
		{"string (not iterable per PHP Twig)", "xyz", nil, false},
		{"int", 42, nil, false},
		{"nil", nil, nil, false},
	})
}

func TestDivisibleBy(t *testing.T) {
	mk := func(n int) []stick.Value { return []stick.Value{n} }
	run(t, testDivisibleBy, []testCase{
		{"9 by 3", 9, mk(3), true},
		{"10 by 3", 10, mk(3), false},
		{"0 by 5", 0, mk(5), true},
		{"division by zero", 4, mk(0), false},
		{"wrong arity", 9, nil, false},
	})
}

func TestSameAs(t *testing.T) {
	wrap := func(v stick.Value) []stick.Value { return []stick.Value{v} }
	run(t, testSameAs, []testCase{
		{"int to int same", 1, wrap(1), true},
		{"int to int different", 1, wrap(2), false},
		{"int to string (different type)", 1, wrap("1"), false},
		{"nil to nil", nil, wrap(nil), true},
		{"string same", "abc", wrap("abc"), true},
		{"slice same via DeepEqual", []int{1, 2}, wrap([]int{1, 2}), true},
		{"slice different", []int{1, 2}, wrap([]int{1, 3}), false},
		{"wrong arity", 1, nil, false},
	})
}

func TestTwigTestsContainsExpectedKeys(t *testing.T) {
	got := TwigTests()
	want := []string{
		"defined", "divisible by", "empty", "even", "iterable",
		"null", "none", "odd", "same as",
	}
	for _, k := range want {
		if _, ok := got[k]; !ok {
			t.Errorf("TwigTests missing %q", k)
		}
	}
}
