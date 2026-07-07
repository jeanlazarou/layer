package parser

import (
	"testing"

	"github.com/jeanlazarou/layer/pkg/ast"
	"github.com/jeanlazarou/layer/pkg/lexer"
)

func TestParseNumber(t *testing.T) {
	tokens, _ := lexer.Tokenize("42")
	program, err := Parse(tokens)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if program.MainExpression == nil {
		t.Fatal("Expected main expression")
	}

	numLit, ok := program.MainExpression.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("Expected NumberLiteral, got %T", program.MainExpression)
	}

	if numLit.Value != int64(42) {
		t.Errorf("Expected value 42, got %v", numLit.Value)
	}
}

func TestParseBinaryOp(t *testing.T) {
	tokens, _ := lexer.Tokenize("2 + 3")
	program, err := Parse(tokens)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	binOp, ok := program.MainExpression.(*ast.BinaryOp)
	if !ok {
		t.Fatalf("Expected BinaryOp, got %T", program.MainExpression)
	}

	if binOp.Operator != "+" {
		t.Errorf("Expected operator +, got %s", binOp.Operator)
	}
}

func TestParseFunctionDef(t *testing.T) {
	source := `
	@layers [core]

	@layer core {
		fn add(x, y) = x + y
	}
	`

	tokens, _ := lexer.Tokenize(source)
	program, err := Parse(tokens)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(program.LayerDeclarations) != 1 {
		t.Fatalf("Expected 1 layer declaration, got %d", len(program.LayerDeclarations))
	}

	layerDecl := program.LayerDeclarations[0]
	if layerDecl.LayerName != "core" {
		t.Errorf("Expected layer name 'core', got '%s'", layerDecl.LayerName)
	}

	if len(layerDecl.Definitions) != 1 {
		t.Fatalf("Expected 1 definition, got %d", len(layerDecl.Definitions))
	}

	funcDef, ok := layerDecl.Definitions[0].(*ast.FunctionDef)
	if !ok {
		t.Fatalf("Expected FunctionDef, got %T", layerDecl.Definitions[0])
	}

	if funcDef.Name != "add" {
		t.Errorf("Expected function name 'add', got '%s'", funcDef.Name)
	}

	if len(funcDef.Params) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(funcDef.Params))
	}
}

func TestParseMatch(t *testing.T) {
	source := `
	match n with
	| 0 -> 1
	| n -> n * 2
	`

	tokens, _ := lexer.Tokenize(source)
	program, err := Parse(tokens)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	matchExpr, ok := program.MainExpression.(*ast.Match)
	if !ok {
		t.Fatalf("Expected Match, got %T", program.MainExpression)
	}

	if len(matchExpr.Cases) != 2 {
		t.Errorf("Expected 2 cases, got %d", len(matchExpr.Cases))
	}
}

func TestParseLetBinding(t *testing.T) {
	source := `let x = 42 in x + 1`

	tokens, _ := lexer.Tokenize(source)
	program, err := Parse(tokens)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	letBinding, ok := program.MainExpression.(*ast.LetBinding)
	if !ok {
		t.Fatalf("Expected LetBinding, got %T", program.MainExpression)
	}

	if letBinding.Name != "x" {
		t.Errorf("Expected name 'x', got '%s'", letBinding.Name)
	}
}

func TestParseFunctionCall(t *testing.T) {
	source := `factorial(5)`

	tokens, _ := lexer.Tokenize(source)
	program, err := Parse(tokens)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	funcCall, ok := program.MainExpression.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("Expected FunctionCall, got %T", program.MainExpression)
	}

	ident, ok := funcCall.Function.(*ast.Identifier)
	if !ok {
		t.Fatalf("Expected Identifier for function, got %T", funcCall.Function)
	}

	if ident.Name != "factorial" {
		t.Errorf("Expected function name 'factorial', got '%s'", ident.Name)
	}

	if len(funcCall.Arguments) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(funcCall.Arguments))
	}
}

func TestParseLambda(t *testing.T) {
	source := `λ(x, y) -> x + y`

	tokens, _ := lexer.Tokenize(source)
	program, err := Parse(tokens)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	lambda, ok := program.MainExpression.(*ast.Lambda)
	if !ok {
		t.Fatalf("Expected Lambda, got %T", program.MainExpression)
	}

	if len(lambda.Params) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(lambda.Params))
	}
}

func TestParseListLiteral(t *testing.T) {
	source := `[1, 2, 3, 4]`

	tokens, _ := lexer.Tokenize(source)
	program, err := Parse(tokens)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	list, ok := program.MainExpression.(*ast.ListLiteral)
	if !ok {
		t.Fatalf("Expected ListLiteral, got %T", program.MainExpression)
	}

	if len(list.Elements) != 4 {
		t.Errorf("Expected 4 elements, got %d", len(list.Elements))
	}
}

func TestParseRecordLiteral(t *testing.T) {
	source := `{name: "Layer", version: 1}`

	tokens, _ := lexer.Tokenize(source)
	program, err := Parse(tokens)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	record, ok := program.MainExpression.(*ast.RecordLiteral)
	if !ok {
		t.Fatalf("Expected RecordLiteral, got %T", program.MainExpression)
	}

	if len(record.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(record.Fields))
	}

	if _, ok := record.Fields["name"]; !ok {
		t.Error("Expected field 'name'")
	}

	if _, ok := record.Fields["version"]; !ok {
		t.Error("Expected field 'version'")
	}
}

func TestParseMemberAccess(t *testing.T) {
	source := `core.factorial`

	tokens, _ := lexer.Tokenize(source)
	program, err := Parse(tokens)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	memberAccess, ok := program.MainExpression.(*ast.MemberAccess)
	if !ok {
		t.Fatalf("Expected MemberAccess, got %T", program.MainExpression)
	}

	if memberAccess.Member != "factorial" {
		t.Errorf("Expected member 'factorial', got '%s'", memberAccess.Member)
	}
}

func TestParseCompleteProgram(t *testing.T) {
	source := `
	@layers [core, debug]

	@layer core {
		fn factorial(n) =
			match n with
			| 0 -> 1
			| n -> n * factorial(n - 1)
	}

	@layer debug {
		fn factorial(n) =
			let _ = log("Computing fib({n})") in
				core.factorial(n)
	}

	factorial(5)
	`

	tokens, _ := lexer.Tokenize(source)
	program, err := Parse(tokens)

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check layers directive
	if program.LayersDirective == nil {
		t.Fatal("Expected layers directive")
	}

	if len(program.LayersDirective.Layers) != 2 {
		t.Errorf("Expected 2 layers, got %d", len(program.LayersDirective.Layers))
	}

	// Check layer declarations
	if len(program.LayerDeclarations) != 2 {
		t.Errorf("Expected 2 layer declarations, got %d", len(program.LayerDeclarations))
	}

	// Check main expression
	if program.MainExpression == nil {
		t.Fatal("Expected main expression")
	}

	_, ok := program.MainExpression.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("Expected FunctionCall for main expression, got %T", program.MainExpression)
	}
}
