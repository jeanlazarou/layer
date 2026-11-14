# Layer Language — Ideas & Future Directions

## 1. Layer Blend Modes

The current layer model only supports "override" (top layer wins / first match). The Photoshop metaphor suggests many more blend strategies:

| Blend Mode | Meaning | Example Use |
|---|---|---|
| `override` (current) | Top layer replaces bottom | debug wraps core.fibonacci |
| `chain` | All layers pipe into each other | Each layer transforms the value |
| `collect` | All layers contribute, results merged | Multiple `log` handlers all fire |
| `fallback` | Try layers top→bottom, use first success | Try cache layer, fallback to data layer |
| `vote` | All layers compute, majority wins | Validation across layers |

Syntax ideas:
```layer
# Global blend mode for the stack
@layers [core, cache, data] blend: fallback

# Per-function blend mode
@layer cache blend: fallback {
  fn fetch(key) = ...
}
```

This is the most distinctive feature of the language. CLOS method combination is the closest precedent, but
there the combination belongs to the generic function; here it belongs to the stack (see POSITIONING.md).

---

## 2. Layer Visibility and Conditional Activation

Photoshop lets you toggle layer visibility. Equivalent in Layer:

```layer
# Conditional layer — only active in certain environments
@layers [core, debug if env == "development", data]

# Runtime layer composition
let result = with_layers([core, debug]) do
  fibonacci(10)
```

Makes layers dynamic rather than static declarations — compose different layer stacks for testing vs. production.

---

## 3. Layer Introspection

Language-level visibility into what each layer contributes:

```layer
@which factorial        # → "debug" (wrapping "core")
@inspect debug          # → {factorial, fibonacci, log}
@diff core debug        # → {added: [], overridden: [factorial, fibonacci]}
```

Useful for education, debugging, and understanding composed programs. The visual editor could display this inline.

---

## 4. Flatten Operation

Photoshop's "flatten layers" equivalent — merge a layer stack into a single resolved layer:

```layer
@flatten [core, debug] as traced_core
```

Creates a new merged layer from the composition. Useful for:
- Understanding what the program "resolves to"
- Optimization (eliminate layer lookup overhead at runtime)
- Shipping a "finalized" production version

---

## 5. Guard Clauses in Pattern Matching

Every real program hits this limitation quickly. Guard clauses make patterns conditional:

```layer
fn classify(n) =
  match n with
  | n when n < 0   -> "negative"
  | 0              -> "zero"
  | n when n > 100 -> "large"
  | n              -> "positive"
```

Low-hanging fruit that significantly expands expressiveness.

---

## 6. Record Pattern Matching

Currently lists can be destructured, but not records:

```layer
fn greet(person) =
  match person with
  | {name: n, age: a} when a >= 18 -> str_concat("Hello, ", n)
  | {name: n}                       -> str_concat("Hi kid, ", n)
```

Natural complement to guard clauses.

---

## 7. Layer Effect Types (Purity System)

Mark layers with the effects they introduce, and enforce it at compile time:

```layer
@layer core : pure {
  fn factorial(n) = ...   # must be pure — compiler enforces
}

@layer data : io {
  fn read(path) = ...     # allowed to do I/O
}

@layer debug : io, state {
  fn log(msg) = ...
}
```

Rules:
- A `pure` layer can only call other `pure` layers
- The layer stack tells you exactly what effects the program has
- The visual editor can color-code by effect type (colors already in use!)

This is the biggest theoretical contribution — algebraic effects expressed spatially.

---

## 8. Visual Editor — Layer Diff View

When a layer is selected in the editor panel, highlight what it changes relative to the layer below:

- Green — functions this layer adds
- Yellow — functions this layer overrides
- Blue — functions this layer wraps (calls through to lower layer)

Turns the visual editor into a powerful teaching and debugging tool.

---

## Implementation Status

1. ~~**Guard clauses**~~ ✅ implemented
2. ~~**Record pattern matching**~~ ✅ implemented
3. ~~**Layer blend modes**~~ ✅ implemented, per stack and per function
4. ~~**Layer introspection**~~ ✅ implemented (`@which`, `@inspect`, `@diff`)
5. ~~**Conditional layer activation**~~ ✅ implemented (`debug if env == "dev"`; CLI `-D env=dev`). Scoped
   activation (`with_layers(...)`) is not.
6. **Flatten operation** — not started
7. ~~**Effect types on layers**~~ ✅ pure/impure split implemented, checked statically and at runtime
   (`@layer<impure>`). Finer-grained effects (io vs state) are not.
8. ~~**Visual editor diff view**~~ ✅ implemented — diff pills on the selected layer card, plus the
   "who answers each call?" panel
9. ~~**Partial application (`?` holes)**~~ ✅ implemented

## Open design questions

These are language-design decisions, deliberately left open rather than picked by default.

**A. Per-function blend modes** — ✅ decided and implemented: a map in the directive,
`blend: {fetch: fallback, total: chain}`, with `_` as the default for unlisted names. It keeps the "the
composition decides, not the function" property. The editor sets them per function in the resolution panel.

**B. Blend modes × recursion** — ✅ decided and implemented: combining modes (chain, collect, fallback) apply
once per outermost call; recursion inside a participating layer stays in that layer. `override` keeps open
recursion. `f(2)` in the chain example is now 20 (it was 110).

**C. A UI model** — ✅ decided and implemented: apps with model / update / view (`app({init, update,
view})`), plain-data messages (`on_click`, `on_input`), Elm record update syntax (`{m | count: 1}`), and pure
ui constructors. The old `onclick` strings and `target` were removed. The editor shows the model and message
history (time travel) and keeps the model across code changes. See LANGUAGE_SPEC "Interactive apps".

**D. Off-main-thread execution.** The editor runs programs on the browser's main thread; a long computation
freezes the page. Running the WASM module in a Web Worker would allow a "stop" button.
