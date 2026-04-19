# Layer Language Philosophy

**A purely functional language with practical effect management**

---

## Table of Contents

- [Core Philosophy](#core-philosophy)
- [The Purity Paradox](#the-purity-paradox)
- [Layer's Solution: Spatial Separation of Effects](#layers-solution-spatial-separation-of-effects)
- [Design Principles](#design-principles)
- [Why Functional Programming?](#why-functional-programming)
- [The Layer System](#the-layer-system)
- [Effect Management](#effect-management)
- [Language Design Choices](#language-design-choices)
- [Practical Benefits](#practical-benefits)

---

## Core Philosophy

Layer is built on a simple but powerful idea:

> **Maximize purity where it matters, isolate effects where they're necessary.**

The language doesn't try to eliminate side effects (impossible for useful programs!), but instead provides a **spatial organization** through the layer system that makes the distinction between pure and impure code explicit and manageable.

### The Central Tension

Functional programming ideally is about:
- **Pure functions**: Same input → Same output, always
- **No side effects**: Functions don't modify external state or depend on it
- **Referential transparency**: Can replace a function call with its result

But reality requires:
- **File I/O**: Reading and writing data
- **Databases**: Persistent storage
- **User interfaces**: Interactive applications
- **Network requests**: Communication with external services
- **Logging**: Debugging and monitoring

**The question**: How do we reconcile these?

**Layer's answer**: The layer system provides spatial separation of concerns, making effects explicit while keeping business logic pure.

---

## The Purity Paradox

As the saying goes:
> *"A purely functional program is a heat generator."*

A program that can't interact with the world is useless. But a program with side effects everywhere is unmaintainable.

### Why Purity Matters

#### 1. Reasoning and Correctness

Pure functions are **predictable**:

```layer
# Pure function - always predictable
fn add(x, y) = x + y

# Can confidently say:
add(2, 3) == add(2, 3)  # Always true!
add(2, 3) == 5          # Always true!
```

vs. impure (in other languages):

```javascript
// Impure - depends on global state
let counter = 0;
function addWithCounter(x, y) {
    counter++;
    return x + y + counter;
}

addWithCounter(2, 3)  // 6
addWithCounter(2, 3)  // 7  ← Different result!
```

**Benefits of purity**:
- ✅ Easy to test (no mocks needed)
- ✅ Easy to parallelize (no shared state)
- ✅ Easy to cache/memoize (deterministic results)
- ✅ Easy to understand (local reasoning)

#### 2. Composition

Pure functions compose beautifully:

```layer
fn double(x) = x * 2
fn addOne(x) = x + 1
fn square(x) = x * x

# Compose them
fn transform(x) = square(addOne(double(x)))

# Can refactor without fear:
fn transform(x) =
    let a = double(x) in
    let b = addOne(a) in
    square(b)

# Both are IDENTICAL - composition is safe!
```

#### 3. Testability

Pure functions don't need mocks:

```layer
# Pure function - easy to test
@layer core {
    fn processOrder(order) =
        match order.status with
        | "pending" -> {...order, status: "processing"}
        | "processing" -> {...order, status: "shipped"}
        | _ -> order
}

# Test (no database, no mocks needed):
assert_equal(
    processOrder({status: "pending", id: 1}),
    {status: "processing", id: 1}
)
```

#### 4. Portability

Pure business logic works everywhere:

```layer
# This code from 2020:
@layer core {
    fn calculateShipping(weight, distance) =
        weight * 0.5 + distance * 0.1
}

# Still works in 2025!
# Works in CLI, web, mobile, embedded
# No changes needed when:
# - Database changes (MySQL → Postgres)
# - UI changes (React → Vue)
# - Infrastructure changes (AWS → Azure)
```

---

## Layer's Solution: Spatial Separation of Effects

Instead of trying to eliminate effects or hide them in monads, Layer uses **layers** to spatially organize code by its purity level.

### The Architecture

```
┌────────────────────────────────┐
│         Pure Core              │
│                                │
│  ┌──────────────────────────┐  │
│  │  Business Logic          │  │
│  │  (all pure functions)    │  │
│  │  - Calculations          │  │
│  │  - Transformations       │  │
│  │  - Validation            │  │
│  └──────────────────────────┘  │
│              ▲                 │
│              │                 │
└──────────────┼─────────────────┘
               │
        ┌──────┴──────┐
        │             │
   ┌────▼────┐   ┌────▼────┐
   │  Input  │   │ Output  │
   │ (Files, │   │ (Files, │
   │  DB,    │   │  UI,    │
   │  UI)    │   │  DB)    │
   └─────────┘   └─────────┘
     Effects       Effects
```

### Layer Stack

Layers stack from bottom to top, with purity decreasing as you go up:

```
┌─────────────────┐
│   Debug Layer   │ ← Logging, tracing (impure)
├─────────────────┤
│   Data Layer    │ ← File I/O, databases (impure)
├─────────────────┤
│   UI Layer      │ ← User interface (impure)
├─────────────────┤
│   Core Layer    │ ← Pure functional primitives (pure)
└─────────────────┘
```

**Key insight**: When you see a function in the `core` layer, you **know** it's pure. When you see a function in the `data` layer, you **expect** effects.

---

## Design Principles

### 1. Expression-Based Everything

Layer has no statements, only expressions. Every construct returns a value:

```layer
# if is an expression
let result = if x > 0 then "positive" else "negative" in result

# match is an expression
let category = match age with
    | a when a < 18 -> "minor"
    | a when a < 65 -> "adult"
    | _ -> "senior"
in category

# Functions are expressions
fn add(x, y) = x + y
```

**Benefits**:
- Simpler mental model (no statement/expression distinction)
- More composable (expressions nest naturally)
- No implicit returns (no special rules)

### 2. Immutability by Default

All bindings are immutable:

```layer
let x = 5 in
let x = 10 in  # New binding, shadows previous
x  # Returns 10
```

**Benefits**:
- Easier to reason about (values don't change)
- Safe parallelization (no race conditions)
- Simpler debugging (no temporal dependencies)

### 3. Explicit Scoping

The `let ... in` syntax makes scoping explicit:

```layer
let a = 1 in
let b = 2 in
let c = a + b in
c * 2
```

**Why `in`?**
- **Parsing clarity**: No ambiguity about where expressions end
- **No semicolons**: The `in` keyword serves as a delimiter
- **Whitespace-insensitive**: Format however you want
- **Explicit scope boundaries**: Clear where variables are available

Example of ambiguity without `in`:

```
# Ambiguous!
let result = if condition
    then value1
    else value2
let next = result + 1
```

Is `let next` part of `result`'s expression or separate?

With `in`, it's unambiguous:

```layer
let result = if condition then value1 else value2 in
let next = result + 1 in
next
```

### 4. Pattern Matching

Pattern matching is the primary control flow mechanism:

```layer
fn factorial(n) =
    match n with
    | 0 -> 1
    | n -> n * factorial(n - 1)
```

Guard clauses add boolean conditions to patterns — the case only fires if the pattern matches *and* the guard is true:

```layer
fn classify(n) =
    match n with
    | n when n < 0   -> "negative"
    | 0              -> "zero"
    | n when n > 100 -> "large"
    | n              -> "positive"
```

**Benefits**:
- More expressive than if/else chains
- Guards eliminate nested conditionals
- Compiler can verify exhaustiveness
- Natural for working with data structures

### 5. Closures and First-Class Functions

Functions can capture their environment:

```layer
fn makeAdder(x) = fn(y) = x + y

let add5 = makeAdder(5) in
add5(3)  # Returns 8
```

**Benefits**:
- Higher-order functions (map, filter, reduce)
- Partial application
- Callback patterns

---

## The Layer System

### Layer as a Composable Unit

A layer is like Photoshop's layers for code - you can stack different aspects of your program:

```layer
@layers [core, ui, debug, data]

@layer core {
    fn factorial(n) =
        match n with
        | 0 -> 1
        | n -> n * factorial(n - 1)
}

@layer debug {
    # Wrap core.factorial with logging
    fn factorial(n) =
        let result = core.factorial(n) in
        debug.log("factorial({n}) = {result}")
        result
}

@layer ui {
    fn calculator() = view({
        title: "Factorial Calculator",
        children: [
            input({name: "n", placeholder: "Enter number", type: "number"}),
            button({label: "Calculate", onclick: "factorial(n)"})
        ]
    })
}
```

### Layer Resolution

When you call `factorial(5)`, the interpreter searches layers from top to bottom:

1. Check `debug` layer → Found! Call `debug.factorial`
2. Inside `debug.factorial`, call `core.factorial` explicitly
3. Pure computation happens in `core` layer
4. Logging happens in `debug` layer

**Explicit layer access**: `core.factorial(5)` bypasses layer override.

### Layer Benefits

1. **Separation of Concerns**: Each layer handles one aspect
2. **Composability**: Layers can be mixed and matched
3. **Aspect-Oriented Programming**: Cross-cutting concerns (logging, caching) are layers
4. **Override Patterns**: Higher layers can wrap lower layers
5. **Visual Organization**: Layers make architecture visible

### Blend Modes — Beyond Override

The Photoshop metaphor extends further than just stacking. Photoshop layers have blend modes (Multiply, Screen, Overlay…). Layer has four:

| Mode | What it does |
|---|---|
| `override` | Top layer wins — the default, for decoration/wrapping |
| `chain` | Result is piped bottom → top through every layer |
| `collect` | All layers contribute; results gathered into a list |
| `fallback` | Top layer first; if it returns `None`, try the next |

```layer
@layers [core, cache, data] blend: fallback
```

This is the most direct expression of the Photoshop metaphor in the language. Instead of asking "which layer wins?", you ask "how should the layers *combine*?" That question is architectural and makes intent explicit in the program's opening line.

---

## Effect Management

Layer takes a **pragmatic approach** to effects: isolate them, don't eliminate them.

### Effect Isolation by Layer

```layer
@layers [core, data]

# PURE CORE - No side effects
@layer core {
    fn calculateDiscount(price, percentage) =
        price * (1 - percentage / 100)

    fn applyDiscountRule(order) =
        match order.total with
        | t when t > 100 -> calculateDiscount(t, 10)
        | t when t > 50  -> calculateDiscount(t, 5)
        | t              -> t
}

# EFFECTS AT THE EDGES
@layer data {
    fn processOrder(orderId) =
        # Effect: Read from database
        let order = data.read_json("orders/{orderId}.json") in

        # Pure computation (calls core layer)
        let newTotal = core.applyDiscountRule(order) in

        # Effect: Write to database
        data.write_json("orders/{orderId}.json", {
            ...order,
            total: newTotal
        })
}
```

**Result**: 90% pure, 10% effects!

### The Option Pattern

Effects that can fail return `Option` types:

```layer
@layer data {
    fn read_file(path) =
        # Returns: Some("content") or None()
}

# Usage:
let result = data.read_file("config.txt") in
match result with
| data.Some(content) -> "Found: {content}"
| data.None() -> "File not found"
```

**Benefits**:
- Explicit error handling (no exceptions)
- Type-safe (can't ignore errors)
- Composable (can map over Options)

### Effect Layers

Layer provides specific layers for different effect types:

- **`core`**: Pure functions (no effects)
- **`ui`**: User interface (render effects)
- **`data`**: File I/O, databases (I/O effects)
- **`debug`**: Logging, tracing (logging effects)

This makes effects **explicit and organized**.

---

## Language Design Choices

### 1. Why `let ... in`?

Borrowed from ML-family languages (Haskell, OCaml, F#):

**Advantages**:
- Clear delimiter (no semicolons needed)
- Whitespace-insensitive
- Simple parser (no lookahead)
- Unambiguous grammar
- Expression-based (no special statement rules)

**Example**:
```layer
let a = 1 in let b = 2 in a + b  # One line

let a = 1 in
let b = 2 in
a + b  # Multiple lines - same result
```

### 2. Why Immutability?

**Immutability eliminates entire classes of bugs**:
- No race conditions (safe parallelism)
- No temporal coupling (order doesn't matter for pure functions)
- Easier debugging (values don't change)
- Simpler mental model (no hidden state mutations)

### 3. Why Pattern Matching?

Pattern matching is more expressive than if/else:

```layer
# With pattern matching
fn describe(value) =
    match value with
    | 0 -> "zero"
    | 1 -> "one"
    | n when n < 0 -> "negative"
    | n -> "positive"

# Without (verbose!)
fn describe(value) =
    if value == 0 then "zero"
    else if value == 1 then "one"
    else if value < 0 then "negative"
    else "positive"
```

### 4. Why Layers Instead of Modules?

**Layers provide spatial organization** that modules don't:

```layer
# Layers (vertical composition)
@layer debug {
    fn factorial(n) =
        debug.log("Computing...")
        core.factorial(n)
}

# Modules (horizontal composition)
module Math {
    fn factorial(n) = ...
}
module Logger {
    fn log(msg) = ...
}
# How do they interact? Not clear!
```

Layers make **cross-cutting concerns** natural:
- Debugging (wraps all functions)
- Caching (intercepts calls)
- Authentication (guards access)

---

## Practical Benefits

### 1. Testing Without Mocks

Pure core = no mocks needed:

```layer
@layer core {
    fn validateEmail(email) =
        match email with
        | e when str_length(e) < 3 -> false
        | e when not (contains(e, "@")) -> false
        | _ -> true
}

# Test (no dependencies!)
assert_equal(validateEmail("user@example.com"), true)
assert_equal(validateEmail("invalid"), false)
```

### 2. Fearless Refactoring

Pure functions can be refactored without fear:

```layer
# Original
fn calculate(x) = (x + 1) * 2

# Refactored
fn calculate(x) =
    let a = x + 1 in
    let b = a * 2 in
    b

# IDENTICAL behavior - guaranteed!
```

### 3. Parallel Execution

Pure functions can run in parallel safely:

```layer
# Can parallelize automatically (no shared state)
let results = map(expensiveComputation, [1, 2, 3, 4, 5])

# Runtime can:
# - Execute in parallel threads
# - Memoize results
# - Reorder operations
```

### 4. Long-Term Maintainability

**Pure business logic never rots**:

```layer
# Written in 2020
@layer core {
    fn calculateTax(amount, rate) = amount * rate
}

# Still works perfectly in 2030!
# - No dependencies on frameworks
# - No dependencies on databases
# - No dependencies on UI libraries
# - Just math!
```

### 5. Cross-Platform Portability

The same Layer code runs:
- **CLI**: Native Go binary
- **Browser**: WebAssembly
- **Server**: HTTP backend
- **Mobile**: Embedded interpreter (future)

**Why?** Pure core is platform-agnostic!

---

## The Pragmatic Middle Ground

Layer aims for **practical functional programming**:

### Not Extremist Purity

Layer is not Haskell. We don't:
- Force all effects through monads
- Require formal proofs
- Hide I/O behind complex type systems

### Not Imperative Chaos

Layer is not JavaScript/Python. We don't:
- Allow mutable state everywhere
- Mix pure and impure code freely
- Leave effects implicit

### The Sweet Spot

```
Pure FP ←────── Layer ──────→ Imperative
(Haskell)      (Sweet Spot)    (JavaScript)

Too rigid       ✅ Just right    Too chaotic
```

**Layer's philosophy**:
- **90% pure code** (core business logic)
- **10% effects** (isolated in specific layers)
- **Explicit separation** (layers make it visible)
- **Practical compromise** (effects when needed, purity when possible)

---

## Comparison with Other Approaches

### Haskell: IO Monad

```haskell
-- Pure function that DESCRIBES an effect
readFile :: FilePath -> IO String

-- Composed effects
processFile :: FilePath -> IO ()
processFile path = do
    content <- readFile path
    let upper = toUpper content  -- Pure!
    writeFile path upper
```

**Pros**: Type system guarantees purity

**Cons**: Steep learning curve, monads are complex

### Elm: Managed Effects

```elm
-- Pure update function
update : Msg -> Model -> (Model, Cmd Msg)
update msg model =
    case msg of
        Increment -> (model + 1, Cmd.none)
        LoadData -> (model, Http.get ...)

-- Runtime executes effects
```

**Pros**: Simple model, impossible states impossible

**Cons**: Limited to UI, can't express arbitrary effects

### Layer: Spatial Layers

```layer
@layer core {
    fn update(model, msg) = ...  -- Pure!
}

@layer data {
    fn loadData() = ...  -- Effect!
}
```

**Pros**: Visual separation, flexible, learnable

**Cons**: Less compile-time guarantees than Haskell

---

## The Goal

Layer aims to make functional programming **accessible and practical**:

```
┌─────────────────────────────────┐
│  90% Pure Code                  │  ← Easy to test, reason, reuse
│  @layer core {                  │
│    Business logic               │
│    Calculations                 │
│    Transformations              │
│    Validation                   │
│  }                              │
├─────────────────────────────────┤
│  10% Impure Code                │  ← Isolated, explicit
│  @layer data {                  │
│    File I/O                     │
│    Database                     │
│  }                              │
│  @layer ui {                    │
│    User interface               │
│  }                              │
│  @layer debug {                 │
│    Logging                      │
│  }                              │
└─────────────────────────────────┘
```

**Philosophy in one sentence**:

> **Layer makes the distinction between pure and impure code explicit through spatial organization, enabling you to maximize the benefits of functional programming without sacrificing practicality.**

---

## Further Reading

- [LANGUAGE_SPEC.md](LANGUAGE_SPEC.md) - Complete language specification
- [TUTORIAL.md](docs/TUTORIAL.md) - Learn Layer by example
- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - Implementation details
- [examples/](examples/) - Real-world examples demonstrating the philosophy

---

**Layer**: *Pure functions, practical effects, spatial organization.*
