# Layer Programming Language

**An experimental purely functional language exploring layer-based composition**

Layer is a functional programming language built around the concept of composable layers. Like Photoshop's layer system for images, Layer allows programmers to stack different aspects of their programs (UI, debugging, async operations) on top of a core functional foundation.

## Key Features

- **Purely Functional**: Immutable data, first-class functions, no side effects
- **Layer-Based Architecture**: Compose programs from stackable, optional layers
- **Visual Composition**: UI inspired by Photoshop's layer paradigm
- **Type Inference**: Hindley-Milner type system
- **Pattern Matching**: Expressive destructuring and matching
- **Layer Override**: Higher layers can wrap or transform lower layer functions

## Quick Example

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
  fn fibonacci(n) = {
    log("Computing fib({n})")
    core.fibonacci(n)
  }
}

@layer ui {
  fn app() =
    view {
      title: "Fibonacci Calculator",
      content: text { value: "fib(10) = {fibonacci(10)}" }
    }
}
```

## Available Layers

- **core** (mandatory) - Functional programming primitives
- **ui** - User interface definitions
- **debug** - Logging and introspection
- **async** - Asynchronous operations
- **data** - Data persistence and transformation

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/layer.git
cd layer

# Install dependencies
pip install -r requirements.txt

# Run the interpreter
python -m layer.repl
```

## Running Layer Programs

```bash
# Execute a .layer file
python -m layer.cli run myprogram.layer

# Interactive REPL
python -m layer.repl

# Visual layer editor (future)
python -m layer.visual
```

## Project Structure

```
layer/
├── README.md
├── LANGUAGE_SPEC.md      # Complete language specification
├── requirements.txt
├── layer/
│   ├── __init__.py
│   ├── lexer.py          # Tokenization
│   ├── parser.py         # AST generation
│   ├── ast_nodes.py      # AST node definitions
│   ├── types.py          # Type system
│   ├── interpreter.py    # Evaluation engine
│   ├── layers/
│   │   ├── __init__.py
│   │   ├── core.py       # Core layer implementation
│   │   ├── ui.py         # UI layer implementation
│   │   └── debug.py      # Debug layer implementation
│   ├── repl.py           # Interactive REPL
│   ├── cli.py            # Command-line interface
│   └── visual/           # Visual layer editor (future)
├── examples/             # Example programs
│   ├── hello.layer
│   ├── fibonacci.layer
│   └── calculator.layer
└── tests/
    ├── test_lexer.py
    ├── test_parser.py
    └── test_interpreter.py
```

## Philosophy

Layer explores the idea that different concerns in programming can be cleanly separated into composable layers:

1. **Core Logic**: Pure functional transformations (mandatory)
2. **UI**: How to present and interact (optional)
3. **Debug**: How to observe and introspect (optional)
4. **Effects**: How to handle I/O and state (optional)

Each layer can:
- Add new functionality
- Override lower layer functions
- Wrap lower layer functions with additional behavior
- Access all lower layers explicitly

## Development Status

Layer is an experimental language in active development. Current status:

- [x] Language specification
- [x] Lexer/Tokenizer
- [x] Parser and AST
- [x] Basic interpreter
- [x] Core layer implementation
- [ ] UI layer (in progress)
- [ ] Debug layer (in progress)
- [ ] Type inference engine
- [ ] Visual layer editor
- [ ] Standard library expansion

## Contributing

This is an experimental project exploring novel programming language concepts. Contributions, ideas, and discussions are welcome!

## License

MIT License - See LICENSE file for details

## Learn More

- [Language Specification](LANGUAGE_SPEC.md)
- [Examples](examples/)
- [Architecture Documentation](docs/ARCHITECTURE.md)

## Inspiration

- Photoshop's layer composition model
- Functional programming (Haskell, OCaml, F#)
- Context-oriented programming
- Aspect-oriented programming
- Feature-oriented programming
