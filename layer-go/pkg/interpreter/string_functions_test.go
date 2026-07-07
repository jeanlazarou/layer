package interpreter

import (
	"testing"

	"github.com/jeanlazarou/layer/pkg/lexer"
	"github.com/jeanlazarou/layer/pkg/parser"
)

func TestStringConcatenation(t *testing.T) {
	source := `"Hello" + " " + "World"`
	tokens, err := lexer.Tokenize(source)
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}

	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	interp := NewInterpreter()
	value, err := interp.Execute(program)
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	result, ok := value.(string)
	if !ok {
		t.Fatalf("Expected string, got %T", value)
	}

	if result != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", result)
	}
}

func TestToString(t *testing.T) {
	source := `to_string(42) + " is the answer"`
	tokens, err := lexer.Tokenize(source)
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}

	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	interp := NewInterpreter()
	value, err := interp.Execute(program)
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	result, ok := value.(string)
	if !ok {
		t.Fatalf("Expected string, got %T", value)
	}

	if result != "42 is the answer" {
		t.Errorf("Expected '42 is the answer', got '%s'", result)
	}
}

func TestToFloat(t *testing.T) {
	source := `to_float("3.14") + 1.0`
	tokens, err := lexer.Tokenize(source)
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}

	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	interp := NewInterpreter()
	value, err := interp.Execute(program)
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	result, ok := value.(float64)
	if !ok {
		t.Fatalf("Expected float64, got %T", value)
	}

	expected := 4.14
	if result < expected-0.001 || result > expected+0.001 {
		t.Errorf("Expected %.2f, got %.2f", expected, result)
	}
}

func TestSplit(t *testing.T) {
	source := `split(" ", "7 8 9")`
	tokens, err := lexer.Tokenize(source)
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}

	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	interp := NewInterpreter()
	value, err := interp.Execute(program)
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	result, ok := value.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", value)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(result))
	}

	expected := []string{"7", "8", "9"}
	for i, exp := range expected {
		if i >= len(result) {
			t.Errorf("Missing element at index %d", i)
			continue
		}
		str, ok := result[i].(string)
		if !ok {
			t.Errorf("Element %d: expected string, got %T", i, result[i])
			continue
		}
		if str != exp {
			t.Errorf("Element %d: expected '%s', got '%s'", i, exp, str)
		}
	}
}
