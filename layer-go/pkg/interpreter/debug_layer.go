package interpreter

import (
	"fmt"
	"time"

	"github.com/jeanlazarou/layer/pkg/layers"
)

// SetupDebugLayer sets up the debug layer with logging and introspection
// functions. Every line they produce goes through interp's log sink, so it
// reaches stderr in the CLI and the Output panel in the visual editor.
func SetupDebugLayer(lc *layers.LayerContext, interp *Interpreter) {
	lc.SetCurrentLayer("debug")

	// ========== Logging Functions ==========

	// log(value) - logs the value and returns it, so it can wrap any
	// expression: log(compute(x))
	lc.Define("log", func(args ...interface{}) (interface{}, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("log expects at least 1 argument")
		}
		interp.emit("[INFO] " + Show(args[0]))
		return args[0], nil
	})

	// trace(name, fn) - wraps a function so each call and result is logged
	lc.Define("trace", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("trace expects 2 arguments (name, function)")
		}

		name := Show(args[0])
		fn, ok := args[1].(*Closure)
		if !ok {
			return nil, fmt.Errorf("second argument to trace must be a function, got %s", typeName(args[1]))
		}

		return &Closure{
			Params: fn.Params,
			Native: func(callArgs ...interface{}) (interface{}, error) {
				argsStr := ""
				for i, arg := range callArgs {
					if i > 0 {
						argsStr += ", "
					}
					argsStr += ShowNested(arg)
				}
				interp.emit(fmt.Sprintf("[TRACE] %s(%s) called", name, argsStr))
				result, err := fn.Call(callArgs...)
				if err != nil {
					return nil, err
				}
				interp.emit(fmt.Sprintf("[TRACE] %s(%s) = %s", name, argsStr, ShowNested(result)))
				return result, nil
			},
		}, nil
	})

	// inspect function - detailed value inspection
	lc.Define("inspect", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("inspect expects 1 argument")
		}
		interp.emit("[DEBUG] Inspect: " + inspectValue(args[0]))
		return args[0], nil
	})

	// ========== Assertion Functions ==========

	lc.Define("assert_equal", func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 || len(args) > 3 {
			return nil, fmt.Errorf("assert_equal expects 2 or 3 arguments (expected, actual, [message])")
		}

		expected := args[0]
		actual := args[1]
		message := "Assertion failed"
		if len(args) == 3 {
			message = Show(args[2])
		}

		if !valuesEqual(expected, actual) {
			errorMsg := fmt.Sprintf("%s: expected %s, got %s", message, ShowNested(expected), ShowNested(actual))
			interp.emit("[ERROR] " + errorMsg)
			return nil, fmt.Errorf("%s", errorMsg)
		}

		interp.emit(fmt.Sprintf("[DEBUG] Assertion passed: %s == %s", ShowNested(expected), ShowNested(actual)))
		return actual, nil
	})

	// ========== Performance Functions ==========

	lc.Define("benchmark", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("benchmark expects 2 arguments (name, function)")
		}

		name := Show(args[0])
		fn, ok := args[1].(*Closure)
		if !ok {
			return nil, fmt.Errorf("second argument to benchmark must be a function, got %s", typeName(args[1]))
		}

		return &Closure{
			Params: fn.Params,
			Native: func(callArgs ...interface{}) (interface{}, error) {
				start := time.Now()
				result, err := fn.Call(callArgs...)
				elapsed := time.Since(start)
				if err != nil {
					return nil, err
				}
				interp.emit(fmt.Sprintf("[BENCH] %s took %.3fms", name, float64(elapsed.Microseconds())/1000.0))
				return result, nil
			},
		}, nil
	})

	// ========== Debugging Utilities ==========

	lc.Define("breakpoint", func(args ...interface{}) (interface{}, error) {
		message := "Breakpoint"
		if len(args) > 0 {
			message = Show(args[0])
		}
		interp.emit("[BREAK] BREAKPOINT: " + message)
		return map[string]interface{}{"type": "None"}, nil
	})

	lc.Define("get_logs", func(args ...interface{}) (interface{}, error) {
		logs := make([]interface{}, len(interp.logs))
		for i, line := range interp.logs {
			logs[i] = line
		}
		return logs, nil
	})

	lc.Define("clear_logs", func(args ...interface{}) (interface{}, error) {
		interp.logs = nil
		return map[string]interface{}{"type": "None"}, nil
	})
}

// inspectValue provides detailed inspection information
func inspectValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return fmt.Sprintf("type: string, length: %d, value: %q", len(v), v)
	case int64:
		return fmt.Sprintf("type: int64, value: %d", v)
	case float64:
		return fmt.Sprintf("type: float64, value: %g", v)
	case bool:
		return fmt.Sprintf("type: bool, value: %t", v)
	case []interface{}:
		return fmt.Sprintf("type: list, length: %d, elements: %v", len(v), v)
	case map[string]interface{}:
		// Check if it's an Option type
		if typeVal, ok := v["type"]; ok {
			if typeVal == "Some" {
				return fmt.Sprintf("type: Option<Some>, value: %v", v["value"])
			} else if typeVal == "None" {
				return "type: Option<None>"
			}
		}
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		return fmt.Sprintf("type: map, size: %d, keys: %v", len(v), keys)
	case *Closure:
		return fmt.Sprintf("type: function, params: %v", v.Params)
	case nil:
		return "type: nil"
	default:
		return fmt.Sprintf("type: %T, value: %v", v, v)
	}
}

// valuesEqual compares two values for equality
func valuesEqual(a, b interface{}) bool {
	// Handle numeric comparisons
	aNum := toFloat64Safe(a)
	bNum := toFloat64Safe(b)
	if aNum != nil && bNum != nil {
		return *aNum == *bNum
	}

	// Handle string comparisons
	aStr, aIsStr := a.(string)
	bStr, bIsStr := b.(string)
	if aIsStr && bIsStr {
		return aStr == bStr
	}

	// Handle bool comparisons
	aBool, aIsBool := a.(bool)
	bBool, bIsBool := b.(bool)
	if aIsBool && bIsBool {
		return aBool == bBool
	}

	// Handle list comparisons
	aList, aIsList := a.([]interface{})
	bList, bIsList := b.([]interface{})
	if aIsList && bIsList {
		if len(aList) != len(bList) {
			return false
		}
		for i := range aList {
			if !valuesEqual(aList[i], bList[i]) {
				return false
			}
		}
		return true
	}

	// Direct comparison
	return a == b
}
