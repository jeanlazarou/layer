# Layer

**An experimental functional language where a program is a stack of layers, like layers in an image editor.**

You write pure logic once, in `core`. Logging, caching, validation, UI and I/O live in layers stacked on top that
wrap or replace functions below them. Hiding, reordering or re-blending layers changes what the program does
without editing the code inside them, and the visual editor makes that stack the main thing you look at.

```layer
@layers [core, debug]

@layer core {
  # Pure business logic. Knows nothing about logging.
  fn fib(n) =
    match n with
    | 0 -> 0
    | 1 -> 1
    | n -> fib(n - 1) + fib(n - 2)
}

@layer debug {
  # Wraps core.fib without touching it.
  fn fib(n) =
    let _ = log("fib({n})") in
      core.fib(n)
}

fib(4)
```

```console
$ layer run fib.layer
[INFO] fib(4)
[INFO] fib(3)
[INFO] fib(2)
...
3
```

The recursive call inside `core.fib` is unqualified, so it goes back through the top of the stack: the debug
layer sees **every** level of the recursion (open recursion, like `this.method()`). Remove `debug` from
`@layers` and the same core code runs silently. `core.fib(n)` is the explicit "call the layer below" escape hatch.

![The visual editor: the layer stack on the left, the debug layer wrapping core.factorial, and the
"who answers each call?" view on the right](docs/images/editor.png)

## Try it

**Online playground:** https://jeanlazarou.github.io/layer/ (built from `main` by
[`.github/workflows/pages.yml`](.github/workflows/pages.yml)). Pick a program from the Examples menu, toggle
layers on and off, and watch the output change.

**Locally:**

```bash
# Editor (needs Go to build the interpreter to WebAssembly, and Node + pnpm)
cd layer-go && ./build-wasm.sh
cd ../visual-editor/frontend && pnpm install && pnpm dev     # http://localhost:5173

# Command line
cd layer-go
go run ./cmd/layer run ../examples/fibonacci.layer
go run ./cmd/layer check ../examples/calculator.layer         # static checks only
go run ./cmd/layer run -D env=dev app.layer                    # for @layers [core, debug if env == "dev"]
```

## What's different

Layer borrows from context-oriented programming, aspects, CLOS method combination and middleware stacks
([docs/POSITIONING.md](docs/POSITIONING.md) compares them in detail). Three ideas are its own:

1. **Blend modes.** How definitions from different layers combine is a property of the *stack*, not of the
   functions: `override` (default), `chain` (pipe results upward), `collect` (call all, return a list),
   `fallback` (first result that isn't `None`).

   ```layer
   @layers [core, data, database, cache] blend: fallback

   @layer database {
     fn fetch(key) = data.Some("computed:{key}")
   }

   @layer<impure> cache {
     fn fetch(key) = data.get(key)
   }

   let _ = data.set("greeting", "cached hello") in
   [fetch("greeting"), fetch("other")]
   # → [Some("cached hello"), Some("computed:other")]
   ```

2. **Purity is a property of where code lives.** `core` and custom layers are pure; `ui`, `debug` and `data`
   are impure, and custom layers opt in with `@layer<impure>`. A pure layer cannot reference an effect, and
   this is checked *before* the program runs, including branches that would never execute:

   ```layer
   # Rejected before running:
   @layers [core, debug]
   @layer core {
     fn area(r) = if r < 0 then log("negative radius") else 3.14159 * r * r
   }
   area(2)
   # check error at 3:30: purity violation: pure layer 'core' cannot call 'log',
   #   which is only defined in impure layer 'debug'
   ```

   Impure layers above may still *wrap* pure functions (that is how the tracing above works): pure code
   cannot request effects, but the outer program may observe it.

3. **The stack is the architecture you look at.** The editor shows the layers as the main panel, a
   "who answers each call?" view, and live diagnostics. Inside the language, `@which fib`, `@inspect debug`
   and `@diff core debug` query the stack.

## The language in one screen

```layer
@layers [core]

@layer core {
  fn sum(xs) =
    match xs with
    | []           -> 0
    | [h, ...rest] -> h + sum(rest)

  fn describe(p) =
    match p with
    | {name: n, age: a} when a >= 18 -> "{n} is an adult"
    | {name: n}                      -> "{n} is a minor"

  fn below(limit, x) = x < limit
}

let add10 = λ(x) -> x + 10 in
[sum(map(add10, [1, 2, 3])), describe({name: "Ada", age: 36}), filter(below(3, ?), [1, 5, 2])]
# → [36, "Ada is an adult", [1, 2]]
```

Functions, lambdas, `let … in`, `if … then … else`, pattern matching (lists, records, guards), string
interpolation in double-quoted strings, and partial application with `?`. Everything is an expression; a
function body is one expression (sequence steps with `let _ = step in next`). The complete reference is
[LANGUAGE_SPEC.md](LANGUAGE_SPEC.md), and [docs/TUTORIAL.md](docs/TUTORIAL.md) walks through it.

The `ui` layer returns declarative components (`view`, `text`, `button`, `input`, `row`, `column`) that the
editor renders as live, interactive widgets. A button's `onclick` is a Layer expression evaluated on top of
the current stack, with input values bound by name.

## Status

This is a research language, not a production tool. Implemented and covered by tests:

- Tree-walking interpreter in Go, compiled natively (CLI) and to WebAssembly (editor)
- Layers, open recursion, explicit layer access, conditional activation (`debug if env == "dev"`)
- Blend modes: override, chain, collect, fallback
- Purity: runtime enforcement plus a static check before execution
- Static checks for undefined names and unknown/inactive layers, reported with positions
- Runtime errors with positions and a Layer call trace; a recursion limit instead of a crash
- Built-in `core`, `ui`, `debug` and `data` layers (the data layer uses an Option type)

Not implemented: type inference (specified, not built), modules, a REPL, per-function blend modes.

## Repository

```
examples/           runnable .layer programs (all run in CI; "# Result:" comments are checked)
layer-go/           the interpreter
  cmd/layer/        command-line runner
  cmd/wasm/         WebAssembly bridge used by the editor
  pkg/lexer, parser, ast, interpreter, layers, runner, sourcefile
visual-editor/      React + Monaco editor; runs entirely in the browser
docs/               tutorial, architecture, positioning, ideas
LANGUAGE_SPEC.md    language reference
PHILOSOPHY.md       design rationale
```

## Development

```bash
cd layer-go
go test ./...                 # interpreter, CLI, every example, and every program in this README
./build-wasm.sh && node test_wasm_node.js   # WebAssembly build + smoke test

cd ../visual-editor/frontend
pnpm typecheck && pnpm lint && pnpm build
pnpm exec playwright install chromium        # once
pnpm test:e2e                                # browser tests of the editor (Playwright)
```

CI ([`.github/workflows/ci.yml`](.github/workflows/ci.yml)) runs all of the above on every push, and the
Pages workflow re-runs the browser tests against the deployed site after each deploy.

## License

MIT — see [LICENSE](LICENSE).
