"""
Layer Language AST Node Definitions

Defines the Abstract Syntax Tree node types for the Layer language.
"""

from dataclasses import dataclass
from typing import List, Optional, Any


@dataclass
class ASTNode:
    """Base class for all AST nodes."""
    line: int
    column: int


# Literals

@dataclass
class NumberLiteral(ASTNode):
    """Numeric literal node."""
    value: float


@dataclass
class StringLiteral(ASTNode):
    """String literal node."""
    value: str


@dataclass
class BooleanLiteral(ASTNode):
    """Boolean literal node."""
    value: bool


@dataclass
class ListLiteral(ASTNode):
    """List literal node."""
    elements: List[ASTNode]


@dataclass
class RecordLiteral(ASTNode):
    """Record/object literal node."""
    fields: dict  # {field_name: ASTNode}


@dataclass
class TupleLiteral(ASTNode):
    """Tuple literal node."""
    elements: List[ASTNode]


# Variables and References

@dataclass
class Identifier(ASTNode):
    """Variable/identifier reference."""
    name: str


@dataclass
class MemberAccess(ASTNode):
    """Member access (e.g., core.factorial)."""
    object: ASTNode
    member: str


# Functions

@dataclass
class Lambda(ASTNode):
    """Lambda/anonymous function."""
    params: List[str]
    body: ASTNode


@dataclass
class FunctionDef(ASTNode):
    """Function definition."""
    name: str
    params: List[str]
    body: ASTNode
    type_annotation: Optional[Any] = None


@dataclass
class FunctionCall(ASTNode):
    """Function call."""
    function: ASTNode
    arguments: List[ASTNode]


# Control Flow

@dataclass
class LetBinding(ASTNode):
    """Let binding expression."""
    name: str
    value: ASTNode
    body: ASTNode


@dataclass
class MatchCase(ASTNode):
    """A single case in a match expression."""
    pattern: ASTNode
    body: ASTNode


@dataclass
class Match(ASTNode):
    """Pattern matching expression."""
    value: ASTNode
    cases: List[MatchCase]


# Patterns

@dataclass
class NumberPattern(ASTNode):
    """Number pattern for matching."""
    value: float


@dataclass
class StringPattern(ASTNode):
    """String pattern for matching."""
    value: str


@dataclass
class BooleanPattern(ASTNode):
    """Boolean pattern for matching (true/false)."""
    value: bool


@dataclass
class IdentifierPattern(ASTNode):
    """Variable binding pattern."""
    name: str


@dataclass
class ListPattern(ASTNode):
    """List destructuring pattern."""
    elements: List[ASTNode]
    rest: Optional[str] = None  # For ...rest patterns


# Operators

@dataclass
class BinaryOp(ASTNode):
    """Binary operation."""
    operator: str
    left: ASTNode
    right: ASTNode


@dataclass
class UnaryOp(ASTNode):
    """Unary operation."""
    operator: str
    operand: ASTNode


# Layer-specific

@dataclass
class LayerDeclaration(ASTNode):
    """Declaration of a layer block."""
    layer_name: str
    definitions: List[ASTNode]


@dataclass
class LayersDirective(ASTNode):
    """Directive to enable specific layers."""
    layer_names: List[str]


# Program

@dataclass
class Block(ASTNode):
    """Block of statements/expressions."""
    statements: List[ASTNode]


@dataclass
class Program(ASTNode):
    """Root program node."""
    layers_directive: Optional[LayersDirective]
    layer_declarations: List[LayerDeclaration]
    main_expression: Optional[ASTNode] = None


def ast_to_dict(node: ASTNode) -> dict:
    """Convert AST node to dictionary for debugging/visualization."""
    if node is None:
        return None

    result = {
        'type': type(node).__name__,
        'line': node.line,
        'column': node.column
    }

    if isinstance(node, NumberLiteral):
        result['value'] = node.value
    elif isinstance(node, StringLiteral):
        result['value'] = node.value
    elif isinstance(node, BooleanLiteral):
        result['value'] = node.value
    elif isinstance(node, Identifier):
        result['name'] = node.name
    elif isinstance(node, FunctionDef):
        result['name'] = node.name
        result['params'] = node.params
        result['body'] = ast_to_dict(node.body)
    elif isinstance(node, FunctionCall):
        result['function'] = ast_to_dict(node.function)
        result['arguments'] = [ast_to_dict(arg) for arg in node.arguments]
    elif isinstance(node, BinaryOp):
        result['operator'] = node.operator
        result['left'] = ast_to_dict(node.left)
        result['right'] = ast_to_dict(node.right)
    elif isinstance(node, LetBinding):
        result['name'] = node.name
        result['value'] = ast_to_dict(node.value)
        result['body'] = ast_to_dict(node.body)
    elif isinstance(node, Match):
        result['value'] = ast_to_dict(node.value)
        result['cases'] = [
            {
                'pattern': ast_to_dict(case.pattern),
                'body': ast_to_dict(case.body)
            }
            for case in node.cases
        ]
    elif isinstance(node, LayerDeclaration):
        result['layer_name'] = node.layer_name
        result['definitions'] = [ast_to_dict(d) for d in node.definitions]
    elif isinstance(node, Program):
        result['layers_directive'] = ast_to_dict(node.layers_directive) if node.layers_directive else None
        result['layer_declarations'] = [ast_to_dict(ld) for ld in node.layer_declarations]
        result['main_expression'] = ast_to_dict(node.main_expression) if node.main_expression else None

    return result
