# Layer Language Architecture

## Overview

Layer is implemented as a tree-walk interpreter written in Python. The architecture follows a traditional compiler pipeline with additional support for layer composition.

## Components

### 1. Lexer (`layer/lexer.py`)

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

### 2. Parser (`layer/parser.py`)

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

### 3. AST Nodes (`layer/ast_nodes.py`)

Defines all AST node types using Python dataclasses.

**Node Categories:**
- Literals: Numbers, strings, booleans, lists, records, tuples
- Variables: Identifiers, member access
- Functions: Lambda, function definitions, function calls
- Control flow: Let bindings, pattern matching
- Operators: Binary and unary operations
- Layers: Layer declarations and directives

### 4. Layer System (`layer/layers/`)

Manages layer composition and function resolution.

**LayerContext** (`layer/layers/__init__.py`):
- Maintains layer stack (bottom to top)
- Resolves function lookups across layers
- Supports layer-specific lookups (e.g., `core.factorial`)
- Enables layer override behavior

**Built-in Layers:**

**Core Layer** (`layer/layers/core.py`):
- Always enabled
- Provides fundamental operations
- Arithmetic, comparison, logical operators
- List operations: map, filter, reduce, etc.
- String operations

**Debug Layer** (`layer/layers/debug.py`):
- Logging and tracing
- Function call inspection
- Benchmarking
- Assertions

**UI Layer** (`layer/layers/ui.py`):
- Declarative UI components
- Component tree building
- Rendering abstraction

### 5. Interpreter (`layer/interpreter.py`)

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

**CLI** (`layer/cli.py`):
- File execution
- Error handling and reporting

**REPL** (`layer/repl.py`):
- Interactive evaluation
- Command support (help, layers, clear)
- Error recovery

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
│   UI Layer      │ ← Can use core functions
├─────────────────┤
│   Core Layer    │ ← Base functionality
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

Pattern matching is implemented via the `match_pattern` method in the interpreter:

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

```python
class Closure:
    params: List[str]      # Parameter names
    body: ASTNode          # Function body AST
    env: Dict[str, Any]    # Captured environment
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

## Future Enhancements

1. **Type Inference**: Hindley-Milner type checking
2. **Module System**: Import/export between files
3. **Effect System**: Track and manage side effects
4. **Optimization**: Bytecode compilation, JIT
5. **Visual Editor**: Photoshop-like layer UI
6. **Hot Reloading**: Live layer updates
7. **Debugger**: Step-through execution, breakpoints
8. **Package Manager**: Layer package ecosystem
