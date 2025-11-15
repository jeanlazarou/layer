"""
Layer Language Parser

Parses tokens into an Abstract Syntax Tree (AST).
"""

from typing import List, Optional
from layer.lexer import Token, TokenType
from layer.ast_nodes import *


class Parser:
    """Recursive descent parser for Layer language."""

    def __init__(self, tokens: List[Token]):
        self.tokens = [t for t in tokens if t.type != TokenType.COMMENT]  # Filter comments
        self.pos = 0

    def error(self, message: str):
        """Raise a parser error with position information."""
        token = self.current()
        raise SyntaxError(f"Parser error at {token.line}:{token.column}: {message}")

    def current(self) -> Token:
        """Get the current token."""
        if self.pos < len(self.tokens):
            return self.tokens[self.pos]
        return self.tokens[-1]  # EOF

    def peek(self, offset: int = 0) -> Token:
        """Look ahead at a token."""
        pos = self.pos + offset
        if pos < len(self.tokens):
            return self.tokens[pos]
        return self.tokens[-1]  # EOF

    def advance(self) -> Token:
        """Consume and return the current token."""
        token = self.current()
        if self.pos < len(self.tokens) - 1:
            self.pos += 1
        return token

    def expect(self, token_type: TokenType) -> Token:
        """Expect a specific token type and consume it."""
        token = self.current()
        if token.type != token_type:
            self.error(f"Expected {token_type.name}, got {token.type.name}")
        return self.advance()

    def match(self, *token_types: TokenType) -> bool:
        """Check if current token matches any of the given types."""
        return self.current().type in token_types

    # Parser methods

    def parse(self) -> Program:
        """Parse the entire program."""
        layers_directive = None
        layer_declarations = []
        main_expression = None

        # Parse @layers directive if present
        if self.match(TokenType.AT):
            if self.peek(1).type == TokenType.LAYERS:
                layers_directive = self.parse_layers_directive()

        # Parse layer declarations
        while self.match(TokenType.AT) and self.peek(1).type == TokenType.LAYER:
            layer_declarations.append(self.parse_layer_declaration())

        # Parse main expression if present
        if not self.match(TokenType.EOF):
            main_expression = self.parse_expression()

        return Program(
            layers_directive=layers_directive,
            layer_declarations=layer_declarations,
            main_expression=main_expression,
            line=1,
            column=1
        )

    def parse_layers_directive(self) -> LayersDirective:
        """Parse @layers [core, ui, debug] directive."""
        start = self.expect(TokenType.AT)
        self.expect(TokenType.LAYERS)
        self.expect(TokenType.LBRACKET)

        layer_names = []
        while not self.match(TokenType.RBRACKET):
            name_token = self.expect(TokenType.IDENTIFIER)
            layer_names.append(name_token.value)

            if self.match(TokenType.COMMA):
                self.advance()

        self.expect(TokenType.RBRACKET)

        return LayersDirective(
            layer_names=layer_names,
            line=start.line,
            column=start.column
        )

    def parse_layer_declaration(self) -> LayerDeclaration:
        """Parse @layer core { ... } declaration."""
        start = self.expect(TokenType.AT)
        self.expect(TokenType.LAYER)
        name_token = self.expect(TokenType.IDENTIFIER)
        self.expect(TokenType.LBRACE)

        definitions = []
        while not self.match(TokenType.RBRACE):
            definitions.append(self.parse_definition())

        self.expect(TokenType.RBRACE)

        return LayerDeclaration(
            layer_name=name_token.value,
            definitions=definitions,
            line=start.line,
            column=start.column
        )

    def parse_definition(self) -> ASTNode:
        """Parse a top-level definition (function, etc.)."""
        if self.match(TokenType.FN):
            return self.parse_function_def()
        else:
            # Could be a value binding or expression
            return self.parse_expression()

    def parse_function_def(self) -> FunctionDef:
        """Parse function definition: fn name(params) = body"""
        start = self.expect(TokenType.FN)
        name_token = self.expect(TokenType.IDENTIFIER)

        # Parse parameters
        self.expect(TokenType.LPAREN)
        params = []

        while not self.match(TokenType.RPAREN):
            param = self.expect(TokenType.IDENTIFIER)
            params.append(param.value)

            if self.match(TokenType.COMMA):
                self.advance()

        self.expect(TokenType.RPAREN)
        self.expect(TokenType.ASSIGN)

        # Parse body
        body = self.parse_expression()

        return FunctionDef(
            name=name_token.value,
            params=params,
            body=body,
            line=start.line,
            column=start.column
        )

    def parse_expression(self) -> ASTNode:
        """Parse an expression."""
        # Let binding
        if self.match(TokenType.LET):
            return self.parse_let_binding()

        # Match expression
        if self.match(TokenType.MATCH):
            return self.parse_match()

        # Lambda
        if self.match(TokenType.LAMBDA):
            return self.parse_lambda()

        # Binary operations and other expressions
        return self.parse_binary_op()

    def parse_let_binding(self) -> LetBinding:
        """Parse let binding: let x = value in body"""
        start = self.expect(TokenType.LET)
        name_token = self.expect(TokenType.IDENTIFIER)
        self.expect(TokenType.ASSIGN)
        value = self.parse_expression()
        self.expect(TokenType.IN)
        body = self.parse_expression()

        return LetBinding(
            name=name_token.value,
            value=value,
            body=body,
            line=start.line,
            column=start.column
        )

    def parse_match(self) -> Match:
        """Parse match expression."""
        start = self.expect(TokenType.MATCH)
        value = self.parse_expression()
        self.expect(TokenType.WITH)

        cases = []
        while self.match(TokenType.PIPE):
            self.advance()  # consume |
            pattern = self.parse_pattern()
            self.expect(TokenType.ARROW)
            body = self.parse_expression()

            cases.append(MatchCase(
                pattern=pattern,
                body=body,
                line=pattern.line,
                column=pattern.column
            ))

        return Match(
            value=value,
            cases=cases,
            line=start.line,
            column=start.column
        )

    def parse_pattern(self) -> ASTNode:
        """Parse a pattern in a match expression."""
        token = self.current()

        # Number pattern
        if self.match(TokenType.NUMBER):
            num_token = self.advance()
            return NumberPattern(
                value=num_token.value,
                line=num_token.line,
                column=num_token.column
            )

        # String pattern
        if self.match(TokenType.STRING):
            str_token = self.advance()
            return StringPattern(
                value=str_token.value,
                line=str_token.line,
                column=str_token.column
            )

        # List pattern
        if self.match(TokenType.LBRACKET):
            return self.parse_list_pattern()

        # Identifier pattern (variable binding or wildcard)
        if self.match(TokenType.IDENTIFIER):
            id_token = self.advance()
            return IdentifierPattern(
                name=id_token.value,
                line=id_token.line,
                column=id_token.column
            )

        self.error(f"Expected pattern, got {token.type.name}")

    def parse_list_pattern(self) -> ListPattern:
        """Parse list pattern: [a, b, ...rest]"""
        start = self.expect(TokenType.LBRACKET)
        elements = []
        rest = None

        while not self.match(TokenType.RBRACKET):
            if self.match(TokenType.ELLIPSIS):
                self.advance()
                rest_token = self.expect(TokenType.IDENTIFIER)
                rest = rest_token.value
                break
            else:
                elements.append(self.parse_pattern())

            if self.match(TokenType.COMMA):
                self.advance()

        self.expect(TokenType.RBRACKET)

        return ListPattern(
            elements=elements,
            rest=rest,
            line=start.line,
            column=start.column
        )

    def parse_lambda(self) -> Lambda:
        """Parse lambda: λ(x, y) -> body"""
        start = self.expect(TokenType.LAMBDA)
        self.expect(TokenType.LPAREN)

        params = []
        while not self.match(TokenType.RPAREN):
            param = self.expect(TokenType.IDENTIFIER)
            params.append(param.value)

            if self.match(TokenType.COMMA):
                self.advance()

        self.expect(TokenType.RPAREN)
        self.expect(TokenType.ARROW)

        body = self.parse_expression()

        return Lambda(
            params=params,
            body=body,
            line=start.line,
            column=start.column
        )

    def parse_binary_op(self, precedence: int = 0) -> ASTNode:
        """Parse binary operations with precedence climbing."""
        left = self.parse_unary_op()

        while True:
            op_token = self.current()
            op_precedence = self.get_precedence(op_token.type)

            # Stop if not a binary operator (precedence 0) or precedence too low
            if op_precedence == 0 or op_precedence < precedence:
                break

            operator = self.advance()
            right = self.parse_binary_op(op_precedence + 1)

            left = BinaryOp(
                operator=operator.value,
                left=left,
                right=right,
                line=operator.line,
                column=operator.column
            )

        return left

    def parse_unary_op(self) -> ASTNode:
        """Parse unary operations."""
        if self.match(TokenType.MINUS, TokenType.NOT):
            op_token = self.advance()
            operand = self.parse_unary_op()

            return UnaryOp(
                operator=op_token.value,
                operand=operand,
                line=op_token.line,
                column=op_token.column
            )

        return self.parse_postfix()

    def parse_postfix(self) -> ASTNode:
        """Parse postfix operations (function calls, member access)."""
        expr = self.parse_primary()

        while True:
            # Function call
            if self.match(TokenType.LPAREN):
                self.advance()
                arguments = []

                while not self.match(TokenType.RPAREN):
                    arguments.append(self.parse_expression())

                    if self.match(TokenType.COMMA):
                        self.advance()

                close_paren = self.expect(TokenType.RPAREN)

                expr = FunctionCall(
                    function=expr,
                    arguments=arguments,
                    line=expr.line,
                    column=expr.column
                )

            # Member access
            elif self.match(TokenType.DOT):
                self.advance()
                member_token = self.expect(TokenType.IDENTIFIER)

                expr = MemberAccess(
                    object=expr,
                    member=member_token.value,
                    line=expr.line,
                    column=expr.column
                )

            else:
                break

        return expr

    def parse_primary(self) -> ASTNode:
        """Parse primary expressions (literals, identifiers, parenthesized)."""
        token = self.current()

        # Number literal
        if self.match(TokenType.NUMBER):
            num_token = self.advance()
            return NumberLiteral(
                value=num_token.value,
                line=num_token.line,
                column=num_token.column
            )

        # String literal
        if self.match(TokenType.STRING):
            str_token = self.advance()
            return StringLiteral(
                value=str_token.value,
                line=str_token.line,
                column=str_token.column
            )

        # Boolean literal
        if self.match(TokenType.BOOLEAN):
            bool_token = self.advance()
            return BooleanLiteral(
                value=bool_token.value,
                line=bool_token.line,
                column=bool_token.column
            )

        # List literal
        if self.match(TokenType.LBRACKET):
            return self.parse_list_literal()

        # Record literal
        if self.match(TokenType.LBRACE):
            return self.parse_record_literal()

        # Identifier
        if self.match(TokenType.IDENTIFIER):
            id_token = self.advance()
            return Identifier(
                name=id_token.value,
                line=id_token.line,
                column=id_token.column
            )

        # Parenthesized expression or tuple
        if self.match(TokenType.LPAREN):
            self.advance()

            # Empty tuple
            if self.match(TokenType.RPAREN):
                close = self.advance()
                return TupleLiteral(elements=[], line=token.line, column=token.column)

            first = self.parse_expression()

            # Tuple
            if self.match(TokenType.COMMA):
                elements = [first]
                while self.match(TokenType.COMMA):
                    self.advance()
                    if not self.match(TokenType.RPAREN):
                        elements.append(self.parse_expression())

                self.expect(TokenType.RPAREN)
                return TupleLiteral(
                    elements=elements,
                    line=token.line,
                    column=token.column
                )

            # Parenthesized expression
            self.expect(TokenType.RPAREN)
            return first

        self.error(f"Unexpected token: {token.type.name}")

    def parse_list_literal(self) -> ListLiteral:
        """Parse list literal: [1, 2, 3]"""
        start = self.expect(TokenType.LBRACKET)
        elements = []

        while not self.match(TokenType.RBRACKET):
            elements.append(self.parse_expression())

            if self.match(TokenType.COMMA):
                self.advance()

        self.expect(TokenType.RBRACKET)

        return ListLiteral(
            elements=elements,
            line=start.line,
            column=start.column
        )

    def parse_record_literal(self) -> RecordLiteral:
        """Parse record literal: {name: "Layer", version: 1}"""
        start = self.expect(TokenType.LBRACE)
        fields = {}

        while not self.match(TokenType.RBRACE):
            key_token = self.expect(TokenType.IDENTIFIER)
            self.expect(TokenType.COLON)
            value = self.parse_expression()

            fields[key_token.value] = value

            if self.match(TokenType.COMMA):
                self.advance()

        self.expect(TokenType.RBRACE)

        return RecordLiteral(
            fields=fields,
            line=start.line,
            column=start.column
        )

    def get_precedence(self, token_type: TokenType) -> int:
        """Get operator precedence."""
        precedence_table = {
            TokenType.OR: 1,
            TokenType.AND: 2,
            TokenType.EQ: 3,
            TokenType.NEQ: 3,
            TokenType.LT: 4,
            TokenType.GT: 4,
            TokenType.LTE: 4,
            TokenType.GTE: 4,
            TokenType.PLUS: 5,
            TokenType.MINUS: 5,
            TokenType.STAR: 6,
            TokenType.SLASH: 6,
            TokenType.PERCENT: 6,
            TokenType.CARET: 7,
        }
        return precedence_table.get(token_type, 0)


def parse(tokens: List[Token]) -> Program:
    """Convenience function to parse tokens."""
    parser = Parser(tokens)
    return parser.parse()


if __name__ == "__main__":
    from layer.lexer import tokenize

    test_code = """
    @layers [core, debug]

    @layer core {
        fn factorial(n) =
            match n with
            | 0 -> 1
            | n -> n * factorial(n - 1)
    }

    factorial(5)
    """

    tokens = tokenize(test_code)
    ast = parse(tokens)

    from layer.ast_nodes import ast_to_dict
    import json
    print(json.dumps(ast_to_dict(ast), indent=2))
