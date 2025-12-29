# Layer Visual Editor

A beautiful, Photoshop-inspired visual interface for the Layer programming language, built with React, Material UI, and WebAssembly.

## Features

- 🎨 **Visual Layer Stack**: Drag-and-drop layer management with Material Design cards
- ✏️ **Monaco Code Editor**: VS Code's editor for a professional coding experience
- 🎯 **Real-time Execution**: Run Layer programs instantly with WebAssembly
- 👁️ **Layer Toggle**: Enable/disable layers to see how composition works
- 🎭 **Color-coded Layers**: Each layer type has a distinct color
- 📊 **Output Console**: Terminal-style console for results and debug logs
- 🎪 **Interactive UI Preview**: See UI components render live as Material UI elements
- 🦀 **WebAssembly Powered**: Near-native performance, runs entirely in browser
- 💾 **Browser Storage**: File I/O using localStorage (no backend needed!)

## Architecture

```
visual-editor/
└── frontend/          # React + Material UI + WASM
    ├── src/
    │   ├── App.tsx           # Main application
    │   ├── components/
    │   │   ├── LayerStack.tsx    # Layer panel (left)
    │   │   ├── CodeEditor.tsx    # Monaco editor (center)
    │   │   ├── OutputConsole.tsx # Output panel (bottom)
    │   │   └── UIPreview.tsx     # Interactive UI rendering
    │   └── wasm/
    │       ├── layer.wasm        # Go interpreter (WASM)
    │       ├── wasm_exec.js      # Go WASM runtime
    │       └── browserFileSystem.ts # localStorage file system
    └── package.json
```

## Getting Started

### Prerequisites

- Node.js 16+ and npm

**That's it!** No Python, no backend server needed.

### Installation

```bash
cd visual-editor/frontend
npm install
```

### Running the Editor

```bash
cd visual-editor/frontend
npm run dev
```

Visual editor will open at `http://localhost:5173` (powered by ⚡ Vite!)

## Usage

### Creating a Layer Program

1. **Select a Layer**: Click on a layer card in the left panel
2. **Edit Code**: Write your Layer code in the Monaco editor
3. **Toggle Layers**: Click the eye icon to enable/disable layers
4. **Execute**: Click the play button ▶️ in the top toolbar
5. **View Output**: See results in the console at the bottom
6. **Interactive UI**: If your code uses the UI layer, see live components!

### Example: Factorial with Debug

1. Select the **core** layer (green)
2. Edit the factorial function
3. Select the **debug** layer (orange)
4. Add logging with `log("message")`
5. Click Execute to see both result and debug logs

### Example: Interactive Calculator

```layer
@layers [core, ui]

@layer core {
  fn square(x) = x * x
  fn double(x) = x + x
}

@layer<ui> ui {
  fn app() = view({
    title: "Math Tools",
    children: [
      input({name: "num", type: "number", placeholder: "Enter number"}),
      button({label: "Square", onclick: "square(num)"}),
      button({label: "Double", onclick: "double(num)"})
    ]
  })
}

render(app())
```

This will create **actual interactive buttons and inputs** that execute Layer code when clicked!

### Layer Colors

- 🟢 **Core** (Green): Pure functional logic
- 🔵 **UI** (Blue): User interface components
- 🟠 **Debug** (Orange): Logging and tracing
- 🟣 **Data** (Purple): File I/O and localStorage
- 🟣 **Main** (Purple): Entry point expression

## Tech Stack

### Frontend
- **React 18**: Modern UI framework
- **TypeScript**: Type-safe development
- **Material UI 5**: Google's Material Design components
- **Monaco Editor**: VS Code's editor component
- **Vite**: Lightning-fast build tool
- **Axios**: HTTP client (for future features)

### Interpreter
- **Go**: Language implementation
- **WebAssembly**: Near-native performance in browser
- **syscall/js**: Go/JavaScript interop

### Storage
- **localStorage**: Browser-native file system
- **IndexedDB**: Future enhancement for larger files

## Project Structure

```
src/
├── App.tsx                  # Main application component
├── components/
│   ├── LayerStack.tsx       # Layer management panel
│   ├── CodeEditor.tsx       # Monaco editor wrapper
│   ├── OutputConsole.tsx    # Output display with UI preview
│   └── UIPreview.tsx        # Renders Layer UI as React components
├── wasm/
│   ├── layer.wasm           # Compiled interpreter
│   ├── wasm_exec.js         # Go WASM runtime
│   └── browserFileSystem.ts # File I/O abstraction
└── index.tsx                # Entry point
```

## Development

### Running Development Server

```bash
npm run dev      # Start Vite dev server with hot reload
```

### Building for Production

```bash
npm run build    # Build optimized bundle
npm run preview  # Preview production build locally
```

### Linting and Formatting

```bash
npm run lint     # ESLint check
npm run format   # Prettier formatting
```

### Rebuilding WASM Module

If you modify the Go interpreter:

```bash
cd ../../layer-go
./build-wasm.sh
```

This compiles the Go interpreter to WebAssembly and copies it to the frontend.

## Features in Detail

### Layer Stack Panel

- **Visual Cards**: Each layer shown as Material UI card
- **Eye Icon**: Toggle layer on/off
- **Lock Icon**: Indicates built-in layers (cannot delete)
- **Color Coding**: Instant visual recognition
- **Reordering**: Drag layers to change composition order (future)

### Code Editor

- **Syntax Highlighting**: Layer language syntax
- **Autocomplete**: Context-aware suggestions
- **Error Highlighting**: Real-time syntax checking
- **Vim Keybindings**: Optional (future)
- **Multiple Files**: Tab-based editing (future)

### Output Console

- **Terminal Style**: Familiar console appearance
- **ANSI Colors**: Syntax-highlighted output
- **Clear Button**: Reset console
- **Copy to Clipboard**: Easy result sharing
- **UI Preview**: Live rendering of UI components

### Interactive UI Rendering

The visual editor can render Layer UI components as **actual interactive React components**:

- `view({...})` → Material UI Box with border
- `text({content: "..."})` → Typography component
- `button({label: "...", onclick: "..."})` → Interactive Button
- `input({name: "...", type: "..."})` → TextField with state
- `container({layout: "..."})` → Flexbox layout

When you click a button, the `onclick` expression executes in the Layer interpreter with current input values!

## Browser File System

The visual editor provides a virtual file system using localStorage:

```layer
@layers [core, data]

# Write a file
data.write_file("notes.txt", "Hello, Layer!")

# Read it back
let content = data.read_file("notes.txt") in
data.unwrap(content, "File not found")
```

Files persist across sessions in your browser!

## Deployment

The visual editor is a **pure static site** - deploy anywhere:

### GitHub Pages

```bash
npm run build
# Deploy 'dist' folder to gh-pages branch
```

### Netlify / Vercel / Cloudflare Pages

```bash
npm run build
# Deploy 'dist' folder
```

### Configuration

- **Build Command**: `npm run build`
- **Output Directory**: `dist`
- **Node Version**: 16+

## Troubleshooting

### "WASM module failed to load"

**Solution**: Ensure the WASM file is in `public/wasm/`:
```bash
cd ../../layer-go
./build-wasm.sh
```

### "Port 5173 already in use"

**Solution**: Use a different port
```bash
PORT=5174 npm run dev
```

### "localStorage quota exceeded"

**Solution**: Clear browser storage for the site or use smaller files

### "UI components not rendering"

**Solution**: Check that your Layer code returns valid XML from `render()` function

## Roadmap

- [x] Monaco code editor integration
- [x] Layer stack visualization
- [x] Real-time code execution
- [x] WebAssembly interpreter
- [x] Interactive UI rendering
- [x] Browser file system
- [ ] Multiple file tabs
- [ ] Layer drag-and-drop reordering
- [ ] Vim keybindings
- [ ] Theme customization (dark mode)
- [ ] Collaborative editing
- [ ] Share programs via URL
- [ ] More UI components (tables, charts, forms)

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## License

See [LICENSE](../../LICENSE) for details.

## Learn More

- [Language Specification](../../LANGUAGE_SPEC.md)
- [Philosophy](../../PHILOSOPHY.md)
- [Example Programs](../../examples/)
- [Quick Start Guide](QUICKSTART.md)

Happy layer programming! 🎨✨
