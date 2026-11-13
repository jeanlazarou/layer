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
3. ~~**Layer blend modes**~~ ✅ implemented (global per stack — see open question A)
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

**A. Per-function blend modes.** Today `blend:` applies to every name defined in more than one layer, so a
program cannot use `fallback` for `fetch` and `collect` for `notify` at the same time. Candidate syntaxes:

```layer
@layers [core, cache, audit] blend: {fetch: fallback, notify: collect}   # per name, in the directive
@layer cache blend: fallback { ... }                                      # per layer (ambiguous: which pairs?)
@blend fallback fn fetch(key) = ...                                       # on the definition
```

The directive form keeps the "composition decides, not the function" property that makes blend modes novel.

**B. Chain × open recursion.** With `blend: chain`, an unqualified recursive call inside the bottom layer
re-enters the whole chain, so `f(2)` with `core: f(n) = 1 + f(n-1)` and `scale: f(x) = x * 10` is 110, not 20
(locked by `TestChainReappliesAtEveryRecursiveCall`). That is consistent with override's open recursion but
rarely what a pipeline means. Options: keep it (and document qualified recursion as the fix), or make the
chain apply only at the outermost call.

**C. A UI model.** The UI layer is a view function plus `onclick` strings evaluated against input values.
An Elm-style `init / update(msg, model) / view(model)` would fit a pure language better: state lives in Layer,
handlers are values instead of strings, and the pure/impure split maps naturally (view and update pure, the
runtime impure).

**D. Off-main-thread execution.** The editor runs programs on the browser's main thread; a long computation
freezes the page. Running the WASM module in a Web Worker would allow a "stop" button.
