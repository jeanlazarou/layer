"""
Core layer implementation for the Layer language.

Provides fundamental functional programming primitives.
"""

import operator
from functools import reduce
from typing import Any, List, Callable


def setup_core_layer(layer_context):
    """Set up the core layer with built-in functions."""

    # Arithmetic operators
    def add(x, y):
        return x + y

    def subtract(x, y):
        return x - y

    def multiply(x, y):
        return x * y

    def divide(x, y):
        if y == 0:
            raise ZeroDivisionError("Division by zero")
        return x / y

    def modulo(x, y):
        return x % y

    def power(x, y):
        return x ** y

    # Comparison operators
    def equal(x, y):
        return x == y

    def not_equal(x, y):
        return x != y

    def less_than(x, y):
        return x < y

    def greater_than(x, y):
        return x > y

    def less_equal(x, y):
        return x <= y

    def greater_equal(x, y):
        return x >= y

    # Logical operators
    def and_op(x, y):
        return x and y

    def or_op(x, y):
        return x or y

    def not_op(x):
        return not x

    # List operations
    def list_head(lst):
        """Get the first element of a list."""
        if not lst:
            raise IndexError("head of empty list")
        return lst[0]

    def list_tail(lst):
        """Get all elements except the first."""
        if not lst:
            raise IndexError("tail of empty list")
        return lst[1:]

    def list_length(lst):
        """Get the length of a list."""
        return len(lst)

    def list_map(fn, lst):
        """Map a function over a list."""
        return [fn(x) for x in lst]

    def list_filter(fn, lst):
        """Filter a list by a predicate."""
        return [x for x in lst if fn(x)]

    def list_reduce(fn, initial, lst):
        """Reduce a list with a binary function."""
        return reduce(lambda acc, x: fn(acc, x), lst, initial)

    def list_fold(fn, lst, initial):
        """Fold a list (alias for reduce)."""
        return list_reduce(fn, initial, lst)

    def list_concat(lst1, lst2):
        """Concatenate two lists."""
        return lst1 + lst2

    def list_take(n, lst):
        """Take the first n elements."""
        return lst[:n]

    def list_drop(n, lst):
        """Drop the first n elements."""
        return lst[n:]

    # String operations
    def string_concat(s1, s2):
        """Concatenate two strings."""
        return str(s1) + str(s2)

    def string_length(s):
        """Get string length."""
        return len(s)

    def string_split(delimiter, s):
        """Split string by delimiter."""
        return s.split(delimiter)

    def string_join(delimiter, lst):
        """Join list of strings."""
        return delimiter.join(str(x) for x in lst)

    # Type conversion
    def to_int(x):
        """Convert to integer."""
        return int(x)

    def to_float(x):
        """Convert to float."""
        return float(x)

    def to_string(x):
        """Convert to string."""
        return str(x)

    # Utility functions
    def identity(x):
        """Identity function."""
        return x

    def const(x):
        """Constant function - returns a function that always returns x."""
        return lambda y: x

    def compose(f, g):
        """Function composition."""
        return lambda x: f(g(x))

    def print_value(x):
        """Print a value and return it."""
        print(x)
        return x

    # Register all core functions
    layer_context.set_current_layer('core')

    core_functions = {
        # Arithmetic
        '+': add,
        '-': subtract,
        '*': multiply,
        '/': divide,
        '%': modulo,
        '^': power,

        # Comparison
        '==': equal,
        '!=': not_equal,
        '<': less_than,
        '>': greater_than,
        '<=': less_equal,
        '>=': greater_equal,

        # Logical
        'and': and_op,
        'or': or_op,
        'not': not_op,

        # List operations
        'head': list_head,
        'tail': list_tail,
        'length': list_length,
        'map': list_map,
        'filter': list_filter,
        'reduce': list_reduce,
        'fold': list_fold,
        'concat': list_concat,
        'take': list_take,
        'drop': list_drop,

        # String operations
        'str_concat': string_concat,
        'str_length': string_length,
        'split': string_split,
        'join': string_join,

        # Type conversion
        'to_int': to_int,
        'to_float': to_float,
        'to_string': to_string,

        # Utility
        'id': identity,
        'const': const,
        'compose': compose,
        'print': print_value,
    }

    for name, func in core_functions.items():
        layer_context.define(name, func, 'core')

    return layer_context
