package interpreter

import (
	"strings"
	"testing"

	"github.com/jeanlazarou/layer/pkg/lexer"
	"github.com/jeanlazarou/layer/pkg/parser"
)

// executeExpectError runs source and returns the runtime error, failing the
// test if execution unexpectedly succeeds.
func executeExpectError(t *testing.T, source string) error {
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
	if err == nil {
		t.Fatalf("expected an error, got result: %v", result)
	}
	return err
}

func TestPureLayerCannotCallQualifiedImpure(t *testing.T) {
	source := `
	@layers [core, debug, logic]
	@layer logic {
	  fn compute(x) =
	    let _ = debug.log("computing") in
	      x * 2
	}
	compute(5)
	`
	err := executeExpectError(t, source)
	if !strings.Contains(err.Error(), "purity violation") || !strings.Contains(err.Error(), "debug.log") {
		t.Errorf("expected purity violation mentioning debug.log, got: %v", err)
	}
}

func TestPureLayerCannotCallUnqualifiedImpure(t *testing.T) {
	// log is only defined in the impure debug layer, so calling it
	// unqualified from a pure layer is an explicit effect.
	source := `
	@layers [core, debug, logic]
	@layer logic {
	  fn compute(x) =
	    let _ = log("computing") in
	      x * 2
	}
	compute(5)
	`
	err := executeExpectError(t, source)
	if !strings.Contains(err.Error(), "purity violation") {
		t.Errorf("expected purity violation, got: %v", err)
	}
}

func TestPureLayerCannotCallDataIO(t *testing.T) {
	source := `
	@layers [core, data, logic]
	@layer logic {
	  fn load() = data.read_file("input.txt")
	}
	load()
	`
	err := executeExpectError(t, source)
	if !strings.Contains(err.Error(), "purity violation") || !strings.Contains(err.Error(), "data.read_file") {
		t.Errorf("expected purity violation mentioning data.read_file, got: %v", err)
	}
}

func TestImpureAnnotationAllowsEffects(t *testing.T) {
	source := `
	@layers [core, debug, logging]
	@layer<impure> logging {
	  fn compute(x) =
	    let _ = debug.log("computing") in
	      x * 2
	}
	compute(5)
	`
	if result := mustExecute(t, source); result != int64(10) {
		t.Errorf("expected 10, got %v", result)
	}
}

func TestPureLayerCanUseOptionConstructors(t *testing.T) {
	// data.Some / data.None construct values without performing I/O, so
	// they are pure exceptions inside the impure data layer.
	source := `
	@layers [core, data, logic]
	@layer logic {
	  fn find(x) =
	    match x with
	    | 0 -> data.None()
	    | n -> data.Some(n)
	}
	data.unwrap(find(7), 0)
	`
	if result := mustExecute(t, source); result != int64(7) {
		t.Errorf("expected 7, got %v", result)
	}
}

func TestOpenRecursionThroughImpureWrapperIsAllowed(t *testing.T) {
	// core is pure and recurses unqualified; the impure debug layer
	// intercepts each recursive call. The interception is the outer
	// program's choice, not an effect requested by core, so it is legal.
	source := `
	@layers [core, debug]
	@layer core {
	  fn fib(n) =
	    match n with
	    | 0 -> 0
	    | 1 -> 1
	    | n -> fib(n - 1) + fib(n - 2)
	}
	@layer<impure> debug {
	  fn fib(n) =
	    let _ = log("fib({n})") in
	      core.fib(n)
	}
	fib(6)
	`
	if result := mustExecute(t, source); result != int64(8) {
		t.Errorf("expected 8, got %v", result)
	}
}

func TestTopLevelCodeIsUnrestricted(t *testing.T) {
	source := `
	@layers [core, debug]
	let _ = debug.log("hello from top level") in
	  42
	`
	if result := mustExecute(t, source); result != int64(42) {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestCoreCannotBeMarkedImpure(t *testing.T) {
	source := `
	@layers [core]
	@layer<impure> core {
	  fn f() = 1
	}
	f()
	`
	err := executeExpectError(t, source)
	if !strings.Contains(err.Error(), "core layer is always pure") {
		t.Errorf("expected core-purity error, got: %v", err)
	}
}

func TestBuiltinImpureLayerCannotBeMarkedPure(t *testing.T) {
	source := `
	@layers [core, debug]
	@layer<pure> debug {
	  fn f() = 1
	}
	f()
	`
	err := executeExpectError(t, source)
	if !strings.Contains(err.Error(), "always impure") {
		t.Errorf("expected fixed-impurity error, got: %v", err)
	}
}

func TestExplicitPureAnnotationEnforced(t *testing.T) {
	source := `
	@layers [core, debug, logic]
	@layer<pure> logic {
	  fn compute(x) =
	    let _ = debug.log("nope") in
	      x
	}
	compute(1)
	`
	err := executeExpectError(t, source)
	if !strings.Contains(err.Error(), "purity violation") {
		t.Errorf("expected purity violation, got: %v", err)
	}
}

func TestPureLayerCanCallPureLayer(t *testing.T) {
	source := `
	@layers [core, validation]
	@layer core {
	  fn calculate(x, y) = x + y
	}
	@layer validation {
	  fn safe_calculate(x, y) = core.calculate(x, y)
	}
	safe_calculate(2, 3)
	`
	if result := mustExecute(t, source); result != int64(5) {
		t.Errorf("expected 5, got %v", result)
	}
}
