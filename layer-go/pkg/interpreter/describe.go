package interpreter

import (
	"sort"

	"github.com/jeanlazarou/layer/pkg/ast"
)

// LayerInfo describes one active layer of a loaded program.
type LayerInfo struct {
	Name    string   `json:"name"`
	Pure    bool     `json:"pure"`
	Defines []string `json:"defines"` // names this program's source defines in the layer (sorted)
}

// Resolution describes how one user-defined name resolves through the stack.
type Resolution struct {
	Name string `json:"name"`
	// Layers lists every active layer defining the name, top of the stack
	// first. With the override blend mode, Layers[0] answers unqualified
	// calls; other modes combine all of them.
	Layers []string `json:"layers"`
}

// Description is a static picture of a loaded program's layer stack, used
// by the visual editor's resolution view.
type Description struct {
	BlendMode   string       `json:"blendMode"`
	Layers      []LayerInfo  `json:"layers"` // bottom of the stack first
	Resolutions []Resolution `json:"resolutions"`
}

// Describe reports the layer stack of a program previously passed to Load
// (or Execute).
func (i *Interpreter) Describe(program *ast.Program) Description {
	lc := i.layerContext

	defined := make(map[string]map[string]bool) // layer → names defined in source
	allNames := make(map[string]bool)
	active := make(map[string]bool)
	for _, l := range lc.GetLayers() {
		active[l] = true
	}
	for _, decl := range program.LayerDeclarations {
		if !active[decl.LayerName] {
			continue
		}
		for _, def := range decl.Definitions {
			if fd, ok := def.(*ast.FunctionDef); ok {
				if defined[decl.LayerName] == nil {
					defined[decl.LayerName] = make(map[string]bool)
				}
				defined[decl.LayerName][fd.Name] = true
				allNames[fd.Name] = true
			}
		}
	}

	mode := lc.GetBlendMode()
	if mode == "" {
		mode = "override"
	}
	d := Description{BlendMode: mode, Layers: []LayerInfo{}, Resolutions: []Resolution{}}

	for _, name := range lc.GetLayers() {
		info := LayerInfo{Name: name, Pure: lc.IsPureLayer(name), Defines: []string{}}
		for fn := range defined[name] {
			info.Defines = append(info.Defines, fn)
		}
		sort.Strings(info.Defines)
		d.Layers = append(d.Layers, info)
	}

	names := make([]string, 0, len(allNames))
	for n := range allNames {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		d.Resolutions = append(d.Resolutions, Resolution{Name: n, Layers: lc.LayersDefining(n)})
	}
	return d
}
