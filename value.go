package stick

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/shopspring/decimal"
)

// A Value represents some value, scalar or otherwise, able to be passed into
// and used by a Stick template.
type Value interface{}

// A SafeValue represents a value that has already been sanitized and escaped.
type SafeValue interface {
	// Value returns the value stored in the SafeValue.
	Value() Value

	// IsSafe returns true if the value is safely escaped for content of type typ.
	IsSafe(typ string) bool

	// SafeFor returns the content types this value is safe for.
	SafeFor() []string
}

// NewSafeValue wraps the given value and returns a SafeValue.
func NewSafeValue(val Value, types ...string) SafeValue {
	safeFor := make(map[string]bool)
	for _, k := range types {
		safeFor[k] = true
	}
	if v, ok := val.(SafeValue); ok {
		for _, k := range v.SafeFor() {
			safeFor[k] = true
		}
		return safeValue{safeFor, v.Value()}
	}
	return safeValue{safeFor, val}
}

type safeValue struct {
	safeFor map[string]bool
	val     Value
}

func (v safeValue) Value() Value {
	return v.val
}

func (v safeValue) IsSafe(typ string) bool {
	_, ok := v.safeFor[typ]
	return ok
}

func (v safeValue) SafeFor() []string {
	r := make([]string, len(v.safeFor))
	i := 0
	for k := range v.safeFor {
		r[i] = k
		i++
	}
	return r
}

// Stringer is implemented by any value that has a String method.
type Stringer interface {
	fmt.Stringer
}

// Number is implemented by any value that has a Number method.
type Number interface {
	// Number returns a float64 representation of the type.
	Number() float64
}

// Boolean is implemented by any value that has a Boolean method.
type Boolean interface {
	// Boolean returns a boolean representation of the type.
	Boolean() bool
}

// CoerceBool coerces the given value into a boolean. Boolean false is returned
// if the value cannot be coerced.
func CoerceBool(v Value) bool {
	switch vc := v.(type) {
	case SafeValue:
		return CoerceBool(vc.Value())
	case bool:
		return vc
	case Boolean:
		return vc.Boolean()
	case uint:
		return vc > 0
	case uint8:
		return vc > 0
	case uint16:
		return vc > 0
	case uint32:
		return vc > 0
	case uint64:
		return vc > 0
	case int:
		return vc > 0
	case int8:
		return vc > 0
	case int16:
		return vc > 0
	case int32:
		return vc > 0
	case int64:
		return vc > 0
	case float32:
		return vc > 0
	case float64:
		return vc > 0
	case string:
		return len(vc) > 0
	case decimal.Decimal:
		return vc.GreaterThan(decimal.Zero)
	case Stringer:
		return len(vc.String()) > 0
	case Number:
		return vc.Number() > 0
	}
	return false
}

func stringToFloat(s string) float64 {
	fv, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return fv
}

// CoerceNumber coerces the given value into a number. Zero (0) is returned
// if the value cannot be coerced.
func CoerceNumber(v Value) float64 {
	switch vc := v.(type) {
	case SafeValue:
		return CoerceNumber(vc.Value())
	case Number:
		return vc.Number()
	case uint:
		return float64(vc)
	case uint8:
		return float64(vc)
	case uint16:
		return float64(vc)
	case uint32:
		return float64(vc)
	case uint64:
		return float64(vc)
	case int:
		return float64(vc)
	case int8:
		return float64(vc)
	case int16:
		return float64(vc)
	case int32:
		return float64(vc)
	case int64:
		return float64(vc)
	case float32:
		return float64(vc)
	case float64:
		return vc
	case decimal.Decimal:
		f, _ := vc.Float64()
		return f
	case Stringer:
		return stringToFloat(vc.String())
	case string:
		return stringToFloat(vc)
	case Boolean:
		if vc.Boolean() {
			return 1
		}
	case bool:
		if vc {
			return 1
		}
	}
	return 0
}

// CoerceString coerces the given value into a string. An empty string is returned
// if the value cannot be coerced.
func CoerceString(v Value) string {
	switch vc := v.(type) {
	case SafeValue:
		return CoerceString(vc.Value())
	case string:
		return vc
	case Stringer:
		return vc.String()
	case float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", vc)
	case Number:
		return fmt.Sprintf("%v", vc.Number())
	case Boolean:
		if vc.Boolean() == true {
			return "1" // Twig compatibility (aka PHP compatibility)
		}
	case bool:
		if vc == true {
			return "1" // Twig compatibility (aka PHP compatibility)
		}

	}
	return ""
}

// GetAttr attempts to access the given value and return the specified attribute.
func GetAttr(v Value, attr Value, args ...Value) (Value, error) {
	r := reflect.Indirect(reflect.ValueOf(v))
	if !r.IsValid() {
		return nil, fmt.Errorf("getattr: value does not support attribute lookup: %v", v)
	}
	var retval reflect.Value
	switch r.Kind() {
	case reflect.Struct:
		strval := CoerceString(attr)
		retval = r.FieldByName(strval)
		if !retval.IsValid() {
			var err error
			retval, err = getMethod(v, strval)
			if err != nil {
				return nil, err
			}
		}
	case reflect.Map:
		retval = r.MapIndex(reflect.ValueOf(attr))
	case reflect.Slice, reflect.Array:
		index := int(CoerceNumber(attr))
		if index >= 0 && index < r.Len() {
			retval = r.Index(index)
		}
	}
	if !retval.IsValid() {
		return nil, fmt.Errorf("getattr: unable to locate attribute \"%s\" on \"%v\"", attr, v)
	}
	if retval.Kind() == reflect.Func {
		t := retval.Type()
		if t.NumOut() > 1 {
			return nil, fmt.Errorf("getattr: multiple return values unsupported, called method \"%s\" on \"%v\"", attr, v)
		}
		rargs := make([]reflect.Value, len(args))
		for k, v := range args {
			rargs[k] = reflect.ValueOf(v)
		}
		if t.NumIn() != len(rargs) {
			return nil, fmt.Errorf("getattr: method \"%s\" on \"%v\" expects %d parameter(s), %d given", attr, v, t.NumIn(), len(rargs))
		}
		res := retval.Call(rargs)
		if len(res) == 0 {
			return nil, nil
		}
		retval = res[0]
	}
	return retval.Interface(), nil
}

func getMethod(v Value, name string) (reflect.Value, error) {
	var retVal reflect.Value
	value := reflect.ValueOf(v)
	retVal = value.MethodByName(name) // Match either "value, value receiver" or "ptr, ptr receiver"
	if retVal.IsValid() {
		return retVal, nil
	}

	var ptr reflect.Value
	if value.Kind() == reflect.Ptr {
		ptr = value
		value = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(v))
		temp := ptr.Elem()
		temp.Set(value)
	}
	retVal = ptr.MethodByName(name) // Match "value, ptr receiver" or "ptr, value receiver"
	if retVal.IsValid() {
		return retVal, nil
	}
	return retVal, fmt.Errorf("stick: unable to locate method \"%s\" on \"%v\"", name, v)
}

// An Iteratee is called for each step in a loop.
type Iteratee func(k, v Value, l Loop) (brk bool, err error)

// Loop contains metadata about the current state of a loop.
type Loop struct {
	Last   bool
	Index  int
	Index0 int
}

// IsArray returns true if the given Value is a slice or array.
func IsArray(val Value) bool {
	r := reflect.Indirect(reflect.ValueOf(val))
	switch r.Kind() {
	case reflect.Slice, reflect.Array:
		return true
	}
	return false
}

// IsMap returns true if the given Value is a map.
func IsMap(val Value) bool {
	r := reflect.Indirect(reflect.ValueOf(val))
	return r.Kind() == reflect.Map
}

// IsIterable returns true if the given Value is a slice, array, or map.
func IsIterable(val Value) bool {
	if val == nil {
		return true
	}
	r := reflect.Indirect(reflect.ValueOf(val))
	switch r.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return true
	}
	return false
}

// Iterate calls the Iteratee func for every item in the Value.
func Iterate(val Value, it Iteratee) (int, error) {
	if val == nil {
		return 0, nil
	}
	r := reflect.Indirect(reflect.ValueOf(val))
	switch r.Kind() {
	case reflect.Slice, reflect.Array:
		ln := r.Len()
		l := Loop{ln == 1, 1, 0}
		for i := 0; i < ln; i++ {
			v := r.Index(i)
			brk, err := it(i, v.Interface(), l)
			if brk || err != nil {
				return i + 1, err
			}

			l.Index++
			l.Index0++
			l.Last = ln == l.Index
		}
		return ln, nil
	case reflect.Map:
		keys := r.MapKeys()
		ln := r.Len()
		l := Loop{ln == 1, 1, 0}
		for i, k := range keys {
			v := r.MapIndex(k)
			brk, err := it(k.Interface(), v.Interface(), l)
			if brk || err != nil {
				return i + 1, err
			}

			l.Index++
			l.Index0++
			l.Last = ln == l.Index
		}
		return ln, nil
	default:
		return 0, fmt.Errorf(`stick: unable to iterate over %s "%v"`, r.Kind(), val)
	}
}

// Len returns the length of Value.
func Len(val Value) (int, error) {
	if val == nil {
		return 0, nil
	}
	r := reflect.Indirect(reflect.ValueOf(val))
	switch r.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return r.Len(), nil
	}
	return 0, fmt.Errorf(`stick: could not get length of %s "%v"`, r.Kind(), val)
}

// Equal returns true if the two Values are considered equal.
func Equal(left Value, right Value) bool {
	// TODO: Stop-gap for now, this will need to be much more sophisticated.
	return CoerceString(left) == CoerceString(right)
}

// Contains returns true if the haystack Value contains needle.
func Contains(haystack Value, needle Value) (bool, error) {
	res := false
	_, err := Iterate(haystack, func(k Value, v Value, l Loop) (bool, error) {
		if Equal(v, needle) {
			res = true
			return true, nil // break
		}
		return false, nil
	})
	if err != nil {
		return false, err
	}
	return res, nil
}
