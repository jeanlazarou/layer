package interpreter

import (
	"testing"

	"github.com/jeanlazarou/layer/pkg/lexer"
	"github.com/jeanlazarou/layer/pkg/parser"
)

func TestInterpreterBasic(t *testing.T) {
	source := "42"
	tokens, _ := lexer.Tokenize(source)
	program, _ := parser.Parse(tokens)

	interp := NewInterpreter()
	result, err := interp.Execute(program)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != int64(42) {
		t.Errorf("Expected 42, got %v", result)
	}
}

func TestInterpreterArithmetic(t *testing.T) {
	source := "2 + 3 * 4"
	tokens, _ := lexer.Tokenize(source)
	program, _ := parser.Parse(tokens)

	interp := NewInterpreter()
	result, err := interp.Execute(program)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should be 14 (3 * 4 = 12, 2 + 12 = 14) due to precedence
	if result != int64(14) {
		t.Errorf("Expected 14, got %v", result)
	}
}

func TestInterpreterLetBinding(t *testing.T) {
	source := "let x = 5 in x + 3"
	tokens, _ := lexer.Tokenize(source)
	program, _ := parser.Parse(tokens)

	interp := NewInterpreter()
	result, err := interp.Execute(program)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != int64(8) {
		t.Errorf("Expected 8, got %v", result)
	}
}

func TestInterpreterFunctionCall(t *testing.T) {
	source := `
	@layers [core]

	@layer core {
		fn double(x) = x * 2
	}

	double(5)
	`

	tokens, _ := lexer.Tokenize(source)
	program, _ := parser.Parse(tokens)

	interp := NewInterpreter()
	result, err := interp.Execute(program)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != int64(10) {
		t.Errorf("Expected 10, got %v", result)
	}
}

func TestInterpreterRecursion(t *testing.T) {
	source := `
	@layers [core]

	@layer core {
		fn factorial(n) =
			match n with
			| 0 -> 1
			| n -> n * factorial(n - 1)
	}

	factorial(5)
	`

	tokens, _ := lexer.Tokenize(source)
	program, _ := parser.Parse(tokens)

	interp := NewInterpreter()
	result, err := interp.Execute(program)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != int64(120) {
		t.Errorf("Expected 120 (5!), got %v", result)
	}
}

func TestInterpreterLayerOverride(t *testing.T) {
	source := `
	@layers [core, debug]

	@layer core {
		fn add(x, y) = x + y
	}

	@layer debug {
		fn add(x, y) = core.add(x, y) + 1
	}

	add(2, 3)
	`

	tokens, _ := lexer.Tokenize(source)
	program, _ := parser.Parse(tokens)

	interp := NewInterpreter()
	result, err := interp.Execute(program)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// debug.add wraps core.add and adds 1
	// core.add(2, 3) = 5, debug.add adds 1 = 6
	if result != int64(6) {
		t.Errorf("Expected 6, got %v", result)
	}
}

func TestInterpreterListOperations(t *testing.T) {
	source := `
	let list = [1, 2, 3, 4] in
	let doubled = map(λ(x) -> x * 2, list) in
	head(doubled)
	`

	tokens, _ := lexer.Tokenize(source)
	program, _ := parser.Parse(tokens)

	interp := NewInterpreter()
	result, err := interp.Execute(program)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != int64(2) {
		t.Errorf("Expected 2, got %v", result)
	}
}

func TestInterpreterMatch(t *testing.T) {
	source := `
	let value = 0 in
	match value with
	| 0 -> "zero"
	| 1 -> "one"
	| n -> "other"
	`

	tokens, _ := lexer.Tokenize(source)
	program, _ := parser.Parse(tokens)

	interp := NewInterpreter()
	result, err := interp.Execute(program)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != "zero" {
		t.Errorf("Expected 'zero', got %v", result)
	}
}

func TestInterpreterUIComponent(t *testing.T) {
	source := `
	@layers [core, ui]

	@layer ui {
		fn page() = view({
			title: "Test",
			children: [text({content: "Hello"})]
		})
	}

	page()
	`

	tokens, _ := lexer.Tokenize(source)
	program, _ := parser.Parse(tokens)

	interp := NewInterpreter()
	result, err := interp.Execute(program)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should return a UIComponent
	if _, ok := result.(*UIComponent); !ok {
		t.Errorf("Expected UIComponent, got %T", result)
	}
}
