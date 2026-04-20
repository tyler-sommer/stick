// Package fn provides built-in functions for Twig-compatibility, mirroring
// the Twig 3.x built-in functions at https://twig.symfony.com/doc/3.x/functions/ .
package fn // import "github.com/tyler-sommer/stick/twig/fn"

import "github.com/tyler-sommer/stick"

// TwigFunctions returns PHP Twig 3.x's built-in function set. A fresh map
// is returned on each call so callers can mutate it without affecting
// others.
func TwigFunctions() map[string]stick.Func {
	return map[string]stick.Func{
		"max": fnMax,
		"min": fnMin,
	}
}

// fnMax returns the largest of its arguments. Mirrors PHP-Twig's max():
//
//	max(1, 2, 3)          → 3
//	max([1, 3, 2])        → 3
//	max({'a':1,'b':3})    → 3 (hash values)
//	max()                 → nil
//
// Numeric comparison uses stick.CoerceNumber, matching PHP's loose-typing
// behavior. String comparison is not implemented here — both sides coerce
// to 0 and the first argument wins. Hashtagueule doesn't exercise that
// path; add it if a real theme needs it.
func fnMax(_ stick.Context, args ...stick.Value) stick.Value {
	return extremum(args, true)
}

// fnMin returns the smallest of its arguments. See fnMax for semantics.
func fnMin(_ stick.Context, args ...stick.Value) stick.Value {
	return extremum(args, false)
}

func extremum(args []stick.Value, max bool) stick.Value {
	var best stick.Value
	seen := false
	feed := func(v stick.Value) {
		if !seen {
			best, seen = v, true
			return
		}
		cur := stick.CoerceNumber(v)
		top := stick.CoerceNumber(best)
		if (max && cur > top) || (!max && cur < top) {
			best = v
		}
	}
	if len(args) == 1 && stick.IsIterable(args[0]) {
		_, _ = stick.Iterate(args[0], func(_, v stick.Value, _ stick.Loop) (bool, error) {
			feed(v)
			return false, nil
		})
	} else {
		for _, a := range args {
			feed(a)
		}
	}
	return best
}
