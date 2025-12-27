package interpreter

import (
	"fmt"
	"os"
	"time"

	"github.com/yourusername/layer/pkg/layers"
)

// Global debug logs storage
var debugLogs = make([]string, 0)

// SetupDebugLayer sets up the debug layer with logging and introspection functions
func SetupDebugLayer(lc *layers.LayerContext) {
	lc.SetCurrentLayer("debug")

	// ========== Logging Functions ==========

	// log function - writes to stderr and returns the value
	lc.Define("log", func(args ...interface{}) (interface{}, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("log expects at least 1 argument")
		}

		timestamp := time.Now().Format("15:04:05.000")
		logMsg := fmt.Sprintf("[%s] [INFO] %s", timestamp, formatValue(args[0]))

		// Print to stderr
		fmt.Fprintln(os.Stderr, logMsg)

		// Store in logs
		debugLogs = append(debugLogs, logMsg)

		// Return the first argument for chaining
		return args[0], nil
	})

	// trace function - wraps a function with tracing
	lc.Define("trace", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("trace expects 2 arguments (name, function)")
		}

		name := fmt.Sprintf("%v", args[0])
		fn, ok := args[1].(*Closure)
		if !ok {
			return nil, fmt.Errorf("second argument to trace must be a function")
		}

		// Return a new closure that wraps the original with tracing
		return &Closure{
			Params:      fn.Params,
			Body:        nil, // Special marker for trace function
			Env: map[string]interface{}{
				"__trace_name": name,
				"__trace_fn":   fn,
			},
			Interpreter: fn.Interpreter,
		}, nil
	})

	// inspect function - detailed value inspection
	lc.Define("inspect", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("inspect expects 1 argument")
		}

		timestamp := time.Now().Format("15:04:05.000")
		info := inspectValue(args[0])
		logMsg := fmt.Sprintf("[%s] [DEBUG] Inspect: %s", timestamp, info)

		fmt.Fprintln(os.Stderr, logMsg)
		debugLogs = append(debugLogs, logMsg)

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
			message = fmt.Sprintf("%v", args[2])
		}

		timestamp := time.Now().Format("15:04:05.000")

		if !valuesEqual(expected, actual) {
			errorMsg := fmt.Sprintf("%s: expected %v, got %v", message, expected, actual)
			logMsg := fmt.Sprintf("[%s] [ERROR] %s", timestamp, errorMsg)
			fmt.Fprintln(os.Stderr, logMsg)
			debugLogs = append(debugLogs, logMsg)
			return nil, fmt.Errorf("%s", errorMsg)
		}

		logMsg := fmt.Sprintf("[%s] [DEBUG] Assertion passed: %v == %v", timestamp, expected, actual)
		fmt.Fprintln(os.Stderr, logMsg)
		debugLogs = append(debugLogs, logMsg)

		return actual, nil
	})

	// ========== Performance Functions ==========

	lc.Define("benchmark", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("benchmark expects 2 arguments (name, function)")
		}

		name := fmt.Sprintf("%v", args[0])
		fn, ok := args[1].(*Closure)
		if !ok {
			return nil, fmt.Errorf("second argument to benchmark must be a function")
		}

		// Return a new closure that wraps the original with benchmarking
		return &Closure{
			Params:      fn.Params,
			Body:        nil, // Special marker for benchmark function
			Env: map[string]interface{}{
				"__benchmark_name": name,
				"__benchmark_fn":   fn,
			},
			Interpreter: fn.Interpreter,
		}, nil
	})

	// ========== Debugging Utilities ==========

	lc.Define("breakpoint", func(args ...interface{}) (interface{}, error) {
		message := "Breakpoint"
		if len(args) > 0 {
			message = fmt.Sprintf("%v", args[0])
		}

		timestamp := time.Now().Format("15:04:05.000")
		logMsg := fmt.Sprintf("[%s] [BREAK] BREAKPOINT: %s", timestamp, message)

		fmt.Fprintln(os.Stderr, logMsg)
		debugLogs = append(debugLogs, logMsg)

		return map[string]interface{}{"type": "None"}, nil
	})

	lc.Define("get_logs", func(args ...interface{}) (interface{}, error) {
		// Convert to []interface{} for Layer compatibility
		logs := make([]interface{}, len(debugLogs))
		for i, log := range debugLogs {
			logs[i] = log
		}
		return logs, nil
	})

	lc.Define("clear_logs", func(args ...interface{}) (interface{}, error) {
		debugLogs = make([]string, 0)
		return map[string]interface{}{"type": "None"}, nil
	})
}

// formatValue formats a value for display
func formatValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case []interface{}:
		return fmt.Sprintf("%v", v)
	case map[string]interface{}:
		// Check if it's an Option type
		if typeVal, ok := v["type"]; ok {
			if typeVal == "Some" {
				return fmt.Sprintf("Some(%v)", v["value"])
			} else if typeVal == "None" {
				return "None"
			}
		}
		return fmt.Sprintf("%v", v)
	case *Closure:
		return v.String()
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("%v", v)
	}
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
