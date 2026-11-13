package interpreter

import (
	"errors"
	"strings"
	"testing"

	"github.com/jeanlazarou/layer/pkg/lexer"
	"github.com/jeanlazarou/layer/pkg/parser"
)

// run executes source with logs collected (not printed) and returns the
// interpreter so tests can inspect logs.
func run(t *testing.T, source string) (interface{}, *Interpreter, error) {
	t.Helper()
	tokens, err := lexer.Tokenize(source)
	if err != nil {
		return nil, nil, err
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		return nil, nil, err
	}
	interp := NewInterpreter()
	interp.LogSink = nil
	result, err := interp.Execute(program)
	return result, interp, err
}

func mustFail(t *testing.T, source, wantInError string) error {
	t.Helper()
	_, _, err := run(t, source)
	if err == nil {
		t.Fatalf("expected an error containing %q, got none", wantInError)
	}
	if !strings.Contains(err.Error(), wantInError) {
		t.Fatalf("expected an error containing %q, got: %v", wantInError, err)
	}
	return err
}

// ── String interpolation ──────────────────────────────────────────────────────

func TestInterpolation(t *testing.T) {
	cases := []struct {
		name, source, want string
	}{
		{"variable", `let n = 5 in "n is {n}"`, "n is 5"},
		{"expression", `"sum: {1 + 2}"`, "sum: 3"},
		{"call", "@layers [core]\n@layer core {\n fn sq(x) = x * x\n}\n\"sq(4) = {sq(4)}\"", "sq(4) = 16"},
		{"record inside", `"r: {{a: 1}}"`, "r: {a: 1}"},
		{"list shows nested strings quoted", `"l: {['a', 'b']}"`, `l: ["a", "b"]`},
		{"space after brace is literal", `"{ }"`, "{ }"},
		{"json-like stays literal", `"{\"a\": 1}"`, `{"a": 1}`},
		{"escaped brace", `"\{n}"`, "{n}"},
		{"single quotes never interpolate", `let n = 1 in '{n}'`, "{n}"},
		{"single-quoted string inside", `"x{str_concat('y', '!')}z"`, "xy!z"},
		{"quote right after brace is literal", `"{'y'}"`, "{'y'}"},
		{"several parts", `let a = 1 in let b = 2 in "{a}+{b}={a + b}"`, "1+2=3"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, _, err := run(t, c.source)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestInterpolationErrors(t *testing.T) {
	mustFail(t, `"{1 +}"`, "parser error")
	mustFail(t, `"{abc"`, "missing '}'")
	mustFail(t, `"{f("x")}"`, "write strings with single quotes")
	mustFail(t, "match \"a\" with\n| \"{x}\" -> 1", "cannot contain")

	// Errors inside an interpolation point at their real position.
	err := mustFail(t, "let s =\n  \"value: {missing}\" in s", "undefined name 'missing'")
	if !strings.Contains(err.Error(), "at 2:12") {
		t.Errorf("expected position 2:12, got %v", err)
	}
}

// ── if / then / else ──────────────────────────────────────────────────────────

func TestIfExpression(t *testing.T) {
	source := `
	@layers [core]
	@layer core {
		fn sign(n) = if n < 0 then "negative" else if n == 0 then "zero" else "positive"
	}
	[sign(-3), sign(0), sign(7)]
	`
	got := mustExecute(t, source)
	if ShowNested(got) != `["negative", "zero", "positive"]` {
		t.Errorf("got %s", ShowNested(got))
	}
}

func TestIfOnlyEvaluatesTakenBranch(t *testing.T) {
	got, _, err := run(t, `if true then 1 else 1 / "boom"`)
	if err != nil || got != int64(1) {
		t.Errorf("got %v, %v", got, err)
	}
}

func TestIfConditionMustBeBoolean(t *testing.T) {
	mustFail(t, `if 1 then 2 else 3`, "if condition must be a boolean, got integer")
}

func TestTrailingTokensAfterMainExpressionAreAnError(t *testing.T) {
	mustFail(t, "1 + 1\n2", "unexpected NUMBER after the main expression")
}

func TestIfRequiresElse(t *testing.T) {
	mustFail(t, `if true then 1`, "expected 'else'")
}

// ── Static checks ─────────────────────────────────────────────────────────────

func TestCheckFindsUndefinedNamesInUnexecutedCode(t *testing.T) {
	// g is never called, but its typo is still reported before running.
	err := mustFail(t, `
	@layers [core]
	@layer core {
		fn f(x) = x
		fn g(x) = fff(x)
	}
	f(1)
	`, "undefined name 'fff'")
	var cerr *CheckError
	if !errors.As(err, &cerr) {
		t.Fatalf("expected a *CheckError, got %T", err)
	}
}

func TestCheckReportsEveryIssue(t *testing.T) {
	_, _, err := run(t, "@layers [core]\n@layer core {\n fn a() = x1\n fn b() = x2\n}\nx3")
	var cerr *CheckError
	if !errors.As(err, &cerr) {
		t.Fatalf("expected a *CheckError, got %v", err)
	}
	if len(cerr.Issues) != 3 {
		t.Fatalf("expected 3 issues, got %v", cerr.Issues)
	}
	for idx, line := range []int{3, 4, 6} {
		if cerr.Issues[idx].Line != line {
			t.Errorf("issue %d: line %d, want %d", idx, cerr.Issues[idx].Line, line)
		}
	}
}

func TestCheckScopesHaveNoFalsePositives(t *testing.T) {
	// Parameters, let bindings, lambda params, list/record pattern
	// bindings, guards, partial-application holes, and recursion.
	source := `
	@layers [core]
	@layer core {
		fn sum(l) =
			match l with
			| []           -> 0
			| [h, ...rest] when h > 0 -> h + sum(rest)
			| [_, ...rest] -> sum(rest)
		fn name_of(p) =
			match p with
			| {name: n} -> n
		fn add(a, b) = a + b
		fn go(xs) =
			let add1 = add(1, ?) in
			let f = λ(x) -> add1(x) * 2 in
			map(f, xs)
	}
	[sum([1, -2, 3]), name_of({name: "x"}), go([1, 2])]
	`
	got := mustExecute(t, source)
	if ShowNested(got) != `[4, "x", [4, 6]]` {
		t.Errorf("got %s", ShowNested(got))
	}
}

func TestCheckLetIsNotRecursive(t *testing.T) {
	mustFail(t, `let x = x + 1 in x`, "undefined name 'x'")
}

func TestCheckQualifiedReferences(t *testing.T) {
	mustFail(t, "@layers [core]\n@layer core {\n fn f(x) = x\n}\ncore.g(1)", "'core.g' is not defined")
	mustFail(t, "@layers [core]\nnope.f(1)", "layer 'nope' is not active")
	mustFail(t, "@layers [core]\n@layer core {\n fn f(x) = data.Some(x)\n}\nf(1)", "layer 'data' is not active")
}

func TestCheckExplainsMissingBuiltinLayer(t *testing.T) {
	mustFail(t, "@layers [core]\nlog(1)", "provided by the 'debug' layer, which is not in @layers")
}

func TestCheckPurityInUnexecutedBranch(t *testing.T) {
	// The impure call sits in a branch that never runs; the static check
	// still rejects it.
	mustFail(t, `
	@layers [core, debug]
	@layer core {
		fn f(x) = if x > 100 then log(x) else x
	}
	f(1)
	`, "purity violation: pure layer 'core' cannot call 'log'")
}

func TestCheckSeesSetVar(t *testing.T) {
	tokens, _ := lexer.Tokenize("@layers [core]\nnum * 2")
	program, _ := parser.Parse(tokens)
	interp := NewInterpreter()
	interp.SetVar("num", int64(21))
	got, err := interp.Execute(program)
	if err != nil || got != int64(42) {
		t.Errorf("got %v, %v", got, err)
	}
}

// ── Layer activation ──────────────────────────────────────────────────────────

func TestDeclaredButUnlistedLayerIsInactive(t *testing.T) {
	// LANGUAGE_SPEC: "You can define layers in @layer blocks but they're
	// only active if listed in @layers [...]".
	source := `
	@layers [core]
	@layer core {
		fn greet() = "core"
	}
	@layer loud {
		fn greet() = "LOUD"
	}
	greet()
	`
	if got := mustExecute(t, source); got != "core" {
		t.Errorf("hidden layer should not override: got %v", got)
	}
	mustFail(t, strings.Replace(source, "greet()\n\t", "loud.greet()\n\t", 1), "layer 'loud' is not active")
}

func TestWithoutDirectiveAllDeclaredLayersAreActive(t *testing.T) {
	source := `
	@layer core {
		fn greet() = "core"
	}
	@layer loud {
		fn greet() = "LOUD"
	}
	greet()
	`
	if got := mustExecute(t, source); got != "LOUD" {
		t.Errorf("got %v", got)
	}
}

// ── Runtime errors ────────────────────────────────────────────────────────────

func TestOperatorErrorsHavePositionsAndTrace(t *testing.T) {
	_, _, err := run(t, "@layers [core]\n@layer core {\n fn inner(x) = x + 1\n fn outer(x) = inner(x)\n}\nouter(\"a\")")
	var rerr *RuntimeError
	if !errors.As(err, &rerr) {
		t.Fatalf("expected *RuntimeError, got %v", err)
	}
	if rerr.Line != 3 || rerr.Column != 18 {
		t.Errorf("position %d:%d, want 3:18", rerr.Line, rerr.Column)
	}
	if !strings.Contains(rerr.Message, "operator + needs two numbers or two strings, got string and integer") {
		t.Errorf("message: %s", rerr.Message)
	}
	if rerr.TraceString() != "core.inner ← core.outer" {
		t.Errorf("trace: %s", rerr.TraceString())
	}
}

func TestRecursionTraceIsCollapsed(t *testing.T) {
	_, _, err := run(t, "@layers [core]\n@layer core {\n fn down(n) = if n == 0 then 1 / \"x\" else down(n - 1)\n}\ndown(5)")
	var rerr *RuntimeError
	if !errors.As(err, &rerr) {
		t.Fatalf("expected *RuntimeError, got %v", err)
	}
	if rerr.TraceString() != "core.down (×6)" {
		t.Errorf("trace: %s", rerr.TraceString())
	}
}

func TestOpenRecursionTraceIsCompressed(t *testing.T) {
	_, _, err := run(t, `
	@layers [core, debug]
	@layer core {
		fn fact(n) = if n == 0 then 1 / "x" else n * fact(n - 1)
	}
	@layer debug {
		fn fact(n) = let _ = log(n) in core.fact(n)
	}
	let r = fact(3) in r
	`)
	var rerr *RuntimeError
	if !errors.As(err, &rerr) {
		t.Fatalf("expected *RuntimeError, got %v", err)
	}
	if got := rerr.TraceString(); got != "(core.fact ← debug.fact) ×4" {
		t.Errorf("trace: %s", got)
	}
}

func TestArityErrorNamesTheFunction(t *testing.T) {
	mustFail(t, "@layers [core]\n@layer core {\n fn f(a, b) = a\n}\nf(1)", "core.f expects 2 arguments, got 1")
}

func TestRunawayRecursionIsAnError(t *testing.T) {
	tokens, _ := lexer.Tokenize("@layers [core]\n@layer core {\n fn loop(n) = loop(n + 1)\n}\nloop(0)")
	program, _ := parser.Parse(tokens)
	interp := NewInterpreter()
	interp.MaxCallDepth = 500
	_, err := interp.Execute(program)
	if err == nil || !strings.Contains(err.Error(), "maximum call depth (500) exceeded") {
		t.Errorf("got %v", err)
	}
}

// ── Output ────────────────────────────────────────────────────────────────────

func TestLogsAreCollected(t *testing.T) {
	_, interp, err := run(t, `
	@layers [core, debug]
	@layer debug {
		fn double(x) = let _ = log("double({x})") in x * 2
	}
	let _ = print("start") in
	trace("double", double)(21)
	`)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"start", "[TRACE] double(21) called", "[INFO] double(21)", "[TRACE] double(21) = 42"}
	if strings.Join(interp.Logs(), "|") != strings.Join(want, "|") {
		t.Errorf("logs:\n got %q\nwant %q", interp.Logs(), want)
	}
}

func TestLogsArePerInterpreter(t *testing.T) {
	source := "@layers [core, debug]\nlet _ = log(1) in length(get_logs())"
	for i := 0; i < 2; i++ {
		got, _, err := run(t, source)
		if err != nil || got != int64(1) {
			t.Fatalf("run %d: got %v, %v (logs leaked between runs?)", i, got, err)
		}
	}
}

// ── Built-ins accept any function ─────────────────────────────────────────────

func TestHigherOrderBuiltinsAcceptBuiltins(t *testing.T) {
	got := mustExecute(t, `map(to_string, [1, 2])`)
	if ShowNested(got) != `["1", "2"]` {
		t.Errorf("map: got %s", ShowNested(got))
	}
	got = mustExecute(t, `compose(str_length, to_string)(12345)`)
	if got != int64(5) {
		t.Errorf("compose: got %v", got)
	}
	got = mustExecute(t, `const(7)("ignored")`)
	if got != int64(7) {
		t.Errorf("const: got %v", got)
	}
}

// ── Display ───────────────────────────────────────────────────────────────────

func TestShow(t *testing.T) {
	cases := []struct {
		val  interface{}
		want string
	}{
		{"top-level string", "top-level string"},
		{[]interface{}{int64(1), "a", 2.5, true}, `[1, "a", 2.5, true]`},
		{map[string]interface{}{"b": int64(2), "a": "x"}, `{a: "x", b: 2}`},
		{map[string]interface{}{"type": "Some", "value": "v"}, `Some("v")`},
		{map[string]interface{}{"type": "None"}, "None"},
		{nil, "nil"},
	}
	for _, c := range cases {
		if got := Show(c.val); got != c.want {
			t.Errorf("Show(%#v) = %q, want %q", c.val, got, c.want)
		}
	}
}

func TestUIRenderEscapesAndSortsAttributes(t *testing.T) {
	got := mustExecute(t, `
	@layers [core, ui]
	view({title: "a < b & \"c\"", children: [button({onclick: "x < 1", label: "go"})]})
	`)
	xml := Show(got)
	want := "<view title=\"a &lt; b &amp; &#34;c&#34;\">\n  <button label=\"go\" onclick=\"x &lt; 1\" />\n</view>"
	if xml != want {
		t.Errorf("got:\n%s\nwant:\n%s", xml, want)
	}
}

// ── Blend modes × recursion ───────────────────────────────────────────────────

// Combining modes apply once per outermost call: recursion inside a
// participating layer stays in that layer. (Before this rule, the chain was
// re-applied at every level and f(2) below was 110.)
func TestChainAppliesOncePerOuterCall(t *testing.T) {
	got := mustExecute(t, `
	@layers [core, scale] blend: chain
	@layer core {
		fn f(n) = if n == 0 then 0 else 1 + f(n - 1)
	}
	@layer scale {
		fn f(x) = x * 10
	}
	f(2)
	`)
	if got != int64(20) {
		t.Errorf("got %v, want 20", got)
	}
}

// A call from a function that is not itself being combined (here core.twice)
// starts a fresh chain each time.
func TestChainCalledFromAnotherFunctionAppliesEachTime(t *testing.T) {
	got := mustExecute(t, `
	@layers [core, scale] blend: chain
	@layer core {
		fn f(n) = if n == 0 then 0 else 1 + f(n - 1)
		fn twice(n) = [f(n), f(n + 1)]
	}
	@layer scale {
		fn f(x) = x * 10
	}
	twice(2)
	`)
	if ShowNested(got) != "[20, 30]" {
		t.Errorf("got %s, want [20, 30]", ShowNested(got))
	}
}

func TestCollectWithRecursionStaysInEachLayer(t *testing.T) {
	got := mustExecute(t, `
	@layers [core, audit] blend: collect
	@layer core {
		fn count(n) = if n == 0 then 0 else 1 + count(n - 1)
	}
	@layer audit {
		fn count(n) = "audit {n}"
	}
	count(3)
	`)
	if ShowNested(got) != `["audit 3", 3]` {
		t.Errorf("got %s", ShowNested(got))
	}
}

// Override keeps open recursion: that is what lets a wrapping layer see
// every recursive step.
func TestOverrideKeepsOpenRecursion(t *testing.T) {
	_, interp, err := run(t, `
	@layers [core, debug]
	@layer core {
		fn f(n) = if n == 0 then 0 else 1 + f(n - 1)
	}
	@layer debug {
		fn f(n) = let _ = log(n) in core.f(n)
	}
	f(2)
	`)
	if err != nil {
		t.Fatal(err)
	}
	if len(interp.Logs()) != 3 {
		t.Errorf("expected a log line per recursion level, got %q", interp.Logs())
	}
}

// ── Per-function blend modes ──────────────────────────────────────────────────

func TestPerFunctionBlendModes(t *testing.T) {
	got := mustExecute(t, `
	@layers [core, data, cache, audit] blend: {fetch: fallback, notify: collect}
	@layer core {
		fn fetch(k) = data.Some("db:{k}")
		fn notify(msg) = "core {msg}"
		fn name() = "core"
	}
	@layer cache {
		fn fetch(k) = data.None()
	}
	@layer audit {
		fn notify(msg) = "audit {msg}"
		fn name() = "audit"
	}
	[data.unwrap(fetch("a"), "?"), notify("x"), name()]
	`)
	// fetch: cache misses, core answers. notify: both run, top first.
	// name: not listed, so the default (override) applies.
	if ShowNested(got) != `["db:a", ["audit x", "core x"], "audit"]` {
		t.Errorf("got %s", ShowNested(got))
	}
}

func TestBlendMapDefaultWithException(t *testing.T) {
	got := mustExecute(t, `
	@layers [core, loud] blend: {_: collect, greet: override}
	@layer core {
		fn greet() = "hi"
		fn tag() = "core"
	}
	@layer loud {
		fn greet() = "HI"
		fn tag() = "loud"
	}
	[greet(), tag()]
	`)
	if ShowNested(got) != `["HI", ["loud", "core"]]` {
		t.Errorf("got %s", ShowNested(got))
	}
}

func TestBlendEntryForUnknownFunctionIsReported(t *testing.T) {
	err := mustFail(t, "@layers [core] blend: {fecth: fallback}\n@layer core {\n fn fetch(k) = k\n}\nfetch(1)",
		"blend mode given for 'fecth', but no layer defines a function with that name")
	var cerr *CheckError
	if !errors.As(err, &cerr) || cerr.Issues[0].Line != 1 || cerr.Issues[0].Column != 24 {
		t.Errorf("expected a check error at 1:24, got %v", err)
	}
}

func TestBlendEntryForFunctionInHiddenLayerIsAllowed(t *testing.T) {
	got := mustExecute(t, `
	@layers [core] blend: {extra: collect}
	@layer core {
		fn main() = 1
	}
	@layer hidden {
		fn extra() = 2
	}
	main()
	`)
	if got != int64(1) {
		t.Errorf("got %v", got)
	}
}
