package interpreter

import "testing"

// These tests lock in Layer's open-recursion semantics: unqualified calls
// inside a layer function always resolve through the full layer stack (late
// binding), so a recursive function overridden by a higher layer re-enters
// that higher layer at every recursion level. Explicit layer-qualified calls
// (core.fib) are the escape hatch and never re-enter higher layers on the
// qualified call itself.

func TestOpenRecursionReentersWrappingLayer(t *testing.T) {
	// wrap adds 100 on every entry. With open recursion the unqualified
	// fib(n-1)/fib(n-2) calls inside core re-enter wrap, so the +100 is
	// applied at every level of the recursion tree:
	//   W(n) = core.fib(n) + 100, core.fib(n) = W(n-1) + W(n-2)
	//   W(0)=100, W(1)=101, W(2)=301, W(3)=502
	// Closed recursion would instead give core.fib(3) + 100 = 102.
	source := `
	@layers [core, wrap]
	@layer core {
	  fn fib(n) =
	    match n with
	    | 0 -> 0
	    | 1 -> 1
	    | n -> fib(n - 1) + fib(n - 2)
	}
	@layer wrap {
	  fn fib(n) = core.fib(n) + 100
	}
	fib(3)
	`
	result := mustExecute(t, source)
	if result != int64(502) {
		t.Errorf("expected open recursion result 502, got %v", result)
	}
}

func TestQualifiedRecursionStaysClosed(t *testing.T) {
	// When core recurses through an explicit core.fact reference, the
	// wrapping layer is bypassed for the whole recursion: only the initial
	// unqualified call goes through wrap.
	//   core-closed fact(3) = 6, wrap adds 100 once -> 106
	source := `
	@layers [core, wrap]
	@layer core {
	  fn fact(n) =
	    match n with
	    | 0 -> 1
	    | n -> n * core.fact(n - 1)
	}
	@layer wrap {
	  fn fact(n) = core.fact(n) + 100
	}
	fact(3)
	`
	result := mustExecute(t, source)
	if result != int64(106) {
		t.Errorf("expected closed recursion result 106, got %v", result)
	}
}

func TestOpenRecursionWithoutOverrideIsPlainRecursion(t *testing.T) {
	// With no higher-layer override, open recursion is indistinguishable
	// from ordinary recursion.
	source := `
	@layers [core]
	@layer core {
	  fn fact(n) =
	    match n with
	    | 0 -> 1
	    | n -> n * fact(n - 1)
	}
	fact(5)
	`
	result := mustExecute(t, source)
	if result != int64(120) {
		t.Errorf("expected 120, got %v", result)
	}
}
