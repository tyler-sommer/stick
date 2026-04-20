// Package fn provides built-in functions for Twig-compatibility, mirroring
// the Twig 3.x built-in functions at https://twig.symfony.com/doc/3.x/functions/ .
package fn // import "github.com/tyler-sommer/stick/twig/fn"

import (
	"math/rand"

	"github.com/tyler-sommer/stick"
)

// TwigFunctions returns PHP Twig 3.x's built-in function set. A fresh map
// is returned on each call so callers can mutate it without affecting
// others.
func TwigFunctions() map[string]stick.Func {
	return map[string]stick.Func{
		"attribute": fnAttribute,
		"cycle":     fnCycle,
		"max":       fnMax,
		"min":       fnMin,
		"random":    fnRandom,
		"range":     fnRange,
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

// fnRange returns a list containing an arithmetic progression between two
// endpoints. Mirrors PHP-Twig's range():
//
//	range(0, 3)            → [0, 1, 2, 3]
//	range(0, 6, 2)         → [0, 2, 4, 6]
//	range(10, 1)           → [10, 9, …, 1]   (step defaults to -1 when low > high)
//	range('a', 'c')        → ["a", "b", "c"]
//
// A step of 0 (or one whose sign disagrees with high-low) is normalised
// to ±1 to avoid infinite loops.
func fnRange(_ stick.Context, args ...stick.Value) stick.Value {
	if len(args) < 2 {
		return []stick.Value{}
	}
	if lo, ok := args[0].(string); ok {
		if hi, ok2 := args[1].(string); ok2 {
			loRunes, hiRunes := []rune(lo), []rune(hi)
			if len(loRunes) == 1 && len(hiRunes) == 1 {
				return runeRange(loRunes[0], hiRunes[0], rangeStep(args, 0))
			}
		}
	}
	return numRange(stick.CoerceNumber(args[0]), stick.CoerceNumber(args[1]), rangeStep(args, 0))
}

// rangeStep extracts an explicit step argument or returns 0 (caller
// normalises by direction).
func rangeStep(args []stick.Value, def float64) float64 {
	if len(args) < 3 {
		return def
	}
	return stick.CoerceNumber(args[2])
}

func numRange(lo, hi, step float64) stick.Value {
	if step == 0 {
		if lo <= hi {
			step = 1
		} else {
			step = -1
		}
	}
	// Reject misdirected steps (e.g. step=1 with lo>hi) by flipping
	// sign — matches PHP's silent flip.
	if (lo < hi && step < 0) || (lo > hi && step > 0) {
		step = -step
	}
	out := []stick.Value{}
	intMode := lo == float64(int64(lo)) && hi == float64(int64(hi)) && step == float64(int64(step))
	if step > 0 {
		for v := lo; v <= hi; v += step {
			if intMode {
				out = append(out, int(v))
			} else {
				out = append(out, v)
			}
		}
	} else {
		for v := lo; v >= hi; v += step {
			if intMode {
				out = append(out, int(v))
			} else {
				out = append(out, v)
			}
		}
	}
	return out
}

func runeRange(lo, hi rune, step float64) stick.Value {
	istep := int(step)
	if istep == 0 {
		if lo <= hi {
			istep = 1
		} else {
			istep = -1
		}
	}
	if (lo < hi && istep < 0) || (lo > hi && istep > 0) {
		istep = -istep
	}
	out := []stick.Value{}
	if istep > 0 {
		for r := lo; r <= hi; r += rune(istep) {
			out = append(out, string(r))
		}
	} else {
		for r := lo; r >= hi; r += rune(istep) {
			out = append(out, string(r))
		}
	}
	return out
}

// fnCycle returns the value at position % len(values) in the values list,
// wrapping. Mirrors PHP-Twig's cycle().
//
//	cycle(['a', 'b', 'c'], 4)   → 'b'    (4 mod 3 = 1)
//	cycle(['x', 'y'], -1)       → 'y'    (negative positions wrap)
//	cycle([], anything)         → nil
func fnCycle(_ stick.Context, args ...stick.Value) stick.Value {
	if len(args) < 2 || !stick.IsIterable(args[0]) {
		return nil
	}
	n, _ := stick.Len(args[0])
	if n == 0 {
		return nil
	}
	idx := int(stick.CoerceNumber(args[1])) % n
	if idx < 0 {
		idx += n
	}
	var out stick.Value
	_, _ = stick.Iterate(args[0], func(_, v stick.Value, l stick.Loop) (bool, error) {
		if l.Index0 == idx {
			out = v
			return true, nil
		}
		return false, nil
	})
	return out
}

// fnRandom mirrors PHP-Twig's random() per the docs:
//
//	random()           → a random non-negative int (math/rand)
//	random(N)          → a random int in [0, N] (inclusive)
//	random("abc")      → a random single character
//	random([x,y,z])    → a random element of the iterable
//
// Uses math/rand — Twig's random() is a templating helper, not a
// cryptographic primitive. Go 1.20+ auto-seeds the global source.
func fnRandom(_ stick.Context, args ...stick.Value) stick.Value {
	if len(args) == 0 {
		return rand.Int63()
	}
	v := args[0]
	switch x := v.(type) {
	case string:
		runes := []rune(x)
		if len(runes) == 0 {
			return ""
		}
		return string(runes[rand.Intn(len(runes))])
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		max := int64(stick.CoerceNumber(v))
		if max <= 0 {
			return int64(0)
		}
		return rand.Int63n(max + 1)
	}
	if stick.IsIterable(v) {
		n, _ := stick.Len(v)
		if n == 0 {
			return nil
		}
		idx := rand.Intn(n)
		var out stick.Value
		_, _ = stick.Iterate(v, func(_, vv stick.Value, l stick.Loop) (bool, error) {
			if l.Index0 == idx {
				out = vv
				return true, nil
			}
			return false, nil
		})
		return out
	}
	return v
}

// fnAttribute does a dynamic attribute / method lookup, matching
// PHP-Twig's attribute() function — useful when the attribute name is
// computed at template time:
//
//	{{ attribute(person, attr_name) }}
//	{{ attribute(person, "greet", ["Bob"]) }}
//
// Returns nil if the attribute can't be resolved (matches PHP-Twig's
// behaviour of returning null on missing attribute).
func fnAttribute(_ stick.Context, args ...stick.Value) stick.Value {
	if len(args) < 2 {
		return nil
	}
	var callArgs []stick.Value
	if len(args) >= 3 {
		if l, ok := args[2].([]stick.Value); ok {
			callArgs = l
		}
	}
	v, err := stick.GetAttr(args[0], args[1], callArgs...)
	if err != nil {
		return nil
	}
	return v
}
