package stick

import (
	"fmt"
	"reflect"
	"strconv"
)

// A Value represents some value, scalar or otherwise, able to be passed into
// and used by a Stick template.
type Value interface{}

// A SafeValue represents a value that has already been sanitized and escaped.
type SafeValue interface {
	// Value returns the value stored in the SafeValue.
	Value() Value

	// Prevent anyone else from implementing SafeValue.
	safe()
}

// NewSafeValue wraps the given value and returns a SafeValue.
func NewSafeValue(val Value) SafeValue {
	if v, ok := val.(SafeValue); ok {
		return v
	}
	return safeValue{val}
}

type safeValue struct {
	val Value
}

func (v safeValue) Value() Value {
	return v.val
}

func (v safeValue) safe() {}

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
	case int:
		return vc > 0
	case float64:
		return vc > 0
	case string:
		return len(vc) > 0
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
	case float64:
		return vc
	case int:
		return float64(vc)
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
	case float64, int, uint:
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
func GetAttr(v Value, attr string, args ...Value) (Value, error) {
	r := reflect.Indirect(reflect.ValueOf(v))
	var retval reflect.Value
	switch r.Kind() {
	case reflect.Struct:
		retval = r.FieldByName(attr)
		if !retval.IsValid() {
			var err error
			retval, err = getMethod(v, attr)
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
	default:
		return nil, fmt.Errorf("getattr: type \"%s\" does not support attribute lookup", r.Type())
	}
	if !retval.IsValid() {
		return nil, fmt.Errorf("getattr: unable to locate attribute \"%s\" on \"%v\"", attr, v)
	} else if retval.Kind() == reflect.Func {
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

type iterator func(k Value, v Value, l loop) (brk bool, err error)

type loop struct {
	Last   bool
	Index  int
	Index0 int
}

func iterate(val Value, it iterator) (int, error) {
	r := reflect.Indirect(reflect.ValueOf(val))
	switch r.Kind() {
	case reflect.Slice, reflect.Array:
		ln := r.Len()
		l := loop{ln == 1, 1, 0}
		for i := 0; i < ln; i++ {
			v := r.Index(i)
			brk, err := it(i, v.Interface(), l)
			if err != nil {
				return 0, err
			} else if brk {
				return ln, nil
			}

			l.Index++
			l.Index0++
			l.Last = ln == l.Index
		}
		return ln, nil
	case reflect.Map:
		keys := r.MapKeys()
		ln := r.Len()
		l := loop{ln == 1, 1, 0}
		for _, k := range keys {
			v := r.MapIndex(k)
			brk, err := it(k.Interface(), v.Interface(), l)
			if err != nil {
				return 0, err
			} else if brk {
				return ln, nil
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

func equal(left Value, right Value) bool {
	// TODO: Stop-gap for now, this will need to be much more sophisticated.
	return CoerceString(left) == CoerceString(right)
}

func contains(haystack Value, needle Value) (bool, error) {
	res := false
	_, err := iterate(haystack, func(k Value, v Value, l loop) (bool, error) {
		if equal(v, needle) {
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
