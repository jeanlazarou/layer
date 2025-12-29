# Layer Visual Editor - Quick Start Guide

Get the visual editor running in 2 minutes! ⚡

## Prerequisites Check

```bash
# Check Node.js (need 16+)
node --version

# Check npm
npm --version
```

**Note**: No Python needed! The interpreter runs in WebAssembly.

## Step-by-Step Setup

### 1. Install Frontend

```bash
cd visual-editor/frontend
npm install
```

This will take 1-2 minutes. Expected output:
```
added 300+ packages...
```

**Note**: We use:
- ⚡ **Vite** for 10-30x faster builds than CRA
- 📘 **TypeScript** for type safety and better developer experience
- 🦀 **WebAssembly** for near-native performance

### 2. Start Visual Editor

```bash
cd visual-editor/frontend
npm run dev
```

You should see:
```
  ➜  Local:   http://localhost:5173/
  ➜  Network: use --host to expose
  ➜  ready in 350 ms ⚡
```

**Your browser will automatically open to http://localhost:5173**

## First Run

You should now see:

```
┌─────────────────────────────────────────────────────────┐
│  Layer Visual Editor                          ▶ 💾 ❓   │
├────────────┬──────────────────────────────────────────────┤
│ LAYERS     │  CODE EDITOR                                │
│            │                                              │
│ ┌────────┐ │  # Core layer - pure functional logic       │
│ │ debug  │░│  fn factorial(n) =                          │
│ │ 👁 🔒  │░│    match n with                             │
│ └────────┘░│    | 0 -> 1                                 │
│            │    | n -> n * factorial(n - 1)              │
│ ┌────────┐ │                                              │
│ │   ui   │ │                                              │
│ │ 👁 🔒  │ │                                              │
│ └────────┘ │                                              │
│            │                                              │
│ ┌────────┐ │──────────────────────────────────────────────│
│ │  core  │ │  OUTPUT CONSOLE                             │
│ │ 👁 🔒  │ │  > Ready to execute...                      │
│ └────────┘ │                                              │
└────────────┴──────────────────────────────────────────────┘
```

## Try It Out!

1. **Click the Play Button** ▶️ in the top toolbar
2. See the output: `Result: 120`
3. **Click on the "debug" layer** in the left panel
4. **Click Play** again to see debug logs!

## What's Happening?

- **Left Panel**: Layer stack (like Photoshop layers)
- **Center**: Monaco code editor (VS Code editor)
- **Bottom**: Terminal-style output console + UI preview
- **Top Bar**: Execute, Save, Help buttons
- **Under the Hood**: WebAssembly interpreter running locally in your browser!

## Common Issues

### "Port 5173 already in use"

**Solution**: Use a different port
```bash
PORT=5174 npm run dev
```

### "Module not found"

**Solution**: Reinstall dependencies
```bash
rm -rf node_modules package-lock.json
npm install
```

### "WASM loading error"

**Solution**: Rebuild the WASM module
```bash
cd ../../layer-go
./build-wasm.sh
```

## Next Steps

✅ Try editing the factorial function
✅ Toggle layers on/off with the eye icon
✅ Create your own layer code
✅ Experiment with debug logging
✅ Try the UI layer with interactive components!

See full [README.md](README.md) for more features!

## Need Help?

- Read the [full documentation](README.md)
- Check the [Layer language spec](../../LANGUAGE_SPEC.md)
- See [example programs](../../examples/)

Happy layer programming! 🎨✨
