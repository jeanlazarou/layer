package interpreter

import (
	"testing"

	"github.com/jeanlazarou/layer/pkg/lexer"
	"github.com/jeanlazarou/layer/pkg/parser"
)

func TestTake(t *testing.T) {
	source := `take(3, [1, 2, 3, 4, 5])`
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

	expected := []int64{1, 2, 3}
	for i, exp := range expected {
		if i >= len(result) {
			t.Errorf("Missing element at index %d", i)
			continue
		}
		val, ok := result[i].(int64)
		if !ok {
			t.Errorf("Element %d: expected int64, got %T", i, result[i])
			continue
		}
		if val != exp {
			t.Errorf("Element %d: expected %d, got %d", i, exp, val)
		}
	}
}

func TestDrop(t *testing.T) {
	source := `drop(2, [1, 2, 3, 4, 5])`
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

	expected := []int64{3, 4, 5}
	for i, exp := range expected {
		if i >= len(result) {
			t.Errorf("Missing element at index %d", i)
			continue
		}
		val, ok := result[i].(int64)
		if !ok {
			t.Errorf("Element %d: expected int64, got %T", i, result[i])
			continue
		}
		if val != exp {
			t.Errorf("Element %d: expected %d, got %d", i, exp, val)
		}
	}
}

func TestJoin(t *testing.T) {
	source := `join(", ", ["apple", "banana", "orange"])`
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

	expected := "apple, banana, orange"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestStrLength(t *testing.T) {
	source := `str_length("hello world")`
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

	result, ok := value.(int64)
	if !ok {
		t.Fatalf("Expected int64, got %T", value)
	}

	if result != 11 {
		t.Errorf("Expected 11, got %d", result)
	}
}

func TestPrint(t *testing.T) {
	source := `print(42)`
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

	result, ok := value.(int64)
	if !ok {
		t.Fatalf("Expected int64, got %T", value)
	}

	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

func TestId(t *testing.T) {
	source := `id(42)`
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

	result, ok := value.(int64)
	if !ok {
		t.Fatalf("Expected int64, got %T", value)
	}

	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

func TestConst(t *testing.T) {
	source := `
		let always42 = const(42) in
		always42(99)
	`
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

	result, ok := value.(int64)
	if !ok {
		t.Fatalf("Expected int64, got %T", value)
	}

	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

func TestCompose(t *testing.T) {
	source := `
@layers [core]

@layer core {
	fn add1(x) = x + 1
	fn mul2(x) = x * 2
	fn add1ThenMul2() = compose(mul2, add1)
}

add1ThenMul2()(5)
	`
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

	result, ok := value.(int64)
	if !ok {
		t.Fatalf("Expected int64, got %T", value)
	}

	// compose(mul2, add1)(5) = mul2(add1(5)) = mul2(6) = 12
	if result != 12 {
		t.Errorf("Expected 12, got %d", result)
	}
}

func TestFold(t *testing.T) {
	// fold is an alias for reduce
	source := `
@layers [core]

@layer core {
	fn add(acc, x) = acc + x
}

fold(add, 0, [1, 2, 3, 4, 5])
	`
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

	result, ok := value.(int64)
	if !ok {
		t.Fatalf("Expected int64, got %T", value)
	}

	if result != 15 {
		t.Errorf("Expected 15, got %d", result)
	}
}
