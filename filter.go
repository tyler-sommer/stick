package stick

import (
	"math"
	"strings"
)

func BuiltInFilters() map[string]Filter {
	filters := make(map[string]Filter)
	filters["abs"] = FilterAbs
	filters["default"] = FilterDefault
	filters["batch"] = FilterBatch
	filters["capitalize"] = FilterCapitalize
	filters["convert_encoding"] = FilterConvertEncoding
	filters["date"] = FilterDate
	filters["date_modify"] = FilterDateModify
	filters["escape"] = FilterEscape
	filters["first"] = FilterFirst
	filters["format"] = FilterFormat
	filters["join"] = FilterJoin
	filters["json_encode"] = FilterJsonEncode
	filters["keys"] = FilterKeys
	filters["last"] = FilterLast
	filters["length"] = FilterLength
	filters["lower"] = FilterLower
	filters["merge"] = FilterMerge
	filters["nl2br"] = FilterNl2Br
	filters["number_format"] = FilterNumberFormat
	filters["raw"] = FilterRaw
	filters["replace"] = FilterReplace
	filters["reverse"] = FilterReverse
	filters["round"] = FilterRound
	filters["slice"] = FilterSlice
	filters["sort"] = FilterSort
	filters["split"] = FilterSplit
	filters["striptags"] = FilterStripTags
	filters["title"] = FilterTitle
	filters["trim"] = FilterTrim
	filters["upper"] = FilterUpper
	filters["url_encode"] = FilterUrlEncode
	return filters
}

func FilterAbs(ctx Context, val Value, args ...Value) Value {
	n := CoerceNumber(val)
	if 0 == n {
		return val
	}
	return math.Abs(n)
}

func FilterBatch(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}

func FilterCapitalize(ctx Context, val Value, args ...Value) Value {
	s := strings.ToLower(val.(string))
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
	return strings.ToLower(val.(string))
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
	return strings.Title(val.(string))
}

func FilterTrim(ctx Context, val Value, args ...Value) Value {
	return strings.TrimSpace(val.(string))
}

func FilterUpper(ctx Context, val Value, args ...Value) Value {
	return strings.ToUpper(val.(string))
}

func FilterUrlEncode(ctx Context, val Value, args ...Value) Value {
	// TODO: Implement Me
	return val
}
