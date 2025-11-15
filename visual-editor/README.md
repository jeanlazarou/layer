# Layer Visual Editor

A beautiful, Photoshop-inspired visual interface for the Layer programming language, built with React and Material UI.

## Features

- 🎨 **Visual Layer Stack**: Drag-and-drop layer management with Material Design cards
- ✏️ **Monaco Code Editor**: VS Code's editor for a professional coding experience
- 🎯 **Real-time Execution**: Run Layer programs and see output instantly
- 👁️ **Layer Toggle**: Enable/disable layers to see how composition works
- 🎭 **Color-coded Layers**: Each layer type has a distinct color
- 📊 **Output Console**: Terminal-style console for results and debug logs

## Architecture

```
visual-editor/
├── frontend/          # React + Material UI app
│   ├── src/
│   │   ├── App.js            # Main application
│   │   ├── components/
│   │   │   ├── LayerStack.js    # Layer panel (left)
│   │   │   ├── CodeEditor.js    # Monaco editor (center)
│   │   │   └── OutputConsole.js # Output panel (bottom)
│   │   └── api/
│   │       └── layerApi.js      # Backend communication
│   └── package.json
└── backend/           # Flask API
    ├── app.py         # REST API for executing Layer code
    └── requirements.txt
```

## Getting Started

### Prerequisites

- Node.js 16+ and npm
- Python 3.8+
- Layer language installed (from parent directory)

### Installation

**1. Install Backend Dependencies:**

```bash
cd visual-editor/backend
pip install -r requirements.txt
```

**2. Install Frontend Dependencies:**

```bash
cd visual-editor/frontend
npm install
```

### Running the Editor

**1. Start the Backend (Terminal 1):**

```bash
cd visual-editor/backend
python app.py
```

Backend will run on `http://localhost:5000`

**2. Start the Frontend (Terminal 2):**

```bash
cd visual-editor/frontend
npm run dev
```

Frontend will open at `http://localhost:3000` (powered by ⚡ Vite!)

## Usage

### Creating a Layer Program

1. **Select a Layer**: Click on a layer card in the left panel
2. **Edit Code**: Write your Layer code in the Monaco editor
3. **Toggle Layers**: Click the eye icon to enable/disable layers
4. **Execute**: Click the play button ▶️ in the top toolbar
5. **View Output**: See results in the console at the bottom

### Example: Factorial with Debug

1. Select the **core** layer (green)
2. Edit the factorial function
3. Select the **debug** layer (orange)
4. Add logging with `log("message")`
5. Click Execute to see both result and debug logs

### Layer Colors

- 🟢 **Core** (Green): Pure functional logic
- 🔵 **UI** (Blue): User interface components
- 🟠 **Debug** (Orange): Logging and tracing
- 🟣 **Main** (Purple): Entry point expression

## Technology Stack

### Frontend
- **⚡ Vite**: Lightning-fast build tool and dev server (10-30x faster than CRA!)
- **React 18**: UI framework
- **Material UI 5**: Component library and design system
- **Monaco Editor**: Code editor (VS Code's editor)
- **Axios**: HTTP client
- **@emotion**: CSS-in-JS styling

### Backend
- **Flask**: Web framework
- **Flask-CORS**: Cross-origin resource sharing
- **Layer Interpreter**: Python implementation

## Features in Detail

### Layer Stack Panel
- Visual representation of layer composition
- Drag handles for reordering (coming soon)
- Enable/disable toggles with eye icons
- Lock icon for immutable layers (core)
- Color-coded layer indicators
- Selected layer highlighting

### Code Editor
- Syntax highlighting
- Line numbers
- Word wrap
- Auto-completion (coming soon)
- Layer-specific code editing
- Real-time sync with layer stack

### Output Console
- Terminal-style dark theme
- Syntax highlighting for errors
- Separate display of logs vs results
- Auto-scroll to latest output
- Clear formatting

## Keyboard Shortcuts (Coming Soon)

- `Ctrl/Cmd + Enter`: Execute code
- `Ctrl/Cmd + S`: Save program
- `Ctrl/Cmd + /`: Toggle layer visibility
- `Ctrl/Cmd + K`: Clear console

## Roadmap

### Phase 1 (Current)
- [x] Basic UI layout with Material UI
- [x] Layer stack visualization
- [x] Monaco editor integration
- [x] Backend API for execution
- [x] Output console
- [x] Layer enable/disable

### Phase 2 (Next)
- [ ] Drag-and-drop layer reordering
- [ ] Add/remove custom layers
- [ ] Save/load programs
- [ ] Export to .layer file
- [ ] Syntax highlighting for Layer language

### Phase 3 (Future)
- [ ] Visual function flow diagram
- [ ] Layer blending modes
- [ ] Collaborative editing
- [ ] Time-travel debugging
- [ ] Layer templates library

## Development

### Frontend Development

```bash
cd visual-editor/frontend
npm run dev        # Development server (⚡ with Vite HMR)
npm run build      # Production build
npm run preview    # Preview production build
npm run lint       # Lint code
```

### Backend Development

```bash
cd visual-editor/backend
python app.py      # Development server with auto-reload
```

### Code Style

Frontend follows React best practices and Material UI patterns.
Backend follows PEP 8 Python style guide.

## Troubleshooting

**Backend not connecting:**
- Ensure Flask is running on port 5000
- Check CORS is enabled
- Verify proxy setting in `package.json`

**Frontend not starting:**
- Run `npm install` to install dependencies
- Clear cache: `npm cache clean --force`
- Delete `node_modules` and reinstall

**Monaco editor not loading:**
- Check internet connection (loads from CDN)
- Verify `@monaco-editor/react` is installed

## Contributing

We welcome contributions! Areas of interest:
- UI/UX improvements
- Layer templates
- Documentation
- Bug fixes

## License

MIT License - Same as Layer language

## Credits

Built with ❤️ using:
- React and Material UI
- Monaco Editor by Microsoft
- Layer programming language

---

**Enjoy visual layer programming!** 🎨✨
