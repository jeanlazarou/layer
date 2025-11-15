"""
Layer implementations for the Layer language.
"""

from typing import Dict, Any, Callable


class LayerContext:
    """
    Manages the layer stack and function resolution.

    Layers are stacked from bottom to top, with higher layers
    able to override or wrap lower layer functions.
    """

    def __init__(self):
        self.layers: Dict[str, Dict[str, Any]] = {}
        self.layer_order = []  # Bottom to top
        self.current_layer = None

    def add_layer(self, name: str):
        """Add a layer to the stack."""
        if name not in self.layers:
            self.layers[name] = {}
            self.layer_order.append(name)

    def set_current_layer(self, name: str):
        """Set the layer currently being defined."""
        self.current_layer = name

    def define(self, name: str, value: Any, layer: str = None):
        """Define a value in a specific layer."""
        target_layer = layer or self.current_layer
        if target_layer is None:
            raise ValueError("No current layer set")

        if target_layer not in self.layers:
            self.add_layer(target_layer)

        self.layers[target_layer][name] = value

    def lookup(self, name: str, layer: str = None) -> Any:
        """
        Look up a value, searching from top layer down.

        If layer is specified, look up in that specific layer.
        """
        if layer:
            if layer in self.layers and name in self.layers[layer]:
                return self.layers[layer][name]
            raise NameError(f"Name '{name}' not found in layer '{layer}'")

        # Search from top to bottom
        for layer_name in reversed(self.layer_order):
            if name in self.layers[layer_name]:
                return self.layers[layer_name][name]

        raise NameError(f"Name '{name}' not found in any layer")

    def get_layer_value(self, layer_name: str, value_name: str) -> Any:
        """Get a specific value from a specific layer."""
        if layer_name not in self.layers:
            raise NameError(f"Layer '{layer_name}' not found")

        if value_name not in self.layers[layer_name]:
            raise NameError(f"Name '{value_name}' not found in layer '{layer_name}'")

        return self.layers[layer_name][value_name]

    def has(self, name: str, layer: str = None) -> bool:
        """Check if a name exists."""
        try:
            self.lookup(name, layer)
            return True
        except NameError:
            return False

    def get_all_names(self) -> set:
        """Get all defined names across all layers."""
        names = set()
        for layer in self.layers.values():
            names.update(layer.keys())
        return names

    def __repr__(self):
        return f"LayerContext(layers={list(self.layers.keys())}, order={self.layer_order})"
