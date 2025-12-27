package lexer

import (
	"testing"
)

func TestLexerBasic(t *testing.T) {
	source := `
	# This is a comment
	fn factorial(n) =
		match n with
		| 0 -> 1
		| n -> n * factorial(n - 1)
	`

	tokens, err := Tokenize(source)
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	// Check we got some tokens
	if len(tokens) == 0 {
		t.Fatal("Expected tokens, got none")
	}

	// Check for EOF
	if tokens[len(tokens)-1].Type != EOF {
		t.Error("Expected EOF token at end")
	}
}

func TestNumbers(t *testing.T) {
	tests := []struct {
		input string
		want  interface{}
	}{
		{"42", int64(42)},
		{"3.14", 3.14},
		{"0", int64(0)},
	}

	for _, tt := range tests {
		tokens, err := Tokenize(tt.input)
		if err != nil {
			t.Errorf("Tokenize(%q) error: %v", tt.input, err)
			continue
		}

		if len(tokens) < 2 {
			t.Errorf("Tokenize(%q) returned too few tokens", tt.input)
			continue
		}

		if tokens[0].Type != NUMBER {
			t.Errorf("Tokenize(%q) expected NUMBER, got %v", tt.input, tokens[0].Type)
			continue
		}

		if tokens[0].Value != tt.want {
			t.Errorf("Tokenize(%q) = %v, want %v", tt.input, tokens[0].Value, tt.want)
		}
	}
}

func TestStrings(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"hello"`, "hello"},
		{`'world'`, "world"},
		{`"hello\nworld"`, "hello\nworld"},
		{`"escaped\"quote"`, `escaped"quote`},
	}

	for _, tt := range tests {
		tokens, err := Tokenize(tt.input)
		if err != nil {
			t.Errorf("Tokenize(%q) error: %v", tt.input, err)
			continue
		}

		if len(tokens) < 2 {
			t.Errorf("Tokenize(%q) returned too few tokens", tt.input)
			continue
		}

		if tokens[0].Type != STRING {
			t.Errorf("Tokenize(%q) expected STRING, got %v", tt.input, tokens[0].Type)
			continue
		}

		if tokens[0].Value != tt.want {
			t.Errorf("Tokenize(%q) = %q, want %q", tt.input, tokens[0].Value, tt.want)
		}
	}
}

func TestKeywords(t *testing.T) {
	tests := []struct {
		input string
		want  TokenType
	}{
		{"fn", FN},
		{"let", LET},
		{"in", IN},
		{"match", MATCH},
		{"with", WITH},
		{"true", BOOLEAN},
		{"false", BOOLEAN},
		{"and", AND},
		{"or", OR},
		{"not", NOT},
		{"layer", LAYER},
		{"layers", LAYERS},
	}

	for _, tt := range tests {
		tokens, err := Tokenize(tt.input)
		if err != nil {
			t.Errorf("Tokenize(%q) error: %v", tt.input, err)
			continue
		}

		if len(tokens) < 2 {
			t.Errorf("Tokenize(%q) returned too few tokens", tt.input)
			continue
		}

		if tokens[0].Type != tt.want {
			t.Errorf("Tokenize(%q) = %v, want %v", tt.input, tokens[0].Type, tt.want)
		}
	}
}

func TestOperators(t *testing.T) {
	tests := []struct {
		input string
		want  TokenType
	}{
		{"+", PLUS},
		{"-", MINUS},
		{"*", STAR},
		{"/", SLASH},
		{"==", EQ},
		{"!=", NEQ},
		{"<", LT},
		{">", GT},
		{"<=", LTE},
		{">=", GTE},
		{"->", ARROW},
		{"...", ELLIPSIS},
	}

	for _, tt := range tests {
		tokens, err := Tokenize(tt.input)
		if err != nil {
			t.Errorf("Tokenize(%q) error: %v", tt.input, err)
			continue
		}

		if len(tokens) < 2 {
			t.Errorf("Tokenize(%q) returned too few tokens", tt.input)
			continue
		}

		if tokens[0].Type != tt.want {
			t.Errorf("Tokenize(%q) = %v, want %v", tt.input, tokens[0].Type, tt.want)
		}
	}
}
