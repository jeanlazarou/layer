# Layer Language Specification

## Overview

Layer is an experimental purely functional programming language built around the concept of "layer programming". The language allows developers to compose programs from multiple layers, each providing different aspects of functionality.

## Core Concepts

### Layers

A Layer program consists of one or more layers stacked on top of each other. Each layer can:
- Define new functions and values
- Transform or extend definitions from lower layers
- Access and compose with lower layer functionality

**Available Layers:**
- `core` (mandatory) - Base functional programming primitives
- `ui` - User interface definitions and rendering
- `debug` - Debugging utilities, logging, and introspection
- `async` - Asynchronous operations and effects
- `data` - Data persistence and transformation

### Layer Composition

Layers are composed from bottom to top, similar to Photoshop layers:
```
┌─────────────────┐
│     debug       │ ← Top layer (optional)
├─────────────────┤
│       ui        │ ← Middle layer (optional)
├─────────────────┤
│      core       │ ← Base layer (mandatory)
└─────────────────┘
```

Each layer can:
- Override functions from lower layers
- Wrap lower layer functions with additional behavior
- Add new functionality

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

### Layer Declarations

```layer
# Enable layers
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

## Type System

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

The UI layer provides a declarative syntax for building user interfaces:

```layer
@layer ui {
  fn app() =
    view {
      layout: "vertical",
      children: [
        text { content: "Hello, Layer!" },
        button {
          label: "Click me",
          onClick: λ() -> log("Clicked!")
        }
      ]
    }
}
```

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

## Execution Model

1. Parse Layer source file
2. Identify enabled layers
3. Build layer stack (core always at bottom)
4. Evaluate each layer's definitions in order
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
- Layer inheritance and composition operators
- Visual layer editor (Photoshop-like UI)
- Hot reloading of individual layers
- Layer versioning and compatibility
