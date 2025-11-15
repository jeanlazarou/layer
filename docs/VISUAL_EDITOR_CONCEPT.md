# Layer Visual Editor Concept

## Overview

The Layer Visual Editor is inspired by Photoshop's layer paradigm, providing a visual interface for composing and managing Layer programs.

## Interface Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  Layer Editor - fibonacci.layer                          [×] [-]│
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────┐  ┌────────────────────────────────────┐  │
│  │   LAYER STACK    │  │        CODE EDITOR                 │  │
│  │                  │  │                                     │  │
│  │  ┌────────────┐  │  │  fn fibonacci(n) =                │  │
│  │  │   debug    │░ │  │    match n with                   │  │
│  │  │  👁  🔒     │░ │  │    | 0 -> 0                       │  │
│  │  └────────────┘░ │  │    | 1 -> 1                       │  │
│  │   ░░░░░░░░░░░░░  │  │    | n -> fib(n-1) + fib(n-2)    │  │
│  │                  │  │                                     │  │
│  │  ┌────────────┐  │  │                                     │  │
│  │  │     ui     │  │  │  # When debug layer is selected:   │  │
│  │  │  👁  🔒     │  │  │  fn fibonacci(n) =                │  │
│  │  └────────────┘  │  │    log("Computing fib({n})")      │  │
│  │                  │  │    core.fibonacci(n)              │  │
│  │  ┌────────────┐  │  │                                     │  │
│  │  │    core    │  │  │                                     │  │
│  │  │  👁  🔒     │  │  └────────────────────────────────────┘  │
│  │  └────────────┘  │                                          │
│  │                  │  ┌────────────────────────────────────┐  │
│  │  [+ Add Layer]   │  │        OUTPUT / CONSOLE            │  │
│  │                  │  │                                     │  │
│  │  [Enable All]    │  │  > fibonacci(10)                  │  │
│  │  [Disable All]   │  │  [TRACE] Computing fib(10)        │  │
│  │                  │  │  [TRACE] Computing fib(9)         │  │
│  └──────────────────┘  │  ...                              │  │
│                        │  Result: 55                       │  │
│                        └────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Key Features

### 1. Layer Stack Panel (Left)

**Visual Layer Representation:**
- Stacked from top to bottom (matching execution order)
- Each layer is a draggable card
- Visual indicators:
  - 👁 Eye icon: Toggle layer visibility (enable/disable)
  - 🔒 Lock icon: Prevent editing
  - Color coding by layer type

**Layer Interactions:**
- **Drag & Drop**: Reorder layers (affects composition)
- **Click**: Select layer to edit
- **Double-click**: Rename layer
- **Right-click**: Context menu (duplicate, delete, export)

**Layer Types:**
- **core** (green): Always visible, cannot be removed
- **ui** (blue): User interface layer
- **debug** (orange): Debugging and logging
- **custom** (purple): User-defined layers

### 2. Code Editor Panel (Center-Right)

**Context-Sensitive Editing:**
- Shows code for currently selected layer
- Syntax highlighting specific to Layer language
- Auto-completion for layer-aware functions
- Inline hints showing which layer a function comes from

**Layer-Aware Features:**
- `core.function()` calls shown with link to source
- Hover over function to see which layer defines it
- Visual indicators for overridden functions
- Quick jump to function definition in another layer

### 3. Output/Console Panel (Bottom-Right)

**Real-time Execution:**
- REPL for testing expressions
- Live output as code changes
- Colored output based on layer (debug logs in orange, etc.)
- Stack trace with layer information

### 4. Layer Blending Modes

Inspired by Photoshop's blend modes, different ways layers compose:

**Override Mode** (default):
- Upper layer completely replaces lower layer function
- Example: Debug layer replaces core.fibonacci

**Wrap Mode**:
- Upper layer wraps lower layer function
- Automatically available via `core.function()`
- Visual indicator shows wrapping relationship

**Extend Mode**:
- Upper layer adds new functions
- Doesn't override existing functions
- Prevents accidental override

**Compose Mode**:
- Function result from lower layer is piped into upper layer
- Functional composition chain

### 5. Layer Properties Panel

When a layer is selected:

```
┌──────────────────────────┐
│  Layer: debug            │
├──────────────────────────┤
│  Mode: [Wrap ▼]          │
│  Opacity: [████░░] 80%   │
│  Enabled: [✓]            │
│                          │
│  Functions:              │
│  • fibonacci (wraps)     │
│  • log (new)             │
│  • trace (new)           │
│                          │
│  Dependencies:           │
│  • core                  │
│                          │
│  [Export Layer]          │
│  [Duplicate Layer]       │
│  [Delete Layer]          │
└──────────────────────────┘
```

### 6. Visual Function Flow

Visualize how functions compose across layers:

```
fibonacci(10) called
       ↓
┌──────────────────┐
│  debug layer     │  → log("Computing...")
│  fibonacci(n)    │      ↓
└──────────────────┘      Calls core.fibonacci(n)
       ↓
┌──────────────────┐
│  core layer      │  → Recursive computation
│  fibonacci(n)    │
└──────────────────┘
       ↓
    Result: 55
```

### 7. Layer Templates

Pre-built layer templates for common patterns:

- **Logging Layer**: Automatically wrap all functions with logging
- **Memoization Layer**: Add caching to expensive functions
- **Type Checking Layer**: Runtime type validation
- **Performance Layer**: Benchmarking and profiling
- **Testing Layer**: Assertions and test cases

### 8. Live Preview

Split view showing:
- **Code View**: Traditional text editor
- **Visual View**: Graphical representation of layer composition
- **Flow View**: Data flow through layers
- **Output View**: Real-time execution results

## Interaction Patterns

### Adding a New Layer

1. Click "+ Add Layer"
2. Choose layer type or create custom
3. Layer appears in stack
4. Editor switches to new layer
5. Define functions
6. See real-time updates in output

### Disabling a Layer

1. Click eye icon on layer
2. Layer becomes semi-transparent
3. Code re-evaluates without that layer
4. Output updates immediately
5. Can compare results with/without layer

### Reordering Layers

1. Drag layer in stack
2. Drop at new position
3. Preview shows composition change
4. Release to apply
5. Code re-evaluates with new order

### Debugging Flow

1. Click on a function call in output
2. Visual flow diagram appears
3. Shows path through layers
4. Highlights each layer that was involved
5. Step through execution layer by layer

## Technical Implementation

### Architecture

```
┌─────────────────────────────────────┐
│   Web-based UI (React/Vue)          │
├─────────────────────────────────────┤
│   Layer Editor Components           │
│   • LayerStack                      │
│   • CodeEditor (Monaco/CodeMirror)  │
│   • OutputPanel                     │
│   • FlowVisualizer                  │
├─────────────────────────────────────┤
│   Layer Language Server (Python)    │
│   • Parser/Lexer                    │
│   • Interpreter                     │
│   • Layer Composition Engine        │
│   • WebSocket for real-time eval    │
└─────────────────────────────────────┘
```

### Features to Implement

**Phase 1: Basic Editor**
- Layer stack visualization
- Basic code editing
- Enable/disable layers
- Simple REPL

**Phase 2: Advanced Composition**
- Drag & drop reordering
- Layer blending modes
- Visual function flow
- Layer templates

**Phase 3: Collaboration**
- Share layers
- Layer marketplace
- Version control integration
- Real-time collaboration

**Phase 4: Advanced Features**
- Visual debugging
- Performance profiling
- Type visualization
- AI-assisted layer suggestion

## Mockup Screenshots (Conceptual)

### Main Interface
```
[Layer List] [Code Editor      ] [Output        ]
[          ] [                  ] [              ]
[core ✓    ] [fn factorial(n)=  ] [> factorial(5)]
[ui   ✓    ] [  match n with    ] [Result: 120   ]
[debug✓    ] [  | 0 -> 1        ] [              ]
[+ Add     ] [  | n -> n*fact..] [              ]
```

### Layer Blend Preview
```
    debug.fibonacci(5)
         ↓
    [Wrap mode - adds logging]
         ↓
    core.fibonacci(5)
         ↓
    [Pure computation]
         ↓
       Result
```

## Future Vision

The visual editor transforms Layer from a text-based language into a visual programming environment where:

1. **Layers are first-class visual objects**
2. **Composition is drag-and-drop**
3. **Function flow is visualized**
4. **Debugging is interactive and graphical**
5. **Collaboration happens in real-time**

This bridges the gap between:
- Traditional programming (powerful, precise)
- Visual programming (intuitive, accessible)

Making functional programming with layers as intuitive as editing photos in Photoshop.
