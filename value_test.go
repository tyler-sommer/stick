package stick

import (
	"testing"
)

var stringTests = map[Value]string{
	"string": "string",
	true:     "1",
	false:    "",
	3:        "3",
	3.14:     "3.14",
}

var boolTests = map[Value]bool{
	true:   true,
	false:  false,
	1:      true,
	0:      false,
	"true": true,
	"":     false,
}

var numberTests = map[Value]float64{
	3:     3.0,
	"3":   3.0,
	true:  1,
	false: 0,
}

func TestValue(t *testing.T) {
	for val, expected := range stringTests {
		actual := CoerceString(val)
		if actual != expected {
			t.Errorf("CoerceString(%v): got \"%v\" expected \"%v\"", val, actual, expected)
		}
	}

	for val, expected := range boolTests {
		actual := CoerceBool(val)
		if actual != expected {
			t.Errorf("CoerceBool(%v): got \"%v\" expected \"%v\"", val, actual, expected)
		}
	}

	for val, expected := range numberTests {
		actual := CoerceNumber(val)
		if actual != expected {
			t.Errorf("CoerceNumber(%v): got \"%v\" expected \"%v\"", val, actual, expected)
		}
	}
}

type getAttrTest struct {
	name string
	cont Value
	attr string
	expected string
	args []Value
}

func newGetAttrTest(name string, cont Value, attr, expected string) getAttrTest {
	return getAttrTest{name, cont, attr, expected, []Value{}}
}

func newGetAttrMethodTest(name string, cont Value, args []Value, attr, expected string) getAttrTest {
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
	return "modified:"+in
}

type propStruct struct {
	Name string
}

var getAttrTests = []getAttrTest{
	newGetAttrTest("anon struct property", struct{Name string}{"Tyler"}, "Name", "Tyler"),
	newGetAttrTest("struct property", propStruct{"Jackie"}, "Name", "Jackie"),
	newGetAttrTest("struct method (value, ptr receiver)", testStruct{"John"}, "Name", "John"),
	newGetAttrTest("struct method (ptr, ptr receiver)", &testStruct{"Adam"}, "Name", "Adam"),
	newGetAttrTest("struct method (value, value receiver)", testStruct{"Sam"}, "VName", "Sam"),
	newGetAttrTest("struct method (ptr, value receiver)", &testStruct{"Rex"}, "VName", "Rex"),
	newGetAttrMethodTest("method with parameters", testStruct{"Ray"}, []Value{"Meow"}, "Modify", "modified:Meow"),
	newGetAttrTest("map (string key)", map[string]Value{"name":"Amy"}, "name", "Amy"),
	newGetAttrTest("array", []Value{"World","Hello"}, "1", "Hello"),
}

func evaluateGetAttrTest(t *testing.T, test getAttrTest) {
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

func TestGetAttr(t *testing.T) {
	for _, test := range getAttrTests {
		evaluateGetAttrTest(t, test)
	}
}
