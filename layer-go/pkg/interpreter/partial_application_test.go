package interpreter

import (
	"testing"

	"github.com/yourusername/layer/pkg/lexer"
	"github.com/yourusername/layer/pkg/parser"
)

// ── Partial application (? holes) ─────────────────────────────────────────────

func TestPartialApplicationSingleHole(t *testing.T) {
	// add(5, ?) returns a function; calling it with 3 gives 8
	source := `
	@layers [core]
	@layer core { fn add(x, y) = x + y }
	let add5 = add(5, ?) in
	add5(3)
	`
	result := mustExecute(t, source)
	if result != int64(8) {
		t.Errorf("expected 8, got %v", result)
	}
}

func TestPartialApplicationFirstHole(t *testing.T) {
	// add(?, 3) — hole in first position
	source := `
	@layers [core]
	@layer core { fn add(x, y) = x + y }
	let addTo3 = add(?, 3) in
	addTo3(5)
	`
	result := mustExecute(t, source)
	if result != int64(8) {
		t.Errorf("expected 8, got %v", result)
	}
}

func TestPartialApplicationMultipleHoles(t *testing.T) {
	// f(?, b, ?) — two holes produce a 2-param function
	source := `
	@layers [core]
	@layer core { fn combine(a, b, c) = a + b + c }
	let f = combine(?, 10, ?) in
	f(1, 100)
	`
	result := mustExecute(t, source)
	if result != int64(111) {
		t.Errorf("expected 111, got %v", result)
	}
}

func TestPartialApplicationWithMap(t *testing.T) {
	// map(add(1, ?), list) — partial application inside a HOF
	source := `
	@layers [core]
	@layer core { fn add(x, y) = x + y }
	map(add(1, ?), [10, 20, 30])
	`
	result := mustExecute(t, source)
	list, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected list, got %T", result)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list))
	}
	if list[0] != int64(11) || list[1] != int64(21) || list[2] != int64(31) {
		t.Errorf("expected [11 21 31], got %v", list)
	}
}

func TestPartialApplicationWithFilter(t *testing.T) {
	// filter(lt(?, 5), list) — keep elements less than 5
	source := `
	@layers [core]
	@layer core { fn lt(x, y) = x < y }
	filter(lt(?, 5), [1, 3, 7, 2, 8])
	`
	result := mustExecute(t, source)
	list, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected list, got %T", result)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 elements, got %d", list)
	}
}

func TestPartialApplicationChained(t *testing.T) {
	// Partial application result can itself be partially applied
	source := `
	@layers [core]
	@layer core { fn add3(a, b, c) = a + b + c }
	let step1 = add3(1, ?, ?) in
	let step2 = step1(2, ?) in
	step2(3)
	`
	result := mustExecute(t, source)
	if result != int64(6) {
		t.Errorf("expected 6, got %v", result)
	}
}

func TestPartialApplicationAllHoles(t *testing.T) {
	// add(?, ?) is equivalent to the original function
	source := `
	@layers [core]
	@layer core { fn mul(x, y) = x * y }
	let f = mul(?, ?) in
	f(6, 7)
	`
	result := mustExecute(t, source)
	if result != int64(42) {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestPartialApplicationInLetBody(t *testing.T) {
	// ? used inline, not stored in a let binding
	source := `
	@layers [core]
	@layer core {
		fn scale(x, factor) = x * factor
		fn double(x) = scale(x, ?)
	}
	double(5)(2)
	`
	// double(5) returns scale(5, ?), i.e. fn(__h0) = scale(5, __h0)
	// calling that with 2 → scale(5, 2) = 10
	result := mustExecute(t, source)
	if result != int64(10) {
		t.Errorf("expected 10, got %v", result)
	}
}

func TestPartialApplicationLexerTokenisesQuestion(t *testing.T) {
	// Verify ? is tokenised as QUESTION, not an error
	tokens, err := lexer.Tokenize("add(?, 3)")
	if err != nil {
		t.Fatalf("unexpected lex error: %v", err)
	}
	found := false
	for _, tok := range tokens {
		if tok.Type == lexer.QUESTION {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected QUESTION token in token stream")
	}
}

func TestPartialApplicationParserDesugarsToLambda(t *testing.T) {
	// add(5, ?) should parse without error and execute correctly
	source := `
	@layers [core]
	@layer core { fn add(x, y) = x + y }
	add(5, ?)(3)
	`
	result := mustExecute(t, source)
	if result != int64(8) {
		t.Errorf("expected 8, got %v", result)
	}
}

// mustExecuteError is defined in introspection_test.go
// mustExecute is defined in features_test.go

func TestPartialApplicationNoConflictWithUserVars(t *testing.T) {
	// A user variable that would have conflicted with the old __h0 scheme.
	// With gensym names starting with \x00, there is no conflict possible.
	source := `
	@layers [core]
	@layer core { fn add(x, y) = x + y }
	let __h0 = 99 in
	let f = add(?, 1) in
	f(10)
	`
	// f(10) should be add(10, 1) = 11, not add(99, 1) = 100
	result := mustExecute(t, source)
	if result != int64(11) {
		t.Errorf("expected 11 (no conflict with __h0=99), got %v", result)
	}
}

func TestPartialApplicationNoHolesUnchanged(t *testing.T) {
	// A call without ? behaves identically to before
	tokens, err := lexer.Tokenize("add(5, 3)")
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	_ = program // just verifying no panic/error
}
