package stick

import (
	"math"
	"strings"
)

func BuiltInFilters() map[string]Filter {
	filters := map[string]Filter{
		"abs": FilterAbs,
		"default": FilterDefault,
		"batch": FilterBatch,
		"capitalize": FilterCapitalize,
		"convert_encoding": FilterConvertEncoding,
		"date": FilterDate,
		"date_modify": FilterDateModify,
		"escape": FilterEscape,
		"first": FilterFirst,
		"format": FilterFormat,
		"join": FilterJoin,
		"json_encode": FilterJsonEncode,
		"keys": FilterKeys,
		"last": FilterLast,
		"length": FilterLength,
		"lower": FilterLower,
		"merge": FilterMerge,
		"nl2br": FilterNl2Br,
		"number_format": FilterNumberFormat,
		"raw": FilterRaw,
		"replace": FilterReplace,
		"reverse": FilterReverse,
		"round": FilterRound,
		"slice": FilterSlice,
		"sort": FilterSort,
		"split": FilterSplit,
		"striptags": FilterStripTags,
		"title": FilterTitle,
		"trim": FilterTrim,
		"upper": FilterUpper,
		"url_encode": FilterUrlEncode,
	}
	return filters
}

func FilterAbs(ctx Context, val Value, args ...Value) Value {
	n := CoerceNumber(val)
	if 0 == n {
		return val
	}
	return math.Abs(n)
}

// Arg1: length, Arg2: Blank Fill Item
func FilterBatch(ctx Context, val Value, args ...Value) Value {
	if 2 != len(args) {
		// need 2 arguments
		return args
	}
	sl, ok := val.([]Value)
	if !ok {
		// not a slice of Values
		return val
	}

	numItemsPerSlice := int(CoerceNumber(args[0]))
	if 0 == numItemsPerSlice || 1 == numItemsPerSlice {
		return val
	}

	blankValue := args[1]

	numSlices := int(len(sl) / numItemsPerSlice)

	out := make([][]Value, numSlices)

	location := 0;
	for outter := 0; outter < numSlices; outter++ {
		for inner := 0; inner < numItemsPerSlice; inner++ {
			if location < len(sl) {
				out[outter][inner] = sl[location]
			} else {
				out[outter][inner] = blankValue
			}
			location++
		}
	}
	return out
}

func FilterCapitalize(ctx Context, val Value, args ...Value) Value {
	s := strings.ToLower(CoerceString(val))
	return strings.ToUpper(s[:1]) + s[1:]
}

func FilterConvertEncoding(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterDate(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterDateModify(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterDefault(ctx Context, val Value, args ...Value) Value {
	var d Value
	if len(args) == 0 {
		d = nil
	} else {
		d = args[0]
	}
	if CoerceString(val) == "" {
		return d
	}
	return val
}

func FilterEscape(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterFirst(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterFormat(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterJoin(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterJsonEncode(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterKeys(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterLast(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterLength(ctx Context, val Value, args ...Value) Value {
	return 0
}

func FilterLower(ctx Context, val Value, args ...Value) Value {
	return strings.ToLower(CoerceString(val))
}

func FilterMerge(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterNl2Br(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterNumberFormat(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterRaw(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterReplace(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterReverse(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterRound(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterSlice(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterSort(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterSplit(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterStripTags(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterTitle(ctx Context, val Value, args ...Value) Value {
	return strings.Title(CoerceString(val))
}

func FilterTrim(ctx Context, val Value, args ...Value) Value {
	return strings.TrimSpace(CoerceString(val))
}

func FilterUpper(ctx Context, val Value, args ...Value) Value {
	return strings.ToUpper(CoerceString(val))
}

func FilterUrlEncode(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}
