# Layer Language Specification

## Overview

Layer is an experimental purely functional programming language built around the concept of "layer programming". The language allows developers to compose programs from multiple layers, each providing different aspects of functionality.

## Core Concepts

### Layers

A Layer program consists of one or more layers stacked on top of each other. Each layer can:
- Define new functions and values
- Transform or extend definitions from lower layers
- Access and compose with lower layer functionality

**Built-in Layers:**
- `core` (mandatory) - Base functional programming primitives (36 functions)
- `ui` - User interface components and rendering (5 components)
- `debug` - Debugging utilities, logging, and introspection (8 functions)
- `data` - File I/O, JSON, and localStorage with Option pattern (21 functions)

**You can also create your own custom layers!** See [Creating Custom Layers](#creating-custom-layers) below.

### The `@layers` Directive

The `@layers` directive has **four purposes**:

1. **Enable/Disable**: Only layers listed in `@layers [...]` are active
2. **Set Order**: The order in the list determines search priority
3. **Conditional activation** (optional): Each layer can have an `if <expr>` guard
4. **Set Blend Mode** (optional): Controls how functions defined in multiple layers are combined

```layer
@layers [core, middleware, debug]
#        └─ These 3 layers are enabled
#        └─ Search order: debug → middleware → core (right to left)

@layer core { ... }       # ✅ Enabled (in @layers list)
@layer middleware { ... } # ✅ Enabled (in @layers list)
@layer debug { ... }      # ✅ Enabled (in @layers list)
@layer cache { ... }      # ❌ Disabled (NOT in @layers list)
```

### Layer Composition and Ordering

Layers are composed from **bottom to top**, similar to Photoshop layers. The order matters!

```
┌─────────────────┐
│   Top Layer     │ ← Searched FIRST (rightmost in @layers)
├─────────────────┤
│  Middle Layer   │ ← Searched second
├─────────────────┤
│  Bottom Layer   │ ← Searched last (leftmost in @layers)
└─────────────────┘
```

**Layer resolution order**: When you call a function, Layer searches from **right to left** in the `@layers` list:

```layer
@layers [core, middleware, debug]  # Order matters!

@layer core {
  fn compute(x) = x * 2
}

@layer middleware {
  fn compute(x) = core.compute(x) + 1  # Wraps core
}

@layer debug {
  fn compute(x) =
    let result = middleware.compute(x) in
    debug.log("Result: {result}")
    result
}

compute(5)  # Calls debug.compute → middleware.compute → core.compute
            # Result: 11 (5 * 2 + 1)
```

**Function resolution**:
1. Call `compute(5)`
2. Search **debug layer** first → Found! Call `debug.compute(5)`
3. Inside debug, calls `middleware.compute(5)` → Found in middleware layer
4. Inside middleware, calls `core.compute(5)` → Found in core layer
5. Core returns `10`, middleware adds `1` = `11`, debug logs and returns `11`

**Each layer can**:
- **Override** functions from lower layers
- **Wrap** lower layer functions with additional behavior (logging, validation, caching)
- **Add** new functionality not present in lower layers
- **Access** any lower layer explicitly (e.g., `core.factorial(n)`)

### Purely Functional

Layer enforces pure functional programming:
- Immutable data structures
- First-class functions
- No side effects in core (effects handled through layers)
- Pattern matching
- Lazy evaluation support

## Syntax

### Basic Expressions

```layer
# Comments start with #

# Function definition
fn add(x, y) = x + y

# Lambda expressions
λ(x) -> x * 2

# Let bindings
let result = add(5, 3) in
  result * 2

# Pattern matching
match value with
| 0 -> "zero"
| n -> "non-zero"

# Pattern matching with guard clauses
match value with
| n when n < 0 -> "negative"
| 0            -> "zero"
| n            -> "positive"
```

### Data Types

```layer
# Numbers
42
3.14

# Strings
"hello world"

# Booleans
true
false

# Lists
[1, 2, 3, 4]

# Records
{ name: "Layer", version: 1.0 }

# Tuples
(1, "hello", true)
```

### Pattern Matching

Pattern matching is the primary control flow mechanism. Cases are tried in order; the first matching case runs.

#### Basic patterns

```layer
fn describe(x) =
    match x with
    | 0    -> "zero"
    | 1    -> "one"
    | n    -> "other"   # identifier pattern — binds x to n
    | _    -> "ignored" # wildcard — never reached here, but valid
```

#### Guard clauses

**Status**: ✅ **Implemented**

A `when` clause after the pattern adds a boolean condition. The case only matches if both the pattern matches **and** the guard evaluates to `true`. If the guard is `false`, the next case is tried.

```layer
fn classify(n) =
    match n with
    | n when n < 0   -> "negative"
    | 0              -> "zero"
    | n when n > 100 -> "large"
    | n              -> "positive"
```

Guards can reference variables bound by the pattern:

```layer
fn first_is_one(list) =
    match list with
    | [h, ...rest] when h == 1 -> "starts with one"
    | [h, ...rest]             -> "other"
    | []                       -> "empty"
```

If a guard expression evaluates to a non-boolean value, a runtime error is raised.

#### List patterns

```layer
fn sum(lst) =
    match lst with
    | []           -> 0
    | [h, ...tail] -> h + sum(tail)
```

#### Record patterns

**Status**: ✅ **Implemented**

Record patterns destructure records by field name. Each field value is itself a sub-pattern. Matching is **partial** — the value may have extra fields that the pattern does not mention.

```layer
fn greet(person) =
    match person with
    | {name: n} -> str_concat("Hello, ", n)

greet({name: "Alice", age: 30})  # → "Hello, Alice"
```

Multiple fields:

```layer
fn describe(p) =
    match p with
    | {name: n, age: a} -> str_concat(n, str_concat(" is ", to_string(a)))

describe({name: "Bob", age: 25})  # → "Bob is 25"
```

Field values can be any sub-pattern, including literals:

```layer
match {status: "active", id: 42} with
| {status: "active", id: n} -> n    # → 42
| {status: _}               -> 0
```

Record patterns compose with guard clauses:

```layer
fn classify(person) =
    match person with
    | {age: a} when a < 18 -> "minor"
    | {age: a} when a < 65 -> "adult"
    | {age: a}             -> "senior"
```

If a field named in the pattern does not exist in the value, the case is skipped (falls through to the next case).

### Layer Declarations

The `@layers` directive **enables** specific layers and sets their **order**. Only enabled layers are active!

```layer
# Enable ONLY these layers (in this order)
@layers [core, ui, debug]

# Define layer-specific code
@layer core {
  fn factorial(n) =
    match n with
    | 0 -> 1
    | n -> n * factorial(n - 1)
}

@layer ui {
  # UI layer can reference core functions
  fn render_result(n) =
    view {
      text: "Factorial of {n} is {factorial(n)}"
    }
}

@layer debug {
  # Debug layer can wrap core functions
  fn factorial(n) = {
    log("Computing factorial of {n}")
    let result = core.factorial(n) in {
      log("Result: {result}")
      result
    }
  }
}

# data layer is defined but NOT enabled
@layer data {
  fn save_result(n, result) = ...
}

# Calling data.save_result() will fail - layer not enabled!
```

**Important**: You can define layers in `@layer` blocks but they're only active if listed in `@layers [...]`.

### Main Expression

After defining layers, you write a **main expression** - code that executes outside any layer definition. This is the entry point of your program.

```layer
@layers [core, ui]

@layer core {
  fn factorial(n) = ...
}

@layer ui {
  fn app() = view({...})
}

# Main expression - executes when program runs
render(app())
```

**Key points about the main expression**:

1. **It's where execution starts** - After all layers are defined, the main expression runs
2. **It's impure** - The main expression typically causes side effects (output, rendering, I/O)
3. **It has access to all enabled layers** - Can call functions from any layer in `@layers [...]`
4. **It's evaluated once** - Not inside a function, runs directly when the program executes

**Examples**:

```layer
# Example 1: Output a calculation
@layers [core]

@layer core {
  fn fibonacci(n) = if n <= 1 then n else fibonacci(n-1) + fibonacci(n-2)
}

fibonacci(10)  # Main expression - outputs: 55
```

```layer
# Example 2: Render UI
@layers [core, ui]

@layer core {
  fn square(x) = x * x
}

@layer ui {
  fn calculator() = view({...})
}

render(calculator())  # Main expression - renders UI to screen
```

```layer
# Example 3: Multiple operations in main expression
@layers [core, data, debug]

@layer core {
  fn process(x) = x * 2
}

let input = data.read_file("input.txt") in
let result = core.process(data.unwrap(input, "0")) in
debug.log("Result: {result}")
result  # Main expression returns this value
```

**Purity of main expression**:

The main expression is **inherently impure** because:
- It causes observable effects (output to console, file system, screen)
- It's the boundary between pure computation and the external world
- It's where `render()`, `print()`, file I/O, etc. actually execute

Think of the main expression as Layer's equivalent to:
- Haskell's `main :: IO ()`
- Python's `if __name__ == "__main__":`
- C's `int main() { ... }`

It's the **impure entry point** that orchestrates pure functions.

### Enabling and Disabling Layers

Layers can be selectively enabled for different scenarios:

#### Development vs Production

```layer
# Development mode - all debugging enabled
@layers [core, validation, auth, debug, profiling]

# OR

# Production mode - debugging disabled
@layers [core, validation, auth]

# Same code, different layer configuration!
@layer core {
  fn process_payment(amount, card) = ...
}

@layer validation {
  fn process_payment(amount, card) =
    if amount <= 0 then "Invalid amount"
    else core.process_payment(amount, card)
}

@layer debug {
  fn process_payment(amount, card) =
    debug.log("Processing payment: ${amount}")
    validation.process_payment(amount, card)
}

@layer profiling {
  fn process_payment(amount, card) =
    debug.benchmark("payment",
      fn() = debug.process_payment(amount, card))
}
```

**In development**: Payment is validated, logged, and profiled

**In production**: Payment is only validated (no logging overhead)

#### Testing Different Layer Combinations

```layer
# Test 1: Without caching
@layers [core, validation]

# Test 2: With caching
@layers [core, validation, cache]

# Test 3: With caching but no validation (measure raw performance)
@layers [core, cache]

@layer core {
  fn expensive_query(id) = ...  # Database query
}

@layer validation {
  fn expensive_query(id) =
    if id < 0 then "Invalid ID"
    else core.expensive_query(id)
}

@layer cache {
  fn expensive_query(id) =
    let cached = data.get("query_{id}") in
    match cached with
    | data.Some(result) -> result
    | data.None() ->
        let result = validation.expensive_query(id) in
        data.set("query_{id}", result)
        result
}
```

#### Feature Flags via Layers

```layer
# Feature flag: Enable experimental UI
@layers [core, ui_experimental]

# OR

# Stable UI
@layers [core, ui_stable]

@layer core {
  fn get_users() = [{name: "Alice"}, {name: "Bob"}]
}

@layer ui_stable {
  fn render_users() =
    view({
      children: [text({content: "User List"})]
    })
}

@layer ui_experimental {
  fn render_users() =
    view({
      children: [
        text({content: "User List (Beta)"}),
        button({label: "Export", onclick: "export_users()"})
      ]
    })
}
```

Switch between features just by changing `@layers`!

#### Conditional Layer Loading

```layer
# Mobile version - no data layer (offline-first)
@layers [core, ui, cache]

# OR

# Desktop version - full data sync
@layers [core, ui, cache, data, sync]

@layer sync {
  fn save_document(doc) =
    # Save locally first
    let local_save = cache.save_document(doc) in
    # Then sync to server
    data.write_json("documents/{doc.id}.json", doc)
    local_save
}
```

### Creating Custom Layers

**You can create your own layers for any purpose!** Custom layers let you organize cross-cutting concerns and add domain-specific functionality.

#### Example: Validation Layer

```layer
@layers [core, validation, ui]

@layer core {
  fn divide(x, y) = x / y
  fn process_user(name, age) = {name: name, age: age}
}

@layer validation {
  # Wrap core functions with validation
  fn divide(x, y) =
    if y == 0 then
      "Error: Division by zero"
    else
      core.divide(x, y)

  fn process_user(name, age) =
    if str_length(name) < 2 then
      "Error: Name too short"
    else if age < 0 || age > 150 then
      "Error: Invalid age"
    else
      core.process_user(name, age)
}

# Now all functions are automatically validated!
divide(10, 0)  # "Error: Division by zero"
divide(10, 2)  # 5.0
process_user("A", 25)  # "Error: Name too short"
process_user("Alice", 25)  # {name: "Alice", age: 25}
```

#### Example: Caching Layer

```layer
@layers [core, cache]

@layer core {
  fn expensive_computation(n) =
    # Simulate expensive work
    fold(fn(acc, x) = acc + x, 0, range(0, n))
}

@layer cache {
  # Cache results to avoid recomputation
  # (In real implementation, would use mutable storage at edges)
  fn expensive_computation(n) =
    let cache_key = "compute_{n}" in
    let cached = data.get(cache_key) in
    match cached with
    | data.Some(value) ->
        debug.log("Cache hit for {n}")
        to_float(value)
    | data.None() ->
        debug.log("Cache miss for {n}")
        let result = core.expensive_computation(n) in
        data.set(cache_key, to_string(result))
        result
}

expensive_computation(1000)  # Slow - computes
expensive_computation(1000)  # Fast - from cache!
```

#### Example: Authorization Layer

```layer
@layers [core, auth, ui]

@layer core {
  fn delete_user(user_id) = ...
  fn get_salary(user_id) = ...
  fn get_profile(user_id) = ...
}

@layer auth {
  # Protect sensitive operations
  fn delete_user(user_id) =
    let current_user = data.get("current_user") in
    match current_user with
    | data.Some(user) ->
        if user == "admin" then
          core.delete_user(user_id)
        else
          "Error: Unauthorized"
    | data.None() ->
        "Error: Not logged in"

  fn get_salary(user_id) =
    # Only managers and self can see salary
    let current_user = data.get("current_user") in
    if current_user == data.Some("manager") || current_user == data.Some(user_id) then
      core.get_salary(user_id)
    else
      "Error: Unauthorized"

  # get_profile is public - don't override, use core version
}

delete_user("123")  # "Error: Not logged in" (no override in auth)
get_salary("123")   # Protected by auth layer
get_profile("123")  # Uses core directly (not overridden)
```

#### Example: Internationalization (i18n) Layer

```layer
@layers [core, i18n, ui]

@layer core {
  fn greet(name) = "Hello, {name}!"
  fn error_message(code) =
    match code with
    | 404 -> "Not found"
    | 500 -> "Server error"
    | _ -> "Unknown error"
}

@layer i18n {
  fn get_lang() = data.unwrap(data.get("language"), "en")

  fn greet(name) =
    match get_lang() with
    | "es" -> "¡Hola, {name}!"
    | "fr" -> "Bonjour, {name}!"
    | "de" -> "Hallo, {name}!"
    | _ -> core.greet(name)  # Default to English

  fn error_message(code) =
    let base = core.error_message(code) in
    match get_lang() with
    | "es" ->
        match code with
        | 404 -> "No encontrado"
        | 500 -> "Error del servidor"
        | _ -> "Error desconocido"
    | "fr" ->
        match code with
        | 404 -> "Pas trouvé"
        | 500 -> "Erreur du serveur"
        | _ -> "Erreur inconnue"
    | _ -> base
}

# Change language
data.set("language", "es")
greet("María")  # "¡Hola, María!"
error_message(404)  # "No encontrado"
```

### Layer Ordering Strategies

**Order matters!** Choose the right order based on what you want each layer to see:

#### Strategy 1: Validation Before Caching
```layer
@layers [core, validation, cache]
# validation wraps core, cache wraps validation
# Result: Only valid results get cached
```

#### Strategy 2: Caching Before Validation
```layer
@layers [core, cache, validation]
# cache wraps core, validation wraps cache
# Result: Cache bypasses validation (faster but potentially unsafe)
```

#### Strategy 3: Multiple Aspects
```layer
@layers [core, validation, auth, cache, logging, ui]
# Execution order for function calls:
# ui → logging → cache → auth → validation → core
```

**General principle**:
- Put **effects** (logging, debugging) at the **top** (wrap everything)
- Put **optimizations** (caching) **after** safety layers (validation, auth)
- Put **business logic** (core) at the **bottom**

### Custom Layer Best Practices

1. **Single Responsibility** - Each layer should have one clear purpose
   ```layer
   # ✅ GOOD
   @layer validation { ... }  # Only validates
   @layer cache { ... }       # Only caches

   # ❌ BAD
   @layer helpers { ... }  # Vague, does everything
   ```

2. **Explicit Layer Access** - Use `layer.function()` to avoid ambiguity
   ```layer
   # ✅ GOOD - Explicit
   fn compute(x) = core.compute(x) + 1

   # ⚠️ OK but unclear
   fn compute(x) = compute(x) + 1  # Which compute?
   ```

3. **Document Layer Order** - Make ordering dependencies clear
   ```layer
   # ✅ GOOD - Documented
   @layers [core, validation, cache]
   # Order: cache wraps validation wraps core
   # Ensures only valid results are cached
   ```

4. **Keep Core Pure** - Only core should be pure logic
   ```layer
   # ✅ GOOD
   @layer core {
     fn calculate(x, y) = x + y  # Pure!
   }

   @layer validation {
     fn calculate(x, y) =  # Can have effects
       data.log("Validating...")
       if y == 0 then "error" else core.calculate(x, y)
   }
   ```

### Layer Override/Composition

```layer
# Access specific layer's version
core.factorial(5)

# Use current layer's version (with all active layers applied)
factorial(5)

# Compose layers explicitly
ui.render_result(5)
```

## Conditional Layer Activation

**Status**: ✅ **Implemented**

Each layer in the `@layers` list can carry an optional `if <expr>` guard. When the condition evaluates to `false` the layer is **completely skipped** — its `@layer` declaration block is never processed and it never appears in the function resolution stack. When it evaluates to `true` the layer is enabled as normal. Omitting the guard is the same as `if true`.

```layer
@layers [core, debug if mode == "development", perf if level > 2]
```

### Syntax

```
@layers [<name> [if <expr>], ...]
```

- `<expr>` is any Layer expression — identifier, comparison, logical operator, etc.
- The expression is evaluated **once**, at program start, before any function call.
- The expression must evaluate to a **boolean**; a non-boolean value is a runtime error.

### Examples

**Literal conditions**

```layer
@layers [core, extra if true]   # extra always enabled
@layers [core, extra if false]  # extra always skipped
```

**Variable-based feature flags**

```layer
@layers [core, debug if mode == "development"]

@layer core  { fn run() = "core"  }
@layer debug { fn run() = "debug" }

# With mode = "development":  run() → "debug"
# With mode = "production":   run() → "core"
```

**Compound conditions**

```layer
@layers [core, verbose if level > 0 and level < 3]
```

**Mixed conditional and unconditional layers**

```layer
@layers [core, debug if false, ui]
# debug is skipped entirely; ui is always active
# function resolution: ui → core
```

### Semantics

- A disabled layer leaves **no trace** in the layer stack — `@which`, `@inspect`, and explicit layer access (e.g., `debug.fn`) will not find it.
- Conditions are evaluated in the interpreter's initial environment, so any variable set via the host API (e.g., `SetVar`) is visible.
- If all conditional layers are disabled, the program continues with only the unconditional layers.

### Use Cases

| Pattern | Example |
|---|---|
| Environment flags | `debug if env == "development"` |
| Feature toggles | `payments if beta_enabled` |
| Platform detection | `mobile_ui if platform == "ios"` |
| Threshold tuning | `verbose if log_level > 2` |

## Layer Blend Modes

**Status**: ✅ **Implemented**

By default, when a function is defined in multiple layers, the **top layer wins** (override). Blend modes change this behaviour. They are set on the `@layers` directive and apply to every function call resolved through the layer stack.

```layer
@layers [core, cache, data] blend: fallback
```

### Available Blend Modes

| Mode | Behaviour | Best for |
|---|---|---|
| `override` (default) | Top layer's version wins | Wrapping / decoration patterns |
| `chain` | Result is piped bottom → top through every layer that defines the function | Transformations, middleware pipelines |
| `collect` | Every layer's version is called; results returned as a list (top → bottom) | Gathering contributions from all layers |
| `fallback` | Top layer first; if it returns `None`, try the next layer down | Optional caching, feature flags |

### `override` (default)

Current behaviour — no change needed. The topmost layer that defines a function is the one that runs.

```layer
@layers [core, debug]  # Same as @layers [core, debug] blend: override

@layer core {
    fn process(x) = x + 1
}
@layer debug {
    fn process(x) = core.process(x)  # wraps core
}

process(5)  # → debug.process(5) = core.process(5) = 6
```

### `chain`

Each layer that defines the function is called in turn, bottom → top. The output of one layer becomes the input to the next.

```layer
@layers [core, transform] blend: chain

@layer core {
    fn process(x) = x + 1      # bottom layer: 5 → 6
}
@layer transform {
    fn process(x) = x * 2      # top layer:  6 → 12
}

process(5)  # → core.process(5) = 6 → transform.process(6) = 12
```

Useful for **transformation pipelines** where each layer applies a step.

### `collect`

Every layer that defines the function is called with the same arguments. Results are returned as a list, ordered top → bottom.

```layer
@layers [core, extra] blend: collect

@layer core  { fn tag(x) = "core"  }
@layer extra { fn tag(x) = "extra" }

tag("anything")  # → ["extra", "core"]
```

Useful when multiple layers should all **contribute** to a result — event handlers, validators, collectors.

### `fallback`

Layers are tried top → bottom. The first result that is not `data.None()` is returned. Requires the `data` layer for `Some`/`None`.

```layer
@layers [core, data, cache] blend: fallback

@layer cache {
    fn fetch(key) = data.None()          # cache miss
}
@layer core {
    fn fetch(key) = data.Some("from_db") # database hit
}

fetch("mykey")  # → cache returns None → falls back to core → Some("from_db")
```

When the cache has a value:

```layer
@layer cache {
    fn fetch(key) = data.Some("from_cache")  # cache hit
}

fetch("mykey")  # → Some("from_cache") — core is never called
```

Useful for **layered caching**, **feature flags**, or any **try-first pattern**.

### Blend Mode and Explicit Layer Access

Blend modes only affect calls that go through normal layer resolution. Explicit layer access (e.g., `core.factorial`) always bypasses blend mode and calls that specific layer directly.

```layer
@layers [core, transform] blend: chain

process(5)         # blend mode applies → 12
core.process(5)    # blend mode bypassed → 6
```

### Validation

Unknown blend modes are caught at **parse time**:

```layer
@layers [core] blend: bogus
# Error: unknown blend mode 'bogus': must be override, chain, collect, or fallback
```

## Layer Introspection

**Status**: ✅ **Implemented**

Introspection expressions let you query the live layer stack at runtime. They are first-class expressions — they return values and work anywhere an expression is valid: top-level, inside `let` bindings, inside function bodies.

### `@which name`

Returns the name of the **topmost layer** that currently provides `name`. This is the layer that would actually run if you called `name(...)`.

```layer
@layers [core, debug]
@layer core  { fn factorial(n) = n }
@layer debug { fn factorial(n) = n }

@which factorial   # → "debug"
```

If the name is not defined in any layer, a runtime error is raised.

### `@inspect layer`

Returns a sorted list of all names defined in the named layer (both user-defined functions and built-ins registered there).

```layer
@layers [core]
@layer core {
    fn add(x, y) = x + y
    fn sub(x, y) = x - y
}

@inspect core   # → ["add", "and", "concat", "drop", ..., "sub", ...]
```

If the layer does not exist, a runtime error is raised.

### `@diff layerA layerB`

Returns a record describing what `layerB` contributes relative to `layerA`:

- `added` — names defined in B but **not** in A
- `overridden` — names defined in **both** A and B

```layer
@layers [core, debug]
@layer core  { fn factorial(n) = n }
@layer debug { fn factorial(n) = n
               fn trace(f) = f }

@diff core debug
# → {added: ["trace"], overridden: ["factorial"]}
```

If either layer does not exist, a runtime error is raised. The direction is always A → B ("what does B change on top of A?").

### Using introspection as expressions

Because these are expressions, they compose naturally with the rest of the language:

```layer
let owner = @which factorial in
str_concat("factorial is provided by: ", owner)

let core_names = @inspect core in
length(core_names)

let changes = @diff core debug in
length(changes.overridden)
```

### Error handling

| Situation | Behaviour |
|---|---|
| `@which` on undefined name | Runtime error |
| `@inspect` on unknown layer | Runtime error |
| `@diff` with unknown layer | Runtime error |
| Unknown op (e.g. `@frobnicate`) | Parse error |

## Layer Purity System

**Status**: 🚧 **Planned Feature** - Not yet implemented

Layer provides a **purity enforcement system** that ensures pure layers cannot call impure functions, preventing accidental side effects in business logic.

### Philosophy

The purity system enforces the principle: **"Effects should be explicit and isolated."**

- **Pure layers** contain only pure functional code (no I/O, no mutations, no side effects)
- **Impure layers** can perform effects (file I/O, logging, UI rendering, database access)
- **The compiler prevents** pure layers from calling impure functions

This gives you **compile-time guarantees** that your core business logic is pure.

### Purity Syntax

Layers are marked with **purity annotations**:

```layer
@layer name { ... }               # Pure by default
@layer<pure> name { ... }         # Explicitly pure
@layer<impure> name { ... }       # Explicitly impure
@layer<ui> name { ... }           # Built-in impure (UI effects)
@layer<data> name { ... }         # Built-in impure (I/O effects)
@layer<debug> name { ... }        # Built-in impure (logging effects)
```

**Default behavior**:
- Custom layers are **pure by default** (safest option)
- Built-in layers (`ui`, `data`, `debug`) are **always impure**
- Core layer is **always pure**

### Purity Rules

The compiler enforces these rules:

#### Rule 1: Pure Can Only Call Pure

```layer
@layer<pure> core {
  fn calculate(x, y) = x + y  # ✅ OK - pure arithmetic
}

@layer<pure> validation {
  fn safe_divide(x, y) =
    if y == 0 then
      core.calculate(x, 1)  # ✅ OK - pure calling pure
    else
      x / y
}

@layer<impure> logging {
  fn traced_divide(x, y) =
    let result = validation.safe_divide(x, y) in
    debug.log("Result: {result}")  # ✅ OK - impure can call anything
    result
}

@layer<pure> bad_layer {
  fn broken(x) =
    debug.log("Computing...")  # ❌ COMPILE ERROR!
    # Error: Pure layer 'bad_layer' cannot call impure function 'debug.log'
    # debug.log is in the 'debug' layer which is marked <debug> (impure)
    x * 2
}
```

#### Rule 2: Impure Can Call Anything

```layer
@layer<impure> cache {
  fn cached_compute(key, compute_fn) =
    let cached = data.get(key) in  # ✅ OK - impure calling impure
    match cached with
    | data.Some(value) -> to_float(value)
    | data.None() ->
        let result = compute_fn() in  # ✅ OK - impure calling pure
        data.set(key, to_string(result))
        result
}
```

#### Rule 3: Built-in Layers Have Fixed Purity

Built-in layers (`core`, `ui`, `data`, `debug`) have **predefined purity** and their functions are **implicitly available** - you don't need to define them, they're part of the language runtime.

```layer
# These are ALWAYS impure (cannot be changed)
@layer<ui> ui { ... }       # UI effects (view, text, button, input, container)
@layer<data> data { ... }   # I/O effects (read_file, write_file, read_json, etc.)
@layer<debug> debug { ... } # Logging effects (log, trace, benchmark, etc.)

# Core is ALWAYS pure (cannot be changed)
@layer core { ... }  # Pure functional primitives (map, filter, +, -, etc.)
```

**Important**: You don't write `@layer<data> data { fn read_file(...) = ... }` yourself - these functions are **built into the language**. When you call `data.read_file()`, you're calling the runtime's implementation, which is implicitly part of the `data` layer.

```layer
@layers [core, data, debug]

# You can use built-in functions directly - no need to define them!
let content = data.read_file("input.txt") in
debug.log("File content: {content}")
```

The built-in layers are **always available** once enabled in `@layers [...]`. You can override or wrap their functions in custom layers, but the base implementation is provided by the runtime.

### Complete Example: Pure vs Impure

```layer
@layers [core, validation, cache, logging, ui]

# ✅ Core is always pure
@layer core {
  fn factorial(n) =
    match n with
    | 0 -> 1
    | n -> n * factorial(n - 1)
}

# ✅ Validation is pure (wraps pure with pure logic)
@layer validation {
  fn factorial(n) =
    if n < 0 then
      -1  # Error indicator
    else if n > 20 then
      -1  # Overflow protection
    else
      core.factorial(n)  # ✅ OK - pure calling pure
}

# ✅ Cache is explicitly impure (needs localStorage)
@layer<impure> cache {
  fn factorial(n) =
    let key = "factorial_{n}" in
    let cached = data.get(key) in  # ✅ OK - impure calling impure (data layer)
    match cached with
    | data.Some(value) -> to_float(value)
    | data.None() ->
        let result = validation.factorial(n) in  # ✅ OK - impure calling pure
        data.set(key, to_string(result))
        result
}

# ✅ Logging is explicitly impure (needs debug.log)
@layer<impure> logging {
  fn factorial(n) =
    debug.log("Computing factorial({n})")  # ✅ OK - impure calling impure
    let result = cache.factorial(n) in
    debug.log("Result: {result}")
    result
}

# ✅ UI is built-in impure (renders to screen)
@layer<ui> ui {
  fn app() = view({  # ✅ OK - ui layer is impure
    title: "Factorial Calculator",
    children: [
      input({name: "n", type: "number"}),
      button({label: "Calculate", onclick: "factorial(n)"})
    ]
  })
}
```

### What Pure Layers CANNOT Do

```layer
@layer<pure> business_logic {
  fn process_order(order) =
    # ❌ COMPILE ERROR: Cannot call impure functions

    debug.log("Processing order")
    # Error: Pure layer 'business_logic' cannot call 'debug.log' from impure layer 'debug'

    data.write_json("orders.json", order)
    # Error: Pure layer 'business_logic' cannot call 'data.write_json' from impure layer 'data'

    # ✅ OK - Pure operations only
    let total = order.items.reduce(fn(sum, item) = sum + item.price, 0) in
    if total > 1000 then
      {...order, discount: 0.1}
    else
      order
}
```

### What Impure Layers CAN Do

```layer
@layer<impure> order_processor {
  fn process_order(order) =
    # ✅ OK - Can call pure layers
    let validated = business_logic.validate_order(order) in

    # ✅ OK - Can call impure layers
    debug.log("Processing order {order.id}")
    data.write_json("orders/{order.id}.json", validated)

    # ✅ OK - Can perform effects
    validated
}
```

### Purity Checking Algorithm

When you call a function, the compiler:

1. **Identifies the function's layer**
   ```layer
   debug.log("test")  # Function: log, Layer: debug
   ```

2. **Checks the layer's purity**
   ```layer
   @layer<debug> debug { ... }  # Impure (built-in)
   ```

3. **Identifies the calling layer**
   ```layer
   @layer business_logic {
     fn my_function() = debug.log("test")  # Caller: business_logic
   }
   ```

4. **Enforces the rule**: Pure cannot call impure
   ```
   ❌ ERROR: Pure layer 'business_logic' cannot call impure function 'debug.log'
   ```

### Error Messages

```layer
@layer core {
  fn bad_example() = data.write_file("log.txt", "test")
}

# Compiler output:
# Error: Purity violation in layer 'core'
#   --> myprogram.layer:3:22
#    |
#  3 |   fn bad_example() = data.write_file("log.txt", "test")
#    |                      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
#    |                      Pure layer 'core' cannot call impure function 'data.write_file'
#    |
#    = note: 'data.write_file' is in layer 'data' which is marked <data> (impure)
#    = note: Layer 'core' is pure by default
#    = help: Either mark layer 'core' as <impure> or move this call to an impure layer
```

### Benefits

**Compile-time Safety**:
```layer
@layer core {
  fn calculate(x) =
    let result = x * 2 in
    data.write_file("log.txt", to_string(result))  # ❌ Caught at compile time!
    result
}
```

**Clear Separation**:
```layer
# Pure layers = testable, predictable, reusable
@layer core { ... }
@layer validation { ... }

# Impure layers = effects isolated at edges
@layer<impure> persistence { ... }
@layer<impure> notifications { ... }
```

**Self-Documenting**:
```layer
@layer<impure> api_client {
  # Name says "impure" - you know it has effects
  fn fetch_user(id) = ...
}

@layer business {
  # Pure by default - you know it's safe
  fn calculate_discount(user) = ...
}
```

**Refactoring Confidence**:
```layer
# Moving a function to a pure layer?
# Compiler ensures you didn't accidentally bring side effects with it!

@layer utils {
  fn helper(x) = x * 2  # ✅ Safe - no hidden effects
}
```

### Common Patterns

#### Pattern 1: Pure Core, Impure Edges

```layer
@layers [core, persistence, ui]

@layer core {
  fn calculate_total(items) =
    items.reduce(fn(sum, item) = sum + item.price, 0)
}

@layer<impure> persistence {
  fn save_order(order) =
    let total = core.calculate_total(order.items) in
    data.write_json("orders/{order.id}.json", {...order, total: total})
}

@layer<ui> ui {
  fn checkout_button(order) =
    button({
      label: "Checkout ${core.calculate_total(order.items)}",
      onclick: "persistence.save_order(order)"
    })
}
```

#### Pattern 2: Layered Validation (All Pure)

```layer
@layers [core, syntax_validation, business_validation, ui]

@layer core {
  fn parse_email(email) = ...  # Pure parsing
}

@layer syntax_validation {
  fn validate_email(email) =
    let parsed = core.parse_email(email) in
    if parsed == None then "Invalid format"
    else email
}

@layer business_validation {
  fn validate_email(email) =
    let syntax_check = syntax_validation.validate_email(email) in
    if syntax_check == email then
      if !contains(email, "@company.com") then
        "Must be company email"
      else
        email
    else
      syntax_check
}
```

#### Pattern 3: Caching Wrapper (Pure Core, Impure Cache)

```layer
@layers [core, cache]

@layer core {
  fn expensive_computation(n) =
    # Pure but slow
    range(0, n).reduce(fn(sum, x) = sum + x * x, 0)
}

@layer<impure> cache {
  fn expensive_computation(n) =
    let key = "compute_{n}" in
    let cached = data.get(key) in  # Impure - localStorage access
    match cached with
    | data.Some(value) -> to_float(value)
    | data.None() ->
        let result = core.expensive_computation(n) in  # Call pure
        data.set(key, to_string(result))  # Impure - localStorage write
        result
}
```

### Advanced: Purity Polymorphism (Future)

**Future feature**: Functions that work in both pure and impure contexts:

```layer
# Hypothetical syntax
fn process<P: Purity>(x) =
  match P with
  | Pure -> core.calculate(x)
  | Impure -> debug.log("Processing {x}"); core.calculate(x)

@layer business {
  fn use_it() = process<Pure>(5)  # ✅ OK - pure version
}

@layer<impure> debug_business {
  fn use_it() = process<Impure>(5)  # ✅ OK - impure version
}
```

### Limitations and Edge Cases

#### Limitation 1: Higher-Order Functions

Higher-order functions can accept functions as arguments, which raises the question: what if a pure layer's function receives an impure function as an argument?

```layer
@layer utils {
  fn apply_twice(fn, x) = fn(fn(x))
}

@layer pure_business {
  fn compute(n) =
    # Passes a pure function - no problem
    apply_twice(fn(x) = x * 2, n)  # ✅ OK
}

@layer<impure> logging_business {
  fn compute(n) =
    # Tries to pass an impure function to a pure layer function
    apply_twice(fn(x) = { debug.log(x); x * 2 }, n)  # ❌ ERROR!
    # Error: Cannot pass impure function to pure layer 'utils'
}
```

**Solution: Context-Based Purity Inference**

The purity of a higher-order function call is determined by:
1. The **layer where it's defined** (pure or impure)
2. The **purity of function arguments** passed to it

**Rule**: A pure layer function **cannot accept impure function arguments**, even when called from an impure context.

```layer
# This works - pure layer, pure arguments
@layer utils {
  fn transform(fn, list) = map(fn, list)
}

@layer business {
  fn double_all(data) =
    transform(fn(x) = x * 2, data)  # ✅ Pure fn to pure layer
}

# This fails - pure layer, impure argument
@layer<impure> logging {
  fn log_and_double(data) =
    transform(fn(x) = { debug.log(x); x * 2 }, data)  # ❌ ERROR
    # Cannot pass impure function to utils.transform (pure layer)
}
```

**Alternative: Make the higher-order function impure**

If you need to accept impure functions, define the higher-order function in an impure layer:

```layer
@layer<impure> utils_impure {
  fn transform(fn, list) = map(fn, list)  # Now accepts any function
}

@layer<impure> logging {
  fn log_and_double(data) =
    utils_impure.transform(fn(x) = { debug.log(x); x * 2 }, data)  # ✅ OK
    # Impure layer can accept impure functions
}
```

**Key insight**: The purity of higher-order functions is **statically checkable** at the call site by inspecting:
- The layer's purity (where the HOF is defined)
- The function arguments' purity (whether they call impure operations)

No additional syntax needed - the compiler infers purity from context!

### Implementation Notes

**Compiler Changes Required**:
1. Parse purity annotations (`<pure>`, `<impure>`, `<ui>`, etc.)
2. Track purity for each layer
3. During function call resolution, check caller purity vs callee purity
4. Emit error if pure calls impure
5. Allow impure to call anything

**Type System Integration**:
```layer
# Effect types (future)
fn read_file : String -> IO<Option<String>>
fn calculate : Int -> Int  # No effect = pure

# Purity in function signatures
fn process : <impure> (Int) -> Int
fn validate : <pure> (String) -> Bool
```

### Summary

**Purity enforcement gives you**:
- ✅ Compile-time guarantee that pure layers have no side effects
- ✅ Clear separation between logic (pure) and effects (impure)
- ✅ Self-documenting code (purity visible in layer declarations)
- ✅ Refactoring safety (can't accidentally introduce effects)
- ✅ Testing benefits (pure layers easy to test)

**Syntax**:
```layer
@layer name { ... }              # Pure (default)
@layer<impure> name { ... }      # Explicitly impure
@layer<ui|data|debug> name { ... }  # Built-in impure layers
```

**Rules**:
- Pure can only call pure
- Impure can call anything
- Built-in layers (ui, data, debug) are always impure
- Core is always pure

**Status**: 🚧 Planned for future release

## Type System

**Status**: 🚧 Planned for future release

Layer uses a Hindley-Milner type system with type inference:

```layer
# Type annotations (optional, inferred if omitted)
fn map : (a -> b) -> [a] -> [b]
fn map(f, list) =
  match list with
  | [] -> []
  | [head, ...tail] -> [f(head), ...map(f, tail)]

# Layer-aware types
@layer ui {
  type View = {
    text: String,
    children: [View]
  }
}
```

## Standard Library (Core Layer)

### List Operations
- `map`, `filter`, `reduce`, `fold`
- `head`, `tail`, `length`
- `concat`, `zip`, `take`, `drop`

### Arithmetic
- `+`, `-`, `*`, `/`, `%`, `^`

### Comparison
- `==`, `!=`, `<`, `>`, `<=`, `>=`

### Logic
- `and`, `or`, `not`

### String Operations
- `concat`, `split`, `join`
- `length`, `substring`

## UI Layer

The UI layer provides declarative user interfaces that render as **interactive React components** in the visual editor. UI components are pure data structures that describe the interface - the rendering happens at the edges (browser).

### Purity Model: Pure Functions, Impure Intent

**Important**: The UI layer is marked `@layer<ui>` (impure), but the UI functions themselves (`view`, `text`, `button`, etc.) are **pure functions** that return data structures.

```layer
@layer<ui> ui {
  fn my_ui() = view({...})  # Pure function - returns data!
}

# Calling my_ui() always returns the same data structure
let ui1 = my_ui() in
let ui2 = my_ui() in
ui1 == ui2  # true - same output for same input
```

**Why mark it impure then?**

The UI layer is marked impure because:

1. **Intent**: UI functions are designed to cause visual effects (rendering to screen)
2. **Boundary**: The UI layer represents where pure computation meets impure effects
3. **onclick handlers**: Can execute code from any layer, including impure operations
4. **Conceptual honesty**: Calling a UI function is intended to show something to the user (an effect)

**The side effect happens outside Layer**:

```
Layer Code                     Visual Editor (JavaScript)
──────────────────            ─────────────────────────────
calculator()          →       Returns pure data structure
                      →       ❌ SIDE EFFECT: Renders to DOM
                      →       ❌ SIDE EFFECT: Handles button clicks
                      →       ❌ SIDE EFFECT: Updates screen
```

The UI layer functions are **pure** (referentially transparent), but they describe **effects** that happen at the system boundary. This is similar to Haskell's IO monad - the description is pure, the execution is impure.

**In the purity system**: Pure layers cannot call UI layer functions, even though the UI functions themselves are pure, because they represent the intent to perform effects.

### Core Concept

UI in Layer is a **side effect** - it's isolated in the UI layer while business logic remains pure in the core layer:

```layer
@layers [core, ui]

@layer core {
  # Pure business logic
  fn calculate(x, y, op) =
    match op with
    | "+" -> x + y
    | "-" -> x - y
    | "*" -> x * y
    | "/" -> x / y
    | _ -> 0
}

@layer ui {
  # UI effect - declarative interface
  fn calculator() = view({
    title: "Calculator",
    children: [
      input({name: "x", type: "number", placeholder: "First number"}),
      input({name: "y", type: "number", placeholder: "Second number"}),
      button({label: "+", onclick: "calculate(to_float(x), to_float(y), '+')"}),
      button({label: "-", onclick: "calculate(to_float(x), to_float(y), '-')"})
    ]
  })
}
```

### Available Components

#### `view(props)`
Container component for the entire UI.

**Properties**:
- `title: String` - Title displayed at top
- `children: List` - Child components

**Example**:
```layer
view({
  title: "My App",
  children: [
    text({content: "Welcome!"}),
    button({label: "Start"})
  ]
})
```

#### `text(props)`
Display static or dynamic text.

**Properties**:
- `content: String` - Text to display
- `style: String` - Optional: "h1", "h2", "h3", "body", "caption"

**Example**:
```layer
text({content: "Hello, World!"})
text({content: "Result: {result}", style: "h2"})
```

#### `button(props)`
Interactive button that executes Layer code on click.

**Properties**:
- `label: String` - Button text
- `onclick: String` - Layer expression to execute (as string)

**Example**:
```layer
button({label: "Calculate", onclick: "factorial(to_float(n))"})
button({label: "Reset", onclick: "clear_inputs()"})
```

**Important**: The `onclick` expression is evaluated with access to all input field values by name.

#### `input(props)`
Text input field for user data.

**Properties**:
- `name: String` - **Required** - Variable name for accessing value
- `placeholder: String` - Placeholder text
- `type: String` - Input type (default: "text")
- `value: String` - Initial value
- `width: String` - CSS width (default: "200px")

**Supported types**:
- `"text"` - Default text input
- `"number"` - Numeric input with spinner controls
- `"password"` - Hidden password input
- `"email"` - Email input with validation
- `"date"` - Date picker
- `"time"` - Time picker
- `"color"` - Color picker

**Example**:
```layer
input({name: "username", placeholder: "Enter username"})
input({name: "age", type: "number", placeholder: "Your age"})
input({name: "birthdate", type: "date"})
```

**Automatic Type Conversion**: When you specify `type: "number"`, the visual editor **automatically converts** the input value to a number when executing `onclick` handlers:

```layer
# With type="number" - automatic conversion!
input({name: "n", type: "number", placeholder: "Enter number"})
button({onclick: "factorial(n)"})  # 'n' is already a number!

# With type="text" - remains a string
input({name: "name", type: "text"})
button({onclick: "greet(name)"})  # 'name' is a string
```

**Supported automatic conversions**:
- `type: "number"` → Converts to float (e.g., `"42"` becomes `42.0`)
- `type: "text"` → Remains string (default)
- `type: "email"`, `type: "password"`, etc. → Remain strings

**Benefits**:
- ✅ No manual `to_float()` needed for numeric inputs
- ✅ Better UX (number keyboard on mobile, spinners, validation)
- ✅ Type-safe by default
- ✅ Less verbose code

#### `container(props)`
Layout container for organizing child components.

**Properties**:
- `layout: String` - "horizontal" or "vertical" (default: "vertical")
- `children: List` - Child components

**Example**:
```layer
container({
  layout: "horizontal",
  children: [
    button({label: "Yes"}),
    button({label: "No"}),
    button({label: "Cancel"})
  ]
})
```

#### `render(component)`
Converts a UI component to its string representation (XML format).

**Purpose**: This function is the **bridge between pure UI data and impure output**. It takes a UI component structure and converts it to a string that can be displayed or sent to the visual editor.

**Parameters**:
- `component` - A UI component (returned by `view`, `button`, `text`, etc.)

**Returns**: String (XML representation of the component)

**Example**:
```layer
@layer<ui> ui {
  fn app() = view({
    title: "My App",
    children: [text({content: "Hello!"})]
  })
}

# Main expression - this is where the impure effect happens!
render(app())
```

**Purity Note**:
- The `render()` function itself is **pure** (same input → same output)
- BUT calling it at the **top level** (main expression) causes an **impure effect** (outputs to console/visual editor)
- This is the **boundary** where pure data becomes impure I/O

```layer
# Pure - just returns a string
let ui_string = render(app()) in ui_string

# Impure - outputs to console/visual editor (main expression)
render(app())
```

When `render()` is called as the **main expression** of a program, it outputs the UI string, which is a side effect. This is similar to Haskell's `main :: IO ()` - the entry point where pure code meets the impure world.

### Input State Management

When UI components render, all input fields are tracked in a state object. The `onclick` handler can access inputs by their `name`:

```layer
@layer ui {
  fn app() = view({
    children: [
      input({name: "x", placeholder: "Enter X"}),
      input({name: "y", placeholder: "Enter Y"}),
      # onclick can access 'x' and 'y' by name
      button({onclick: "to_float(x) + to_float(y)"})
    ]
  })
}
```

**How it works**:
1. Visual editor extracts all `input` components and their `name` properties
2. Creates a state object: `{ x: "", y: "" }`
3. Updates state when user types
4. Evaluates `onclick` expression with state values injected as variables

### Complete Example

```layer
@layers [core, ui]

@layer core {
  fn factorial(n) =
    match n with
    | 0 -> 1
    | n -> n * factorial(n - 1)

  fn is_valid_input(s) =
    let num = to_float(s) in
    num >= 0 && num <= 20
}

@layer ui {
  fn app() = view({
    title: "Factorial Calculator",
    children: [
      text({content: "Calculate factorial (0-20)", style: "h2"}),
      input({
        name: "n",
        type: "number",
        placeholder: "Enter a number"
      }),
      container({
        layout: "horizontal",
        children: [
          button({
            label: "Calculate",
            onclick: "if is_valid_input(n) then factorial(to_float(n)) else 'Invalid input'"
          }),
          button({
            label: "Clear",
            onclick: "''"
          })
        ]
      })
    ]
  })
}

app()
```

### How UI Renders

In the **visual editor** (browser):
1. Layer code returns a UI component structure
2. Visual editor parses it into a tree
3. Components render as **actual Material UI React components**:
   - `view` → Material UI Box with styled border
   - `text` → Typography component
   - `button` → Button with click handler
   - `input` → TextField with state binding
   - `container` → Flexbox layout

This means the UI is **fully interactive** - buttons work, inputs capture data, and Layer expressions execute in real-time!

### UI as Pure Data

Despite being interactive, UI components are **pure data structures**:

```layer
# This is just a value - no side effects!
let my_ui = view({
  children: [text({content: "Hello"})]
}) in my_ui

# Can transform UI with pure functions
fn add_border(component) = {
  ...component,
  style: "border: 1px solid black"
}

# Can compose UIs
fn wrap_with_title(title, content) =
  view({
    title: title,
    children: content
  })
```

The **effect** (rendering to screen) only happens when the visual editor receives the UI structure.

### Best Practices

#### 1. Keep Business Logic in Core Layer

```layer
# ✅ GOOD - Pure logic in core
@layer core {
  fn validate_email(email) =
    str_length(email) > 3 && contains(email, "@")
}

@layer ui {
  fn form() = view({
    children: [
      input({name: "email", type: "email"}),
      button({onclick: "if validate_email(email) then 'Valid!' else 'Invalid!'"})
    ]
  })
}

# ❌ BAD - Logic mixed with UI
@layer ui {
  fn form() = view({
    children: [
      input({name: "email"}),
      button({onclick: "if str_length(email) > 3 && contains(email, '@') then 'Valid!' else 'Invalid!'"})
    ]
  })
}
```

#### 2. Use Type Attribute for Automatic Conversion

```layer
# ✅ GOOD - Automatic conversion with type="number"
input({name: "age", type: "number"})
button({onclick: "calculate_retirement(age)"})  # age is a number!

# ⚠️ OK but manual - Works but verbose
input({name: "age"})
button({onclick: "calculate_retirement(to_float(age))"})

# ❌ BAD - Missing type attribute AND conversion
input({name: "age"})
button({onclick: "calculate_retirement(age)"})  # age is a string - will fail!
```

#### 3. Use Type Attribute for Better UX

```layer
# ✅ GOOD - type helps user experience
input({name: "count", type: "number", placeholder: "Quantity"})
input({name: "birthdate", type: "date"})
input({name: "email", type: "email"})

# ❌ OK but not ideal - all text inputs
input({name: "count", placeholder: "Quantity"})
input({name: "birthdate", placeholder: "YYYY-MM-DD"})
input({name: "email", placeholder: "your@email.com"})
```

#### 4. Compose UI Functions

```layer
@layer ui {
  # Reusable components
  fn labeled_input(label, name, type) =
    container({
      children: [
        text({content: label}),
        input({name: name, type: type})
      ]
    })

  # Use them
  fn form() = view({
    children: [
      labeled_input("Name:", "name", "text"),
      labeled_input("Age:", "age", "number"),
      labeled_input("Email:", "email", "email")
    ]
  })
}
```

### Limitations

1. **No nested state** - Input names are flat (no `user.name`, just `name`)
2. **Simple onclick** - Expression evaluated once per click (no reactivity)
3. **No custom styling** - Limited to Material UI defaults
4. **Browser only** - UI layer only works in visual editor, not CLI
5. **Only `number` type converts** - Other types (`date`, `time`) remain strings

### Future Enhancements

Planned improvements:
- **More type conversions** - Auto-convert dates, booleans, etc.
- **Nested state** - Support `{user: {name: "Alice"}}`
- **Reactive expressions** - Auto-update on input change
- **Custom styles** - CSS-in-Layer
- **More components** - Tables, charts, forms, etc.

## Debug Layer

The debug layer adds introspection and logging:

```layer
@layer debug {
  # Wrap any function with logging
  fn trace(name, fn) =
    λ(...args) -> {
      log("{name} called with {args}")
      let result = fn(...args) in {
        log("{name} returned {result}")
        result
      }
    }

  # Inspect value and structure
  fn inspect(value) = ...

  # Breakpoint
  fn breakpoint(message) = ...
}
```

## Data Layer

The data layer provides file I/O, JSON operations, and localStorage functionality using the Option pattern for error handling:

### Option Pattern

The data layer uses an Option type to handle errors functionally:

```layer
# Option type (Some/None)
data.Some(value)    # Successful result
data.None()         # Error or absence of value

# Helper functions
data.is_some(option)           # Check if Some
data.is_none(option)           # Check if None
data.unwrap(option, default)   # Extract value or return default
data.map_option(fn, option)    # Map function over Option
```

### File Operations

```layer
@layer data {
  # Read a file (returns Some(content) or None)
  data.read_file(path)

  # Write to a file (returns Some(content) or None)
  data.write_file(path, content)

  # Check if file exists (returns boolean)
  data.exists(path)

  # Delete a file (returns Some(path) or None)
  data.delete(path)

  # List files in directory (returns Some([files]) or None)
  data.list_files(directory)
}
```

### JSON Operations

```layer
@layer data {
  # Read and parse JSON file
  data.read_json(path)           # Returns Some(data) or None

  # Write data to JSON file
  data.write_json(path, data)    # Returns Some(data) or None

  # Parse JSON string
  data.parse_json(json_string)   # Returns Some(data) or None

  # Convert to JSON string
  data.to_json(data)             # Returns Some(string) or None
}
```

### localStorage Operations

```layer
@layer data {
  # Get value from localStorage
  data.get(key)           # Returns Some(value) or None

  # Set value in localStorage
  data.set(key, value)    # Returns Some(value)

  # Delete key from localStorage
  data.delete_key(key)    # Returns Some(key) or None

  # Get all keys
  data.keys()             # Returns [keys]

  # Clear all storage
  data.clear()            # Returns None
}
```

### Example Usage

```layer
@layers [core, data, debug]

# File I/O with error handling
let file_content = data.read_file("input.txt") in
let result = data.unwrap(file_content, "File not found") in
debug.log(result)

# JSON processing
let users = data.read_json("users.json") in
let processed = data.map_option(
  fn(user_list) -> filter(fn(u) -> u.age >= 18, user_list),
  users
) in
data.write_json("adults.json", data.unwrap(processed, []))

# localStorage
let _ = data.set("username", "alice") in
let username = data.get("username") in
debug.log(data.unwrap(username, "No user"))
```

## Execution Model

1. Parse Layer source file
2. Evaluate `if <expr>` conditions in `@layers`; exclude disabled layers
3. Build layer stack from the remaining enabled layers (core always at bottom)
4. Evaluate each enabled layer's definitions in order
5. Execute main expression with composed layer context

## File Format

Layer source files use `.layer` extension:

```
myprogram.layer
```

## Future Considerations

- Module system for organizing large programs
- Effect system for managing side effects
- Dependent types
- ~~Layer introspection~~ ✅ implemented — see [Layer Introspection](#layer-introspection)
- ~~Conditional layer activation~~ ✅ implemented — see [Conditional Layer Activation](#conditional-layer-activation)
- Flatten operation — merge a layer stack into a single resolved layer
- Layer effect types and purity enforcement (see [Layer Purity System](#layer-purity-system))
- Hot reloading of individual layers
- Layer versioning and compatibility
- Visual layer diff view in the editor
