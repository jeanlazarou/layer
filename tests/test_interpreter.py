"""
Tests for the Layer interpreter.
"""

import pytest
from layer.interpreter import interpret


def test_arithmetic():
    """Test basic arithmetic operations."""
    assert interpret("5 + 3") == 8
    assert interpret("10 - 4") == 6
    assert interpret("6 * 7") == 42
    assert interpret("20 / 4") == 5


def test_factorial():
    """Test factorial function."""
    code = """
    @layers [core]

    @layer core {
        fn factorial(n) =
            match n with
            | 0 -> 1
            | n -> n * factorial(n - 1)
    }

    factorial(5)
    """
    assert interpret(code) == 120


def test_fibonacci():
    """Test fibonacci function."""
    code = """
    @layers [core]

    @layer core {
        fn fib(n) =
            match n with
            | 0 -> 0
            | 1 -> 1
            | n -> fib(n - 1) + fib(n - 2)
    }

    fib(10)
    """
    assert interpret(code) == 55


def test_let_binding():
    """Test let bindings."""
    code = """
    let x = 5 in
    let y = 10 in
        x + y
    """
    assert interpret(code) == 15


def test_lambda():
    """Test lambda functions."""
    code = """
    @layers [core]

    @layer core {
        fn apply(f, x) = f(x)
    }

    apply(λ(x) -> x * 2, 21)
    """
    assert interpret(code) == 42


def test_list_operations():
    """Test list operations."""
    code = """
    @layers [core]

    @layer core {
        fn sum(lst) =
            match lst with
            | [] -> 0
            | [head, ...tail] -> head + sum(tail)
    }

    sum([1, 2, 3, 4, 5])
    """
    assert interpret(code) == 15


def test_layer_override():
    """Test layer override behavior."""
    code = """
    @layers [core, debug]

    @layer core {
        fn add(x, y) = x + y
    }

    @layer debug {
        fn add(x, y) =
            let result = core.add(x, y) in
                result
    }

    add(5, 3)
    """
    assert interpret(code) == 8
