package stick

import (
	"strconv"
	"fmt"
)

// A Value represents some value, scalar or otherwise, able to be passed into
// and used by a Stick template.
type Value interface{}

func CoerceBool(v Value) bool {
	switch vc := v.(type) {
	case bool:
		return vc
	case uint:
		return vc > 0
	case int:
		return vc > 0
	case float64:
		return vc > 0
	case string:
		return len(vc) > 0
	}
	return false
}

func CoerceNumber(v Value) float64 {
	switch vc := v.(type) {
	case string:
		fv, err := strconv.ParseFloat(vc, 64)
		if err != nil {
			return 0
		}
		return fv
	case float64:
		return vc
	case int:
		return float64(vc)
	case bool:
		if vc {
			return 1
		}
	}
	return 0
}

func CoerceString(v Value) string {
	switch vc := v.(type) {
	case string:
		return vc
	case float64, int, uint:
		return fmt.Sprintf("%v", vc)
	case bool:
		if vc == true {
			// Twig compatibility (aka PHP compatibility)
			return "1"
		}
	}
	return ""
}
