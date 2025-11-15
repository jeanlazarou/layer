"""
UI layer implementation for the Layer language.

Provides declarative UI components and rendering.
This is a conceptual implementation showing how UI could work.
"""

from typing import Any, Dict, List
from dataclasses import dataclass


@dataclass
class UIComponent:
    """Base class for UI components."""
    type: str
    props: Dict[str, Any]
    children: List['UIComponent'] = None

    def __init__(self, type: str, props: Dict[str, Any], children=None):
        self.type = type
        self.props = props
        self.children = children or []

    def render(self, indent=0) -> str:
        """Render the component to a string (for visualization)."""
        indent_str = "  " * indent
        result = f"{indent_str}<{self.type}"

        # Add props
        for key, value in self.props.items():
            if key != 'children':
                result += f" {key}=\"{value}\""

        if not self.children:
            result += " />"
            return result

        result += ">\n"

        # Render children
        for child in self.children:
            if isinstance(child, UIComponent):
                result += child.render(indent + 1) + "\n"
            else:
                result += f"{indent_str}  {child}\n"

        result += f"{indent_str}</{self.type}>"
        return result

    def __repr__(self):
        return f"UIComponent({self.type}, {self.props})"


def setup_ui_layer(layer_context):
    """Set up the UI layer with component constructors."""

    def view(props):
        """Create a view/container component."""
        if not isinstance(props, dict):
            raise TypeError("view expects a record/dict")

        children = props.get('children', [])
        if not isinstance(children, list):
            children = [children]

        return UIComponent('view', props, children)

    def text(props):
        """Create a text component."""
        if not isinstance(props, dict):
            # Allow shorthand: text("Hello")
            if isinstance(props, str):
                props = {'content': props}
            else:
                raise TypeError("text expects a record/dict or string")

        return UIComponent('text', props)

    def button(props):
        """Create a button component."""
        if not isinstance(props, dict):
            raise TypeError("button expects a record/dict")

        return UIComponent('button', props)

    def input_field(props):
        """Create an input field component."""
        if not isinstance(props, dict):
            raise TypeError("input expects a record/dict")

        return UIComponent('input', props)

    def container(layout, children):
        """Create a container with specific layout."""
        return UIComponent('container', {'layout': layout}, children)

    def row(children):
        """Create a horizontal layout."""
        return container('horizontal', children)

    def column(children):
        """Create a vertical layout."""
        return container('vertical', children)

    def render_ui(component):
        """
        Render a UI component to string.

        In a real implementation, this would render to an actual UI framework.
        """
        if isinstance(component, UIComponent):
            return component.render()
        return str(component)

    # Register UI functions
    layer_context.set_current_layer('ui')

    ui_functions = {
        'view': view,
        'text': text,
        'button': button,
        'input': input_field,
        'row': row,
        'column': column,
        'render': render_ui,
    }

    for name, func in ui_functions.items():
        layer_context.define(name, func, 'ui')

    return layer_context
