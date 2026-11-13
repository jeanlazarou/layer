package interpreter

import (
	"fmt"
	"math"

	"github.com/jeanlazarou/layer/pkg/layers"
)

// SetupCoreLayer sets up the core layer with built-in functions and operators
func SetupCoreLayer(lc *layers.LayerContext, interp *Interpreter) {
	lc.SetCurrentLayer("core")

	// Arithmetic operators (+ supports both numeric and string concatenation)
	lc.Define("+", makePlusOp())
	lc.Define("-", makeNumericOp("-", func(a, b float64) float64 { return a - b }))
	lc.Define("*", makeNumericOp("*", func(a, b float64) float64 { return a * b }))
	lc.Define("/", makeNumericOp("/", func(a, b float64) float64 { return a / b }))
	lc.Define("%", makeNumericOp("%", func(a, b float64) float64 { return math.Mod(a, b) }))
	lc.Define("^", makeNumericOp("^", func(a, b float64) float64 { return math.Pow(a, b) }))

	// Comparison operators
	lc.Define("==", makeComparisonOp(func(a, b interface{}) bool { return equals(a, b) }))
	lc.Define("!=", makeComparisonOp(func(a, b interface{}) bool { return !equals(a, b) }))
	lc.Define("<", makeNumericComparisonOp("<", func(a, b float64) bool { return a < b }))
	lc.Define(">", makeNumericComparisonOp(">", func(a, b float64) bool { return a > b }))
	lc.Define("<=", makeNumericComparisonOp("<=", func(a, b float64) bool { return a <= b }))
	lc.Define(">=", makeNumericComparisonOp(">=", func(a, b float64) bool { return a >= b }))

	// Logical operators
	lc.Define("and", makeLogicalOp("and", func(a, b bool) bool { return a && b }))
	lc.Define("or", makeLogicalOp("or", func(a, b bool) bool { return a || b }))
	lc.Define("not", makeUnaryLogicalOp(func(a bool) bool { return !a }))

	// List operations
	lc.Define("map", makeMapFunc())
	lc.Define("filter", makeFilterFunc())
	lc.Define("reduce", makeReduceFunc())
	lc.Define("fold", makeFoldFunc()) // Alias for reduce
	lc.Define("head", makeHeadFunc())
	lc.Define("tail", makeTailFunc())
	lc.Define("length", makeLengthFunc())
	lc.Define("concat", makeConcatFunc())
	lc.Define("take", makeTakeFunc())
	lc.Define("drop", makeDropFunc())

	// String operations
	lc.Define("split", makeSplitFunc())
	lc.Define("str_concat", makeStrConcatFunc())
	lc.Define("str_length", makeStrLengthFunc())
	lc.Define("join", makeJoinFunc())

	// Type conversion functions
	lc.Define("to_string", makeToStringFunc())
	lc.Define("to_float", makeToFloatFunc())
	lc.Define("to_int", makeToIntFunc())

	// Utility functions
	lc.Define("id", makeIdFunc())
	lc.Define("const", makeConstFunc())
	lc.Define("compose", makeComposeFunc())
	lc.Define("print", makePrintFunc(interp))
}

// Helper functions to create operator functions

func makePlusOp() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
		}

		// Try string concatenation first
		if s1, ok1 := args[0].(string); ok1 {
			if s2, ok2 := args[1].(string); ok2 {
				return s1 + s2, nil
			}
		}

		// Fall back to numeric addition
		a := toFloat64Safe(args[0])
		b := toFloat64Safe(args[1])

		if a == nil || b == nil {
			return nil, fmt.Errorf("operator + needs two numbers or two strings, got %s and %s", typeName(args[0]), typeName(args[1]))
		}

		result := *a + *b

		// Return int64 if both inputs were integers and result is whole
		if isInt(args[0]) && isInt(args[1]) && result == math.Floor(result) {
			return int64(result), nil
		}

		return result, nil
	}
}

func makeNumericOp(name string, op func(float64, float64) float64) func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
		}

		a := toFloat64Safe(args[0])
		b := toFloat64Safe(args[1])

		if a == nil || b == nil {
			return nil, fmt.Errorf("operator %s needs two numbers, got %s and %s", name, typeName(args[0]), typeName(args[1]))
		}

		result := op(*a, *b)

		// Return int64 if both inputs were integers and result is whole
		if isInt(args[0]) && isInt(args[1]) && result == math.Floor(result) {
			return int64(result), nil
		}

		return result, nil
	}
}

func makeComparisonOp(op func(interface{}, interface{}) bool) func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
		}
		return op(args[0], args[1]), nil
	}
}

func makeNumericComparisonOp(name string, op func(float64, float64) bool) func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
		}

		a := toFloat64Safe(args[0])
		b := toFloat64Safe(args[1])

		if a == nil || b == nil {
			return nil, fmt.Errorf("operator %s needs two numbers, got %s and %s", name, typeName(args[0]), typeName(args[1]))
		}

		return op(*a, *b), nil
	}
}

func makeLogicalOp(name string, op func(bool, bool) bool) func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
		}

		a, aOk := args[0].(bool)
		b, bOk := args[1].(bool)

		if !aOk || !bOk {
			return nil, fmt.Errorf("operator %s needs two booleans, got %s and %s", name, typeName(args[0]), typeName(args[1]))
		}

		return op(a, b), nil
	}
}

func makeUnaryLogicalOp(op func(bool) bool) func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument, got %d", len(args))
		}

		a, ok := args[0].(bool)
		if !ok {
			return nil, fmt.Errorf("operator not needs a boolean, got %s", typeName(args[0]))
		}

		return op(a), nil
	}
}

// List functions

func makeMapFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("map expects 2 arguments (function, list)")
		}

		fn := args[0]
		if !isCallable(fn) {
			return nil, fmt.Errorf("first argument to map must be a function, got %s", typeName(fn))
		}

		list, ok := args[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("second argument to map must be a list")
		}

		result := make([]interface{}, len(list))
		for i, elem := range list {
			val, err := callValue(fn, elem)
			if err != nil {
				return nil, err
			}
			result[i] = val
		}

		return result, nil
	}
}

func makeFilterFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("filter expects 2 arguments (function, list)")
		}

		fn := args[0]
		if !isCallable(fn) {
			return nil, fmt.Errorf("first argument to filter must be a function, got %s", typeName(fn))
		}

		list, ok := args[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("second argument to filter must be a list")
		}

		result := make([]interface{}, 0)
		for _, elem := range list {
			val, err := callValue(fn, elem)
			if err != nil {
				return nil, err
			}

			if b, ok := val.(bool); ok && b {
				result = append(result, elem)
			}
		}

		return result, nil
	}
}

func makeReduceFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("reduce expects 3 arguments (function, initial, list)")
		}

		fn := args[0]
		if !isCallable(fn) {
			return nil, fmt.Errorf("first argument to reduce must be a function, got %s", typeName(fn))
		}

		acc := args[1]

		list, ok := args[2].([]interface{})
		if !ok {
			return nil, fmt.Errorf("third argument to reduce must be a list")
		}

		for _, elem := range list {
			val, err := callValue(fn, acc, elem)
			if err != nil {
				return nil, err
			}
			acc = val
		}

		return acc, nil
	}
}

func makeHeadFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("head expects 1 argument (list)")
		}

		list, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("head requires a list argument")
		}

		if len(list) == 0 {
			return nil, fmt.Errorf("head called on empty list")
		}

		return list[0], nil
	}
}

func makeTailFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("tail expects 1 argument (list)")
		}

		list, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("tail requires a list argument")
		}

		if len(list) == 0 {
			return nil, fmt.Errorf("tail called on empty list")
		}

		return list[1:], nil
	}
}

func makeLengthFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("length expects 1 argument")
		}

		switch v := args[0].(type) {
		case []interface{}:
			return int64(len(v)), nil
		case string:
			return int64(len(v)), nil
		default:
			return nil, fmt.Errorf("length requires a list or string argument")
		}
	}
}

func makeConcatFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("concat expects 2 arguments")
		}

		// String concatenation
		if s1, ok1 := args[0].(string); ok1 {
			if s2, ok2 := args[1].(string); ok2 {
				return s1 + s2, nil
			}
		}

		// List concatenation
		if l1, ok1 := args[0].([]interface{}); ok1 {
			if l2, ok2 := args[1].([]interface{}); ok2 {
				result := make([]interface{}, len(l1)+len(l2))
				copy(result, l1)
				copy(result[len(l1):], l2)
				return result, nil
			}
		}

		return nil, fmt.Errorf("concat requires two strings or two lists")
	}
}

// Helper functions

func toFloat64Safe(val interface{}) *float64 {
	switch v := val.(type) {
	case int64:
		f := float64(v)
		return &f
	case float64:
		return &v
	case int:
		f := float64(v)
		return &f
	default:
		return nil
	}
}

func isInt(val interface{}) bool {
	_, ok := val.(int64)
	if ok {
		return true
	}
	_, ok = val.(int)
	return ok
}

func equals(a, b interface{}) bool {
	// Try numeric comparison
	aNum := toFloat64Safe(a)
	bNum := toFloat64Safe(b)
	if aNum != nil && bNum != nil {
		return math.Abs(*aNum-*bNum) < 1e-9
	}

	// Direct comparison
	return a == b
}

// String operations

func makeSplitFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("split expects 2 arguments (delimiter, string)")
		}

		delimiter, ok1 := args[0].(string)
		if !ok1 {
			return nil, fmt.Errorf("first argument to split must be a string")
		}

		str, ok2 := args[1].(string)
		if !ok2 {
			return nil, fmt.Errorf("second argument to split must be a string")
		}

		parts := splitString(str, delimiter)
		result := make([]interface{}, len(parts))
		for i, part := range parts {
			result[i] = part
		}

		return result, nil
	}
}

func splitString(str, delimiter string) []string {
	if delimiter == "" {
		// Split into characters
		result := make([]string, 0, len(str))
		for _, ch := range str {
			result = append(result, string(ch))
		}
		return result
	}

	// Use simple split algorithm
	var result []string
	start := 0
	for i := 0; i <= len(str)-len(delimiter); i++ {
		if str[i:i+len(delimiter)] == delimiter {
			result = append(result, str[start:i])
			start = i + len(delimiter)
			i += len(delimiter) - 1
		}
	}
	result = append(result, str[start:])
	return result
}

func makeStrConcatFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("str_concat expects 2 arguments")
		}

		s1, ok1 := args[0].(string)
		if !ok1 {
			return nil, fmt.Errorf("first argument to str_concat must be a string")
		}

		s2, ok2 := args[1].(string)
		if !ok2 {
			return nil, fmt.Errorf("second argument to str_concat must be a string")
		}

		return s1 + s2, nil
	}
}

// Type conversion functions

func makeToStringFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("to_string expects 1 argument")
		}

		switch v := args[0].(type) {
		case string:
			return v, nil
		case int64:
			return fmt.Sprintf("%d", v), nil
		case int:
			return fmt.Sprintf("%d", v), nil
		case float64:
			// Remove trailing zeros and decimal point if not needed
			s := fmt.Sprintf("%f", v)
			// Trim trailing zeros
			for len(s) > 0 && s[len(s)-1] == '0' {
				s = s[:len(s)-1]
			}
			// Trim trailing decimal point
			if len(s) > 0 && s[len(s)-1] == '.' {
				s = s[:len(s)-1]
			}
			return s, nil
		case bool:
			if v {
				return "true", nil
			}
			return "false", nil
		default:
			return fmt.Sprintf("%v", v), nil
		}
	}
}

func makeToFloatFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("to_float expects 1 argument")
		}

		switch v := args[0].(type) {
		case float64:
			return v, nil
		case int64:
			return float64(v), nil
		case int:
			return float64(v), nil
		case string:
			// Simple string to float parser
			result, err := parseFloat(v)
			if err != nil {
				return nil, fmt.Errorf("cannot convert '%s' to float: %v", v, err)
			}
			return result, nil
		default:
			return nil, fmt.Errorf("cannot convert %T to float", v)
		}
	}
}

func makeToIntFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("to_int expects 1 argument")
		}

		switch v := args[0].(type) {
		case int64:
			return v, nil
		case int:
			return int64(v), nil
		case float64:
			return int64(v), nil
		case string:
			// Simple string to int parser
			result, err := parseInt(v)
			if err != nil {
				return nil, fmt.Errorf("cannot convert '%s' to int: %v", v, err)
			}
			return result, nil
		default:
			return nil, fmt.Errorf("cannot convert %T to int", v)
		}
	}
}

// Simple string to number parsers (avoiding strconv dependency for WASM)

func parseFloat(s string) (float64, error) {
	s = trimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}

	negative := false
	if s[0] == '-' {
		negative = true
		s = s[1:]
	} else if s[0] == '+' {
		s = s[1:]
	}

	var result float64
	var fraction float64
	var divisor float64 = 1
	inFraction := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '.' {
			if inFraction {
				return 0, fmt.Errorf("multiple decimal points")
			}
			inFraction = true
			continue
		}

		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid character '%c'", ch)
		}

		digit := float64(ch - '0')
		if inFraction {
			divisor *= 10
			fraction = fraction*10 + digit
		} else {
			result = result*10 + digit
		}
	}

	result += fraction / divisor

	if negative {
		result = -result
	}

	return result, nil
}

func parseInt(s string) (int64, error) {
	s = trimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}

	negative := false
	if s[0] == '-' {
		negative = true
		s = s[1:]
	} else if s[0] == '+' {
		s = s[1:]
	}

	var result int64
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid character '%c'", ch)
		}
		result = result*10 + int64(ch-'0')
	}

	if negative {
		result = -result
	}

	return result, nil
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

// Additional list operations

func makeFoldFunc() func(...interface{}) (interface{}, error) {
	// fold is just an alias for reduce
	return makeReduceFunc()
}

func makeTakeFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("take expects 2 arguments (n, list)")
		}

		n, ok := args[0].(int64)
		if !ok {
			return nil, fmt.Errorf("first argument to take must be an integer")
		}

		list, ok := args[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("second argument to take must be a list")
		}

		if n < 0 {
			n = 0
		}
		if n > int64(len(list)) {
			n = int64(len(list))
		}

		return list[:n], nil
	}
}

func makeDropFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("drop expects 2 arguments (n, list)")
		}

		n, ok := args[0].(int64)
		if !ok {
			return nil, fmt.Errorf("first argument to drop must be an integer")
		}

		list, ok := args[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("second argument to drop must be a list")
		}

		if n < 0 {
			n = 0
		}
		if n > int64(len(list)) {
			return []interface{}{}, nil
		}

		return list[n:], nil
	}
}

// Additional string operations

func makeStrLengthFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("str_length expects 1 argument")
		}

		str, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("str_length requires a string argument")
		}

		return int64(len(str)), nil
	}
}

func makeJoinFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("join expects 2 arguments (delimiter, list)")
		}

		delimiter, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("first argument to join must be a string")
		}

		list, ok := args[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("second argument to join must be a list")
		}

		parts := make([]string, len(list))
		for i, item := range list {
			parts[i] = fmt.Sprintf("%v", item)
		}

		return joinStrings(parts, delimiter), nil
	}
}

func joinStrings(parts []string, delimiter string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}

	// Calculate total length
	totalLen := 0
	for _, part := range parts {
		totalLen += len(part)
	}
	totalLen += len(delimiter) * (len(parts) - 1)

	// Build result
	result := make([]byte, 0, totalLen)
	result = append(result, parts[0]...)
	for i := 1; i < len(parts); i++ {
		result = append(result, delimiter...)
		result = append(result, parts[i]...)
	}

	return string(result)
}

// Utility functions

func makeIdFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("id expects 1 argument")
		}
		return args[0], nil
	}
}

func makeConstFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("const expects 1 argument")
		}

		value := args[0]

		// A one-argument function that ignores its argument.
		return &Closure{
			Params: []string{"_"},
			Native: func(...interface{}) (interface{}, error) { return value, nil },
		}, nil
	}
}

func makeComposeFunc() func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("compose expects 2 arguments (f, g)")
		}

		f, g := args[0], args[1]
		if !isCallable(f) {
			return nil, fmt.Errorf("first argument to compose must be a function, got %s", typeName(f))
		}
		if !isCallable(g) {
			return nil, fmt.Errorf("second argument to compose must be a function, got %s", typeName(g))
		}

		// compose(f, g)(x) = f(g(x))
		return &Closure{
			Params: []string{"x"},
			Native: func(args ...interface{}) (interface{}, error) {
				inner, err := callValue(g, args[0])
				if err != nil {
					return nil, err
				}
				return callValue(f, inner)
			},
		}, nil
	}
}

// makePrintFunc writes a value to the interpreter's output (stderr for the
// CLI, the Output panel in the editor) and returns it unchanged.
func makePrintFunc(interp *Interpreter) func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("print expects 1 argument")
		}
		interp.emit(Show(args[0]))
		return args[0], nil
	}
}
