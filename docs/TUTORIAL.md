# Layer Language Tutorial

Welcome to Layer! This tutorial will guide you through the basics of programming in Layer.

## Installation

```bash
git clone https://github.com/yourusername/layer.git
cd layer
pip install -r requirements.txt
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
python -m layer.cli run hello.layer
```

Output: `Hello, World!`

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
python -m layer.repl
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
4. **Explicit is better**: Use `core.function()` when you want to bypass upper layers
5. **Think in compositions**: Build complex behavior by stacking simple layers

Happy coding with Layer! 🎨
