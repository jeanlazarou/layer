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

Use `chain` for **transformation pipelines** — each layer applies a step.

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
7. **Choose blend modes intentionally**: `fallback` for caching, `chain` for pipelines, `collect` when all layers matter

Happy coding with Layer! 🎨
