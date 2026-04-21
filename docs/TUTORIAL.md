# Layer Language Tutorial

Welcome to Layer! This tutorial will guide you through the basics of programming in Layer.

## Installation

```bash
git clone https://github.com/yourusername/layer.git
cd layer/layer-go
go build -o layer ./cmd/layer
```

Or use the visual editor (no installation needed - runs in browser):
```bash
cd layer/visual-editor/frontend
npm install
npm run dev
```

## Your First Layer Program

Create a file `hello.layer`:

```layer
@layers [core]

@layer core {
    fn greet(name) =
        str_concat("Hello, ", str_concat(name, "!"))
}

greet("World")
```

Run it:

```bash
cd layer-go
./layer run hello.layer
```

Output: `Hello, World!`

**Or use the visual editor**: Open http://localhost:5173, paste the code, and click ▶️

## Basic Concepts

### 1. Functions

Functions are the heart of Layer. All functions are pure (no side effects).

```layer
# Function with one parameter
fn double(x) = x * 2

# Function with multiple parameters
fn add(x, y) = x + y

# Nested function calls
fn quadruple(x) = double(double(x))
```

### 2. Let Bindings

Use `let` to create local bindings:

```layer
let x = 5 in
let y = 10 in
    x + y
```

### 3. Pattern Matching

Pattern matching is powerful for control flow:

```layer
fn factorial(n) =
    match n with
    | 0 -> 1
    | n -> n * factorial(n - 1)
```

#### Pattern Types

**Literal patterns:**
```layer
match value with
| 0 -> "zero"
| 1 -> "one"
| n -> "other"
```

**List patterns:**
```layer
fn sum(lst) =
    match lst with
    | [] -> 0
    | [head, ...tail] -> head + sum(tail)
```

**Wildcard pattern:**
```layer
match x with
| 0 -> "zero"
| _ -> "not zero"
```

**Record patterns:**

Destructure records by field name. Matching is partial — extra fields in the value are ignored:

```layer
fn greet(person) =
    match person with
    | {name: n, age: a} when a >= 18 -> str_concat("Hello, ", n)
    | {name: n}                       -> str_concat("Hi kid, ", n)

greet({name: "Alice", age: 30})  # → "Hello, Alice"
greet({name: "Tom",   age: 10})  # → "Hi kid, Tom"
```

If a field named in the pattern is absent from the value, the case is skipped.

**Guard clauses (`when`):**

Add a boolean condition after a pattern with `when`. The case only fires if the pattern matches *and* the guard is true:

```layer
fn classify(n) =
    match n with
    | n when n < 0   -> "negative"
    | 0              -> "zero"
    | n when n > 100 -> "large"
    | n              -> "positive"
```

Guards can use variables bound by the pattern — here `n` in the guard refers to the matched value:

```layer
fn greet_age(age) =
    match age with
    | a when a < 13  -> "Hi, kid!"
    | a when a < 18  -> "Hey, teenager!"
    | a when a >= 65 -> "Good day, elder!"
    | _              -> "Hello!"
```

If no case matches (all patterns failed or all guards were false), a runtime error is raised.

### 4. Lists

Lists are immutable and can contain any values:

```layer
let numbers = [1, 2, 3, 4, 5] in
let doubled = map(λ(x) -> x * 2, numbers) in
    doubled
```

**List operations:**
- `head(list)` - first element
- `tail(list)` - all but first
- `length(list)` - list length
- `map(fn, list)` - transform elements
- `filter(fn, list)` - select elements
- `concat(list1, list2)` - join lists

### 5. Lambda Functions

Anonymous functions for quick operations:

```layer
# Lambda syntax
λ(x) -> x * 2

# Using with map
map(λ(x) -> x * x, [1, 2, 3, 4, 5])
```

### 6. Records

Records are like objects or dictionaries:

```layer
let person = {name: "Alice", age: 30} in
    person
```

## Working with Layers

### The Core Layer

The core layer is always enabled and provides:
- Arithmetic: `+`, `-`, `*`, `/`, `%`, `^`
- Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`
- Logic: `and`, `or`, `not`
- Lists, strings, and more

### The Debug Layer

Add debugging to your programs:

```layer
@layers [core, debug]

@layer core {
    fn power(base, exp) =
        match exp with
        | 0 -> 1
        | n -> base * power(base, n - 1)
}

@layer debug {
    fn power(base, exp) =
        let _ = log("Computing power") in
            core.power(base, exp)
}

power(2, 8)
```

**Debug functions:**
- `log(message)` - Log a message
- `trace(name, fn)` - Trace function calls
- `inspect(value)` - Inspect value structure
- `benchmark(name, fn)` - Measure execution time

### The UI Layer

Create user interfaces declaratively:

```layer
@layers [core, ui]

@layer ui {
    fn app() =
        view({
            title: "My App",
            children: [
                text({content: "Hello!"}),
                button({label: "Click me"})
            ]
        })
}

render(app())
```

## Layer Composition

The power of Layer comes from composing layers:

```layer
@layers [core, ui, debug]

@layer core {
    fn fibonacci(n) =
        match n with
        | 0 -> 0
        | 1 -> 1
        | n -> fibonacci(n - 1) + fibonacci(n - 2)
}

@layer debug {
    # Wrap core function with logging
    fn fibonacci(n) =
        trace("fib", core.fibonacci)(n)
}

@layer ui {
    fn app() =
        view({
            children: [
                text({content: "Fib(10):"}),
                text({content: fibonacci(10)})
            ]
        })
}

render(app())
```

In this example:
- Core layer defines the algorithm
- Debug layer adds tracing
- UI layer presents the result
- Each layer builds on the ones below

## Layer Introspection

Layer has three built-in expressions for inspecting the layer stack at runtime. They all start with `@` and return plain values — you can use them anywhere an expression is valid.

### `@which name` — who provides this?

Returns the name of the layer that would actually run if you called `name(...)`:

```layer
@layers [core, debug]
@layer core  { fn fib(n) = n }
@layer debug { fn fib(n) = n }

@which fib   # → "debug"
```

Useful for confirming that the right layer is active.

### `@inspect layer` — what's in this layer?

Returns a sorted list of every name defined in the layer (user-defined functions and built-ins):

```layer
@layers [core]
@layer core {
    fn double(x) = x * 2
    fn triple(x) = x * 3
}

@inspect core   # → [..., "double", ..., "triple", ...]
```

### `@diff layerA layerB` — what does B change?

Returns a record with two fields:

- `added` — names in B that A doesn't have
- `overridden` — names that exist in both

```layer
@layers [core, cache]
@layer core  { fn fetch(k) = "db"    }
@layer cache { fn fetch(k) = "cache"
               fn evict(k) = "ok"    }

@diff core cache
# → {added: ["evict"], overridden: ["fetch"]}
```

### Introspection as expressions

All three return values, so you can use them inside `let` bindings or pass results to functions:

```layer
@layers [core, debug]
@layer core  { fn fib(n) = n }
@layer debug { fn fib(n) = n }

let owner   = @which fib in
let changes = @diff core debug in
str_concat("fib is in: ", owner)
```

## Conditional Layer Activation

Any layer in `@layers` can be given an `if <expr>` guard. If the condition is `false` at startup the layer is **completely excluded** from the stack — as though it was never listed.

### Syntax

```layer
@layers [core, debug if mode == "development"]
```

### Literal conditions

```layer
@layers [core, extra if true]   # extra always active
@layers [core, extra if false]  # extra never active
```

### Variable-based feature flags

The most common use is toggling layers based on configuration variables set by the host application (or the REPL via `SetVar`):

```layer
@layers [core, debug if mode == "development"]

@layer core  { fn run() = "production path" }
@layer debug { fn run() = "debug path"      }

run()
# mode = "development"  →  "debug path"
# mode = "production"   →  "production path"
```

### Compound conditions

Any expression works, including `and`/`or` and comparisons:

```layer
@layers [core, verbose if level > 0 and level < 3]
```

### Mixed conditional and unconditional layers

Unconditional layers are unaffected by the feature:

```layer
@layers [core, debug if false, ui]
# Stack after evaluation: ui → core
# debug is gone — @which, @inspect, and explicit debug.fn all fail
```

### Non-boolean condition is a runtime error

```layer
@layers [core, extra if 42]  # ← ERROR: condition must be boolean
```

### Combining with blend modes

`if` guards and `blend:` can be used together:

```layer
@layers [core, cache if caching_enabled, data] blend: fallback
```

## Layer Blend Modes

By default, the **top layer wins** when multiple layers define the same function (override). You can change this with the `blend:` option on `@layers`:

```layer
@layers [core, cache, data] blend: fallback
```

There are four blend modes:

### `override` (default)

The highest layer in the stack that defines the function is called. This is the standard wrapping/decoration pattern.

### `chain`

The result is piped through every layer that defines the function, from bottom to top. Each layer transforms the output of the layer below it.

```layer
@layers [core, sanitize] blend: chain

@layer core {
    fn process(x) = x + 1        # 5 → 6
}
@layer sanitize {
    fn process(x) = x * 2        # 6 → 12
}

process(5)   # → 12
```

#### Multi-param functions: the first parameter is the pipeline slot

When your function has more than one parameter, `chain` uses **partial application**:

- The **first parameter** receives the piped value (what came out of the layer below).
- **All other parameters** are forwarded unchanged from the original call to every layer.

This means every layer can share the same signature — only the first argument changes as it flows up the stack:

```layer
@layers [core, normalize, clamp] blend: chain

@layer core {
    fn scale(value, factor) = value * factor
    # called with original args: scale(3, 10) → 30
}
@layer normalize {
    fn scale(value, factor) = value / factor
    # value=30 (piped), factor=10 (original context) → 3.0
}
@layer clamp {
    fn scale(value, factor) = if value > 1.0 then 1.0 else value
    # value=3.0 (piped), factor=10 (still available) → 1.0
}

scale(3, 10)   # core: 30 → normalize: 3.0 → clamp: 1.0
```

Think of `chain` as partially applying the original call: the **first parameter flows**, all others are **read-only context** shared by every layer in the pipeline.

Use `chain` for **transformation pipelines** — each layer applies a step, and all layers have access to the original call context.

### `collect`

Every layer that defines the function is called with the same arguments. The results are returned as a list (top layer first).

```layer
@layers [core, extra] blend: collect

@layer core  { fn tag(x) = "core"  }
@layer extra { fn tag(x) = "extra" }

tag("x")   # → ["extra", "core"]
```

Use `collect` when multiple layers should **all contribute** — validators, event handlers, feature collectors.

### `fallback`

Layers are tried top → bottom. The first result that is not `data.None()` wins. Requires the `data` layer.

```layer
@layers [core, data, cache] blend: fallback

@layer cache {
    fn fetch(key) = data.None()            # cache miss → try next layer
}
@layer core {
    fn fetch(key) = data.Some("from_db")   # database hit
}

fetch("mykey")   # → Some("from_db")
```

When the cache has a result:

```layer
@layer cache {
    fn fetch(key) = data.Some("cached!")   # returns immediately
}
fetch("mykey")   # → Some("cached!")
```

Use `fallback` for **caching**, **feature flags**, or any **try-first** pattern.

### Notes

- Explicit layer access (`core.factorial`) always bypasses blend mode.
- Unknown blend modes produce a **parse error** before the program runs.

## Partial Application

Use `?` in any argument position to create a partially applied function on the spot. The call is transformed into a lambda at parse time — no special runtime machinery needed.

### Basic usage

```layer
@layers [core]

@layer core {
    fn add(x, y)      = x + y
    fn multiply(x, y) = x * y
}

let add5    = add(5, ?)    in   # fixes x=5, leaves y open
let double  = multiply(2, ?) in   # fixes x=2, leaves y open

add5(3)     # → 8
double(7)   # → 14
```

### Hole in any position

```layer
let addTo10 = add(?, 10) in   # leaves x open, fixes y=10
addTo10(32)  # → 42
```

### With higher-order functions

This is where partial application really pays off — no need for verbose inline lambdas:

```layer
# Without partial application:
map(fn(x) = add(1, x), [10, 20, 30])

# With partial application:
map(add(1, ?), [10, 20, 30])   # → [11, 21, 31]

filter(lt(?, 5), [1, 3, 7, 2, 8])   # → [1, 3, 2]
```

### Multiple holes

Each `?` becomes a separate parameter:

```layer
@layer core { fn combine(a, b, c) = a + b + c }

let f = combine(?, 10, ?) in
f(1, 100)   # → 111
```

### Chaining

The result of partial application is just a function — it can be partially applied again:

```layer
@layer core { fn add3(a, b, c) = a + b + c }

add3(1, ?, ?)(2, ?)(3)   # → 6
```

### Remember: `?` is not `_`

```layer
match x with
| _ -> "wildcard in pattern — matches anything"

add(5, ?)   # hole in expression — fills in later
```

`_` lives in **pattern position**; `?` lives in **expression position**. They never overlap.

## Examples

### Fibonacci Sequence

```layer
@layers [core]

@layer core {
    fn fib(n) =
        match n with
        | 0 -> 0
        | 1 -> 1
        | n -> fib(n - 1) + fib(n - 2)

    fn fib_list(n) =
        map(fib, take(n, [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10]))
}

fib_list(10)
```

### List Processing

```layer
@layers [core]

@layer core {
    fn square(x) = x * x

    fn is_even(x) =
        match x % 2 with
        | 0 -> true
        | _ -> false

    fn sum(lst) =
        match lst with
        | [] -> 0
        | [h, ...t] -> h + sum(t)
}

let numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10] in
let squares = map(square, numbers) in
let even_squares = filter(is_even, squares) in
    sum(even_squares)
```

### Higher-Order Functions

```layer
@layers [core]

@layer core {
    fn compose(f, g) =
        λ(x) -> f(g(x))

    fn twice(f) =
        λ(x) -> f(f(x))

    fn add_one(x) = x + 1
}

let add_two = twice(add_one) in
    add_two(5)
```

## Interactive REPL

Launch the REPL:

```bash
cd layer-go
./layer
```

Try these commands:

```
layer> fn double(x) = x * 2
layer> double(21)
42
layer> let numbers = [1, 2, 3, 4, 5]
layer> map(double, numbers)
[2, 4, 6, 8, 10]
layer> help
```

## Next Steps

1. Explore the examples in `examples/`
2. Read the full language specification in `LANGUAGE_SPEC.md`
3. Check out the architecture documentation in `docs/ARCHITECTURE.md`
4. Experiment with your own layer combinations!

## Tips

1. **Keep layers focused**: Each layer should have a single concern
2. **Use core for logic**: Put pure algorithms in the core layer
3. **Layer for aspects**: Use layers for cross-cutting concerns (logging, UI, etc.)
4. **Explicit is better**: Use `core.function()` when you want to bypass layer resolution entirely
5. **Think in compositions**: Build complex behavior by stacking simple layers
6. **Use guard clauses instead of nested ifs**: `| n when n < 0 -> ...` is clearer than nested conditionals
7. **Choose blend modes intentionally**: `fallback` for caching, `chain` for pipelines, `collect` when all layers matter — and in `chain`, design your first parameter as the "pipeline value" and the rest as shared context
8. **Use `@which` to debug surprises**: if a function behaves unexpectedly, `@which name` tells you exactly which layer is running it
9. **Use `@diff` to understand a layer**: before wrapping or overriding, run `@diff lower_layer your_layer` to see what you're changing
10. **Put feature flags in `@layers`, not inside functions**: `@layers [core, debug if mode == "dev"]` is cleaner than scattering `if mode == "dev"` checks throughout the code
11. **Use `?` instead of inline lambdas with HOFs**: `map(add(1, ?), list)` is cleaner than `map(fn(x) = add(1, x), list)`

Happy coding with Layer! 🎨
