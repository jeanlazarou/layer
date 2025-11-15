"""
Tests for the Layer lexer.
"""

import pytest
from layer.lexer import tokenize, TokenType


def test_numbers():
    """Test numeric literal tokenization."""
    tokens = tokenize("42 3.14")
    assert tokens[0].type == TokenType.NUMBER
    assert tokens[0].value == 42
    assert tokens[1].type == TokenType.NUMBER
    assert tokens[1].value == 3.14


def test_strings():
    """Test string literal tokenization."""
    tokens = tokenize('"hello" "world"')
    assert tokens[0].type == TokenType.STRING
    assert tokens[0].value == "hello"
    assert tokens[1].type == TokenType.STRING
    assert tokens[1].value == "world"


def test_identifiers():
    """Test identifier tokenization."""
    tokens = tokenize("foo bar_baz")
    assert tokens[0].type == TokenType.IDENTIFIER
    assert tokens[0].value == "foo"
    assert tokens[1].type == TokenType.IDENTIFIER
    assert tokens[1].value == "bar_baz"


def test_keywords():
    """Test keyword tokenization."""
    tokens = tokenize("fn let in match with")
    assert tokens[0].type == TokenType.FN
    assert tokens[1].type == TokenType.LET
    assert tokens[2].type == TokenType.IN
    assert tokens[3].type == TokenType.MATCH
    assert tokens[4].type == TokenType.WITH


def test_operators():
    """Test operator tokenization."""
    tokens = tokenize("+ - * / == != -> λ")
    assert tokens[0].type == TokenType.PLUS
    assert tokens[1].type == TokenType.MINUS
    assert tokens[2].type == TokenType.STAR
    assert tokens[3].type == TokenType.SLASH
    assert tokens[4].type == TokenType.EQ
    assert tokens[5].type == TokenType.NEQ
    assert tokens[6].type == TokenType.ARROW
    assert tokens[7].type == TokenType.LAMBDA


def test_comments():
    """Test comment handling."""
    tokens = tokenize("42 # this is a comment\n 100")
    # Comments are filtered out by parser, but lexer creates them
    values = [t.value for t in tokens if t.type == TokenType.NUMBER]
    assert values == [42, 100]


def test_layer_directives():
    """Test layer directive tokenization."""
    tokens = tokenize("@layers [core]")
    assert tokens[0].type == TokenType.AT
    assert tokens[1].type == TokenType.LAYERS
    assert tokens[2].type == TokenType.LBRACKET
    assert tokens[3].type == TokenType.IDENTIFIER
    assert tokens[4].type == TokenType.RBRACKET
