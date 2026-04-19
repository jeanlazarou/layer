package interpreter

import (
	"testing"

	"github.com/yourusername/layer/pkg/lexer"
	"github.com/yourusername/layer/pkg/parser"
)

// ── Guard Clauses ─────────────────────────────────────────────────────────────

func TestGuardClauseBasic(t *testing.T) {
	source := `
	@layers [core]
	@layer core {
		fn classify(n) =
			match n with
			| n when n < 0   -> "negative"
			| 0              -> "zero"
			| n when n > 100 -> "large"
			| n              -> "positive"
	}
	classify(-5)
	`
	result := mustExecute(t, source)
	if result != "negative" {
		t.Errorf("expected \"negative\", got %v", result)
	}
}

func TestGuardClauseZero(t *testing.T) {
	source := `
	@layers [core]
	@layer core {
		fn classify(n) =
			match n with
			| n when n < 0   -> "negative"
			| 0              -> "zero"
			| n when n > 100 -> "large"
			| n              -> "positive"
	}
	classify(0)
	`
	result := mustExecute(t, source)
	if result != "zero" {
		t.Errorf("expected \"zero\", got %v", result)
	}
}

func TestGuardClauseLarge(t *testing.T) {
	source := `
	@layers [core]
	@layer core {
		fn classify(n) =
			match n with
			| n when n < 0   -> "negative"
			| 0              -> "zero"
			| n when n > 100 -> "large"
			| n              -> "positive"
	}
	classify(200)
	`
	result := mustExecute(t, source)
	if result != "large" {
		t.Errorf("expected \"large\", got %v", result)
	}
}

func TestGuardClausePositive(t *testing.T) {
	source := `
	@layers [core]
	@layer core {
		fn classify(n) =
			match n with
			| n when n < 0   -> "negative"
			| 0              -> "zero"
			| n when n > 100 -> "large"
			| n              -> "positive"
	}
	classify(42)
	`
	result := mustExecute(t, source)
	if result != "positive" {
		t.Errorf("expected \"positive\", got %v", result)
	}
}

func TestGuardClauseUsesPatternBinding(t *testing.T) {
	// Guard can reference the variable bound by the pattern
	source := `
	match [1, 2, 3] with
	| [h, ...rest] when h == 1 -> "starts with one"
	| [h, ...rest]             -> "other"
	`
	result := mustExecute(t, source)
	if result != "starts with one" {
		t.Errorf("expected \"starts with one\", got %v", result)
	}
}

func TestGuardClauseNonBoolError(t *testing.T) {
	source := `
	match 5 with
	| n when n + 1 -> "bad"
	`
	tokens, _ := lexer.Tokenize(source)
	program, _ := parser.Parse(tokens)
	interp := NewInterpreter()
	_, err := interp.Execute(program)
	if err == nil {
		t.Error("expected error for non-boolean guard, got nil")
	}
}

// ── Blend Mode: chain ────────────────────────────────────────────────────────

func TestBlendModeChain(t *testing.T) {
	// chain: bottom layer runs first, result is piped up to top layer.
	// core (bottom): process(x) = x + 1
	// transform (top): process(x) = x * 2
	// chain(process(5)) → core.process(5) = 6 → transform.process(6) = 12
	source := `
	@layers [core, transform] blend: chain

	@layer core {
		fn process(x) = x + 1
	}

	@layer transform {
		fn process(x) = x * 2
	}

	process(5)
	`
	result := mustExecute(t, source)
	if result != int64(12) {
		t.Errorf("expected 12, got %v", result)
	}
}

func TestBlendModeChainSingleLayer(t *testing.T) {
	// chain with only one layer defining the function behaves like override
	source := `
	@layers [core, extra] blend: chain

	@layer core {
		fn double(x) = x * 2
	}

	double(7)
	`
	result := mustExecute(t, source)
	if result != int64(14) {
		t.Errorf("expected 14, got %v", result)
	}
}

// ── Blend Mode: collect ───────────────────────────────────────────────────────

func TestBlendModeCollect(t *testing.T) {
	// collect: all layers' versions are called, results returned as a list (top to bottom)
	source := `
	@layers [core, extra] blend: collect

	@layer core {
		fn tag(x) = "core"
	}

	@layer extra {
		fn tag(x) = "extra"
	}

	tag("anything")
	`
	result := mustExecute(t, source)
	list, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", result)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 results, got %d", len(list))
	}
	// top layer first: extra, then core
	if list[0] != "extra" {
		t.Errorf("expected list[0] = \"extra\", got %v", list[0])
	}
	if list[1] != "core" {
		t.Errorf("expected list[1] = \"core\", got %v", list[1])
	}
}

// ── Blend Mode: fallback ──────────────────────────────────────────────────────

func TestBlendModeFallback(t *testing.T) {
	// fallback: top layer (cache) returns None, so falls back to bottom layer (db).
	// Requires the data layer for None/Some.
	source := `
	@layers [core, data, cache] blend: fallback

	@layer core {
		fn fetch(key) = data.Some("from_db")
	}

	@layer cache {
		fn fetch(key) = data.None()
	}

	fetch("mykey")
	`
	result := mustExecute(t, source)
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["type"] != "Some" {
		t.Errorf("expected Some, got %v", m["type"])
	}
	if m["value"] != "from_db" {
		t.Errorf("expected value \"from_db\", got %v", m["value"])
	}
}

func TestBlendModeFallbackTopSucceeds(t *testing.T) {
	// When the top layer returns a non-None result, it is used directly.
	source := `
	@layers [core, data, cache] blend: fallback

	@layer core {
		fn fetch(key) = data.Some("from_db")
	}

	@layer cache {
		fn fetch(key) = data.Some("from_cache")
	}

	fetch("mykey")
	`
	result := mustExecute(t, source)
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["value"] != "from_cache" {
		t.Errorf("expected \"from_cache\", got %v", m["value"])
	}
}

// ── Invalid blend mode ────────────────────────────────────────────────────────

func TestBlendModeInvalidIsParseError(t *testing.T) {
	source := `@layers [core] blend: bogus`
	tokens, _ := lexer.Tokenize(source)
	_, err := parser.Parse(tokens)
	if err == nil {
		t.Error("expected parse error for unknown blend mode, got nil")
	}
}

// ── helper ────────────────────────────────────────────────────────────────────

func mustExecute(t *testing.T, source string) interface{} {
	t.Helper()
	tokens, err := lexer.Tokenize(source)
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	interp := NewInterpreter()
	result, err := interp.Execute(program)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	return result
}
