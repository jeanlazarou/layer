# Layer Language Architecture

## Overview

Layer is implemented as a tree-walk interpreter written in Go and compiled to WebAssembly. The architecture follows a traditional compiler pipeline with additional support for layer composition.

## Components

### 1. Lexer (`layer-go/pkg/lexer`)

The lexer performs tokenization of Layer source code.

**Responsibilities:**
- Convert source text into tokens
- Handle comments
- Recognize keywords, operators, and literals
- Track line and column numbers for error reporting

**Key Features:**
- Support for Unicode lambda symbol (λ)
- String interpolation preparation
- Multi-character operators (→, ==, !=, etc.)

### 2. Parser (`layer-go/pkg/parser`)

The parser builds an Abstract Syntax Tree (AST) from tokens.

**Responsibilities:**
- Recursive descent parsing
- Operator precedence handling
- Pattern matching syntax
- Layer directive parsing

**Key Features:**
- Precedence climbing for binary operators
- Pattern matching with destructuring
- Record and list literals
- Function definitions with parameters

### 3. AST Nodes (`layer-go/pkg/ast`)

Defines all AST node types using Go structs.

**Node Categories:**
- Literals: Numbers, strings, booleans, lists, records, tuples
- Variables: Identifiers, member access
- Functions: Lambda, function definitions, function calls
- Control flow: Let bindings, pattern matching
- Operators: Binary and unary operations
- Layers: Layer declarations and directives

### 4. Layer System (`layer-go/pkg/layers`)

Manages layer composition and function resolution.

**LayerContext** (`layer-go/pkg/layers/layer_context.go`):
- Maintains layer stack (bottom to top)
- Resolves function lookups across layers
- Supports layer-specific lookups (e.g., `core.factorial`)
- Enables layer override behavior

**Built-in Layers:**

**Core Layer** (`layer-go/pkg/interpreter/core_layer.go`):
- Always enabled
- Provides fundamental operations
- Arithmetic, comparison, logical operators
- List operations: map, filter, reduce, etc.
- String operations

**Debug Layer** (`layer-go/pkg/interpreter/debug_layer.go`):
- Logging and tracing
- Function call inspection
- Benchmarking
- Assertions

**UI Layer** (`layer-go/pkg/interpreter/ui_layer.go`):
- Declarative UI components
- Component tree building
- Rendering abstraction
- XML output for visual editor

**Data Layer** (`layer-go/pkg/interpreter/data_layer_wasm.go`):
- File I/O (via browser localStorage)
- JSON operations
- Option pattern for error handling

### 5. Interpreter (`layer-go/pkg/interpreter/interpreter.go`)

Evaluates AST nodes in the context of layers.

**Responsibilities:**
- AST evaluation
- Environment management
- Pattern matching implementation
- Closure creation and execution
- Layer composition resolution

**Key Features:**
- Lexical scoping with closures
- Pattern matching with destructuring
- Layer-aware function lookup
- Recursive evaluation

### 6. CLI Tools

**CLI** (`layer-go/cmd/layer/main.go`):
- File execution
- REPL mode
- Error handling and reporting

**WASM Interface** (`layer-go/cmd/wasm/main.go`):
- WebAssembly exports for browser
- JavaScript interop
- Browser file system integration

## Execution Flow

```
Source Code
    ↓
[Lexer]
    ↓
Tokens
    ↓
[Parser]
    ↓
AST (Program node)
    ↓
[Interpreter]
    ↓
1. Parse @layers directive
2. Enable specified layers
3. Evaluate @layer declarations
4. Execute main expression
    ↓
Result
```

## Layer Composition

Layers are stacked from bottom to top:

```
┌─────────────────┐
│   Debug Layer   │ ← Can wrap core functions
├─────────────────┤
│   Data Layer    │ ← File I/O, JSON, localStorage
├─────────────────┤
│   UI Layer      │ ← Can use core functions
├─────────────────┤
│   Core Layer    │ ← Base functionality (always enabled)
└─────────────────┘
```

**Function Resolution:**
1. Check local environment (let bindings, parameters)
2. Search layers from top to bottom
3. Return first match found

**Layer Override:**
```layer
@layer core {
  fn factorial(n) = ...  # Original definition
}

@layer debug {
  fn factorial(n) =
    let _ = log("Called") in
      core.factorial(n)  # Explicitly call core version
}
```

## Type System (Future)

The language specification includes a Hindley-Milner type system, but type inference is not yet implemented. Current implementation uses dynamic typing.

## Pattern Matching

Pattern matching is implemented via the `matchPattern` method in the interpreter:

**Supported Patterns:**
- Literal patterns: `0`, `"hello"`
- Variable binding: `n`, `x`
- Wildcard: `_`
- List patterns: `[a, b, ...rest]`

**Future Patterns:**
- Record patterns: `{name: n, age: a}`
- Tuple patterns: `(x, y, z)`
- Type patterns

## Closures

Functions in Layer are closures that capture their environment:

```go
type Closure struct {
    Params      []string           // Parameter names
    Body        ast.Node           // Function body AST
    Env         map[string]interface{} // Captured environment
    Interpreter *Interpreter       // Reference to interpreter
}
```

When a closure is called:
1. Create new environment from captured environment
2. Bind arguments to parameters
3. Evaluate body in new environment

## Error Handling

All errors include position information (line, column):
- **Lexer errors**: Invalid characters, unterminated strings
- **Parser errors**: Unexpected tokens, syntax errors
- **Runtime errors**: Type errors, undefined variables, pattern match failures

## WebAssembly Architecture

The Go interpreter is compiled to WebAssembly for browser execution:

**Compilation:**
```bash
cd layer-go
GOOS=js GOARCH=wasm go build -o layer.wasm ./cmd/wasm
```

**Browser Integration:**
- WASM module loaded in React frontend
- JavaScript interop via `syscall/js` package
- File I/O through browser localStorage
- Synchronous operations to avoid WASM event loop deadlocks

**Performance:**
- Near-native execution speed
- ~3MB WASM binary
- No network latency (runs locally)
- Offline-capable

## Future Enhancements

1. **Purity Enforcement**: `@layer<pure>` vs `@layer<impure>` markers (see LANGUAGE_SPEC.md)
2. **Type Inference**: Hindley-Milner type checking
3. **Module System**: Import/export between files
4. **Effect System**: Track and manage side effects
5. **Optimization**: Bytecode compilation, JIT
6. **Visual Editor Enhancements**: Photoshop-like layer UI
7. **Hot Reloading**: Live layer updates
8. **Debugger**: Step-through execution, breakpoints
9. **Package Manager**: Layer package ecosystem
