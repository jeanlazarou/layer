//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/yourusername/layer/pkg/interpreter"
	"github.com/yourusername/layer/pkg/lexer"
	"github.com/yourusername/layer/pkg/parser"
)

// ExecuteResult represents the result of executing Layer code
type ExecuteResult struct {
	Success bool     `json:"success"`
	Result  string   `json:"result,omitempty"`
	Logs    []string `json:"logs,omitempty"`
	Error   string   `json:"error,omitempty"`
}

var logBuffer []string

// captureLog captures log output
func captureLog(message string) {
	logBuffer = append(logBuffer, message)
}

// executeLayer executes Layer source code
func executeLayer(this js.Value, args []js.Value) interface{} {
	// Reset log buffer
	logBuffer = make([]string, 0)

	if len(args) != 1 {
		return makeErrorResult("execute expects 1 argument (source code)")
	}

	source := args[0].String()

	// Create result object
	result := ExecuteResult{
		Success: true,
		Logs:    make([]string, 0),
	}

	// Tokenize
	tokens, err := lexer.Tokenize(source)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Lexer error: %v", err)
		return resultToJS(result)
	}

	// Parse
	program, err := parser.Parse(tokens)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Parser error: %v", err)
		return resultToJS(result)
	}

	// Interpret
	interp := interpreter.NewInterpreter()
	value, err := interp.Execute(program)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Runtime error: %v", err)
		return resultToJS(result)
	}

	// Format result
	result.Result = formatValue(value)
	result.Logs = logBuffer

	return resultToJS(result)
}

// parseLayerFile parses a .layer file and returns structured data
func parseLayerFile(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return makeErrorResult("parseLayerFile expects 1 argument (source code)")
	}

	source := args[0].String()

	// Extract layer information using text parsing
	result := map[string]interface{}{
		"success": true,
	}

	// Parse @layers directive
	enabledLayers := extractEnabledLayers(source)
	result["enabledLayers"] = enabledLayers

	// Parse @layer blocks
	layerCode := extractLayerCode(source)
	result["layers"] = layerCode

	// Extract main expression
	mainExpr := extractMainExpression(source)
	result["mainExpression"] = mainExpr

	return mapToJS(result)
}

// formatValue formats a value for display
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return "nil"
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
		return fmt.Sprintf("%v", v)
	case *interpreter.Closure:
		return v.String()
	case *interpreter.UIComponent:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// makeErrorResult creates an error result
func makeErrorResult(errMsg string) interface{} {
	result := ExecuteResult{
		Success: false,
		Error:   errMsg,
	}
	return resultToJS(result)
}

// resultToJS converts ExecuteResult to JS object
func resultToJS(result ExecuteResult) interface{} {
	obj := js.Global().Get("Object").New()
	obj.Set("success", result.Success)

	if result.Result != "" {
		obj.Set("result", result.Result)
	}

	if len(result.Logs) > 0 {
		logsArray := js.Global().Get("Array").New()
		for i, log := range result.Logs {
			logsArray.SetIndex(i, log)
		}
		obj.Set("logs", logsArray)
	} else {
		obj.Set("logs", js.Global().Get("Array").New())
	}

	if result.Error != "" {
		obj.Set("error", result.Error)
	}

	return obj
}

// mapToJS converts a Go map to JS object
func mapToJS(m map[string]interface{}) interface{} {
	obj := js.Global().Get("Object").New()
	for key, value := range m {
		switch v := value.(type) {
		case string:
			obj.Set(key, v)
		case []string:
			arr := js.Global().Get("Array").New()
			for i, s := range v {
				arr.SetIndex(i, s)
			}
			obj.Set(key, arr)
		case map[string]string:
			subObj := js.Global().Get("Object").New()
			for k, val := range v {
				subObj.Set(k, val)
			}
			obj.Set(key, subObj)
		case bool:
			obj.Set(key, v)
		default:
			obj.Set(key, fmt.Sprintf("%v", v))
		}
	}
	return obj
}

func main() {
	// Keep the program running
	c := make(chan struct{})

	// Register functions
	js.Global().Set("executeLayer", js.FuncOf(executeLayer))
	js.Global().Set("parseLayerFile", js.FuncOf(parseLayerFile))

	fmt.Println("Layer WASM module loaded successfully!")

	// Wait forever
	<-c
}
