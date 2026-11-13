//go:build js && wasm

// Command wasm exposes the Layer interpreter to JavaScript. Every exported
// function takes source text and returns a JSON string, so the browser side
// only has to JSON.parse it:
//
//	executeLayer(source, varsJSON?) → runner.Result
//	describeLayer(source)          → runner.DescribeResult
//	parseLayerFile(source)         → {success, errors, enabledLayers, declaredLayers, layers, purity, blendMode, mainExpression}
//
// varsJSON is an optional JSON object of top-level variables (the current
// values of UI inputs). Integral numbers become Layer integers.
package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/jeanlazarou/layer/pkg/runner"
	"github.com/jeanlazarou/layer/pkg/sourcefile"
)

func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf(`{"success":false,"errors":[{"kind":"internal","message":%q}]}`, err.Error())
	}
	return string(b)
}

func errorJSON(msg string) string {
	return toJSON(map[string]interface{}{
		"success": false,
		"errors":  []runner.Issue{{Kind: "internal", Message: msg}},
	})
}

// decodeVars parses a JSON object of variables, mapping integral numbers to
// int64 so `square(n)` on a typed-in "5" yields 25, not 25.0.
func decodeVars(s string) (map[string]interface{}, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var raw map[string]interface{}
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}
	vars := make(map[string]interface{}, len(raw))
	for k, v := range raw {
		if n, ok := v.(json.Number); ok {
			if i, err := n.Int64(); err == nil {
				vars[k] = i
			} else if f, err := n.Float64(); err == nil {
				vars[k] = f
			}
			continue
		}
		vars[k] = v
	}
	return vars, nil
}

func executeLayer(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return errorJSON("executeLayer expects source code")
	}
	var vars map[string]interface{}
	if len(args) > 1 && args[1].Type() == js.TypeString {
		var err error
		if vars, err = decodeVars(args[1].String()); err != nil {
			return errorJSON("invalid variables: " + err.Error())
		}
	}
	return toJSON(runner.Run(args[0].String(), runner.Options{Vars: vars}))
}

func describeLayer(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return errorJSON("describeLayer expects source code")
	}
	return toJSON(runner.Describe(args[0].String()))
}

func parseLayerFile(this js.Value, args []js.Value) (result interface{}) {
	if len(args) != 1 {
		return errorJSON("parseLayerFile expects source code")
	}
	defer func() {
		if r := recover(); r != nil {
			result = errorJSON(fmt.Sprintf("parser panic: %v", r))
		}
	}()

	file, err := sourcefile.Split(args[0].String())
	if err != nil {
		return toJSON(map[string]interface{}{"success": false, "errors": runner.Issues(err)})
	}
	return toJSON(struct {
		Success bool           `json:"success"`
		Errors  []runner.Issue `json:"errors"`
		*sourcefile.File
	}{true, []runner.Issue{}, file})
}

func main() {
	js.Global().Set("executeLayer", js.FuncOf(executeLayer))
	js.Global().Set("describeLayer", js.FuncOf(describeLayer))
	js.Global().Set("parseLayerFile", js.FuncOf(parseLayerFile))

	// Keep the Go runtime alive so the exported functions stay callable.
	select {}
}
