"""
Tests for the Layer parser.
"""

import pytest
from layer.lexer import tokenize
from layer.parser import parse
from layer.ast_nodes import *


def test_parse_number():
    """Test parsing number literals."""
    tokens = tokenize("42")
    ast = parse(tokens)
    assert isinstance(ast.main_expression, NumberLiteral)
    assert ast.main_expression.value == 42


def test_parse_function_def():
    """Test parsing function definitions."""
    code = """
    @layers [core]
    @layer core {
        fn add(x, y) = x + y
    }
    """
    tokens = tokenize(code)
    ast = parse(tokens)

    assert len(ast.layer_declarations) == 1
    layer = ast.layer_declarations[0]
    assert layer.layer_name == "core"
    assert len(layer.definitions) == 1

    func_def = layer.definitions[0]
    assert isinstance(func_def, FunctionDef)
    assert func_def.name == "add"
    assert func_def.params == ["x", "y"]


def test_parse_match():
    """Test parsing match expressions."""
    code = """
    match n with
    | 0 -> 1
    | n -> n
    """
    tokens = tokenize(code)
    ast = parse(tokens)

    match_expr = ast.main_expression
    assert isinstance(match_expr, Match)
    assert len(match_expr.cases) == 2
    assert isinstance(match_expr.cases[0].pattern, NumberPattern)
    assert isinstance(match_expr.cases[1].pattern, IdentifierPattern)


def test_parse_let_binding():
    """Test parsing let bindings."""
    code = "let x = 5 in x + 10"
    tokens = tokenize(code)
    ast = parse(tokens)

    let_binding = ast.main_expression
    assert isinstance(let_binding, LetBinding)
    assert let_binding.name == "x"
    assert isinstance(let_binding.value, NumberLiteral)
    assert isinstance(let_binding.body, BinaryOp)


def test_parse_lambda():
    """Test parsing lambda expressions."""
    code = "λ(x) -> x * 2"
    tokens = tokenize(code)
    ast = parse(tokens)

    lambda_expr = ast.main_expression
    assert isinstance(lambda_expr, Lambda)
    assert lambda_expr.params == ["x"]
    assert isinstance(lambda_expr.body, BinaryOp)


def test_parse_list():
    """Test parsing list literals."""
    code = "[1, 2, 3]"
    tokens = tokenize(code)
    ast = parse(tokens)

    list_lit = ast.main_expression
    assert isinstance(list_lit, ListLiteral)
    assert len(list_lit.elements) == 3
