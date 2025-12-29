# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Layer is an experimental purely functional programming language built around **layer-based composition**. Like Photoshop's layer system for images, Layer allows programmers to stack different aspects of their programs (UI, debugging, async operations, data persistence) on top of a core functional foundation.

The project consists of:
- **Go-based interpreter**: Tree-walk interpreter with lexer, parser, and evaluator compiled to WebAssembly
- **Visual Editor**: React + Material UI interface with Monaco editor and **live interactive UI rendering**
- **Layer system**: Composable layers (core, ui, debug, data) that can override/wrap each other

**Key Innovation**: The UI layer outputs declarative UI components that are **actually rendered as interactive React components** in the visual editor. Write functional UI code in Layer, see it come alive instantly with buttons, inputs, and real interactivity!

## Development Commands

### Go Interpreter (Core Language)

```bash
# Build the interpreter
cd layer-go
go build -o layer ./cmd/layer

# Run the REPL
./layer

# Execute a .layer file
./layer run ../examples/fibonacci.layer

# Run tests
go test ./...                   # All tests
go test ./pkg/interpreter       # Specific package
go test -v ./...                # Verbose output
go test -cover ./...            # With coverage report

# Build WebAssembly module
cd layer-go
./build-wasm.sh                 # Outputs layer.wasm
```

### Visual Editor (Frontend Only - No Backend Needed!)

The visual editor runs entirely in the browser using WebAssembly:

```bash
# Frontend (standalone - WASM interpreter embedded)
cd visual-editor/frontend
npm run dev                     # Runs on http://localhost:5173 (Vite dev server)
npm run build                   # Production build
npm run preview                 # Preview production build
npm run lint                    # Lint code
```

## Architecture

### Language Implementation (Go)

The interpreter follows a traditional pipeline with layer-aware extensions:

```
Source Code (.layer file)
    ↓
Lexer (layer-go/pkg/lexer) → Tokens
    ↓
Parser (layer-go/pkg/parser) → AST (layer-go/pkg/ast)
    ↓
Interpreter (layer-go/pkg/interpreter) + Layer System (layer-go/pkg/layers)
    ↓
Result
```

**Key Components:**

- **`layer-go/pkg/lexer`**: Tokenization with support for λ symbol, multi-character operators (→, ==, etc.)
- **`layer-go/pkg/parser`**: Recursive descent parser with precedence climbing for operators
- **`layer-go/pkg/ast`**: AST node definitions using Go structs
- **`layer-go/pkg/interpreter`**: Evaluates AST with closures, pattern matching, and layer-aware lookups
- **`layer-go/pkg/layers`**: Layer implementations (core, ui, debug, data)
  - **`layer_context.go`**: LayerContext manages layer stack and function resolution
  - Layer resolution: searches from top to bottom, first match wins
  - Supports explicit layer access (e.g., `core.factorial`)

**Layer Composition Model:**

Layers stack from bottom to top. Higher layers can override or wrap lower layer functions:

```
┌─────────────────┐
│   Debug Layer   │ ← Can wrap core.fibonacci with logging
├─────────────────┤
│   Data Layer    │ ← File I/O, JSON, localStorage
├─────────────────┤
│   UI Layer      │ ← Declarative UI components
├─────────────────┤
│   Core Layer    │ ← Pure functional primitives (mandatory)
└─────────────────┘
```

### Visual Editor Architecture

```
visual-editor/
└── frontend/
    └── src/
        ├── App.tsx         # Main React application with WASM interpreter
        ├── components/
        │   ├── LayerStack.tsx    # Layer panel (left sidebar)
        │   ├── CodeEditor.tsx    # Monaco editor integration
        │   ├── OutputConsole.tsx # Terminal-style output with UI preview
        │   └── UIPreview.tsx     # Renders Layer UI components as live React UI
        └── wasm/
            ├── layer.wasm        # Go interpreter compiled to WebAssembly
            ├── wasm_exec.js      # Go WASM runtime
            └── browserFileSystem.ts  # Browser storage for file I/O
```

**Tech Stack:**
- Frontend: React 18, Vite, TypeScript, Material UI 5, Monaco Editor
- Interpreter: Go compiled to WebAssembly (embedded in browser)
- Storage: localStorage for file I/O and data persistence

**Interactive UI Rendering:**

The UI layer is not just conceptual - it's fully functional! When Layer code outputs UI components, they are rendered as actual interactive Material UI components:

```layer
@layer ui {
  fn calculator() = view({
    title: "Calculator",
    children: [
      input({name: "x", placeholder: "Enter number"}),
      button({label: "Square", onclick: "x * x"}),
      text({content: "Click to see result"})
    ]
  })
}
```

This renders as:
- Real Material UI TextField for the input
- Real Material UI Button that executes `x * x` when clicked
- Real Typography component for text
- State management handles input values and result display

**Component Mapping** (`visual-editor/frontend/src/components/UIPreview.tsx`):
- `view` → Material UI Box with styled border
- `text` → Typography component
- `button` → Button with onclick handler that executes Layer expressions
- `input` → TextField with state binding
- `container` → Flexbox layout (horizontal/vertical)

When a button is clicked, the `onclick` expression is sent back to the interpreter with current input values, executed as Layer code, and the result updates the UI - creating a **true reactive programming environment**.

### Layer System Deep Dive

The layer system is the core innovation of the language:

1. **Function Resolution**: When calling `fibonacci(5)`, the interpreter searches layers from top to bottom
2. **Explicit Layer Access**: `core.fibonacci(5)` bypasses layer override and calls the core version directly
3. **Wrapping Pattern**: Higher layers can wrap lower layers:
   ```layer
   @layer debug {
     fn fibonacci(n) = {
       log("Computing fib({n})")
       core.fibonacci(n)  # Calls original implementation
     }
   }
   ```

4. **Layer Override**: If a layer defines a function with the same name, it overrides lower layers

### Data Layer Implementation

The data layer (`layer-go/pkg/interpreter/data_layer_wasm.go`) provides side effects through an **Option pattern** (functional error handling):

- All I/O operations return `data.Some(value)` on success or `data.None()` on failure
- Operations: file I/O, JSON parsing, localStorage
- Helper functions: `is_some()`, `is_none()`, `unwrap()`, `map_option()`

## Important Patterns and Conventions

### When Working with the Go Interpreter

1. **Build before testing**: Run `go build` after making changes to ensure code compiles
2. **Parser changes require AST node updates**: If adding new syntax, update both `parser/` and `ast/` packages
3. **Layer functions** are registered in `layer-go/pkg/interpreter/*_layer.go` files (e.g., `core_layer.go`, `ui_layer.go`)
4. **Pattern matching** is implemented in `interpreter.go`'s `matchPattern()` method
5. **Closures capture environment**: Functions in Layer are closures (see `Closure` struct in `interpreter.go`)

### When Working with the Visual Editor

1. **No backend needed**: Visual editor runs standalone with WASM interpreter embedded
2. **Frontend uses Vite**: Much faster than Create React App, uses ES modules
3. **Monaco editor integration**: Loaded via `@monaco-editor/react` package
4. **UI Preview is fully interactive**: The `UIPreview.tsx` component parses XML output from Layer's UI layer and renders actual Material UI components
5. **Layer colors are semantic**:
   - Core: Green (pure functional)
   - UI: Blue (interface)
   - Debug: Orange (logging/tracing)
   - Data: Purple (I/O operations)
   - Main: Purple (entry point)

### Testing Interactive UI

To test the interactive UI rendering:

1. Write a Layer program with a UI layer that returns components
2. Execute the program
3. The output console will show both text output and a live UI preview
4. Interact with rendered buttons and inputs - they execute Layer code in real-time

Example interactive calculator:
```layer
@layers [core, ui]

@layer core {
  fn square(x) = x * x
  fn double(x) = x + x
}

@layer ui {
  fn app() = view({
    title: "Math Tools",
    children: [
      input({name: "num", placeholder: "Enter a number"}),
      button({label: "Square", onclick: "square(num)"}),
      button({label: "Double", onclick: "double(num)"})
    ]
  })
}

app()
```

### File Locations

- **Example programs**: `examples/*.layer`
- **Language spec**: `LANGUAGE_SPEC.md` (complete syntax and semantics reference)
- **Architecture docs**: `docs/ARCHITECTURE.md`
- **Visual editor docs**: `visual-editor/README.md`, `visual-editor/QUICKSTART.md`

## Testing

Tests use Go's testing framework and are located in `layer-go/`:
- `pkg/lexer/*_test.go`: Tokenization tests
- `pkg/parser/*_test.go`: Parsing and AST generation tests
- `pkg/interpreter/*_test.go`: Evaluation and runtime tests

Run all tests:
```bash
cd layer-go
go test ./...
```

**Note**: Type inference is specified but not yet implemented (current implementation uses dynamic typing).

## Common Workflows

### Adding a New Built-in Function to a Layer

1. Open the appropriate layer file (e.g., `layer-go/pkg/interpreter/core_layer.go`)
2. Add the function using `lc.Define()` in the `Setup*Layer()` function
3. Update tests in `layer-go/pkg/interpreter/*_test.go`
4. Document in `LANGUAGE_SPEC.md` if it's a public API
5. Rebuild WASM if needed: `cd layer-go && ./build-wasm.sh`

### Adding a New Layer

1. Create new file in `layer-go/pkg/interpreter/` (e.g., `async_layer.go`)
2. Define `SetupAsyncLayer(lc *layers.LayerContext)` function
3. Register layer functions using `lc.Define()`
4. Call the setup function in `interpreter.go`
5. Update `LANGUAGE_SPEC.md` with layer documentation
6. Add layer color mapping in visual editor (`visual-editor/frontend/src/components/LayerStack.tsx`)

### Debugging Layer Programs

Use the debug layer to trace execution:
```layer
@layers [core, debug]

@layer debug {
  fn fibonacci(n) = {
    log("fib({n})")
    core.fibonacci(n)
  }
}
```

Or use Go's debugging tools:
```bash
# Run with verbose output
cd layer-go
./layer run --verbose ../examples/myprogram.layer

# Use Delve debugger
dlv debug ./cmd/layer -- run ../examples/myprogram.layer
```

## Project Status

This is an experimental language exploring novel programming concepts. Current implementation status:
- ✅ Lexer, parser, interpreter (Go implementation)
- ✅ Core, UI, Debug, Data layers
- ✅ Visual editor with Monaco integration
- ✅ **Interactive UI rendering** - UI layer components render as live, interactive React components
- ✅ **WebAssembly compilation** - Go implementation compiled to WASM for standalone deployment!
- ❌ Type inference (Hindley-Milner specified but not implemented)
- ❌ Module system
- ❌ Effect system
- ❌ Purity enforcement system (planned - see LANGUAGE_SPEC.md)

## ✅ WebAssembly Architecture (COMPLETED!)

**Status**: The Layer interpreter is implemented in Go and compiled to WebAssembly!

**What's Complete**:
- ✅ Full Go implementation in `layer-go/` (lexer, parser, interpreter, layer system)
- ✅ All tests passing
- ✅ WASM module compiled (`layer.wasm` - ~3 MB)
- ✅ Frontend integrated (no backend needed!)
- ✅ Near-native performance

**Architecture**:
```
Standalone (Current):
React Frontend ←(Direct JS calls)→ Layer WASM Module (Go compiled)
```

**How to Use**:
```bash
# Build WASM module
cd layer-go
./build-wasm.sh

# Run frontend (NO BACKEND NEEDED!)
cd ../visual-editor/frontend
npm run dev
# Opens at http://localhost:5173
```

**Deployment**:
```bash
cd visual-editor/frontend
npm run build
# Deploy 'dist' folder to:
# - GitHub Pages
# - Netlify
# - Vercel
# - CloudFlare Pages
# - Any static hosting
```

**Benefits Achieved**:
- ✅ No Flask backend required
- ✅ Deploy as pure static site
- ✅ Works offline after first load
- ✅ Near-native performance (10x faster)
- ✅ All interactive UI features preserved
- ✅ Scales infinitely via CDN

**Documentation**:
- See `MIGRATION_SUCCESS.md` for complete migration details
- See `layer-go/WASM_COMPLETE.md` for WASM integration guide
- See `layer-go/INTERPRETER_COMPLETE.md` for Go implementation details
