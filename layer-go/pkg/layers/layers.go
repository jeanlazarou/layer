package layers

import (
	"fmt"
)

// LayerContext manages the layer stack and function resolution
// Layers are stacked from bottom to top, with higher layers
// able to override or wrap lower layer functions.
type LayerContext struct {
	layers       map[string]map[string]interface{}
	layerOrder   []string // Bottom to top
	currentLayer string
	blendMode    string // "override" (default/""), "chain", "collect", "fallback"
}

// NewLayerContext creates a new layer context
func NewLayerContext() *LayerContext {
	return &LayerContext{
		layers:     make(map[string]map[string]interface{}),
		layerOrder: make([]string, 0),
	}
}

// AddLayer adds a layer to the stack
func (lc *LayerContext) AddLayer(name string) {
	if _, exists := lc.layers[name]; !exists {
		lc.layers[name] = make(map[string]interface{})
		lc.layerOrder = append(lc.layerOrder, name)
	}
}

// SetCurrentLayer sets the layer currently being defined
func (lc *LayerContext) SetCurrentLayer(name string) {
	lc.currentLayer = name
}

// Define defines a value in a specific layer
func (lc *LayerContext) Define(name string, value interface{}, layer ...string) error {
	targetLayer := lc.currentLayer
	if len(layer) > 0 && layer[0] != "" {
		targetLayer = layer[0]
	}

	if targetLayer == "" {
		return fmt.Errorf("no current layer set")
	}

	if _, exists := lc.layers[targetLayer]; !exists {
		lc.AddLayer(targetLayer)
	}

	lc.layers[targetLayer][name] = value
	return nil
}

// Lookup looks up a value, searching from top layer down
// If layer is specified, look up in that specific layer
func (lc *LayerContext) Lookup(name string, layer ...string) (interface{}, error) {
	if len(layer) > 0 && layer[0] != "" {
		targetLayer := layer[0]
		if layerMap, exists := lc.layers[targetLayer]; exists {
			if value, found := layerMap[name]; found {
				return value, nil
			}
		}
		return nil, fmt.Errorf("name '%s' not found in layer '%s'", name, targetLayer)
	}

	// Search from top to bottom
	for i := len(lc.layerOrder) - 1; i >= 0; i-- {
		layerName := lc.layerOrder[i]
		if value, found := lc.layers[layerName][name]; found {
			return value, nil
		}
	}

	return nil, fmt.Errorf("name '%s' not found in any layer", name)
}

// GetLayerValue gets a specific value from a specific layer
func (lc *LayerContext) GetLayerValue(layerName, valueName string) (interface{}, error) {
	if _, exists := lc.layers[layerName]; !exists {
		return nil, fmt.Errorf("layer '%s' not found", layerName)
	}

	if value, found := lc.layers[layerName][valueName]; found {
		return value, nil
	}

	return nil, fmt.Errorf("name '%s' not found in layer '%s'", valueName, layerName)
}

// Has checks if a name exists
func (lc *LayerContext) Has(name string, layer ...string) bool {
	_, err := lc.Lookup(name, layer...)
	return err == nil
}

// GetAllNames returns all defined names across all layers
func (lc *LayerContext) GetAllNames() []string {
	nameSet := make(map[string]bool)
	for _, layerMap := range lc.layers {
		for name := range layerMap {
			nameSet[name] = true
		}
	}

	names := make([]string, 0, len(nameSet))
	for name := range nameSet {
		names = append(names, name)
	}
	return names
}

// GetLayers returns the layer order
func (lc *LayerContext) GetLayers() []string {
	return lc.layerOrder
}

// SetBlendMode sets the blend mode for function resolution
func (lc *LayerContext) SetBlendMode(mode string) {
	lc.blendMode = mode
}

// GetBlendMode returns the current blend mode
func (lc *LayerContext) GetBlendMode() string {
	return lc.blendMode
}

// LookupAll returns all matching values from top layer down to bottom.
// Only includes layers that actually define the name.
func (lc *LayerContext) LookupAll(name string) []interface{} {
	var values []interface{}
	for i := len(lc.layerOrder) - 1; i >= 0; i-- {
		if value, found := lc.layers[lc.layerOrder[i]][name]; found {
			values = append(values, value)
		}
	}
	return values
}

// String returns a string representation of the layer context
func (lc *LayerContext) String() string {
	return fmt.Sprintf("LayerContext(layers=%v, order=%v)", lc.GetLayerNames(), lc.layerOrder)
}

// GetLayerNames returns all layer names
func (lc *LayerContext) GetLayerNames() []string {
	names := make([]string, 0, len(lc.layers))
	for name := range lc.layers {
		names = append(names, name)
	}
	return names
}
