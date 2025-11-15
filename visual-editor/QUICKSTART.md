# Layer Visual Editor - Quick Start Guide

Get the visual editor running in 3 minutes! ⚡

## Prerequisites Check

```bash
# Check Node.js (need 16+)
node --version

# Check Python (need 3.8+)
python3 --version

# Check npm
npm --version
```

## Step-by-Step Setup

### 1. Install Backend

```bash
cd visual-editor/backend
pip install -r requirements.txt
```

Expected output:
```
Successfully installed Flask-3.0.0 flask-cors-4.0.0
```

### 2. Install Frontend

```bash
cd ../frontend
npm install
```

This will take 1-2 minutes. Expected output:
```
added 300+ packages...
```

**Note**: We use ⚡ Vite instead of Create React App for 10x faster builds!

### 3. Start Backend Server

In **Terminal 1**:

```bash
cd visual-editor/backend
python app.py
```

You should see:
```
🚀 Layer Visual Editor Backend Starting...
📡 API available at http://localhost:5000
🔗 Frontend should connect to this server
 * Running on http://0.0.0.0:5000
```

**Keep this terminal running!**

### 4. Start Frontend

In **Terminal 2** (new terminal):

```bash
cd visual-editor/frontend
npm run dev
```

You should see:
```
  ➜  Local:   http://localhost:3000/
  ➜  Network: use --host to expose
  ➜  ready in 350 ms ⚡
```

**Your browser will automatically open to http://localhost:3000**

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
- **Bottom**: Terminal-style output console
- **Top Bar**: Execute, Save, Help buttons

## Common Issues

### "Backend not responding"

**Solution**: Check if backend is running on port 5000
```bash
curl http://localhost:5000/api/health
```

Should return: `{"status":"healthy"}`

### "Port 3000 already in use"

**Solution**: Use a different port
```bash
PORT=3001 npm start
```

### "Module not found"

**Solution**: Reinstall dependencies
```bash
rm -rf node_modules package-lock.json
npm install
```

## Next Steps

✅ Try editing the factorial function
✅ Toggle layers on/off with the eye icon
✅ Create your own layer code
✅ Experiment with debug logging

See full [README.md](README.md) for more features!

## Need Help?

- Read the [full documentation](README.md)
- Check the [Layer language tutorial](../../docs/TUTORIAL.md)
- See [example programs](../../examples/)

Happy layer programming! 🎨✨
