package lexer

import "fmt"

// TokenType represents the type of a token
type TokenType int

const (
	// Literals
	NUMBER TokenType = iota
	STRING
	BOOLEAN
	TEMPLATE // double-quoted string with {expr} interpolation; Value is []StringPart

	// Identifiers and Keywords
	IDENTIFIER

	// Keywords
	FN
	LET
	IN
	MATCH
	WITH
	LAMBDA
	TRUE
	FALSE

	// Layer keywords
	LAYER
	LAYERS
	AT

	// Operators
	PLUS
	MINUS
	STAR
	SLASH
	PERCENT
	CARET

	// Comparison
	EQ
	NEQ
	LT
	GT
	LTE
	GTE

	// Logical
	AND
	OR
	NOT

	// Delimiters
	LPAREN
	RPAREN
	LBRACKET
	RBRACKET
	LBRACE
	RBRACE

	// Punctuation
	COMMA
	COLON
	SEMICOLON
	DOT
	ARROW
	ASSIGN
	PIPE
	ELLIPSIS
	QUESTION // ? — partial application hole

	// Special
	NEWLINE
	EOF
	COMMENT
)

var tokenNames = map[TokenType]string{
	NUMBER:     "NUMBER",
	STRING:     "STRING",
	BOOLEAN:    "BOOLEAN",
	TEMPLATE:   "TEMPLATE",
	IDENTIFIER: "IDENTIFIER",
	FN:         "FN",
	LET:        "LET",
	IN:         "IN",
	MATCH:      "MATCH",
	WITH:       "WITH",
	LAMBDA:     "LAMBDA",
	TRUE:       "TRUE",
	FALSE:      "FALSE",
	LAYER:      "LAYER",
	LAYERS:     "LAYERS",
	AT:         "AT",
	PLUS:       "PLUS",
	MINUS:      "MINUS",
	STAR:       "STAR",
	SLASH:      "SLASH",
	PERCENT:    "PERCENT",
	CARET:      "CARET",
	EQ:         "EQ",
	NEQ:        "NEQ",
	LT:         "LT",
	GT:         "GT",
	LTE:        "LTE",
	GTE:        "GTE",
	AND:        "AND",
	OR:         "OR",
	NOT:        "NOT",
	LPAREN:     "LPAREN",
	RPAREN:     "RPAREN",
	LBRACKET:   "LBRACKET",
	RBRACKET:   "RBRACKET",
	LBRACE:     "LBRACE",
	RBRACE:     "RBRACE",
	COMMA:      "COMMA",
	COLON:      "COLON",
	SEMICOLON:  "SEMICOLON",
	DOT:        "DOT",
	ARROW:      "ARROW",
	ASSIGN:     "ASSIGN",
	PIPE:       "PIPE",
	ELLIPSIS:   "ELLIPSIS",
	QUESTION:   "QUESTION",
	NEWLINE:    "NEWLINE",
	EOF:        "EOF",
	COMMENT:    "COMMENT",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", t)
}

// Token represents a single token in the source code
type Token struct {
	Type   TokenType
	Value  interface{}
	Line   int
	Column int
}

// StringPart is one segment of a TEMPLATE token: either literal text or the
// source of an interpolated expression. Line and Column locate the first
// character of Expr in the original source, so the parser can report
// errors inside interpolations at their real position.
type StringPart struct {
	Text   string
	Expr   string
	IsExpr bool
	Line   int
	Column int
}

// NewToken creates a new token
func NewToken(tokenType TokenType, value interface{}, line, column int) Token {
	return Token{
		Type:   tokenType,
		Value:  value,
		Line:   line,
		Column: column,
	}
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%s, %v, %d:%d)", t.Type, t.Value, t.Line, t.Column)
}
