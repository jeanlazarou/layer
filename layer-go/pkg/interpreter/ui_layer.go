package interpreter

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"

	"github.com/jeanlazarou/layer/pkg/layers"
)

// UIComponent represents a UI component
type UIComponent struct {
	Type     string
	Props    map[string]interface{}
	Children []interface{}
}

// Render renders the component to XML. Attribute values and text are
// escaped, and attributes are emitted in sorted order so the output is
// deterministic. Event handlers (on_click, on_input, ...) are written as
// JSON so the runtime can send the exact message back.
func (u *UIComponent) Render(indent int) string {
	indentStr := strings.Repeat("  ", indent)
	var b strings.Builder
	b.WriteString(indentStr + "<" + u.Type)

	keys := make([]string, 0, len(u.Props))
	for key := range u.Props {
		if key != "children" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := Show(u.Props[key])
		if isHandler(key) {
			if encoded, err := EncodeData(u.Props[key]); err == nil {
				value = encoded
			}
		}
		b.WriteString(" " + key + "=\"" + xmlEscape(value) + "\"")
	}

	if len(u.Children) == 0 {
		b.WriteString(" />")
		return b.String()
	}

	b.WriteString(">\n")
	for _, child := range u.Children {
		if comp, ok := child.(*UIComponent); ok {
			b.WriteString(comp.Render(indent+1) + "\n")
		} else {
			b.WriteString(indentStr + "  " + xmlEscape(Show(child)) + "\n")
		}
	}
	b.WriteString(indentStr + "</" + u.Type + ">")
	return b.String()
}

func xmlEscape(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func (u *UIComponent) String() string {
	return u.Render(0)
}

func isHandler(prop string) bool {
	return strings.HasPrefix(prop, "on_")
}

// checkProps validates the props of a component: event handlers must be
// messages (plain data), and the props of the removed string-handler model
// get an explanation instead of silently doing nothing.
func checkProps(component string, props map[string]interface{}) error {
	if _, ok := props["onclick"]; ok {
		return fmt.Errorf("%s: onclick code strings were removed — give the button a message instead, e.g. button({label: \"+\", on_click: \"inc\"}), and handle it in an app's update (see LANGUAGE_SPEC, Interactive apps)", component)
	}
	if _, ok := props["target"]; ok {
		return fmt.Errorf("%s: target was removed — in an app, update decides how a message changes the model", component)
	}
	for key, value := range props {
		if !isHandler(key) {
			continue
		}
		if err := CheckData(value, key); err != nil {
			return fmt.Errorf("%s: %s must be a message (plain data): %v", component, key, err)
		}
	}
	if tag, ok := props["on_input"]; ok {
		if _, isString := tag.(string); !isString {
			return fmt.Errorf("%s: on_input must be a string tag — typing sends the message {tag: text}, got %s", component, typeName(tag))
		}
	}
	return nil
}

// App is the value of app({init, update, view}). The runtime (the editor,
// or the CLI with --send) drives it: model = init; show view(model); for
// each message, model = update(msg, model) and show the view again. Models
// and messages are plain data.
type App struct {
	Init   interface{}
	Update interface{}
	View   interface{}
}

func (a *App) String() string {
	return "<app>"
}

// uiPureFunctions are the ui layer's constructors. They only build values,
// so pure layers may call them; the effects (rendering, events) belong to
// the runtime.
var uiPureFunctions = []string{"view", "text", "button", "input", "container", "row", "column", "render", "app"}

// SetupUILayer sets up the UI layer with component constructors
func SetupUILayer(lc *layers.LayerContext) {
	lc.SetCurrentLayer("ui")

	propsComponent := func(name string) func(args ...interface{}) (interface{}, error) {
		return func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("%s expects 1 argument (props)", name)
			}
			props, ok := args[0].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("%s expects a record of props, got %s", name, typeName(args[0]))
			}
			if err := checkProps(name, props); err != nil {
				return nil, err
			}
			return &UIComponent{Type: name, Props: props, Children: []interface{}{}}, nil
		}
	}

	// view({title, children}) — a titled panel
	lc.Define("view", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("view expects 1 argument (props)")
		}
		props, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("view expects a record of props, got %s", typeName(args[0]))
		}
		if err := checkProps("view", props); err != nil {
			return nil, err
		}
		children := make([]interface{}, 0)
		if childrenVal, found := props["children"]; found {
			if childList, ok := childrenVal.([]interface{}); ok {
				children = childList
			} else {
				children = []interface{}{childrenVal}
			}
		}
		return &UIComponent{Type: "view", Props: props, Children: children}, nil
	})

	// text("Hello") or text({content: "Hello"})
	lc.Define("text", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("text expects 1 argument")
		}
		props := make(map[string]interface{})
		if str, ok := args[0].(string); ok {
			props["content"] = str
		} else if propsMap, ok := args[0].(map[string]interface{}); ok {
			props = propsMap
		} else {
			props["content"] = Show(args[0])
		}
		if err := checkProps("text", props); err != nil {
			return nil, err
		}
		return &UIComponent{Type: "text", Props: props, Children: []interface{}{}}, nil
	})

	// button({label, on_click: message})
	lc.Define("button", propsComponent("button"))

	// input({name, value, type, placeholder, label, on_input: tag})
	lc.Define("input", propsComponent("input"))

	// container(layout, children)
	lc.Define("container", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("container expects 2 arguments (layout, children)")
		}
		layout, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("container layout must be a string")
		}
		children, ok := args[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("container children must be a list")
		}
		return &UIComponent{Type: "container", Props: map[string]interface{}{"layout": layout}, Children: children}, nil
	})

	box := func(name, layout string) func(args ...interface{}) (interface{}, error) {
		return func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("%s expects 1 argument (children)", name)
			}
			children, ok := args[0].([]interface{})
			if !ok {
				return nil, fmt.Errorf("%s children must be a list, got %s", name, typeName(args[0]))
			}
			return &UIComponent{Type: "container", Props: map[string]interface{}{"layout": layout}, Children: children}, nil
		}
	}
	lc.Define("row", box("row", "horizontal"))
	lc.Define("column", box("column", "vertical"))

	// render(component) — the component as XML text
	lc.Define("render", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("render expects 1 argument (component)")
		}
		if comp, ok := args[0].(*UIComponent); ok {
			return comp.Render(0), nil
		}
		return Show(args[0]), nil
	})

	// app({init: model, update: fn(msg, model), view: fn(model)})
	lc.Define("app", func(args ...interface{}) (interface{}, error) {
		const usage = "app expects {init: <initial model>, update: fn(msg, model), view: fn(model)}"
		if len(args) != 1 {
			return nil, fmt.Errorf("%s", usage)
		}
		spec, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%s, got %s", usage, typeName(args[0]))
		}
		for key := range spec {
			if key != "init" && key != "update" && key != "view" {
				return nil, fmt.Errorf("unknown app field '%s' — %s", key, usage)
			}
		}
		init, hasInit := spec["init"]
		update, hasUpdate := spec["update"]
		view, hasView := spec["view"]
		if !hasInit || !hasUpdate || !hasView {
			return nil, fmt.Errorf("%s (missing %s)", usage, missingFields(hasInit, hasUpdate, hasView))
		}
		if isCallable(init) {
			return nil, fmt.Errorf("app init must be the initial model, got a function — did you mean init()?")
		}
		if err := CheckData(init, "init"); err != nil {
			return nil, fmt.Errorf("app init must be plain data: %v", err)
		}
		if !isCallable(update) {
			return nil, fmt.Errorf("app update must be a function fn(msg, model), got %s", typeName(update))
		}
		if !isCallable(view) {
			return nil, fmt.Errorf("app view must be a function fn(model), got %s", typeName(view))
		}
		return &App{Init: init, Update: update, View: view}, nil
	})
}

func missingFields(hasInit, hasUpdate, hasView bool) string {
	var missing []string
	for _, f := range []struct {
		name string
		has  bool
	}{{"init", hasInit}, {"update", hasUpdate}, {"view", hasView}} {
		if !f.has {
			missing = append(missing, f.name)
		}
	}
	return strings.Join(missing, ", ")
}
