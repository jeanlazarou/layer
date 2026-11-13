# Layer Language Architecture

Layer is a tree-walking interpreter written in Go. The same packages are built natively (the `layer` CLI)
and to WebAssembly (the visual editor).

## Pipeline

```
source text
  │  pkg/lexer        tokens; double-quoted strings with {expr} become TEMPLATE tokens
  ▼
tokens
  │  pkg/parser       recursive descent + precedence climbing; desugars f(a, ?) into a lambda
  ▼
AST (pkg/ast)         Program = @layers directive + @layer declarations + one main expression
  │
  │  Interpreter.Load (pkg/interpreter)
  │    1. evaluate @layers conditions, enable listed layers (core is always first/bottom)
  │    2. apply purity annotations (@layer<impure>)
  │    3. evaluate each active layer's fn definitions into closures
  │    4. static checks (check.go): undefined names, unknown/inactive layers,
  │       missing members, purity violations — every function body, all at once
  ▼
  │  Interpreter.Execute: evaluate the main expression
  ▼
result value + log lines, or a located error
```

`pkg/runner` wraps the whole pipeline (`Run`, `Describe`) and converts any failure into located `Issue`s
(`lexer | parser | check | runtime | internal`). It recovers from interpreter panics, which matters under
WebAssembly where a panic would end the Go runtime for every later call. The CLI, the WASM bridge and the
tests all go through it.

## Packages

| Package | Role |
|---|---|
| `pkg/lexer` | Tokens with line/column. Soft keywords (`if`, `then`, `else`, `when`, `blend`) stay identifiers. |
| `pkg/parser` | AST construction. Layer bodies accept only `fn` definitions; a program has one main expression. |
| `pkg/ast` | Node types. |
| `pkg/layers` | `LayerContext`: layer order, definitions per layer, purity flags, blend mode, lookups. |
| `pkg/interpreter` | Evaluation, built-in layers, static checks, errors, display (`Show`), stack description. |
| `pkg/runner` | End-to-end runs with structured results. |
| `pkg/sourcefile` | Splits a `.layer` file into editor parts (per-layer code, purity, blend mode, main). |
| `cmd/layer` | CLI: `layer run [-D name=value] file`, `layer check file`. |
| `cmd/wasm` | WebAssembly exports returning JSON: `executeLayer`, `describeLayer`, `parseLayerFile`. |

## Name resolution

1. Local environment (parameters, `let`, pattern bindings, variables seeded with `SetVar`/`-D`).
2. The layer stack, top to bottom, combined according to the name's blend mode (its per-function mode
   from `blend: {name: mode}`, else the stack default):
   - `override`: the topmost definition.
   - `chain`: bottom definition first, each higher one receives the previous result as its first argument.
   - `collect`: every definition, results as a list (top first).
   - `fallback`: top-down, first result that is not `None`; failing layers are skipped with a `[WARN]` log.

`layer.name` (e.g. `core.fib`) reads one layer directly. Under `override`, unqualified calls are late-bound,
so recursion inside a lower layer re-enters the top of the stack (open recursion); qualified recursion stays
inside the layer. The combining modes apply once per outermost call: `Interpreter.combining` counts the
combined calls running per name, and while one runs, an unqualified call from a layer that defines the name
resolves to that layer's own definition.

With an `@layers` directive, declared but unlisted layers are inactive; their definitions are never loaded.

## Purity

`core` and custom layers are pure, `ui`/`debug`/`data` impure (a few `data` helpers such as `Some`/`None` are
pure exceptions), and `@layer<impure> name` opts a custom layer in. A pure layer may not reference a name
defined only in impure layers, nor `impure.name` explicitly. This is checked statically before execution and
again at runtime. An impure layer *wrapping* a pure function is allowed: pure code cannot request effects,
but the outer program may observe it.

## Values

Go values represent Layer values: `int64`, `float64`, `string`, `bool`, `[]interface{}` (lists and tuples),
`map[string]interface{}` (records; Options are `{type: "Some", value}` / `{type: "None"}`), `*Closure`, Go
built-ins `func(...interface{}) (interface{}, error)`, and `*UIComponent`. `Show` renders them in Layer
literal syntax for results, logs and interpolation; `typeName` names their type in error messages.

`Closure` holds parameters, body, captured environment, the defining layer (for purity), an optional name
(for traces), and optionally a `Native` Go implementation (used by `compose`, `const`, `trace`, `benchmark`).

## Errors

- Lexer/parser errors: `… error at line:col: message`.
- Static check errors: a `*CheckError` with one positioned issue per problem.
- Runtime errors: `*RuntimeError` with the position of the failing expression (errors from Go built-ins are
  positioned at their call site) and the trace of named Layer functions, innermost first, with recursion
  compressed: `(core.fact ← debug.fact) ×5`.
- Recursion is bounded (`MaxCallDepth`): 100,000 natively, 20,000 under WebAssembly.

## Output

Built-ins never print. `log`, `trace`, `inspect`, `print`, `assert_equal`, `benchmark` and fallback warnings
call `Interpreter.emit`, which stores the line (see `Logs()`, `get_logs()`) and forwards it to `LogSink`:
stderr in the CLI, collected into the result for the editor.

## WebAssembly

`./build-wasm.sh` builds `cmd/wasm` and copies the matching `wasm_exec.js` (from `$GOROOT/lib/wasm`) into
`visual-editor/frontend/public/`. The browser's native call stack (~1 MB) is much smaller than Go's: Go
functions are WebAssembly functions, so deep Layer recursion used to overflow it at ~2,000 calls and kill the
Go runtime. Go's wasm port unwinds the native stack whenever a goroutine parks, so `Closure.Call` evaluates on
a fresh goroutine every 200 calls (`limits_js.go`), which bounds native stack use regardless of depth.

The data layer's WebAssembly build stores files in `localStorage` (or a real directory through the File System
Access API) via `visual-editor/frontend/src/wasm/browserFileSystem.ts`.

## Visual editor

React + MUI + Monaco (`visual-editor/frontend`). Editor state (layers top-first, main expression, blend mode)
is assembled into a program by `src/program.ts`, which also maps interpreter positions back to a layer and
line. `describeLayer` runs on every edit (debounced) for live diagnostics and the "who answers each call?"
panel; `executeLayer` runs programs and UI button handlers (input values passed as JSON variables). See
[visual-editor/README.md](../visual-editor/README.md).

## Not implemented

Type inference (Hindley-Milner is specified in LANGUAGE_SPEC.md only), modules, a REPL, tuple patterns, and
running programs off the browser's main thread.
