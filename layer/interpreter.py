"""
Layer Language Interpreter

Evaluates Layer AST nodes in the context of a layer stack.
"""

from typing import Any, Dict, List
from layer.ast_nodes import *
from layer.layers import LayerContext
from layer.layers.core import setup_core_layer
from layer.layers.debug import setup_debug_layer
from layer.layers.ui import setup_ui_layer
from layer.layers.data import setup_data_layer


class Closure:
    """Represents a function closure."""

    def __init__(self, params: List[str], body: ASTNode, env: Dict[str, Any], interpreter=None):
        self.params = params
        self.body = body
        self.env = env  # Captured environment
        self.interpreter = interpreter

    def __call__(self, *args):
        """Make closure callable."""
        if len(args) != len(self.params):
            raise TypeError(f"Expected {len(self.params)} arguments, got {len(args)}")

        # Create new environment with parameters bound
        new_env = self.env.copy()
        for param, arg in zip(self.params, args):
            new_env[param] = arg

        # Evaluate body in new environment
        if self.interpreter:
            return self.interpreter.eval(self.body, new_env)
        else:
            # Fallback for old behavior
            return ('closure_call', new_env, self.body)

    def __repr__(self):
        return f"<function({', '.join(self.params)})>"


class Interpreter:
    """Interprets Layer programs."""

    def __init__(self):
        self.layer_context = LayerContext()
        self.local_env: Dict[str, Any] = {}  # Local variable environment

        # Set up built-in layers
        self.layer_context.add_layer('core')
        setup_core_layer(self.layer_context)

    def enable_layer(self, layer_name: str):
        """Enable a layer."""
        self.layer_context.add_layer(layer_name)

        # Set up built-in layer implementations
        if layer_name == 'debug':
            setup_debug_layer(self.layer_context)
        elif layer_name == 'ui':
            setup_ui_layer(self.layer_context)
        elif layer_name == 'data':
            setup_data_layer(self.layer_context)

    def error(self, node: ASTNode, message: str):
        """Raise an interpreter error."""
        raise RuntimeError(f"Runtime error at {node.line}:{node.column}: {message}")

    def eval(self, node: ASTNode, env: Dict[str, Any] = None) -> Any:
        """Evaluate an AST node."""
        if env is None:
            env = self.local_env

        # Literals
        if isinstance(node, NumberLiteral):
            return node.value

        if isinstance(node, StringLiteral):
            return node.value

        if isinstance(node, BooleanLiteral):
            return node.value

        if isinstance(node, ListLiteral):
            return [self.eval(elem, env) for elem in node.elements]

        if isinstance(node, TupleLiteral):
            return tuple(self.eval(elem, env) for elem in node.elements)

        if isinstance(node, RecordLiteral):
            return {key: self.eval(value, env) for key, value in node.fields.items()}

        # Identifiers
        if isinstance(node, Identifier):
            # Look up in local environment first
            if node.name in env:
                return env[node.name]

            # Then in layer context
            try:
                return self.layer_context.lookup(node.name)
            except NameError:
                self.error(node, f"Undefined variable: {node.name}")

        # Member access (layer.function)
        if isinstance(node, MemberAccess):
            # For now, assume object is a layer name
            if isinstance(node.object, Identifier):
                layer_name = node.object.name
                try:
                    return self.layer_context.get_layer_value(layer_name, node.member)
                except NameError as e:
                    self.error(node, str(e))
            else:
                self.error(node, "Member access only supported on layer names currently")

        # Functions
        if isinstance(node, Lambda):
            return Closure(node.params, node.body, env.copy(), self)

        if isinstance(node, FunctionDef):
            closure = Closure(node.params, node.body, env.copy(), self)
            # Store in current layer
            self.layer_context.define(node.name, closure)
            return closure

        if isinstance(node, FunctionCall):
            func = self.eval(node.function, env)
            args = [self.eval(arg, env) for arg in node.arguments]

            # Handle all callables (both closures and built-in functions)
            if callable(func):
                try:
                    return func(*args)
                except Exception as e:
                    self.error(node, f"Error calling function: {e}")

            self.error(node, f"Not a function: {func}")

        # Binary operations
        if isinstance(node, BinaryOp):
            left = self.eval(node.left, env)
            right = self.eval(node.right, env)

            # Look up operator function
            try:
                op_func = self.layer_context.lookup(node.operator)
                return op_func(left, right)
            except NameError:
                self.error(node, f"Unknown operator: {node.operator}")

        # Unary operations
        if isinstance(node, UnaryOp):
            operand = self.eval(node.operand, env)

            if node.operator == '-':
                return -operand
            elif node.operator == 'not':
                op_func = self.layer_context.lookup('not')
                return op_func(operand)
            else:
                self.error(node, f"Unknown unary operator: {node.operator}")

        # Let binding
        if isinstance(node, LetBinding):
            value = self.eval(node.value, env)
            new_env = env.copy()
            new_env[node.name] = value
            return self.eval(node.body, new_env)

        # Pattern matching
        if isinstance(node, Match):
            value = self.eval(node.value, env)

            for case in node.cases:
                match_result = self.match_pattern(case.pattern, value, env)
                if match_result is not None:
                    # Pattern matched, evaluate body with bindings
                    new_env = env.copy()
                    new_env.update(match_result)
                    return self.eval(case.body, new_env)

            self.error(node, f"No pattern matched for value: {value}")

        # Block
        if isinstance(node, Block):
            result = None
            for stmt in node.statements:
                result = self.eval(stmt, env)
            return result

        self.error(node, f"Unknown node type: {type(node).__name__}")

    def match_pattern(self, pattern: ASTNode, value: Any, env: Dict[str, Any]) -> Dict[str, Any]:
        """
        Match a pattern against a value.

        Returns a dictionary of bindings if match succeeds, None otherwise.
        """
        # Number pattern
        if isinstance(pattern, NumberPattern):
            if pattern.value == value:
                return {}
            return None

        # String pattern
        if isinstance(pattern, StringPattern):
            if pattern.value == value:
                return {}
            return None

        # Boolean pattern
        if isinstance(pattern, BooleanPattern):
            if pattern.value == value:
                return {}
            return None

        # Identifier pattern (binds variable or wildcard)
        if isinstance(pattern, IdentifierPattern):
            if pattern.name == '_':
                return {}  # Wildcard, always matches
            return {pattern.name: value}

        # List pattern
        if isinstance(pattern, ListPattern):
            if not isinstance(value, list):
                return None

            # Match elements
            bindings = {}

            if pattern.rest:
                # Pattern like [a, b, ...rest]
                if len(value) < len(pattern.elements):
                    return None

                # Match fixed elements
                for i, elem_pattern in enumerate(pattern.elements):
                    elem_bindings = self.match_pattern(elem_pattern, value[i], env)
                    if elem_bindings is None:
                        return None
                    bindings.update(elem_bindings)

                # Bind rest
                rest_value = value[len(pattern.elements):]
                bindings[pattern.rest] = rest_value

            else:
                # Exact length match
                if len(value) != len(pattern.elements):
                    return None

                for elem_pattern, elem_value in zip(pattern.elements, value):
                    elem_bindings = self.match_pattern(elem_pattern, elem_value, env)
                    if elem_bindings is None:
                        return None
                    bindings.update(elem_bindings)

            return bindings

        return None

    def execute(self, program: Program) -> Any:
        """Execute a complete Layer program."""
        # Enable layers
        if program.layers_directive:
            for layer_name in program.layers_directive.layer_names:
                self.enable_layer(layer_name)

        # Evaluate layer declarations
        for layer_decl in program.layer_declarations:
            self.layer_context.set_current_layer(layer_decl.layer_name)

            for definition in layer_decl.definitions:
                self.eval(definition)

        # Evaluate main expression
        if program.main_expression:
            return self.eval(program.main_expression)

        return None


def interpret(source: str) -> Any:
    """Convenience function to interpret Layer source code."""
    from layer.lexer import tokenize
    from layer.parser import parse

    tokens = tokenize(source)
    ast = parse(tokens)
    interpreter = Interpreter()
    return interpreter.execute(ast)


if __name__ == "__main__":
    test_code = """
    @layers [core, debug]

    @layer core {
        fn factorial(n) =
            match n with
            | 0 -> 1
            | n -> n * factorial(n - 1)
    }

    @layer debug {
        fn factorial(n) =
            let _ = log("Computing factorial of {n}") in
                core.factorial(n)
    }

    factorial(5)
    """

    result = interpret(test_code)
    print(f"Result: {result}")
