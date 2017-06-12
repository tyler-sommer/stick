package stick

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

type testType struct{}

func (t testType) String() string {
	return "some string"
}

func (t testType) Boolean() bool {
	return true
}

func (t testType) Number() float64 {
	return 42
}

func TestValue(t *testing.T) {
	var stringTests = map[Value]string{
		testType{}: "some string",
		"string":   "string",

		true:  "1",
		false: "",

		int(3):   "3",
		int8(3):  "3",
		int32(3): "3",
		int64(3): "3",

		uint(3):   "3",
		uint8(3):  "3",
		uint16(3): "3",
		uint32(3): "3",
		uint64(3): "3",

		float64(3.14): "3.14",
		float32(3.14): "3.14",

		decimal.NewFromFloat(3.1415): "3.1415",
	}
	for val, expected := range stringTests {
		actual := CoerceString(val)
		if actual != expected {
			t.Errorf("CoerceString(%v): got \"%v\" expected \"%v\"", val, actual, expected)
		}
	}

	var boolTests = map[Value]bool{
		testType{}: true,
		true:       true,
		false:      false,

		int(1):   true,
		int(0):   false,
		int8(1):  true,
		int8(0):  false,
		int16(1): true,
		int16(0): false,
		int32(1): true,
		int32(0): false,
		int64(1): true,
		int64(0): false,

		uint(1):   true,
		uint(0):   false,
		uint8(1):  true,
		uint8(0):  false,
		uint16(1): true,
		uint16(0): false,
		uint32(1): true,
		uint32(0): false,
		uint64(1): true,
		uint64(0): false,

		float32(1): true,
		float32(0): false,
		float64(1): true,
		float64(0): false,

		"true": true,
		"":     false,
	}
	for val, expected := range boolTests {
		actual := CoerceBool(val)
		if actual != expected {
			t.Errorf("CoerceBool(%v): got \"%v\" expected \"%v\"", val, actual, expected)
		}
	}

	var numberTests = map[Value]float64{
		testType{}: 42,
		"3":        3.0,

		int(3):   3.0,
		int8(3):  3.0,
		int16(3): 3.0,
		int32(3): 3.0,
		int64(3): 3.0,

		uint(3):   3.0,
		uint8(3):  3.0,
		uint16(3): 3.0,
		uint32(3): 3.0,
		uint64(3): 3.0,

		float32(3.14): 3.14,
		float64(3.14): 3.14,

		true:  1,
		false: 0,
	}
	for val, expected := range numberTests {
		actual := CoerceNumber(val)
		if math.Abs(actual-expected) >= 0.000001 {
			t.Errorf("CoerceNumber(%v) - %T: got \"%v\" expected \"%v\"", val, val, actual, expected)
		}
	}
}

type getAttrTest struct {
	name     string
	cont     Value
	attr     Value
	expected string
	args     []Value
}

func newGetAttrTest(name string, cont, attr Value, expected string) getAttrTest {
	return getAttrTest{name, cont, attr, expected, []Value{}}
}

func newGetAttrMethodTest(name string, cont Value, args []Value, attr Value, expected string) getAttrTest {
	return getAttrTest{name, cont, attr, expected, args}
}

type testStruct struct {
	name string
}

func (t *testStruct) Name() string {
	return t.name
}

func (t testStruct) VName() string {
	return t.name
}

func (t *testStruct) Modify(in string) string {
	return "modified:" + in
}

type propStruct struct {
	Name string
}

func TestGetAttr(t *testing.T) {
	var getAttrTests = []getAttrTest{
		newGetAttrTest("map with non-string keys", map[int]string{1:"test"}, 1, "test"),
		newGetAttrTest("anon struct property", struct{ Name string }{"Tyler"}, "Name", "Tyler"),
		newGetAttrTest("struct property", propStruct{"Jackie"}, "Name", "Jackie"),
		newGetAttrTest("struct method (value, ptr receiver)", testStruct{"John"}, "Name", "John"),
		newGetAttrTest("struct method (ptr, ptr receiver)", &testStruct{"Adam"}, "Name", "Adam"),
		newGetAttrTest("struct method (value, value receiver)", testStruct{"Sam"}, "VName", "Sam"),
		newGetAttrTest("struct method (ptr, value receiver)", &testStruct{"Rex"}, "VName", "Rex"),
		newGetAttrMethodTest("method with parameters", testStruct{"Ray"}, []Value{"Meow"}, "Modify", "modified:Meow"),
		newGetAttrTest("map (string key)", map[string]Value{"name": "Amy"}, "name", "Amy"),
		newGetAttrTest("array", []Value{"World", "Hello"}, "1", "Hello"),
	}

	for _, test := range getAttrTests {
		res, err := GetAttr(test.cont, test.attr, test.args...)
		if err != nil {
			t.Errorf("getattr: %s: unexpected error:\n\t%v", test.name, err)
			return
		}
		actual := CoerceString(res)
		if actual != test.expected {
			t.Errorf("getattr: %s: got \"%s\" expected \"%s\"", test.name, actual, test.expected)
		}
	}
}

func TestIsIterable(t *testing.T) {
	ts := []struct {
		name     string
		input    Value
		expected bool
	}{
		{"is iterable nil", nil, true},
		{"is iterable array", [4]int{}, true},
		{"is iterable slice", []int{}, true},
		{"is iterable map", map[string]string{}, true},
		{"is iterable string", "a string", false},
		{"is iterable struct", struct{ name string }{"world"}, false},
	}
	for _, test := range ts {
		actual := IsIterable(test.input)
		if actual != test.expected {
			t.Errorf("%s:\n\texpected: %v\n\tgot: %v", test.name, test.expected, actual)
		}
	}
}

func TestIsMap(t *testing.T) {
	ts := []struct {
		name     string
		input    Value
		expected bool
	}{
		{"is map nil", nil, false},
		{"is map array", [4]int{}, false},
		{"is map slice", []int{}, false},
		{"is map map", map[string]string{}, true},
		{"is map string", "a string", false},
		{"is map struct", struct{ name string }{"world"}, false},
	}
	for _, test := range ts {
		actual := IsMap(test.input)
		if actual != test.expected {
			t.Errorf("%s:\n\texpected: %v\n\tgot: %v", test.name, test.expected, actual)
		}
	}
}

func TestIsArray(t *testing.T) {
	ts := []struct {
		name     string
		input    Value
		expected bool
	}{
		{"is array nil", nil, false},
		{"is array array", [4]int{}, true},
		{"is array slice", []int{}, true},
		{"is array map", map[string]string{}, false},
		{"is array string", "a string", false},
		{"is array struct", struct{ name string }{"world"}, false},
	}
	for _, test := range ts {
		actual := IsArray(test.input)
		if actual != test.expected {
			t.Errorf("%s:\n\texpected: %v\n\tgot: %v", test.name, test.expected, actual)
		}
	}
}

func TestLen(t *testing.T) {
	ts := []struct {
		name     string
		input    Value
		expected int
		err      bool
	}{
		{"len nil", nil, 0, false},
		{"len array", [4]int{}, 4, false},
		{"len empty slice", []int{}, 0, false},
		{"len empty map", map[string]string{}, 0, false},
		{"len map", map[string]string{"a": "A", "b": "B"}, 2, false},
		{"len empty string", "", 0, true},
		{"len string", "a string", 0, true},
		{"len struct", struct{ name string }{"world"}, 0, true},
	}
	for _, test := range ts {
		actual, err := Len(test.input)
		if err == nil && test.err {
			t.Errorf("%s:\n\texpected error, got none.", test.name)
		} else if err != nil && !test.err {
			t.Errorf("%s:\n\tunexpected error: %v", test.name, err)
		}
		if actual != test.expected {
			t.Errorf("%s:\n\texpected: %v\n\tgot: %v", test.name, test.expected, actual)
		}
	}
}

func TestIterate(t *testing.T) {
	noError := ""
	ts := []struct {
		name          string
		input         Value
		expectedError string
	}{
		{"iterate string", "a string", "unable to iterate over string"},
		{"iterate map", map[string]string{"a": "A", "b": "B"}, noError},
		{"iterate slice", []string{"a", "b", "c"}, noError},
		{"iterate array", [3]string{"a", "b", "c"}, noError},
		{"iterate struct", struct{ name string }{"world"}, "unable to iterate over struct"},
	}
	for _, test := range ts {
		n, err := Iterate(test.input, func(k, v Value, l Loop) (bool, error) {
			expected, err := GetAttr(test.input, k)
			if err != nil {
				return true, fmt.Errorf("%s:\n\tunexpected error: %s", test.name, err)
			}
			if v != expected {
				return true, fmt.Errorf("%s:\n\texpected: %v\n\tgot: %v\n", test.name, v, expected)
			}
			return false, nil
		})
		if test.expectedError != noError {
			if err == nil {
				t.Errorf("%s:\n\texpected error: %s", test.name, err)
			} else if !strings.Contains(err.Error(), test.expectedError) {
				t.Errorf("%s: got error\n\t%+v\nexpected error\n\t%v", test.name, err, test.expectedError)
			}
		} else if err != nil {
			t.Errorf("%s:\n\tunexpected error: %s", test.name, err)
		}
		l, _ := Len(test.input)
		if n != l {
			t.Errorf("%s:\n\texpected to iterate over %d chars, got %d", test.name, l, n)
		}
	}
}

func TestIterate_breakSlice(t *testing.T) {
	vals := []string{"hello", "world", "!"}
	res := []string{}
	n, err := Iterate(vals, func(k, v Value, l Loop) (bool, error) {
		res = append(res, CoerceString(v))
		if l.Index == 2 {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Errorf("expected to iterate over 2 items, got %d", n)
	}
	if v := strings.Join(res, " "); v != "hello world" {
		t.Errorf("expected 'hello world' got '%s'", v)
	}
}

func TestIterate_breakMap(t *testing.T) {
	vals := map[string]string{"hello": "world", "exclaim": "!"}
	res := []string{}
	n, err := Iterate(vals, func(k, v Value, l Loop) (bool, error) {
		res = append(res, CoerceString(k), CoerceString(v))
		if l.Index == 1 {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("expected to iterate over 1 item, got %d", n)
	}
	exp := []string{"hello world", "exclaim !"}
	if v := strings.Join(res, " "); v != exp[0] && v != exp[1] {
		t.Errorf("expected one of '%s' or '%s', got '%s'", exp[0], exp[1], v)
	}
}

func TestIterate_error(t *testing.T) {
	vals := []string{"hello", "world", "!"}
	res := []string{}
	n, err := Iterate(vals, func(k, v Value, l Loop) (bool, error) {
		if CoerceString(v) == "!" {
			return false, fmt.Errorf("error on '!'")
		}
		res = append(res, CoerceString(v))
		return false, nil
	})
	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != "error on '!'" {
		t.Errorf("unexpected error: %v\n\texpected error: %v", err, "error on '!'")
	}
	if n != 3 {
		t.Errorf("expected to iterate over 3 items, got %d", n)
	}
	if v := strings.Join(res, " "); v != "hello world" {
		t.Errorf("expected 'hello world' got '%s'", v)
	}
}
