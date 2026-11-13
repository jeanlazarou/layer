package runner

import "testing"

func TestRunReportsKinds(t *testing.T) {
	cases := []struct{ src, kind string }{
		{"1 + 1", "value"},
		{"@layers [core, ui]\nview({title: \"t\"})", "ui"},
		{"@layers [core, ui]\nrender(view({title: \"t\"}))", "ui"},
		{"@layers [core]\n@layer core {\n fn f() = 1\n}", "none"},
	}
	for _, c := range cases {
		res := Run(c.src, Options{})
		if !res.Success || res.Kind != c.kind {
			t.Errorf("%q: success=%v kind=%q errors=%v, want kind %q", c.src, res.Success, res.Kind, res.Errors, c.kind)
		}
	}
}

func TestIssuesAreLocated(t *testing.T) {
	cases := []struct {
		src       string
		kind      string
		line, col int
	}{
		{"\"unterminated", "lexer", 1, 14},
		{"@layers [core]\n(1 + ", "parser", 2, 6},
		{"@layers [core]\nnope", "check", 2, 1},
		{"@layers [core]\n1 + \"a\"", "runtime", 2, 3},
	}
	for _, c := range cases {
		res := Run(c.src, Options{})
		if res.Success || len(res.Errors) != 1 {
			t.Fatalf("%q: expected one error, got %+v", c.src, res)
		}
		is := res.Errors[0]
		if is.Kind != c.kind || is.Line != c.line || is.Column != c.col {
			t.Errorf("%q: got %s at %d:%d, want %s at %d:%d (%s)", c.src, is.Kind, is.Line, is.Column, c.kind, c.line, c.col, is.Message)
		}
	}
}

func TestRunFailsCleanlyInsideBuiltins(t *testing.T) {
	// Calling a non-function from inside a builtin is a clean error, not a crash.
	res := Run("@layers [core]\nreduce(λ(a, b) -> a(b), 1, [2])", Options{})
	if res.Success {
		t.Fatalf("expected failure, got %+v", res)
	}
}

func TestDescribe(t *testing.T) {
	res := Describe("@layers [core, debug]\n@layer core {\n fn f(x) = x\n}\n@layer debug {\n fn f(x) = core.f(x)\n}\nf(1)")
	if !res.Success || res.Description == nil {
		t.Fatalf("%+v", res)
	}
	r := res.Description.Resolutions
	if len(r) != 1 || r[0].Name != "f" || len(r[0].Layers) != 2 || r[0].Layers[0] != "debug" {
		t.Errorf("resolutions: %+v", r)
	}

	// Check errors still produce a description (the editor keeps showing
	// the stack while you fix a typo).
	res = Describe("@layers [core]\n@layer core {\n fn f(x) = typo(x)\n}\n1")
	if !res.Success || res.Description == nil || len(res.Errors) != 1 {
		t.Errorf("with a check error: %+v", res)
	}
}
