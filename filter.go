package stick

func BuiltInFilters() map[string]Filter {
	filters := make(map[string]Filter)
	filters["default"] = FilterDefault

	return filters
}

func FilterDefault(ctx Context, val Value, args ...Value) Value {
	if CoerceBool(val) {
		return val
	}

	if len(args) >= 1 {
		return args[0]
	} else {
		return ""
	}

}
