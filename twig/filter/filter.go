// Package filter provides built-in filters for Twig-compatibility.
package filter // import "github.com/tyler-sommer/stick/twig/filter"

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"reflect"
	"time"

	"github.com/tyler-sommer/stick"
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
		// if the value is a string of "now" then we can use the current time
		// @doc https://twig.symfony.com/doc/3.x/filters/date.html
		valString := stick.CoerceString(val)
		if strings.ToLower(valString) == "now" {
			dt = time.Now()
		} else {
			// TODO: trigger runtime error
			return nil
		}
	}

	if l := len(args); l >= 1 {
		requestedLayout = stick.CoerceString(args[0])
	}

	// build a golang date string
	table := map[string]string{
		"d": "02",
		"D": "Mon",
		"j": "2",
		"l": "Monday",
		"N": "", // TODO: ISO-8601 numeric representation of the day of the week (added in PHP 5.1.0)
		"S": "ZZZ",
		"w": "", // TODO: Numeric representation of the day of the week
		"z": "", // TODO: The day of the year (starting from 0)
		"W": "", // TODO: ISO-8601 week number of year, weeks starting on Monday (added in PHP 4.1.0)
		"F": "January",
		"m": "01",
		"M": "Jan",
		"n": "1",
		"t": "", // TODO: Number of days in the given month
		"L": "", // TODO: Whether it's a leap year
		"o": "", // TODO: ISO-8601 year number. This has the same value as Y, except that if the ISO week number (W) belongs to the previous or next year, that year is used instead. (added in PHP 5.1.0)
		"Y": "2006",
		"y": "06",
		"a": "pm",
		"A": "PM",
		"B": "", // TODO: Swatch Internet time (is this even still a thing?!)
		"g": "3",
		"G": "15",
		"h": "03",
		"H": "15",
		"i": "04",
		"s": "05",
		"u": "000000",
		"e": "", // TODO: Timezone identifier (added in PHP 5.1.0)
		"I": "", // TODO: Whether or not the date is in daylight saving time
		"O": "-0700",
		"P": "-07:00",
		"T": "MST",
		"c": "2006-01-02T15:04:05-07:00",
		"r": "Mon, 02 Jan 2006 15:04:05 -0700",
		"U": "", // TODO: Seconds since the Unix Epoch (January 1 1970 00:00:00 GMT)
	}
	var layout string

	maxLen := len(requestedLayout)
	for i := 0; i < maxLen; i++ {
		char := string(requestedLayout[i])
		if t, ok := table[char]; ok {
			layout += t
			continue
		}
		if "\\" == char && i < maxLen-1 {
			layout += string(requestedLayout[i+1])
			continue
		}
		layout += char
	}

	toReturn := dt.Format(layout)

	if strings.Contains(toReturn, "ZZZ") {
		replace := "th"
		dayIs := dt.Format("02")
		if dayIs == "01" || dayIs == "21" || dayIs == "31" {
			replace = "st"
		} else if dayIs == "02" || dayIs == "22" {
			replace = "nd"
		} else if dayIs == "03" || dayIs == "23" {
			replace = "rd"
		}
		toReturn = strings.Replace(toReturn, "ZZZ", replace, 1)
	}

	return toReturn
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
		runes := []rune(s)
		return string(runes[0])
	}

	return nil
}

func filterFormat(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	// TODO: Implement Me
	return val
}

func filterJoin(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	if !stick.IsIterable(val) {
		// Twig returns the value itself when a non-array is passed to join
		return stick.CoerceString(val)
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
	// TODO: implement flags
	jsonData, err := json.Marshal(val)
	if err != nil {
		// TODO: Report error
		return nil
	}

	return string(jsonData)
}

func filterKeys(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	r := reflect.Indirect(reflect.ValueOf(val))
	switch r.Kind() {
	case reflect.Slice, reflect.Array:
		ln := r.Len()
		res := make([]int, 0)
		for i := 0; i < ln; i++ {
			res = append(res, i)
		}
		return res
	case reflect.Map:
		keys := r.MapKeys()
		res := make([]string, 0)
		for _, k := range keys {
			res = append(res, fmt.Sprintf("%v", k))
		}
		sort.Strings(res)
		return res
	default:
		return []string{}
	}
}

func filterLast(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	if stick.IsArray(val) {
		arr := reflect.ValueOf(val)
		return arr.Index(arr.Len() - 1).Interface()
	}

	if stick.IsMap(val) {
		// TODO: Trigger runtime error, Golang randomises map keys so getting the "Last" does not make sense
		return nil
	}

	if s := stick.CoerceString(val); s != "" {
		runes := []rune(s)
		return string(runes[len(runes)-1])
	}

	return nil
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
	if !stick.IsIterable(val) {
		return nil
	}

	if len(args) != 1 {
		return nil
	}

	outMap, isObject := val.(map[string]stick.Value)

	if isObject {
		argMap, ok := args[0].(map[string]stick.Value)

		if ok {
			for k, v := range argMap {
				outMap[k] = v
			}
		}

		return outMap
	} else {
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
}

// nl2brRe matches every newline form in one pass: CRLF, bare CR, bare LF.
// Each is preserved and prefixed with <br /> via the capture group.
var nl2brRe = regexp.MustCompile(`(\r\n|\r|\n)`)

// filterNL2BR replaces newline characters with HTML <br /> tags. Mirrors
// PHP's nl2br: the original newline character(s) are preserved after the tag.
func filterNL2BR(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	return nl2brRe.ReplaceAllString(stick.CoerceString(val), "<br />$1")
}

// filterNumberFormat formats a number with grouped thousands. PHP-Twig
// signature: |number_format(decimals=0, decimal_point='.', thousands_separator=',').
func filterNumberFormat(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	n := stick.CoerceNumber(val)
	decimals := 0
	decPoint := "."
	thouSep := ","
	if len(args) >= 1 {
		decimals = int(stick.CoerceNumber(args[0]))
	}
	if len(args) >= 2 {
		decPoint = stick.CoerceString(args[1])
	}
	if len(args) >= 3 {
		thouSep = stick.CoerceString(args[2])
	}
	if decimals < 0 {
		decimals = 0
	}
	raw := strconv.FormatFloat(n, 'f', decimals, 64)
	sign := ""
	if strings.HasPrefix(raw, "-") {
		sign = "-"
		raw = raw[1:]
	}
	intPart, fracPart := raw, ""
	if dot := strings.IndexByte(raw, '.'); dot >= 0 {
		intPart = raw[:dot]
		fracPart = raw[dot+1:]
	}
	if len(intPart) > 3 && thouSep != "" {
		var b strings.Builder
		// Insert thousands separator from the right.
		for i, r := range intPart {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				b.WriteString(thouSep)
			}
			b.WriteRune(r)
		}
		intPart = b.String()
	}
	out := sign + intPart
	if decimals > 0 {
		out += decPoint + fracPart
	}
	return out
}

func filterRaw(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	return stick.NewSafeValue(stick.CoerceString(val), "html", "html_attr", "js", "css", "url")
}

func filterReplace(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	if len(args) != 1 {
		return val
	}

	res := stick.CoerceString(val)

	if stick.IsMap(args[0]) {
		replaces := make([]string, 0)
		stick.Iterate(args[0], func(k, v stick.Value, l stick.Loop) (bool, error) {
			replaces = append(replaces, stick.CoerceString(k))
			replaces = append(replaces, stick.CoerceString(v))
			return false, nil
		})

		replacer := strings.NewReplacer(replaces...)
		res = replacer.Replace(res)
	}

	return res
}

func filterReverse(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	if stick.IsArray(val) {
		arr := reflect.ValueOf(val)
		res := make([]interface{}, 0)
		for i := arr.Len() - 1; i >= 0; i-- {
			res = append(res, arr.Index(i).Interface())
		}
		return res
	}

	if stick.IsMap(val) {
		return val
	}

	if s := stick.CoerceString(val); s != "" {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	}

	return nil
}

func filterRound(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	input := stick.CoerceNumber(val)
	precision := 0
	algo := ""
	if len(args) > 0 {
		precision = int(math.Round(stick.CoerceNumber(args[0])))
	}
	if precision < 0 {
		precision = 0
	}
	if len(args) > 1 {
		algo = stick.CoerceString(args[1])
	}

	mult := math.Pow10(precision)
	switch algo {
	case "ceil":
		return math.Ceil(input*mult) / mult
	case "floor":
		return math.Floor(input*mult) / mult
	default:
		return math.Round(input*mult) / mult
	}
}

// filterSlice extracts a slice of an array or string. Mirrors PHP-Twig
// |slice(start, length=null, preserve_keys=false). preserve_keys is
// accepted for compatibility but ignored — stick collections don't have
// the PHP "string-keyed array" notion that flag was designed for.
//
// `start` may be negative to count from the end. If `length` is omitted,
// slices through to the end. Out-of-range indices are clamped silently
// (matching PHP-Twig's permissive behavior).
func filterSlice(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	if len(args) == 0 {
		return val
	}
	start := int(stick.CoerceNumber(args[0]))
	hasLength := len(args) >= 2
	length := 0
	if hasLength {
		length = int(stick.CoerceNumber(args[1]))
	}

	rv := reflect.Indirect(reflect.ValueOf(val))
	if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
		n := rv.Len()
		s, e := normalizeSliceBounds(start, length, hasLength, n)
		out := make([]stick.Value, 0, e-s)
		for i := s; i < e; i++ {
			out = append(out, rv.Index(i).Interface())
		}
		return out
	}

	// Fall back to string slicing (rune-aware so multi-byte chars stay intact).
	runes := []rune(stick.CoerceString(val))
	n := len(runes)
	s, e := normalizeSliceBounds(start, length, hasLength, n)
	return string(runes[s:e])
}

// normalizeSliceBounds resolves PHP-Twig's permissive slice indexing
// semantics: negative start counts from the end, negative length means
// "stop that many elements before the end", and indices are clamped to
// the valid [0, n] range without erroring.
func normalizeSliceBounds(start, length int, hasLength bool, n int) (s, e int) {
	if start < 0 {
		start = n + start
	}
	if start < 0 {
		start = 0
	}
	if start > n {
		start = n
	}
	if !hasLength {
		return start, n
	}
	end := start + length
	if length < 0 {
		end = n + length
	}
	if end < start {
		end = start
	}
	if end > n {
		end = n
	}
	return start, end
}

// filterSort returns val sorted in ascending order. Strings are compared
// lexicographically; numbers numerically; mixed values fall back to their
// stringified form. Maps are treated as their value sequence (keys
// discarded), matching PHP-Twig's sort semantics on associative arrays
// (which this Go port can't preserve since map iteration is unordered).
func filterSort(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	if val == nil {
		return []stick.Value{}
	}
	rv := reflect.Indirect(reflect.ValueOf(val))
	var items []stick.Value
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		items = make([]stick.Value, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			items[i] = rv.Index(i).Interface()
		}
	case reflect.Map:
		items = make([]stick.Value, 0, rv.Len())
		for _, k := range rv.MapKeys() {
			items = append(items, rv.MapIndex(k).Interface())
		}
	default:
		return val
	}
	sort.SliceStable(items, func(i, j int) bool {
		return sortLess(items[i], items[j])
	})
	return items
}

// sortLess compares two stick.Values: numbers numerically when both
// coerce to a non-zero number (or one is exactly zero), otherwise as
// their string representation. Good enough for typical Twig use.
func sortLess(a, b stick.Value) bool {
	if isNumeric(a) && isNumeric(b) {
		return stick.CoerceNumber(a) < stick.CoerceNumber(b)
	}
	return stick.CoerceString(a) < stick.CoerceString(b)
}

func isNumeric(v stick.Value) bool {
	switch v.(type) {
	case int, int32, int64, float32, float64:
		return true
	}
	return false
}

// filterSplit splits a string by a delimiter, returning a []string.
// PHP-Twig signature: |split(delimiter, limit=null).
//
//   - empty delimiter splits per character (rune-aware).
//   - positive limit caps the number of returned segments; the last one
//     contains the remainder of the string (matches PHP's explode/strings.SplitN).
//   - negative limit returns all segments except the last |limit|.
//   - omitted limit returns every segment.
func filterSplit(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	s := stick.CoerceString(val)
	sep := ""
	if len(args) >= 1 {
		sep = stick.CoerceString(args[0])
	}
	limit := 0
	hasLimit := len(args) >= 2
	if hasLimit {
		limit = int(stick.CoerceNumber(args[1]))
	}

	if sep == "" {
		// Split per character.
		out := []string{}
		for _, r := range s {
			out = append(out, string(r))
		}
		if hasLimit && limit > 0 && limit < len(out) {
			head := append([]string{}, out[:limit-1]...)
			head = append(head, strings.Join(out[limit-1:], ""))
			return head
		}
		if hasLimit && limit < 0 {
			drop := -limit
			if drop >= len(out) {
				return []string{}
			}
			return out[:len(out)-drop]
		}
		return out
	}

	if hasLimit && limit > 0 {
		return strings.SplitN(s, sep, limit)
	}
	all := strings.Split(s, sep)
	if hasLimit && limit < 0 {
		drop := -limit
		if drop >= len(all) {
			return []string{}
		}
		return all[:len(all)-drop]
	}
	return all
}

// stripTagsRe matches anything between < and > non-greedily, including
// multi-line tags. Mirrors PHP's strip_tags for the common case.
var stripTagsRe = regexp.MustCompile(`(?s)<[^>]*>`)

// filterStripTags removes HTML/XML tags from a string. PHP-Twig signature
// allows an `allowed_tags` argument; for now this implementation strips
// everything (the argument is accepted but not honored). Stub-state was
// "return val unchanged" so even partial implementation is a strict win.
func filterStripTags(ctx stick.Context, val stick.Value, args ...stick.Value) stick.Value {
	return stripTagsRe.ReplaceAllString(stick.CoerceString(val), "")
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
	return url.QueryEscape(stick.CoerceString(val))
}
