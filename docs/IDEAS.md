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

This is the single most unique feature of the language — no other language does this.

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

## Implementation Priority

1. **Guard clauses** — unblocks writing real programs today
2. **Record pattern matching** — natural companion to guard clauses
3. **Layer blend modes** — the novel contribution that differentiates the language
4. **Layer introspection** — makes the layer concept tangible and inspectable
5. **Conditional layer activation** — makes layers dynamic
6. **Flatten operation** — optimization + understanding
7. **Effect types on layers** — the deep theoretical work
8. **Visual editor diff view** — UX complement to introspection
