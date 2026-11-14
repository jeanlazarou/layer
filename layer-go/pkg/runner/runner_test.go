package runner

import (
	"strings"
	"testing"
)

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

const counterApp = `@layers [core, ui]
@layer core {
  fn update(msg, m) =
    match msg with
    | "inc"     -> {m | n: m.n + 1}
    | {set: v}  -> {m | n: v}
    | "bad"     -> {m | n: λ(x) -> x}
    | "boom"    -> {m | n: m.n + "x"}
}
@layer ui {
  fn page(m) = text("n = {m.n}")
}
app({init: {n: 0}, update: update, view: page})`

func TestAppStartsFromInit(t *testing.T) {
	res := Run(counterApp, Options{})
	if !res.Success || res.Kind != "app" || string(res.Model) != `{"n":0}` || res.ModelText != "{n: 0}" ||
		res.Display != `<text content="n = 0" />` {
		t.Errorf("%+v", res)
	}
}

func TestAppResumesFromModelAndAppliesMessages(t *testing.T) {
	res := Run(counterApp, Options{Model: `{"n":10}`, Messages: []string{`"inc"`, `"inc"`}})
	if !res.Success || string(res.Model) != `{"n":12}` || res.Display != `<text content="n = 12" />` {
		t.Errorf("%+v", res)
	}
	// Floats survive the round trip as floats.
	res = Run(counterApp, Options{Messages: []string{`{"set":2.0}`}})
	if string(res.Model) != `{"n":2.0}` {
		t.Errorf("model %s", res.Model)
	}
}

func TestAppErrors(t *testing.T) {
	cases := []struct {
		opts Options
		want string
	}{
		{Options{Messages: []string{`"bad"`}}, `update returned a model that is not plain data after message "bad": model.n is a function`},
		{Options{Messages: []string{`"boom"`}}, "operator + needs two numbers or two strings"},
		{Options{Model: `{"n":`}, "invalid model"},
		{Options{Send: []string{`λ(x) -> x`}}, "must be plain data"},
	}
	for _, c := range cases {
		res := Run(counterApp, c.opts)
		if res.Success || len(res.Errors) == 0 || !strings.Contains(res.Errors[0].Message, c.want) {
			t.Errorf("%+v: got %+v, want %q", c.opts, res.Errors, c.want)
		}
	}
	// A runtime error inside update is located in the program.
	res := Run(counterApp, Options{Messages: []string{`"boom"`}})
	if res.Errors[0].Line != 8 || res.Errors[0].Trace != "core.update" {
		t.Errorf("location: %+v", res.Errors[0])
	}
}

func TestAppViewMustReturnUI(t *testing.T) {
	src := "@layers [core, ui]\n@layer core {\n fn update(msg, m) = m\n fn page(m) = 42\n}\napp({init: {}, update: update, view: page})"
	res := Run(src, Options{})
	if res.Success || !strings.Contains(res.Errors[0].Message, "an app's view must return a UI component") {
		t.Errorf("%+v", res.Errors)
	}
}

func TestAppViewListBecomesColumn(t *testing.T) {
	src := `@layers [core, ui, extra] blend: {page: collect}
@layer core {
  fn update(msg, m) = m
}
@layer ui {
  fn page(m) = text("main")
}
@layer extra {
  fn page(m) = text("extra")
}
app({init: {}, update: update, view: page})`
	res := Run(src, Options{})
	want := "<container layout=\"vertical\">\n  <text content=\"extra\" />\n  <text content=\"main\" />\n</container>"
	if !res.Success || res.Display != want {
		t.Errorf("display %q errors %+v", res.Display, res.Errors)
	}
}
