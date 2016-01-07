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
