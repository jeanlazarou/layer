package interpreter

import (
	"testing"

	"github.com/jeanlazarou/layer/pkg/lexer"
	"github.com/jeanlazarou/layer/pkg/parser"
)

// mustExecuteError asserts that a program fails at runtime.
func mustExecuteError(t *testing.T, source string) {
	t.Helper()
	tokens, err := lexer.Tokenize(source)
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	_, err = NewInterpreter().Execute(program)
	if err == nil {
		t.Error("expected a runtime error, got nil")
	}
}

// ── @which ────────────────────────────────────────────────────────────────────

func TestWhichBasic(t *testing.T) {
	source := `
	@layers [core]
	@layer core { fn factorial(n) = n }
	@which factorial
	`
	if result := mustExecute(t, source); result != "core" {
		t.Errorf("expected \"core\", got %v", result)
	}
}

func TestWhichTopLayerWins(t *testing.T) {
	// Both layers define factorial — debug is on top, so @which returns "debug"
	source := `
	@layers [core, debug]
	@layer core  { fn factorial(n) = n }
	@layer debug { fn factorial(n) = n }
	@which factorial
	`
	if result := mustExecute(t, source); result != "debug" {
		t.Errorf("expected \"debug\", got %v", result)
	}
}

func TestWhichUndefinedIsError(t *testing.T) {
	mustExecuteError(t, `
	@layers [core]
	@which nonexistent
	`)
}

// ── @inspect ──────────────────────────────────────────────────────────────────

func TestInspectReturnsNames(t *testing.T) {
	source := `
	@layers [core]
	@layer core {
		fn add(x, y) = x + y
		fn sub(x, y) = x - y
	}
	@inspect core
	`
	result := mustExecute(t, source)
	list, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", result)
	}
	found := map[string]bool{}
	for _, v := range list {
		found[v.(string)] = true
	}
	for _, want := range []string{"add", "sub"} {
		if !found[want] {
			t.Errorf("expected '%s' in @inspect core result", want)
		}
	}
}

func TestInspectUnknownLayerIsError(t *testing.T) {
	mustExecuteError(t, `
	@layers [core]
	@inspect nosuchlayer
	`)
}

// ── @diff ─────────────────────────────────────────────────────────────────────

func TestDiffAdded(t *testing.T) {
	source := `
	@layers [core, extra]
	@layer core  { fn process(x) = x }
	@layer extra { fn process(x) = x
	               fn helper(x)  = x }
	@diff core extra
	`
	result := mustExecute(t, source)
	rec := result.(map[string]interface{})

	added := toStringSlice(t, rec["added"])
	if !containsStr(added, "helper") {
		t.Errorf("expected \"helper\" in added, got %v", added)
	}
	overridden := toStringSlice(t, rec["overridden"])
	if !containsStr(overridden, "process") {
		t.Errorf("expected \"process\" in overridden, got %v", overridden)
	}
}

func TestDiffNoOverrides(t *testing.T) {
	source := `
	@layers [core, extra]
	@layer core  { fn a(x) = x }
	@layer extra { fn b(x) = x }
	@diff core extra
	`
	result := mustExecute(t, source)
	rec := result.(map[string]interface{})

	overridden := toStringSlice(t, rec["overridden"])
	if len(overridden) != 0 {
		t.Errorf("expected empty overridden, got %v", overridden)
	}
	added := toStringSlice(t, rec["added"])
	if !containsStr(added, "b") {
		t.Errorf("expected \"b\" in added, got %v", added)
	}
}

func TestDiffUnknownLayerIsError(t *testing.T) {
	mustExecuteError(t, `
	@layers [core]
	@diff core nosuchlayer
	`)
}

// ── introspection as a first-class expression ─────────────────────────────────

func TestIntrospectionInsideLetBinding(t *testing.T) {
	source := `
	@layers [core]
	@layer core { fn foo(x) = x }
	let layer_name = @which foo in
	layer_name
	`
	if result := mustExecute(t, source); result != "core" {
		t.Errorf("expected \"core\", got %v", result)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func toStringSlice(t *testing.T, v interface{}) []string {
	t.Helper()
	list, ok := v.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", v)
	}
	out := make([]string, len(list))
	for i, item := range list {
		out[i] = item.(string)
	}
	return out
}

func containsStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
