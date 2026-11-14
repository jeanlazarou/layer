# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Layer is an experimental purely functional programming language built around **layer-based composition**:
a program is a stack of layers (core, ui, debug, data, or custom) where higher layers override or wrap
functions of lower ones, like layers in an image editor.

- **Go interpreter** (`layer-go/`): lexer → parser → static checks → tree-walking evaluator. Built natively
  (CLI) and to WebAssembly (editor).
- **Visual editor** (`visual-editor/frontend/`): React + MUI + Monaco, runs the WASM interpreter in the browser.
  The layer stack is the main UI; `ui`-layer results render as live widgets.

## Commands

```bash
cd layer-go
go test ./...                          # everything: interpreter, CLI, examples/, README programs
go test ./pkg/interpreter -run TestX   # one test
go vet ./... && GOOS=js GOARCH=wasm go vet ./cmd/wasm ./pkg/...
gofmt -l .                             # CI fails on unformatted files
go run ./cmd/layer run ../examples/fibonacci.layer     # run a program
go run ./cmd/layer check file.layer                    # load + static checks only
go run ./cmd/layer run -D env=dev file.layer           # seed variables for @layers conditions
go run ./cmd/layer run --send '"inc"' app.layer        # send messages to an app, print model + view
./build-wasm.sh                        # writes ../visual-editor/frontend/public/{layer.wasm,wasm_exec.js}
node test_wasm_node.js                 # smoke-test the WASM build (run after build-wasm.sh)

cd visual-editor/frontend              # package manager: pnpm
pnpm dev                               # http://localhost:5173
pnpm typecheck && pnpm lint && pnpm build   # typecheck/lint also cover e2e/
pnpm test:e2e                          # Playwright: builds, serves dist/ on :4173, runs e2e/*.spec.ts
E2E_BASE_URL=https://jeanlazarou.github.io/layer/ pnpm test:e2e   # against the deployed site
```

CI (`.github/workflows/ci.yml`) runs all of the above (the editor job rebuilds the WASM first, so browser
tests exercise the interpreter in the same commit); `pages.yml` deploys the editor to GitHub Pages and then
runs the browser tests against the live URL.

## Architecture

```
source → lexer (pkg/lexer) → parser (pkg/parser) → AST (pkg/ast)
       → Interpreter.Load: activate @layers, purity annotations, evaluate fn definitions,
                           static checks (pkg/interpreter/check.go)
       → Interpreter.Execute: evaluate the main expression
pkg/runner   wraps the whole pipeline: Run / Describe → structured Result with located Issues,
             never panics. Used by cmd/layer, cmd/wasm and tests.
pkg/sourcefile  splits a .layer file into editor parts (layers, purity, blend mode, main).
pkg/layers   LayerContext: layer order, per-layer definitions, purity flags, blend mode.
```

Key interpreter files (`layer-go/pkg/interpreter/`):
- `interpreter.go` — `Closure`, `Eval`, blend modes (`lookupWithBlend`), `Load`/`Execute`, runtime purity checks
- `check.go` — static checks: undefined names, unknown/inactive layers, purity violations (all reported at once)
- `values.go` — `RuntimeError` (position + Layer call trace), `CheckError`, `Show` (display format), `typeName`
- `describe.go` — stack/resolution description used by the editor's "who answers each call?" panel
- `limits_js.go` / `limits_other.go` — recursion limits; under WASM, deep calls hop goroutines (see below)
- `core_layer.go`, `debug_layer.go`, `ui_layer.go`, `data_layer_native.go` / `data_layer_wasm.go` — built-ins

## Language semantics worth knowing

- **Resolution**: unqualified names resolve top-down through active layers (open recursion — recursive calls
  re-enter the top of the stack). `core.f(x)` targets one layer. `core` is always the bottom layer.
- **Activation**: with an `@layers [...]` directive, only listed layers are active; a declared but unlisted
  `@layer x { }` block is inactive (the editor's "hidden" layers). Without a directive, all declared layers are
  active. Conditions (`debug if env == "dev"`) are evaluated at load time from `SetVar` / CLI `-D` values.
- **Layer bodies contain only `fn` definitions.** A function body is a single expression; sequence steps with
  `let _ = step in next`. There are no `{ stmt; stmt }` blocks (`{` starts a record).
- **Strings**: double-quoted strings interpolate `{expr}` (unless `{` is followed by whitespace, a quote, `\` or
  `}`; `\{` escapes). Single-quoted strings never interpolate.
- **`if c then a else b`** — `if`/`then`/`else`/`when`/`blend` are soft keywords (lexed as identifiers).
- **Records**: `r.field` is field access when `r` is a local value, layer access otherwise (`core.f`); locals
  shadow layer names. `{r | f: v}` (Elm) copies with replaced fields, which must already exist.
- **Apps**: `app({init, update, view})` (ui built-in) returns `*interpreter.App`. Models and messages are plain
  data (`CheckData`/`EncodeData`/`DecodeData` in `data.go`; floats keep a `.0` so ints/floats round-trip).
  `runner.Run` drives apps statelessly: `Options.Model` (JSON) + `Messages` (JSON) / `Send` (Layer expressions)
  → `Result.Model`/`ModelText`/`Display` (view XML; a list of components becomes a column). Handlers render as
  JSON attributes (`on_click`, `on_input`). The ui constructors are pure exceptions, so pure layers can build
  UI. `onclick`/`target` props are errors that explain the replacement.
- **Purity**: `core` and custom layers are pure; `ui`/`debug`/`data` impure; `@layer<impure> name` opts in.
  Enforced statically (check.go) and at runtime.
- **Blend modes**: `@layers [...] blend: chain` (all names) or `blend: {fetch: fallback, _: chain}` (per
  function; `_` = default). `LayerContext.BlendModeFor(name)` resolves the mode. Combining modes
  (chain/collect/fallback) apply once per outermost call: `Interpreter.combining` tracks names being combined,
  and inside one an unqualified call from a layer defining the name resolves to that layer (recursion stays
  local). `override` keeps open recursion. `fallback` skips a failing layer but logs a `[WARN]` line.
- **Output**: `log`, `trace`, `print`, ... go through `Interpreter.emit` → `LogSink` (stderr in the CLI;
  collected into `Result.Logs` for the editor). Never write to `os.Stdout`/`os.Stderr` from built-ins.
- **Errors**: build them with `i.error(node, msg)` (positioned `*RuntimeError`); errors from Go built-ins are
  positioned at the call site by `i.wrap`. Messages should use `typeName(v)` (Layer type names), not `%T`.
- **WASM stack**: the browser's native stack is ~1 MB. `Closure.Call` evaluates on a fresh goroutine every
  `stackHopEvery` calls under WASM (parking unwinds the native stack); without this, ~2000-deep recursion
  killed the Go runtime permanently. Keep this in mind before restructuring `Closure.Call`.

## Common workflows

**Add a built-in function**: define it in the right `Setup*Layer` with `lc.Define(...)` (use `callValue` /
`isCallable` to accept any function value, `Show` to format values); add tests in
`pkg/interpreter/*_test.go`; document it in LANGUAGE_SPEC.md; if it lives in an optional layer, add it to
`builtinProviders` in `check.go` (better "undefined name" message) and to `BUILTINS` in
`visual-editor/frontend/src/language/layerLanguage.ts` (highlighting/completion); rebuild WASM.

**Change syntax**: update `pkg/lexer` / `pkg/parser` / `pkg/ast`, then evaluation in `interpreter.go`, the
walker in `check.go` (or names will be reported as undefined or skipped), and the Monaco tokenizer.

**Add a built-in layer**: `Setup<Name>Layer(lc, interp)`, register it in `Interpreter.EnableLayer` (with its
purity), add a color in `visual-editor/frontend/src/program.ts` (`LAYER_COLORS`) and the name to
`BUILTIN_LAYERS` in `src/types/index.ts`.

**Examples and docs are tested**: every `examples/*.layer` must run (a `# Result: ...` comment is compared
with the output; an app's `# Send: <msg>` lines are sent in order and the final model is compared with
`# Model: ...`), and every ```` ```layer ```` block in README.md that starts with `@layers` must run
(`layer-go/cmd/layer/main_test.go`). Start an intentionally failing README snippet with a comment line.

## Visual editor notes

- `src/program.ts` assembles editor state into a program (`buildProgram`) and maps interpreter positions back
  to a layer + line (`locateIssue`); the editor's layer order is top-first, the program's is bottom-first.
- Live diagnostics call `describeLayer` (debounced) on every edit; runs call `executeLayer`.
- Apps: `App.tsx` keeps an `AppSession` (every step: message, model, view) for the inspector and time travel;
  each message is `executeLayer(source, {model, messages: [msg]})` — never splice user text into source.
  App operations go through a queue so rapid events apply in order. Run/structural changes re-render with the
  current model (hot reload); Restart starts from init. Input keystrokes are merged into one history entry
  (`AppStep.group`).
- Layer colors: core green, ui blue, debug orange, data purple.
- Browser tests live in `e2e/editor.spec.ts`. They drive Monaco through `window.monaco` (`setCode` helper),
  navigate to `'./'` (never `'/'`, which breaks under the Pages sub-path), and find controls by role/label —
  keep MUI `Tooltip` titles and button texts stable or update the tests with them.

## Status

Implemented: interpreter, CLI, WASM editor, core/ui/debug/data layers, open recursion, blend modes, purity
(static + runtime), conditional layers, introspection (`@which`, `@inspect`, `@diff`), partial application.
Per-function blend modes, records (field access/update), interactive apps (model/update/view). Not implemented:
type inference (Hindley-Milner is specified only), modules, REPL.

## Docs

`LANGUAGE_SPEC.md` (reference), `docs/TUTORIAL.md`, `docs/ARCHITECTURE.md`, `docs/POSITIONING.md` (prior art),
`docs/IDEAS.md` (future work), `PHILOSOPHY.md`, `visual-editor/README.md`.
