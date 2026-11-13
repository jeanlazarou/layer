//go:build js && wasm

package interpreter

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/jeanlazarou/layer/pkg/layers"
)

// In-memory localStorage for fallback
var localStorage = make(map[string]interface{})

// SetupDataLayer sets up the data layer with browser-based I/O and storage functions
func SetupDataLayer(lc *layers.LayerContext) {
	lc.SetCurrentLayer("data")

	// ========== Option Type Constructors ==========

	lc.Define("Some", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("Some expects 1 argument")
		}
		return map[string]interface{}{
			"type":  "Some",
			"value": args[0],
		}, nil
	})

	lc.Define("None", func(args ...interface{}) (interface{}, error) {
		return map[string]interface{}{
			"type": "None",
		}, nil
	})

	// ========== Option Helpers ==========

	lc.Define("is_some", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("is_some expects 1 argument")
		}

		if opt, ok := args[0].(map[string]interface{}); ok {
			return opt["type"] == "Some", nil
		}
		return false, nil
	})

	lc.Define("is_none", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("is_none expects 1 argument")
		}

		if opt, ok := args[0].(map[string]interface{}); ok {
			return opt["type"] == "None", nil
		}
		return false, nil
	})

	lc.Define("unwrap", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("unwrap expects 2 arguments (option, default)")
		}

		if opt, ok := args[0].(map[string]interface{}); ok {
			if opt["type"] == "Some" {
				return opt["value"], nil
			}
		}
		return args[1], nil
	})

	lc.Define("map_option", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("map_option expects 2 arguments (fn, option)")
		}

		fn := args[0]
		if !isCallable(fn) {
			return nil, fmt.Errorf("first argument to map_option must be a function, got %s", typeName(fn))
		}

		if opt, ok := args[1].(map[string]interface{}); ok {
			if opt["type"] == "Some" {
				result, err := callValue(fn, opt["value"])
				if err != nil {
					return map[string]interface{}{"type": "None"}, nil
				}
				return map[string]interface{}{
					"type":  "Some",
					"value": result,
				}, nil
			}
		}
		return map[string]interface{}{"type": "None"}, nil
	})

	lc.Define("flatmap_option", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("flatmap_option expects 2 arguments (fn, option)")
		}

		fn := args[0]
		if !isCallable(fn) {
			return nil, fmt.Errorf("first argument to flatmap_option must be a function, got %s", typeName(fn))
		}

		if opt, ok := args[1].(map[string]interface{}); ok {
			if opt["type"] == "Some" {
				result, err := callValue(fn, opt["value"])
				if err != nil {
					return map[string]interface{}{"type": "None"}, nil
				}
				return result, nil
			}
		}
		return map[string]interface{}{"type": "None"}, nil
	})

	// ========== Browser Storage Mode Control ==========

	lc.Define("request_file_access", func(args ...interface{}) (interface{}, error) {
		fs := js.Global().Get("layerFileSystem")
		if fs.IsUndefined() {
			return false, fmt.Errorf("layerFileSystem not available")
		}

		promise := fs.Call("requestAccess")
		result := await(promise)

		return result.Bool(), nil
	})

	lc.Define("set_storage_mode", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("set_storage_mode expects 1 argument")
		}

		mode := fmt.Sprintf("%v", args[0])
		if mode != "virtual" && mode != "real" {
			return nil, fmt.Errorf("mode must be 'virtual' or 'real'")
		}

		fs := js.Global().Get("layerFileSystem")
		if fs.IsUndefined() {
			return map[string]interface{}{"type": "None"}, nil
		}

		fs.Call("setMode", mode)
		return map[string]interface{}{"type": "None"}, nil
	})

	lc.Define("get_storage_mode", func(args ...interface{}) (interface{}, error) {
		fs := js.Global().Get("layerFileSystem")
		if fs.IsUndefined() {
			return "virtual", nil
		}

		mode := fs.Call("getMode")
		return mode.String(), nil
	})

	// ========== File Operations (Browser-based) ==========

	lc.Define("read_file", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("read_file expects 1 argument (path)")
		}

		path := fmt.Sprintf("%v", args[0])

		fs := js.Global().Get("layerFileSystem")
		if fs.IsUndefined() {
			return map[string]interface{}{"type": "None"}, nil
		}

		// Use synchronous version to avoid deadlock
		result := fs.Call("readFileSync", path)

		return jsValueToOption(result), nil
	})

	lc.Define("write_file", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("write_file expects 2 arguments (path, content)")
		}

		path := fmt.Sprintf("%v", args[0])
		content := fmt.Sprintf("%v", args[1])

		fs := js.Global().Get("layerFileSystem")
		if fs.IsUndefined() {
			return map[string]interface{}{"type": "None"}, nil
		}

		// Use synchronous version to avoid deadlock
		result := fs.Call("writeFileSync", path, content)

		return jsValueToOption(result), nil
	})

	lc.Define("exists", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("exists expects 1 argument (path)")
		}

		path := fmt.Sprintf("%v", args[0])

		fs := js.Global().Get("layerFileSystem")
		if fs.IsUndefined() {
			return false, nil
		}

		// Use synchronous version to avoid deadlock
		result := fs.Call("existsSync", path)

		return result.Bool(), nil
	})

	lc.Define("delete", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("delete expects 1 argument (path)")
		}

		path := fmt.Sprintf("%v", args[0])

		fs := js.Global().Get("layerFileSystem")
		if fs.IsUndefined() {
			return map[string]interface{}{"type": "None"}, nil
		}

		// Use synchronous version to avoid deadlock
		result := fs.Call("deleteFileSync", path)

		return jsValueToOption(result), nil
	})

	lc.Define("list_files", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("list_files expects 1 argument (directory)")
		}

		path := fmt.Sprintf("%v", args[0])

		fs := js.Global().Get("layerFileSystem")
		if fs.IsUndefined() {
			return map[string]interface{}{"type": "None"}, nil
		}

		// Use synchronous version to avoid deadlock
		result := fs.Call("listFilesSync", path)

		return jsValueToOption(result), nil
	})

	// ========== JSON Operations ==========

	lc.Define("read_json", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("read_json expects 1 argument (path)")
		}

		path := fmt.Sprintf("%v", args[0])

		fs := js.Global().Get("layerFileSystem")
		if fs.IsUndefined() {
			return map[string]interface{}{"type": "None"}, nil
		}

		// Use synchronous version to avoid deadlock
		result := fs.Call("readFileSync", path)

		option := jsValueToOption(result)
		if opt, ok := option.(map[string]interface{}); ok && opt["type"] == "Some" {
			jsonStr := fmt.Sprintf("%v", opt["value"])

			var data interface{}
			err := json.Unmarshal([]byte(jsonStr), &data)
			if err != nil {
				return map[string]interface{}{"type": "None"}, nil
			}

			return map[string]interface{}{
				"type":  "Some",
				"value": data,
			}, nil
		}

		return map[string]interface{}{"type": "None"}, nil
	})

	lc.Define("write_json", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("write_json expects 2 arguments (path, data)")
		}

		path := fmt.Sprintf("%v", args[0])
		data := args[1]

		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		fs := js.Global().Get("layerFileSystem")
		if fs.IsUndefined() {
			return map[string]interface{}{"type": "None"}, nil
		}

		// Use synchronous version to avoid deadlock
		result := fs.Call("writeFileSync", path, string(jsonBytes))

		option := jsValueToOption(result)
		if opt, ok := option.(map[string]interface{}); ok && opt["type"] == "Some" {
			return map[string]interface{}{
				"type":  "Some",
				"value": data,
			}, nil
		}

		return map[string]interface{}{"type": "None"}, nil
	})

	lc.Define("parse_json", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("parse_json expects 1 argument (json_string)")
		}

		jsonStr := fmt.Sprintf("%v", args[0])

		var data interface{}
		err := json.Unmarshal([]byte(jsonStr), &data)
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		return map[string]interface{}{
			"type":  "Some",
			"value": data,
		}, nil
	})

	lc.Define("to_json", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("to_json expects 1 argument (data)")
		}

		jsonBytes, err := json.MarshalIndent(args[0], "", "  ")
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		return map[string]interface{}{
			"type":  "Some",
			"value": string(jsonBytes),
		}, nil
	})

	// ========== localStorage Operations (In-memory fallback) ==========

	lc.Define("get", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("get expects 1 argument (key)")
		}

		key := fmt.Sprintf("%v", args[0])

		if value, ok := localStorage[key]; ok {
			return map[string]interface{}{
				"type":  "Some",
				"value": value,
			}, nil
		}

		return map[string]interface{}{"type": "None"}, nil
	})

	lc.Define("set", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("set expects 2 arguments (key, value)")
		}

		key := fmt.Sprintf("%v", args[0])
		value := args[1]

		localStorage[key] = value

		return map[string]interface{}{
			"type":  "Some",
			"value": value,
		}, nil
	})

	lc.Define("delete_key", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("delete_key expects 1 argument (key)")
		}

		key := fmt.Sprintf("%v", args[0])

		if _, ok := localStorage[key]; ok {
			delete(localStorage, key)
			return map[string]interface{}{
				"type":  "Some",
				"value": key,
			}, nil
		}

		return map[string]interface{}{"type": "None"}, nil
	})

	lc.Define("keys", func(args ...interface{}) (interface{}, error) {
		keys := make([]interface{}, 0, len(localStorage))
		for k := range localStorage {
			keys = append(keys, k)
		}
		return keys, nil
	})

	lc.Define("clear", func(args ...interface{}) (interface{}, error) {
		localStorage = make(map[string]interface{})
		return map[string]interface{}{"type": "None"}, nil
	})
}

// Helper: Await a JavaScript Promise (simplified synchronous version)
func await(promise js.Value) js.Value {
	// Create buffered channels to prevent blocking
	done := make(chan js.Value, 1)

	// Create callbacks and ensure they're released
	var thenFunc, catchFunc js.Func

	thenFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			done <- args[0]
		} else {
			done <- js.Undefined()
		}
		// Release callbacks after use
		thenFunc.Release()
		catchFunc.Release()
		return nil
	})

	catchFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			fmt.Printf("Promise error: %v\n", args[0])
		}
		done <- js.Undefined()
		// Release callbacks after use
		thenFunc.Release()
		catchFunc.Release()
		return nil
	})

	// Attach handlers to promise
	promise.Call("then", thenFunc).Call("catch", catchFunc)

	// Wait for result
	return <-done
}

// Helper: Convert JavaScript Option object to Go map
func jsValueToOption(jsVal js.Value) interface{} {
	if jsVal.IsUndefined() || jsVal.IsNull() {
		return map[string]interface{}{"type": "None"}
	}

	typ := jsVal.Get("type")
	if typ.IsUndefined() {
		return map[string]interface{}{"type": "None"}
	}

	typeStr := typ.String()

	if typeStr == "None" {
		return map[string]interface{}{"type": "None"}
	}

	if typeStr == "Some" {
		value := jsVal.Get("value")
		return map[string]interface{}{
			"type":  "Some",
			"value": jsValueToGo(value),
		}
	}

	return map[string]interface{}{"type": "None"}
}

// Helper: Convert JavaScript value to Go value
func jsValueToGo(jsVal js.Value) interface{} {
	switch jsVal.Type() {
	case js.TypeString:
		return jsVal.String()
	case js.TypeNumber:
		return jsVal.Float()
	case js.TypeBoolean:
		return jsVal.Bool()
	case js.TypeObject:
		// Check if it's an array
		if jsVal.Get("length").Type() == js.TypeNumber {
			length := jsVal.Get("length").Int()
			arr := make([]interface{}, length)
			for i := 0; i < length; i++ {
				arr[i] = jsValueToGo(jsVal.Index(i))
			}
			return arr
		}
		// Otherwise it's an object - convert to map
		obj := make(map[string]interface{})
		keys := js.Global().Get("Object").Call("keys", jsVal)
		keysLen := keys.Get("length").Int()
		for i := 0; i < keysLen; i++ {
			key := keys.Index(i).String()
			obj[key] = jsValueToGo(jsVal.Get(key))
		}
		return obj
	default:
		return nil
	}
}
