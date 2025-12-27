// +build !js,!wasm
//go:build !js && !wasm

package interpreter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourusername/layer/pkg/layers"
)

// In-memory localStorage for WASM/non-browser environments
var localStorage = make(map[string]interface{})

// SetupDataLayer sets up the data layer with I/O and storage functions
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

		fn, ok := args[0].(*Closure)
		if !ok {
			return nil, fmt.Errorf("first argument to map_option must be a function")
		}

		if opt, ok := args[1].(map[string]interface{}); ok {
			if opt["type"] == "Some" {
				result, err := fn.Call(opt["value"])
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

		fn, ok := args[0].(*Closure)
		if !ok {
			return nil, fmt.Errorf("first argument to flatmap_option must be a function")
		}

		if opt, ok := args[1].(map[string]interface{}); ok {
			if opt["type"] == "Some" {
				result, err := fn.Call(opt["value"])
				if err != nil {
					return map[string]interface{}{"type": "None"}, nil
				}
				return result, nil
			}
		}
		return map[string]interface{}{"type": "None"}, nil
	})

	// ========== File Operations ==========

	lc.Define("read_file", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("read_file expects 1 argument (path)")
		}

		path := fmt.Sprintf("%v", args[0])
		absPath, err := filepath.Abs(path)
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		content, err := os.ReadFile(absPath)
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		return map[string]interface{}{
			"type":  "Some",
			"value": string(content),
		}, nil
	})

	lc.Define("write_file", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("write_file expects 2 arguments (path, content)")
		}

		path := fmt.Sprintf("%v", args[0])
		content := fmt.Sprintf("%v", args[1])

		absPath, err := filepath.Abs(path)
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		// Create directory if it doesn't exist
		dir := filepath.Dir(absPath)
		if dir != "" {
			os.MkdirAll(dir, 0755)
		}

		err = os.WriteFile(absPath, []byte(content), 0644)
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		return map[string]interface{}{
			"type":  "Some",
			"value": content,
		}, nil
	})

	lc.Define("exists", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("exists expects 1 argument (path)")
		}

		path := fmt.Sprintf("%v", args[0])
		absPath, err := filepath.Abs(path)
		if err != nil {
			return false, nil
		}

		_, err = os.Stat(absPath)
		return err == nil, nil
	})

	lc.Define("delete", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("delete expects 1 argument (path)")
		}

		path := fmt.Sprintf("%v", args[0])
		absPath, err := filepath.Abs(path)
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		// Check if it's a file
		info, err := os.Stat(absPath)
		if err != nil || info.IsDir() {
			return map[string]interface{}{"type": "None"}, nil
		}

		err = os.Remove(absPath)
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		return map[string]interface{}{
			"type":  "Some",
			"value": path,
		}, nil
	})

	lc.Define("list_files", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("list_files expects 1 argument (directory)")
		}

		path := fmt.Sprintf("%v", args[0])
		absPath, err := filepath.Abs(path)
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		// Check if it's a directory
		info, err := os.Stat(absPath)
		if err != nil || !info.IsDir() {
			return map[string]interface{}{"type": "None"}, nil
		}

		entries, err := os.ReadDir(absPath)
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		files := make([]interface{}, 0, len(entries))
		for _, entry := range entries {
			files = append(files, entry.Name())
		}

		return map[string]interface{}{
			"type":  "Some",
			"value": files,
		}, nil
	})

	// ========== JSON Operations ==========

	lc.Define("read_json", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("read_json expects 1 argument (path)")
		}

		path := fmt.Sprintf("%v", args[0])
		absPath, err := filepath.Abs(path)
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		content, err := os.ReadFile(absPath)
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		var data interface{}
		err = json.Unmarshal(content, &data)
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		return map[string]interface{}{
			"type":  "Some",
			"value": data,
		}, nil
	})

	lc.Define("write_json", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("write_json expects 2 arguments (path, data)")
		}

		path := fmt.Sprintf("%v", args[0])
		data := args[1]

		absPath, err := filepath.Abs(path)
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		// Create directory if it doesn't exist
		dir := filepath.Dir(absPath)
		if dir != "" {
			os.MkdirAll(dir, 0755)
		}

		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		err = os.WriteFile(absPath, jsonBytes, 0644)
		if err != nil {
			return map[string]interface{}{"type": "None"}, nil
		}

		return map[string]interface{}{
			"type":  "Some",
			"value": data,
		}, nil
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

	// ========== localStorage Operations ==========

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
