// Package filter provides built-in filters for Twig-compatibility.
package filter

import (
	"math"
	"strings"
	"unicode/utf8"

	"github.com/tyler-sommer/stick"
	"reflect"
	"time"
)

// builtInFilters returns a map containing all built-in Twig filters,
// with the exception of "escape", which is provided by the AutoEscapeExtension.
func TwigFilters() map[string]stick.Filter {
	return map[string]stick.Filter{
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
func filterAbs(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	n := stick.CoerceNumber(val)
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
func filterBatch(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	perSlice := 1
	var blankValue stick.Value
	if l := len(args); l >= 1 {
		perSlice = int(stick.CoerceNumber(args[0]))
		if l >= 2 {
			blankValue = args[1]
		}
	}
	if !stick.IsIterable(val) {
		// TODO: This would trigger an E_WARNING in PHP.
		return nil
	}
	if perSlice <= 1 {
		// TODO: This would trigger an E_WARNING in PHP.
		return nil
	}
	l, _ := stick.Len(val)
	numSlices := int(math.Ceil(float64(l) / float64(perSlice)))
	out := make([][]stick.Value, numSlices)
	curr := []stick.Value{}
	i := 0
	j := 0
	_, err := stick.Iterate(val, func(k, v stick.Value, l stick.Loop) (bool, error) {
		// Use a variable length slice and append(). This maintains
		// correct compatibility with Twig when the fill value is nil.
		curr = append(curr, v)
		j++
		if j == perSlice {
			out[i] = curr
			curr = []stick.Value{}
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
func filterCapitalize(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	s := stick.CoerceString(val)
	return strings.ToUpper(s[:1]) + s[1:]
}

func filterConvertEncoding(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

func filterDate(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	var requestedLayout string
	dt, ok := val.(time.Time)
	if !ok {
		// TODO: trigger runtime error
		return nil
	}

	if l := len(args); l >= 1 {
		requestedLayout = stick.CoerceString(args[0])
	}

	// build a golang date string
	table := map[string]string{
		"d" : "02",
		"D" : "Mon",
		"j" : "2",
		"l" : "Monday",
		"N" : "", // TODO: ISO-8601 numeric representation of the day of the week (added in PHP 5.1.0)
		"S" : "", // TODO: English ordinal suffix for the day of the month, 2 characters
		"w" : "", // TODO: Numeric representation of the day of the week
		"z" : "", // TODO: The day of the year (starting from 0)
		"W" : "", // TODO: ISO-8601 week number of year, weeks starting on Monday (added in PHP 4.1.0)
		"F" : "January",
		"m" : "01",
		"M" : "Jan",
		"n" : "1",
		"t" : "", // TODO: Number of days in the given month
		"L" : "", // TODO: Whether it's a leap year
		"o" : "", // TODO: ISO-8601 year number. This has the same value as Y, except that if the ISO week number (W) belongs to the previous or next year, that year is used instead. (added in PHP 5.1.0)
		"Y" : "2006",
		"y" : "06",
		"a" : "pm",
		"A" : "PM",
		"B" : "", // TODO: Swatch Internet time (is this even still a thing?!)
		"g" : "3",
		"G" : "15",
		"h" : "03",
		"H" : "15",
		"i" : "04",
		"s" : "05",
		"u" : "000000",
		"e" : "", // TODO: Timezone identifier (added in PHP 5.1.0)
		"I" : "", // TODO: Whether or not the date is in daylight saving time
		"O" : "-0700",
		"P" : "-07:00",
		"T" : "MST",
		"c" : "2006-01-02T15:04:05-07:00",
		"r" : "Mon, 02 Jan 2006 15:04:05 -0700",
		"U" : "", // TODO: Seconds since the Unix Epoch (January 1 1970 00:00:00 GMT)
	}
	var layout string

	maxLen := len(requestedLayout);
	for i := 0; i < maxLen; i++ {
		char := string(requestedLayout[i])
		if t, ok := table[char]; ok {
			layout += t
			continue
		}
		if "\\" == char && i < maxLen-1{
			layout += string(requestedLayout[i+1])
			continue
		}
		layout += char
	}

	return dt.Format(layout)
}

func filterDateModify(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

// filterDefault takes one argument, the default value. If val is empty,
// the default value will be returned.
func filterDefault(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	var d stick.Value
	if len(args) > 0 {
		d = args[0]
	}
	if stick.CoerceString(val) == "" {
		return d
	}
	return val
}

func filterFirst(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	if stick.IsArray(val) {
		arr := reflect.ValueOf(val)
		return arr.Index(0).Interface()
	}

	if stick.IsMap(val) {
		// TODO: Trigger runtime error, Golang randomises map keys so getting the "First" does not make sense
		return nil
	}

	if s := stick.CoerceString(val); s != "" {
		return string(s[0])
	}

	return nil
}

func filterFormat(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

func filterJoin(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	if ! stick.IsIterable(val) {
		return nil
	}

	separator := ``
	if len(args) == 1 {
		separator = stick.CoerceString(args[0])
	}

	var slice []string
	stick.Iterate(val, func(k, v stick.Value, l stick.Loop) (bool, error) {
		slice = append(slice, stick.CoerceString(v))
		return false, nil
	})

	return strings.Join(slice, separator)
}

func filterJSONEncode(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

func filterKeys(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

func filterLast(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

// filterLength returns the length of val.
func filterLength(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	if v, ok := val.(string); ok {
		return utf8.RuneCountInString(v)
	}
	l, _ := stick.Len(val)
	// TODO: Report error
	return l
}

// filterLower returns val transformed to lower-case.
func filterLower(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	return strings.ToLower(stick.CoerceString(val))
}

func filterMerge(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	if ! stick.IsIterable(val) {
		return nil
	}

	if len(args) != 1 {
		return nil
	}

	var out []stick.Value

	stick.Iterate(val, func(k, v stick.Value, l stick.Loop) (bool, error) {
		out = append(out, v)
		return false, nil
	})

	stick.Iterate(args[0], func(k, v stick.Value, l stick.Loop) (bool, error) {
		out = append(out, v)
		return false, nil
	})

	return out
}

func filterNL2BR(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

func filterNumberFormat(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

func filterRaw(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

func filterReplace(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

func filterReverse(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

func filterRound(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

func filterSlice(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

func filterSort(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

func filterSplit(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

func filterStripTags(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

// filterTitle returns val with the first character of each word capitalized.
func filterTitle(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	return strings.Title(stick.CoerceString(val))
}

// filterTrim returns val with whitespace trimmed on both left and ride sides.
func filterTrim(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	return strings.TrimSpace(stick.CoerceString(val))
}

// filterUpper returns val in upper-case.
func filterUpper(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	return strings.ToUpper(stick.CoerceString(val))
}

func filterURLEncode(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}