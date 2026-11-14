//go:build js && wasm

// Command wasm exposes the Layer interpreter to JavaScript. Every exported
// function takes source text and returns a JSON string, so the browser side
// only has to JSON.parse it:
//
//	executeLayer(source, appStateJSON?) → runner.Result
//	describeLayer(source)               → runner.DescribeResult
//	parseLayerFile(source)              → {success, errors, enabledLayers, declaredLayers, layers, purity, blendMode, blendModes, mainExpression}
//
// appStateJSON drives an app (a program whose main expression is app(...)):
// {"model": <model from a previous result>, "messages": [<message>, ...]}.
// Without it, an app starts from its init.
package main

import (
	"encoding/json"
	"fmt"
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

// appState is the optional second argument of executeLayer.
type appState struct {
	Model    json.RawMessage   `json:"model"`
	Messages []json.RawMessage `json:"messages"`
}

func executeLayer(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return errorJSON("executeLayer expects source code")
	}
	opts := runner.Options{}
	if len(args) > 1 && args[1].Type() == js.TypeString {
		var state appState
		if err := json.Unmarshal([]byte(args[1].String()), &state); err != nil {
			return errorJSON("invalid app state: " + err.Error())
		}
		if len(state.Model) > 0 && string(state.Model) != "null" {
			opts.Model = string(state.Model)
		}
		for _, msg := range state.Messages {
			opts.Messages = append(opts.Messages, string(msg))
		}
	}
	return toJSON(runner.Run(args[0].String(), opts))
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
