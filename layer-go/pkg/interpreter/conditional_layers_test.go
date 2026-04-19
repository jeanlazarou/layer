package interpreter

import (
	"testing"

	"github.com/yourusername/layer/pkg/lexer"
	"github.com/yourusername/layer/pkg/parser"
)

// helper: execute source with pre-seeded variables
func executeWithVars(t *testing.T, source string, vars map[string]interface{}) interface{} {
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
	for k, v := range vars {
		interp.SetVar(k, v)
	}
	result, err := interp.Execute(program)
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	return result
}

// ── Literal conditions ────────────────────────────────────────────────────────

func TestConditionalLayerTrue(t *testing.T) {
	// Layer is enabled when condition is literal true
	source := `
	@layers [core, extra if true]
	@layer core { fn greet() = "core" }
	@layer extra { fn greet() = "extra" }
	greet()
	`
	if result := mustExecute(t, source); result != "extra" {
		t.Errorf("expected \"extra\", got %v", result)
	}
}

func TestConditionalLayerFalse(t *testing.T) {
	// Layer is skipped when condition is literal false — core version runs
	source := `
	@layers [core, extra if false]
	@layer core { fn greet() = "core" }
	@layer extra { fn greet() = "extra" }
	greet()
	`
	if result := mustExecute(t, source); result != "core" {
		t.Errorf("expected \"core\", got %v", result)
	}
}

// ── Variable-based conditions ─────────────────────────────────────────────────

func TestConditionalLayerVariableEnabled(t *testing.T) {
	source := `
	@layers [core, debug if mode == "development"]
	@layer core  { fn run() = "core" }
	@layer debug { fn run() = "debug" }
	run()
	`
	result := executeWithVars(t, source, map[string]interface{}{"mode": "development"})
	if result != "debug" {
		t.Errorf("expected \"debug\", got %v", result)
	}
}

func TestConditionalLayerVariableDisabled(t *testing.T) {
	source := `
	@layers [core, debug if mode == "development"]
	@layer core  { fn run() = "core" }
	@layer debug { fn run() = "debug" }
	run()
	`
	result := executeWithVars(t, source, map[string]interface{}{"mode": "production"})
	if result != "core" {
		t.Errorf("expected \"core\", got %v", result)
	}
}

// ── Multiple conditions in the same directive ─────────────────────────────────

func TestConditionalLayerOnlyMiddleSkipped(t *testing.T) {
	// Three layers; only the middle one is conditional and disabled.
	// top overrides core, middle is skipped, so top's version runs.
	source := `
	@layers [core, middle if false, top]
	@layer core   { fn f() = "core"   }
	@layer middle { fn f() = "middle" }
	@layer top    { fn f() = "top"    }
	f()
	`
	if result := mustExecute(t, source); result != "top" {
		t.Errorf("expected \"top\", got %v", result)
	}
}

func TestConditionalLayerBothEnabled(t *testing.T) {
	source := `
	@layers [core, a if true, b if true]
	@layer core { fn f() = "core" }
	@layer a    { fn f() = "a"    }
	@layer b    { fn f() = "b"    }
	f()
	`
	// b is topmost enabled → wins
	if result := mustExecute(t, source); result != "b" {
		t.Errorf("expected \"b\", got %v", result)
	}
}

// ── Condition expression using operators ─────────────────────────────────────

func TestConditionalLayerCompoundCondition(t *testing.T) {
	source := `
	@layers [core, debug if level > 0 and level < 3]
	@layer core  { fn run() = "core"  }
	@layer debug { fn run() = "debug" }
	run()
	`
	result := executeWithVars(t, source, map[string]interface{}{"level": int64(1)})
	if result != "debug" {
		t.Errorf("expected \"debug\", got %v", result)
	}
}

// ── Non-boolean condition is a runtime error ──────────────────────────────────

func TestConditionalLayerNonBoolIsError(t *testing.T) {
	source := `
	@layers [core, extra if 42]
	@layer core  { fn f() = "core" }
	@layer extra { fn f() = "extra" }
	f()
	`
	tokens, _ := lexer.Tokenize(source)
	program, _ := parser.Parse(tokens)
	_, err := NewInterpreter().Execute(program)
	if err == nil {
		t.Error("expected runtime error for non-boolean condition, got nil")
	}
}

// ── Unconditional layers are unaffected ───────────────────────────────────────

func TestUnconditionalLayersStillWork(t *testing.T) {
	// Mix of conditional and unconditional — make sure nothing regresses
	source := `
	@layers [core, debug if false, ui]
	@layer core  { fn tag() = "core" }
	@layer debug { fn tag() = "debug" }
	@layer ui    { fn tag() = "ui"   }
	tag()
	`
	// debug skipped, ui active → ui wins
	if result := mustExecute(t, source); result != "ui" {
		t.Errorf("expected \"ui\", got %v", result)
	}
}
