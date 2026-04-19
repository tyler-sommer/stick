// Package tests provides built-in tests for Twig-compatibility, mirroring
// the Twig 3.x built-in tests at https://twig.symfony.com/doc/3.x/tests/ .
//
// The "defined" test is a special case: it must check whether a reference
// resolves in scope without evaluating it (so that `{% set x = null %}`
// still counts as defined). That timing-sensitive behavior is handled by
// an interceptor in the executor; the registration here exists so a
// lookup of Env.Tests["defined"] returns non-nil, but the registered
// function is a fallback that only fires if the interceptor is bypassed
// (e.g. someone builds a binary expression with a TestExpr("defined") at
// AST-construction time).
package tests // import "github.com/tyler-sommer/stick/twig/tests"

import (
	"reflect"

	"github.com/tyler-sommer/stick"
)

// TwigTests returns PHP Twig 3.x's built-in test set. A fresh map is
// returned on each call so callers can mutate it without affecting others.
func TwigTests() map[string]stick.Test {
	isNull := testNull
	return map[string]stick.Test{
		"defined":      testDefinedFallback, // intercepted in the executor; see exec.go
		"divisible by": testDivisibleBy,
		"empty":        testEmpty,
		"even":         testEven,
		"iterable":     testIterable,
		"null":         isNull,
		"none":         isNull, // alias
		"odd":          testOdd,
		"same as":      testSameAs,
	}
}

// testDefinedFallback is only reached when the executor's interceptor is
// bypassed (rare — only via a hand-built TestExpr at AST construction).
// Approximates the semantic with a nil check.
func testDefinedFallback(_ stick.Context, val stick.Value, _ ...stick.Value) bool {
	return val != nil
}

// testEmpty returns true for nil, empty strings, and zero-length slices,
// arrays, maps, and channels. Scalar non-nil values are never empty.
func testEmpty(_ stick.Context, val stick.Value, _ ...stick.Value) bool {
	if val == nil {
		return true
	}
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		return rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	}
	// Bools/numbers: only empty if nil; once here they're non-nil non-zero-length.
	return false
}

// testEven returns true for integer-valued numbers whose value is even.
// Non-numeric inputs are coerced via stick.CoerceNumber.
func testEven(_ stick.Context, val stick.Value, _ ...stick.Value) bool {
	n := int64(stick.CoerceNumber(val))
	return n%2 == 0
}

// testOdd returns true for integer-valued numbers whose value is odd.
func testOdd(_ stick.Context, val stick.Value, _ ...stick.Value) bool {
	n := int64(stick.CoerceNumber(val))
	return n%2 != 0
}

// testNull returns true for nil values only. `none` is a Twig alias.
func testNull(_ stick.Context, val stick.Value, _ ...stick.Value) bool {
	return val == nil
}

// testIterable returns true for slices, arrays, maps, and channels.
// Strings are NOT considered iterable (matches PHP Twig's `is iterable`,
// which checks Traversable + array, both of which exclude strings).
func testIterable(_ stick.Context, val stick.Value, _ ...stick.Value) bool {
	if val == nil {
		return false
	}
	switch reflect.ValueOf(val).Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		return true
	}
	return false
}

// testDivisibleBy takes a single argument n and returns true if
// int(val) % int(n) == 0. Returns false on arity mismatch or when n is 0.
func testDivisibleBy(_ stick.Context, val stick.Value, args ...stick.Value) bool {
	if len(args) != 1 {
		return false
	}
	d := int64(stick.CoerceNumber(args[0]))
	if d == 0 {
		return false
	}
	return int64(stick.CoerceNumber(val))%d == 0
}

// testSameAs performs a strict identity check: val and args[0] must have
// the same dynamic type and value. Matches PHP Twig's `is same as(x)`
// (=== in PHP). Falls back to reflect.DeepEqual for non-comparable types
// so `[]int{1,2} is same as([]int{1,2})` still behaves sensibly — PHP
// `===` would return false for non-identical arrays, but stick can't
// compare identity of values it received through `any`, and
// reflect.DeepEqual is the closest meaningful alternative.
func testSameAs(_ stick.Context, val stick.Value, args ...stick.Value) bool {
	if len(args) != 1 {
		return false
	}
	a, b := val, args[0]
	if a == nil || b == nil {
		return a == b
	}
	ta, tb := reflect.TypeOf(a), reflect.TypeOf(b)
	if ta != tb {
		return false
	}
	if ta.Comparable() {
		return a == b
	}
	return reflect.DeepEqual(a, b)
}
