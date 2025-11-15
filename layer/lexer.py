"""
Layer Language Lexer

Tokenizes Layer source code into a stream of tokens.
"""

import re
from enum import Enum, auto
from dataclasses import dataclass
from typing import List, Optional


class TokenType(Enum):
    """Token types for the Layer language."""

    # Literals
    NUMBER = auto()
    STRING = auto()
    BOOLEAN = auto()

    # Identifiers and Keywords
    IDENTIFIER = auto()

    # Keywords
    FN = auto()
    LET = auto()
    IN = auto()
    MATCH = auto()
    WITH = auto()
    LAMBDA = auto()
    TRUE = auto()
    FALSE = auto()

    # Layer keywords
    LAYER = auto()
    LAYERS = auto()
    AT = auto()

    # Operators
    PLUS = auto()
    MINUS = auto()
    STAR = auto()
    SLASH = auto()
    PERCENT = auto()
    CARET = auto()

    # Comparison
    EQ = auto()
    NEQ = auto()
    LT = auto()
    GT = auto()
    LTE = auto()
    GTE = auto()

    # Logical
    AND = auto()
    OR = auto()
    NOT = auto()

    # Delimiters
    LPAREN = auto()
    RPAREN = auto()
    LBRACKET = auto()
    RBRACKET = auto()
    LBRACE = auto()
    RBRACE = auto()

    # Punctuation
    COMMA = auto()
    COLON = auto()
    SEMICOLON = auto()
    DOT = auto()
    ARROW = auto()
    ASSIGN = auto()
    PIPE = auto()
    ELLIPSIS = auto()

    # Special
    NEWLINE = auto()
    EOF = auto()
    COMMENT = auto()


@dataclass
class Token:
    """Represents a single token in the source code."""
    type: TokenType
    value: any
    line: int
    column: int

    def __repr__(self):
        return f"Token({self.type.name}, {repr(self.value)}, {self.line}:{self.column})"


class Lexer:
    """Tokenizes Layer source code."""

    KEYWORDS = {
        'fn': TokenType.FN,
        'let': TokenType.LET,
        'in': TokenType.IN,
        'match': TokenType.MATCH,
        'with': TokenType.WITH,
        'true': TokenType.TRUE,
        'false': TokenType.FALSE,
        'and': TokenType.AND,
        'or': TokenType.OR,
        'not': TokenType.NOT,
        'layer': TokenType.LAYER,
        'layers': TokenType.LAYERS,
    }

    def __init__(self, source: str):
        self.source = source
        self.pos = 0
        self.line = 1
        self.column = 1
        self.tokens: List[Token] = []

    def error(self, message: str):
        """Raise a lexer error with position information."""
        raise SyntaxError(f"Lexer error at {self.line}:{self.column}: {message}")

    def peek(self, offset: int = 0) -> Optional[str]:
        """Look ahead at a character without consuming it."""
        pos = self.pos + offset
        if pos < len(self.source):
            return self.source[pos]
        return None

    def advance(self) -> Optional[str]:
        """Consume and return the next character."""
        if self.pos >= len(self.source):
            return None

        char = self.source[self.pos]
        self.pos += 1

        if char == '\n':
            self.line += 1
            self.column = 1
        else:
            self.column += 1

        return char

    def skip_whitespace(self):
        """Skip whitespace characters except newlines."""
        while self.peek() and self.peek() in ' \t\r':
            self.advance()

    def read_comment(self) -> Token:
        """Read a comment (starts with #)."""
        start_line = self.line
        start_column = self.column

        self.advance()  # Skip #

        comment = ""
        while self.peek() and self.peek() != '\n':
            comment += self.advance()

        return Token(TokenType.COMMENT, comment.strip(), start_line, start_column)

    def read_number(self) -> Token:
        """Read a numeric literal."""
        start_line = self.line
        start_column = self.column

        num_str = ""
        has_dot = False

        while self.peek() and (self.peek().isdigit() or self.peek() == '.'):
            if self.peek() == '.':
                if has_dot:
                    break
                # Check if it's ellipsis
                if self.peek(1) == '.' and self.peek(2) == '.':
                    break
                has_dot = True
            num_str += self.advance()

        value = float(num_str) if has_dot else int(num_str)
        return Token(TokenType.NUMBER, value, start_line, start_column)

    def read_string(self) -> Token:
        """Read a string literal."""
        start_line = self.line
        start_column = self.column

        quote = self.advance()  # " or '
        string = ""

        while self.peek() and self.peek() != quote:
            if self.peek() == '\\':
                self.advance()
                next_char = self.advance()
                # Handle escape sequences
                escapes = {
                    'n': '\n',
                    't': '\t',
                    'r': '\r',
                    '\\': '\\',
                    '"': '"',
                    "'": "'"
                }
                string += escapes.get(next_char, next_char)
            else:
                string += self.advance()

        if not self.peek():
            self.error("Unterminated string")

        self.advance()  # Closing quote
        return Token(TokenType.STRING, string, start_line, start_column)

    def read_identifier(self) -> Token:
        """Read an identifier or keyword."""
        start_line = self.line
        start_column = self.column

        ident = ""
        while self.peek() and (self.peek().isalnum() or self.peek() == '_'):
            ident += self.advance()

        # Check if it's a keyword
        token_type = self.KEYWORDS.get(ident, TokenType.IDENTIFIER)

        # Handle boolean values
        if token_type == TokenType.TRUE:
            return Token(TokenType.BOOLEAN, True, start_line, start_column)
        elif token_type == TokenType.FALSE:
            return Token(TokenType.BOOLEAN, False, start_line, start_column)

        return Token(token_type, ident, start_line, start_column)

    def tokenize(self) -> List[Token]:
        """Tokenize the entire source code."""
        while self.pos < len(self.source):
            self.skip_whitespace()

            if not self.peek():
                break

            char = self.peek()
            start_line = self.line
            start_column = self.column

            # Comments
            if char == '#':
                self.read_comment()  # Skip comments
                continue

            # Newlines
            if char == '\n':
                self.advance()
                # Skip newlines for now (we can make them significant later)
                continue

            # Numbers
            if char.isdigit():
                self.tokens.append(self.read_number())
                continue

            # Strings
            if char in '"\'':
                self.tokens.append(self.read_string())
                continue

            # Lambda (must come before identifier check since λ.isalpha() is True)
            if char == 'λ':
                self.advance()
                self.tokens.append(Token(TokenType.LAMBDA, 'λ', start_line, start_column))
                continue

            # Identifiers and keywords
            if char.isalpha() or char == '_':
                self.tokens.append(self.read_identifier())
                continue

            # Two-character operators
            if char == '-' and self.peek(1) == '>':
                self.advance()
                self.advance()
                self.tokens.append(Token(TokenType.ARROW, '->', start_line, start_column))
                continue

            if char == '=' and self.peek(1) == '=':
                self.advance()
                self.advance()
                self.tokens.append(Token(TokenType.EQ, '==', start_line, start_column))
                continue

            if char == '!' and self.peek(1) == '=':
                self.advance()
                self.advance()
                self.tokens.append(Token(TokenType.NEQ, '!=', start_line, start_column))
                continue

            if char == '<' and self.peek(1) == '=':
                self.advance()
                self.advance()
                self.tokens.append(Token(TokenType.LTE, '<=', start_line, start_column))
                continue

            if char == '>' and self.peek(1) == '=':
                self.advance()
                self.advance()
                self.tokens.append(Token(TokenType.GTE, '>=', start_line, start_column))
                continue

            if char == '.' and self.peek(1) == '.' and self.peek(2) == '.':
                self.advance()
                self.advance()
                self.advance()
                self.tokens.append(Token(TokenType.ELLIPSIS, '...', start_line, start_column))
                continue

            # Single-character tokens
            single_char_tokens = {
                '+': TokenType.PLUS,
                '-': TokenType.MINUS,
                '*': TokenType.STAR,
                '/': TokenType.SLASH,
                '%': TokenType.PERCENT,
                '^': TokenType.CARET,
                '<': TokenType.LT,
                '>': TokenType.GT,
                '(': TokenType.LPAREN,
                ')': TokenType.RPAREN,
                '[': TokenType.LBRACKET,
                ']': TokenType.RBRACKET,
                '{': TokenType.LBRACE,
                '}': TokenType.RBRACE,
                ',': TokenType.COMMA,
                ':': TokenType.COLON,
                ';': TokenType.SEMICOLON,
                '.': TokenType.DOT,
                '=': TokenType.ASSIGN,
                '|': TokenType.PIPE,
                '@': TokenType.AT,
            }

            if char in single_char_tokens:
                self.advance()
                self.tokens.append(Token(single_char_tokens[char], char, start_line, start_column))
                continue

            self.error(f"Unexpected character: {char}")

        # Add EOF token
        self.tokens.append(Token(TokenType.EOF, None, self.line, self.column))
        return self.tokens


def tokenize(source: str) -> List[Token]:
    """Convenience function to tokenize source code."""
    lexer = Lexer(source)
    return lexer.tokenize()


if __name__ == "__main__":
    # Test the lexer
    test_code = """
    # This is a comment
    fn factorial(n) =
        match n with
        | 0 -> 1
        | n -> n * factorial(n - 1)

    let result = factorial(5) in
        result
    """

    tokens = tokenize(test_code)
    for token in tokens:
        if token.type != TokenType.EOF:
            print(token)
