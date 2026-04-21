package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

// Lexer tokenizes Layer source code
type Lexer struct {
	source  string
	pos     int
	line    int
	column  int
	tokens  []Token
	runes   []rune
}

// Keywords maps keyword strings to their token types
var keywords = map[string]TokenType{
	"fn":     FN,
	"let":    LET,
	"in":     IN,
	"match":  MATCH,
	"with":   WITH,
	"true":   TRUE,
	"false":  FALSE,
	"and":    AND,
	"or":     OR,
	"not":    NOT,
	"layer":  LAYER,
	"layers": LAYERS,
}

// NewLexer creates a new lexer for the given source code
func NewLexer(source string) *Lexer {
	return &Lexer{
		source: source,
		pos:    0,
		line:   1,
		column: 1,
		tokens: make([]Token, 0),
		runes:  []rune(source),
	}
}

// error raises a lexer error with position information
func (l *Lexer) error(message string) error {
	return fmt.Errorf("lexer error at %d:%d: %s", l.line, l.column, message)
}

// peek looks ahead at a rune without consuming it
func (l *Lexer) peek(offset int) rune {
	pos := l.pos + offset
	if pos < len(l.runes) {
		return l.runes[pos]
	}
	return 0
}

// advance consumes and returns the next rune
func (l *Lexer) advance() rune {
	if l.pos >= len(l.runes) {
		return 0
	}

	char := l.runes[l.pos]
	l.pos++

	if char == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}

	return char
}

// skipWhitespace skips whitespace characters except newlines
func (l *Lexer) skipWhitespace() {
	for l.peek(0) != 0 && (l.peek(0) == ' ' || l.peek(0) == '\t' || l.peek(0) == '\r') {
		l.advance()
	}
}

// readComment reads a comment (starts with #)
func (l *Lexer) readComment() Token {
	startLine := l.line
	startColumn := l.column

	l.advance() // Skip #

	var comment strings.Builder
	for l.peek(0) != 0 && l.peek(0) != '\n' {
		comment.WriteRune(l.advance())
	}

	return NewToken(COMMENT, strings.TrimSpace(comment.String()), startLine, startColumn)
}

// readNumber reads a numeric literal
func (l *Lexer) readNumber() (Token, error) {
	startLine := l.line
	startColumn := l.column

	var numStr strings.Builder
	hasDot := false

	for l.peek(0) != 0 && (unicode.IsDigit(l.peek(0)) || l.peek(0) == '.') {
		if l.peek(0) == '.' {
			if hasDot {
				break
			}
			// Check if it's ellipsis
			if l.peek(1) == '.' && l.peek(2) == '.' {
				break
			}
			hasDot = true
		}
		numStr.WriteRune(l.advance())
	}

	numString := numStr.String()
	var value interface{}

	if hasDot {
		// Parse as float
		var f float64
		_, err := fmt.Sscanf(numString, "%f", &f)
		if err != nil {
			return Token{}, l.error(fmt.Sprintf("invalid number: %s", numString))
		}
		value = f
	} else {
		// Parse as int
		var i int64
		_, err := fmt.Sscanf(numString, "%d", &i)
		if err != nil {
			return Token{}, l.error(fmt.Sprintf("invalid number: %s", numString))
		}
		value = i
	}

	return NewToken(NUMBER, value, startLine, startColumn), nil
}

// readString reads a string literal
func (l *Lexer) readString() (Token, error) {
	startLine := l.line
	startColumn := l.column

	quote := l.advance() // " or '
	var str strings.Builder

	for l.peek(0) != 0 && l.peek(0) != quote {
		if l.peek(0) == '\\' {
			l.advance()
			nextChar := l.advance()
			// Handle escape sequences
			switch nextChar {
			case 'n':
				str.WriteRune('\n')
			case 't':
				str.WriteRune('\t')
			case 'r':
				str.WriteRune('\r')
			case '\\':
				str.WriteRune('\\')
			case '"':
				str.WriteRune('"')
			case '\'':
				str.WriteRune('\'')
			default:
				str.WriteRune(nextChar)
			}
		} else {
			str.WriteRune(l.advance())
		}
	}

	if l.peek(0) == 0 {
		return Token{}, l.error("unterminated string")
	}

	l.advance() // Closing quote
	return NewToken(STRING, str.String(), startLine, startColumn), nil
}

// readIdentifier reads an identifier or keyword
func (l *Lexer) readIdentifier() Token {
	startLine := l.line
	startColumn := l.column

	var ident strings.Builder
	for l.peek(0) != 0 && (unicode.IsLetter(l.peek(0)) || unicode.IsDigit(l.peek(0)) || l.peek(0) == '_') {
		ident.WriteRune(l.advance())
	}

	identStr := ident.String()

	// Check if it's a keyword
	if tokenType, ok := keywords[identStr]; ok {
		// Handle boolean values
		if tokenType == TRUE {
			return NewToken(BOOLEAN, true, startLine, startColumn)
		} else if tokenType == FALSE {
			return NewToken(BOOLEAN, false, startLine, startColumn)
		}
		return NewToken(tokenType, identStr, startLine, startColumn)
	}

	return NewToken(IDENTIFIER, identStr, startLine, startColumn)
}

// Tokenize tokenizes the entire source code
func (l *Lexer) Tokenize() ([]Token, error) {
	for l.pos < len(l.runes) {
		l.skipWhitespace()

		if l.peek(0) == 0 {
			break
		}

		char := l.peek(0)
		startLine := l.line
		startColumn := l.column

		// Comments
		if char == '#' {
			l.readComment() // Skip comments
			continue
		}

		// Newlines
		if char == '\n' {
			l.advance()
			// Skip newlines for now (we can make them significant later)
			continue
		}

		// Numbers
		if unicode.IsDigit(char) {
			token, err := l.readNumber()
			if err != nil {
				return nil, err
			}
			l.tokens = append(l.tokens, token)
			continue
		}

		// Strings
		if char == '"' || char == '\'' {
			token, err := l.readString()
			if err != nil {
				return nil, err
			}
			l.tokens = append(l.tokens, token)
			continue
		}

		// Lambda (λ)
		if char == 'λ' {
			l.advance()
			l.tokens = append(l.tokens, NewToken(LAMBDA, "λ", startLine, startColumn))
			continue
		}

		// Identifiers and keywords
		if unicode.IsLetter(char) || char == '_' {
			l.tokens = append(l.tokens, l.readIdentifier())
			continue
		}

		// Two-character operators
		if char == '-' && l.peek(1) == '>' {
			l.advance()
			l.advance()
			l.tokens = append(l.tokens, NewToken(ARROW, "->", startLine, startColumn))
			continue
		}

		if char == '=' && l.peek(1) == '=' {
			l.advance()
			l.advance()
			l.tokens = append(l.tokens, NewToken(EQ, "==", startLine, startColumn))
			continue
		}

		if char == '!' && l.peek(1) == '=' {
			l.advance()
			l.advance()
			l.tokens = append(l.tokens, NewToken(NEQ, "!=", startLine, startColumn))
			continue
		}

		if char == '<' && l.peek(1) == '=' {
			l.advance()
			l.advance()
			l.tokens = append(l.tokens, NewToken(LTE, "<=", startLine, startColumn))
			continue
		}

		if char == '>' && l.peek(1) == '=' {
			l.advance()
			l.advance()
			l.tokens = append(l.tokens, NewToken(GTE, ">=", startLine, startColumn))
			continue
		}

		if char == '.' && l.peek(1) == '.' && l.peek(2) == '.' {
			l.advance()
			l.advance()
			l.advance()
			l.tokens = append(l.tokens, NewToken(ELLIPSIS, "...", startLine, startColumn))
			continue
		}

		// Single-character tokens
		singleCharTokens := map[rune]TokenType{
			'+': PLUS,
			'-': MINUS,
			'*': STAR,
			'/': SLASH,
			'%': PERCENT,
			'^': CARET,
			'<': LT,
			'>': GT,
			'(': LPAREN,
			')': RPAREN,
			'[': LBRACKET,
			']': RBRACKET,
			'{': LBRACE,
			'}': RBRACE,
			',': COMMA,
			':': COLON,
			';': SEMICOLON,
			'.': DOT,
			'=': ASSIGN,
			'|': PIPE,
			'@': AT,
			'?': QUESTION,
		}

		if tokenType, ok := singleCharTokens[char]; ok {
			l.advance()
			l.tokens = append(l.tokens, NewToken(tokenType, string(char), startLine, startColumn))
			continue
		}

		return nil, l.error(fmt.Sprintf("unexpected character: %c", char))
	}

	// Add EOF token
	l.tokens = append(l.tokens, NewToken(EOF, nil, l.line, l.column))
	return l.tokens, nil
}

// Tokenize is a convenience function to tokenize source code
func Tokenize(source string) ([]Token, error) {
	lexer := NewLexer(source)
	return lexer.Tokenize()
}
