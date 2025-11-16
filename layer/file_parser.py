"""
Utility for parsing .layer files into structured data.

This module provides high-level functions to extract layer information
from .layer source files for use by tools like the visual editor.
"""

from typing import Dict, List, Optional
import re


class LayerFileParseError(Exception):
    """Error raised when parsing a .layer file fails."""
    pass


class LayerFileData:
    """Structured data extracted from a .layer file."""

    def __init__(
        self,
        enabled_layers: List[str],
        layer_code: Dict[str, str],
        main_expression: str
    ):
        self.enabled_layers = enabled_layers
        self.layer_code = layer_code
        self.main_expression = main_expression

    def to_dict(self) -> dict:
        """Convert to dictionary for JSON serialization."""
        return {
            'enabledLayers': self.enabled_layers,
            'layers': self.layer_code,
            'mainExpression': self.main_expression
        }


def parse_layer_file(content: str) -> LayerFileData:
    """
    Parse a .layer file and extract structured data.

    Args:
        content: The raw content of a .layer file

    Returns:
        LayerFileData containing enabled layers, layer code, and main expression

    Raises:
        LayerFileParseError: If the file cannot be parsed
    """
    # Parse @layers directive
    layers_match = re.search(r'@layers\s*\[([^\]]+)\]', content)
    if not layers_match:
        raise LayerFileParseError('No @layers directive found')

    enabled_layer_names = [name.strip() for name in layers_match.group(1).split(',')]

    # Parse @layer blocks using brace counting
    parsed_layers = {}
    layer_header_pattern = re.compile(r'@layer\s+(\w+)\s*\{')

    for match in layer_header_pattern.finditer(content):
        layer_name = match.group(1)
        start_pos = match.end()

        # Find matching closing brace by counting
        brace_count = 1
        end_pos = start_pos

        while brace_count > 0 and end_pos < len(content):
            if content[end_pos] == '{':
                brace_count += 1
            elif content[end_pos] == '}':
                brace_count -= 1
            end_pos += 1

        if brace_count == 0:
            # Extract layer code (excluding the closing brace)
            layer_code = content[start_pos:end_pos - 1]

            # Remove indentation (typically 4 spaces)
            parsed_layers[layer_name] = _dedent_code(layer_code)
        else:
            raise LayerFileParseError(f'Unmatched braces in @layer {layer_name}')

    # Extract main expression (everything after last @layer block)
    main_expression = _extract_main_expression(content, layer_header_pattern)

    return LayerFileData(
        enabled_layers=enabled_layer_names,
        layer_code=parsed_layers,
        main_expression=main_expression
    )


def _dedent_code(code: str, indent: str = '    ') -> str:
    """
    Remove common indentation from code block.

    Args:
        code: The code to dedent
        indent: The indentation string to remove (default: 4 spaces)

    Returns:
        Dedented code
    """
    lines = code.split('\n')
    dedented_lines = []

    for line in lines:
        if line.startswith(indent):
            dedented_lines.append(line[len(indent):])
        else:
            dedented_lines.append(line)

    return '\n'.join(dedented_lines).strip()


def _extract_main_expression(content: str, layer_pattern: re.Pattern) -> str:
    """
    Extract the main expression from a .layer file.

    The main expression is everything after the last @layer block,
    excluding leading comments.

    Args:
        content: The full file content
        layer_pattern: Compiled regex pattern for finding @layer headers

    Returns:
        The main expression code
    """
    last_layer_matches = list(layer_pattern.finditer(content))

    if not last_layer_matches:
        # No layer blocks, everything is main expression
        return content.strip()

    last_match = last_layer_matches[-1]
    last_layer_start = last_match.start()

    # Find the closing brace of the last layer
    after_last_layer = content[last_layer_start:]
    layer_start_brace = after_last_layer.index('{') + 1
    brace_count = 1
    pos = layer_start_brace

    while brace_count > 0 and pos < len(after_last_layer):
        if after_last_layer[pos] == '{':
            brace_count += 1
        elif after_last_layer[pos] == '}':
            brace_count -= 1
        pos += 1

    main_start = last_layer_start + pos
    main_expression = content[main_start:].strip()

    # Remove leading comment lines (but preserve inline comments)
    lines = main_expression.split('\n')
    main_lines = []

    for line in lines:
        trimmed = line.strip()
        if trimmed and not trimmed.startswith('#'):
            main_lines.append(line)

    return '\n'.join(main_lines).strip()
