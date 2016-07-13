package stick

import (
	"math"
	"strings"
	"unicode/utf8"
)

// builtInFilters returns a map containing all built-in Twig filters,
// with the exception of "escape", which is provided by the AutoEscapeExtension.
func builtInFilters() map[string]Filter {
	return map[string]Filter{
		"abs":              filterAbs,
		"default":          filterDefault,
		"batch":            filterBatch,
		"capitalize":       filterCapitalize,
		"convert_encoding": filterConvertEncoding,
		"date":             filterDate,
		"date_modify":      filterDateModify,
		"first":            filterFirst,
		"format":           filterFormat,
		"join":             filterJoin,
		"json_encode":      filterJSONEncode,
		"keys":             filterKeys,
		"last":             filterLast,
		"length":           filterLength,
		"lower":            filterLower,
		"merge":            filterMerge,
		"nl2br":            filterNL2BR,
		"number_format":    filterNumberFormat,
		"raw":              filterRaw,
		"replace":          filterReplace,
		"reverse":          filterReverse,
		"round":            filterRound,
		"slice":            filterSlice,
		"sort":             filterSort,
		"split":            filterSplit,
		"striptags":        filterStripTags,
		"title":            filterTitle,
		"trim":             filterTrim,
		"upper":            filterUpper,
		"url_encode":       filterURLEncode,
	}
}

// filterAbs takes no arguments and returns the absolute value of val.
// Value val will be coerced into a number.
func filterAbs(ctx Context, val Value, args ...Value) Value {
	n := CoerceNumber(val)
	if 0 == n {
		return n
	}
	return math.Abs(n)
}

// filterBatch takes 2 arguments and returns a batched version of val.
// Value val must be a map, slice, or array. The filter has two optional arguments: number
// of items per batch (defaults to 1), and the default fill value. If the
// fill value is not specified, the last group of batched values may be smaller than
// the number specified as items per batch.
func filterBatch(ctx Context, val Value, args ...Value) Value {
	perSlice := 1
	var blankValue Value
	if l := len(args); l >= 1 {
		perSlice = int(CoerceNumber(args[0]))
		if l >= 2 {
			blankValue = args[1]
		}
	}
	if !IsIterable(val) {
		// TODO: This would trigger an E_WARNING in PHP.
		return nil
	}
	if perSlice <= 1 {
		// TODO: This would trigger an E_WARNING in PHP.
		return nil
	}
	l, _ := Len(val)
	numSlices := int(math.Ceil(float64(l) / float64(perSlice)))
	out := make([][]Value, numSlices)
	curr := []Value{}
	i := 0
	j := 0
	_, err := Iterate(val, func(k, v Value, l Loop) (bool, error) {
		// Use a variable length slice and append(). This maintains
		// correct compatibility with Twig when the fill value is nil.
		curr = append(curr, v)
		j++
		if j == perSlice {
			out[i] = curr
			curr = []Value{}
			i++
			j = 0
		}
		return false, nil
	})
	if err != nil {
		// TODO: Report error
		return nil
	}
	if i != numSlices {
		for ; blankValue != nil && j < perSlice; j++ {
			curr = append(curr, blankValue)
		}
		out[i] = curr
	}
	return out
}

// filterCapitalize takes no arguments and returns val with the first
// character capitalized.
func filterCapitalize(ctx Context, val Value, args ...Value) Value {
	s := CoerceString(val)
	return strings.ToUpper(s[:1]) + s[1:]
}

func filterConvertEncoding(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterDate(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterDateModify(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

// filterDefault takes one argument, the default value. If val is empty,
// the default value will be returned.
func filterDefault(ctx Context, val Value, args ...Value) Value {
	var d Value
	if len(args) > 0 {
		d = args[0]
	}
	if CoerceString(val) == "" {
		return d
	}
	return val
}

func filterFirst(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterFormat(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterJoin(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterJSONEncode(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterKeys(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterLast(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

// filterLength returns the length of val.
func filterLength(ctx Context, val Value, args ...Value) Value {
	if v, ok := val.(string); ok {
		return utf8.RuneCountInString(v)
	}
	l, _ := Len(val)
	// TODO: Report error
	return l
}

// filterLower returns val transformed to lower-case.
func filterLower(ctx Context, val Value, args ...Value) Value {
	return strings.ToLower(CoerceString(val))
}

func filterMerge(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterNL2BR(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterNumberFormat(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterRaw(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterReplace(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterReverse(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterRound(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterSlice(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterSort(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterSplit(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterStripTags(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

// filterTitle returns val with the first character of each word capitalized.
func filterTitle(ctx Context, val Value, args ...Value) Value {
	return strings.Title(CoerceString(val))
}

// filterTrim returns val with whitespace trimmed on both left and ride sides.
func filterTrim(ctx Context, val Value, args ...Value) Value {
	return strings.TrimSpace(CoerceString(val))
}

// filterUpper returns val in upper-case.
func filterUpper(ctx Context, val Value, args ...Value) Value {
	return strings.ToUpper(CoerceString(val))
}

func filterURLEncode(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}