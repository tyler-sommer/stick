package stick

import (
	"math"
	"strings"
)

func builtInFilters() map[string]Filter {
	filters := map[string]Filter{
		"abs": filterAbs,
		"default": filterDefault,
		"batch": filterBatch,
		"capitalize": filterCapitalize,
		"convert_encoding": filterConvertEncoding,
		"date": filterDate,
		"date_modify": filterDateModify,
		"escape": filterEscape,
		"first": filterFirst,
		"format": filterFormat,
		"join": filterJoin,
		"json_encode": filterJsonEncode,
		"keys": filterKeys,
		"last": filterLast,
		"length": filterLength,
		"lower": filterLower,
		"merge": filterMerge,
		"nl2br": filterNl2Br,
		"number_format": filterNumberFormat,
		"raw": filterRaw,
		"replace": filterReplace,
		"reverse": filterReverse,
		"round": filterRound,
		"slice": filterSlice,
		"sort": filterSort,
		"split": filterSplit,
		"striptags": filterStripTags,
		"title": filterTitle,
		"trim": filterTrim,
		"upper": filterUpper,
		"url_encode": filterUrlEncode,
	}
	return filters
}

func filterAbs(ctx Context, val Value, args ...Value) Value {
	n := CoerceNumber(val)
	if 0 == n {
		return val
	}
	return math.Abs(n)
}

// Arg1: length, Arg2: Blank Fill Item
func filterBatch(ctx Context, val Value, args ...Value) Value {
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

func filterCapitalize(ctx Context, val Value, args ...Value) Value {
	s := strings.ToLower(CoerceString(val))
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

func filterDefault(ctx Context, val Value, args ...Value) Value {
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

func filterEscape(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
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

func filterJsonEncode(ctx Context, val Value, args ...Value) Value {
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

func filterLength(ctx Context, val Value, args ...Value) Value {
	return 0
}

func filterLower(ctx Context, val Value, args ...Value) Value {
	return strings.ToLower(CoerceString(val))
}

func filterMerge(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func filterNl2Br(ctx Context, val Value, args ...Value) Value {
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

func filterTitle(ctx Context, val Value, args ...Value) Value {
	return strings.Title(CoerceString(val))
}

func filterTrim(ctx Context, val Value, args ...Value) Value {
	return strings.TrimSpace(CoerceString(val))
}

func filterUpper(ctx Context, val Value, args ...Value) Value {
	return strings.ToUpper(CoerceString(val))
}

func filterUrlEncode(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}
