# Layer Programming Language

**A purely functional language with practical effect management through spatial layer composition**

Layer is a functional programming language that makes the distinction between pure and impure code explicit through a spatial organization system. Like Photoshop's layer system for images, Layer allows you to stack different aspects of your program (business logic, UI, debugging, I/O) in visually separate layers.

## Philosophy

> **Maximize purity where it matters, isolate effects where they're necessary.**

Layer doesn't try to eliminate side effects (impossible!), but provides **spatial separation** through layers that makes effects explicit while keeping business logic pure.

See [PHILOSOPHY.md](PHILOSOPHY.md) for the complete rationale.

## Key Features

- **Purely Functional Core**: Immutable data, first-class functions, no side effects
- **Layer-Based Architecture**: Spatial separation of pure and impure code
- **Expression-Based**: Everything is an expression, no statements
- **Pattern Matching**: Expressive destructuring and control flow
- **Layer Override**: Higher layers can wrap or transform lower layer functions
- **WebAssembly**: Runs in browser without server (4.3 MB)
- **Interactive UI**: UI layer renders as actual React components

## Quick Example

```layer
@layers [core, ui, debug]

@layer core {
  # Pure business logic
  fn fibonacci(n) =
    match n with
    | 0 -> 0
    | 1 -> 1
    | n -> fibonacci(n - 1) + fibonacci(n - 2)
}

@layer debug {
  # Wrap pure function with logging (effect)
  fn fibonacci(n) =
    let result = core.fibonacci(n) in
    debug.log("fib({n}) = {result}")
    result
}

@layer ui {
  # Interactive UI (effect)
  fn app() = view({
    title: "Fibonacci Calculator",
    children: [
      input({name: "n", placeholder: "Enter number"}),
      button({label: "Calculate", onclick: "fibonacci(n)"})
    ]
  })
}
```

**Result**: 90% pure code (core layer), 10% effects (ui/debug layers) ✅

## Available Layers

- **core** - Pure functional primitives (36 functions)
- **ui** - User interface components (renders as React)
- **debug** - Logging, tracing, benchmarking (8 functions)
- **data** - File I/O, JSON, localStorage (21 functions)

## Installation & Quick Start

### Browser (Visual Editor - No Installation!)

The visual editor runs entirely in your browser using WebAssembly:

```bash
cd visual-editor/frontend
npm install
npm run dev
```

Open http://localhost:5173 - **No backend needed!**

### Native CLI

```bash
cd layer-go
go build ./cmd/layer
./layer ../examples/calculator.layer
```

## Implementation

Layer is implemented in **Go** with compilation to both:
- **Native binaries** - For CLI and server use
- **WebAssembly** - For browser execution

**Previous Python implementation is deprecated** - Go is the primary and only maintained version.

### Why Go?

1. **Performance**: 10x faster than Python
2. **WebAssembly**: Runs in browser without backend
3. **Type Safety**: Easier to maintain and extend
4. **Deployment**: Single binary, no dependencies
5. **Portability**: Same code runs CLI, server, and browser

## Running Layer Programs

### Browser (Recommended)

```bash
cd visual-editor/frontend
npm run dev
```

- Open example files (calculator, data_demo, debug_trace)
- Write code with Monaco editor
- See interactive UI rendered in real-time
- Files stored in localStorage (virtual mode) or real file system (Chrome/Edge)

### Native CLI

```bash
cd layer-go
go build ./cmd/layer

# Run examples
./layer ../examples/calculator.layer
./layer ../examples/fibonacci.layer
./layer ../examples/data_demo.layer
```

### REPL (Coming Soon)

```bash
./layer
> let x = 5 in x * 2
10
```

## Project Structure

```
layer/
├── PHILOSOPHY.md              # Language design philosophy
├── LANGUAGE_SPEC.md           # Complete language specification
├── examples/                  # Example Layer programs
│   ├── calculator.layer       # Interactive calculator
│   ├── fibonacci.layer        # Recursion example
│   ├── data_demo.layer        # File I/O and JSON
│   └── debug_trace.layer      # Tracing and debugging
├── layer-go/                  # Go implementation (PRIMARY)
│   ├── cmd/layer/             # Native CLI
│   ├── cmd/wasm/              # WebAssembly entry point
│   ├── pkg/
│   │   ├── lexer/             # Tokenization
│   │   ├── parser/            # Parsing
│   │   ├── interpreter/       # Evaluation + layers
│   │   └── layers/            # Layer system
│   ├── build-wasm.sh          # Build WASM
│   ├── layer.wasm             # Compiled WASM (4.3 MB)
│   ├── README.md              # Go implementation details
│   ├── CHANGELOG.md           # Implementation history
│   └── TESTING_GUIDE.md       # How to test everything
├── visual-editor/             # Visual editor (React + Vite)
│   └── frontend/
│       ├── src/
│       │   ├── components/    # Monaco, UI preview, etc.
│       │   └── wasm/          # WASM loader + browser FS
│       └── public/
│           └── layer.wasm     # Deployed WASM
└── docs/                      # Additional documentation
    ├── ARCHITECTURE.md
    └── TUTORIAL.md
```

## Examples

All examples work in both browser and CLI:

### Calculator (`examples/calculator.layer`)
Interactive calculator with buttons and display:
- String concatenation for display
- Pattern matching for operations
- Renders as actual Material UI components

### Data Demo (`examples/data_demo.layer`)
File I/O, JSON, and localStorage:
- Read/write files (virtual or real mode)
- Parse and generate JSON
- localStorage key-value storage

### Debug Trace (`examples/debug_trace.layer`)
Function tracing and debugging:
- Trace recursive execution
- Benchmark performance
- Inspect values

### Fibonacci (`examples/fibonacci.layer`)
Classic recursion example:
- Pattern matching
- Pure functional style
- Debug layer wrapping

## Language Highlights

### Expression-Based

Everything is an expression:

```layer
# if is an expression
let result = if x > 0 then "positive" else "negative" in result

# match is an expression
let category = match age with
  | a when a < 18 -> "minor"
  | a when a < 65 -> "adult"
  | _ -> "senior"
in category
```

### Immutability

All bindings are immutable:

```layer
let x = 5 in
let x = 10 in  # New binding, shadows previous
x              # Returns 10
```

### let ... in Syntax

Explicit scoping inspired by ML-family languages:

```layer
let a = 1 in
let b = 2 in
let c = a + b in
c * 2  # Returns 6
```

**Why `in`?**
- No semicolons needed
- Whitespace-insensitive
- Unambiguous grammar
- Clear scope boundaries

### Pattern Matching

```layer
fn describe(list) =
  match list with
  | [] -> "empty"
  | [x] -> "one element"
  | [x, y] -> "two elements"
  | x::xs -> "many elements"
```

### Layer Override

```layer
@layer core {
  fn greet(name) = "Hello, {name}"
}

@layer enthusiastic {
  fn greet(name) = core.greet(name) ++ "!!!"
}

greet("World")  # "Hello, World!!!"
```

## Documentation

- **[PHILOSOPHY.md](PHILOSOPHY.md)** - Why Layer is designed this way
- **[LANGUAGE_SPEC.md](LANGUAGE_SPEC.md)** - Complete syntax and semantics
- **[layer-go/README.md](layer-go/README.md)** - Implementation details
- **[layer-go/CHANGELOG.md](layer-go/CHANGELOG.md)** - Implementation history
- **[layer-go/TESTING_GUIDE.md](layer-go/TESTING_GUIDE.md)** - How to test
- **[visual-editor/README.md](visual-editor/README.md)** - Visual editor guide
- **[docs/TUTORIAL.md](docs/TUTORIAL.md)** - Learn by example

## Development

### Running Tests

```bash
cd layer-go
go test ./pkg/...              # All tests
go test ./pkg/interpreter -v   # Interpreter tests
```

**Current status**: 22/22 tests passing ✅

### Building WASM

```bash
cd layer-go
./build-wasm.sh
# Outputs: layer.wasm (4.3 MB)
# Auto-deploys to ../visual-editor/frontend/public/
```

### Adding Built-in Functions

1. Choose layer (core/data/debug/ui)
2. Add to `pkg/interpreter/[layer]_layer.go`
3. Add tests to `pkg/interpreter/interpreter_test.go`
4. Update LANGUAGE_SPEC.md
5. Rebuild WASM if needed

## Performance

- **Native**: ~10x faster than Python implementation
- **WASM**: Near-native performance
- **Load time**: < 1 second (4.3 MB gzipped)
- **File I/O** (virtual): < 10ms per operation
- **File I/O** (real): < 100ms per operation

## Roadmap

### Short Term
- [ ] REPL implementation
- [ ] Better error messages (line/column numbers)
- [ ] Performance optimizations

### Medium Term
- [ ] Module system
- [ ] Type inference (Hindley-Milner)
- [ ] LSP server for editor support

### Long Term
- [ ] Effect system (algebraic effects)
- [ ] Native code compilation (LLVM)
- [ ] Package manager

## Contributing

Contributions welcome! The codebase is well-tested and documented.

**Getting started**:
1. Read [PHILOSOPHY.md](PHILOSOPHY.md) to understand the design
2. Check [LANGUAGE_SPEC.md](LANGUAGE_SPEC.md) for syntax
3. Look at [examples/](examples/) for usage patterns
4. Run tests: `cd layer-go && go test ./pkg/...`

**Areas for contribution**:
- Performance optimizations
- Better error messages
- REPL implementation
- Module system design
- Documentation improvements
- More examples

## License

MIT License - See LICENSE file

## Acknowledgments

Layer draws inspiration from:
- **Haskell**: `let ... in` syntax, purity philosophy
- **OCaml/F#**: ML-family syntax and pattern matching
- **Elm**: Managed effects and explicit architecture
- **Photoshop**: Layer metaphor for composition

---

**Layer**: Pure functions, practical effects, spatial organization.

**Try it now**: `cd visual-editor/frontend && npm run dev`
