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
// deterministic.
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
		b.WriteString(" " + key + "=\"" + xmlEscape(Show(u.Props[key])) + "\"")
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

// SetupUILayer sets up the UI layer with component constructors
func SetupUILayer(lc *layers.LayerContext) {
	lc.SetCurrentLayer("ui")

	// view component
	lc.Define("view", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("view expects 1 argument (props)")
		}

		props, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("view expects a record/map")
		}

		children := make([]interface{}, 0)
		if childrenVal, found := props["children"]; found {
			if childList, ok := childrenVal.([]interface{}); ok {
				children = childList
			} else {
				children = []interface{}{childrenVal}
			}
		}

		return &UIComponent{
			Type:     "view",
			Props:    props,
			Children: children,
		}, nil
	})

	// text component
	lc.Define("text", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("text expects 1 argument")
		}

		// Allow shorthand: text("Hello") or text({content: "Hello"})
		props := make(map[string]interface{})

		if str, ok := args[0].(string); ok {
			props["content"] = str
		} else if propsMap, ok := args[0].(map[string]interface{}); ok {
			props = propsMap
		} else {
			return nil, fmt.Errorf("text expects a string or record")
		}

		return &UIComponent{
			Type:     "text",
			Props:    props,
			Children: []interface{}{},
		}, nil
	})

	// button component
	lc.Define("button", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("button expects 1 argument (props)")
		}

		props, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("button expects a record/map")
		}

		return &UIComponent{
			Type:     "button",
			Props:    props,
			Children: []interface{}{},
		}, nil
	})

	// input component
	lc.Define("input", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("input expects 1 argument (props)")
		}

		props, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("input expects a record/map")
		}

		return &UIComponent{
			Type:     "input",
			Props:    props,
			Children: []interface{}{},
		}, nil
	})

	// container component
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

		props := map[string]interface{}{
			"layout": layout,
		}

		return &UIComponent{
			Type:     "container",
			Props:    props,
			Children: children,
		}, nil
	})

	// row helper
	lc.Define("row", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("row expects 1 argument (children)")
		}

		children, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("row children must be a list")
		}

		props := map[string]interface{}{
			"layout": "horizontal",
		}

		return &UIComponent{
			Type:     "container",
			Props:    props,
			Children: children,
		}, nil
	})

	// column helper
	lc.Define("column", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("column expects 1 argument (children)")
		}

		children, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("column children must be a list")
		}

		props := map[string]interface{}{
			"layout": "vertical",
		}

		return &UIComponent{
			Type:     "container",
			Props:    props,
			Children: children,
		}, nil
	})

	// render function
	lc.Define("render", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("render expects 1 argument (component)")
		}

		if comp, ok := args[0].(*UIComponent); ok {
			return comp.Render(0), nil
		}

		return fmt.Sprintf("%v", args[0]), nil
	})
}
