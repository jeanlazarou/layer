# Layer Visual Editor - Screenshots & Features

## Main Interface

The Visual Editor provides a clean, modern interface inspired by Photoshop's layer system.

### Layout Overview

```
┌──────────────────────────────────────────────────────────────────┐
│  Layer Visual Editor                              ▶ 💾 ❓  🌙    │  ← App Bar
├─────────────┬────────────────────────────────────────────────────┤
│             │                                                     │
│  LAYER      │              CODE EDITOR                           │
│  STACK      │                                                     │
│             │  ┌──────────────────────────────────────┐          │
│ ┌─────────┐ │  │ Code Editor              [core]     │          │
│ │ debug   │ │  ├──────────────────────────────────────┤          │
│ │ 👁  🔒  │ │  │ # Core layer                         │          │
│ └─────────┘ │  │ fn factorial(n) =                    │          │
│             │  │   match n with                       │          │
│ ┌─────────┐ │  │   | 0 -> 1                           │          │
│ │   ui    │ │  │   | n -> n * factorial(n - 1)       │          │
│ │ 👁  🔒  │ │  │                                      │          │
│ └─────────┘ │  │                                      │          │
│             │  │                                      │          │
│ ┌─────────┐ │  └──────────────────────────────────────┘          │
│ │  core   │◄│  ← Selected                                        │
│ │ 👁  🔒  │ │                                                     │
│ └─────────┘ │                                                     │
│             │                                                     │
│ ┌─────────┐ │                                                     │
│ │  main   │ │                                                     │
│ └─────────┘ │                                                     │
│             │                                                     │
│ [+ Add      │                                                     │
│    Layer]   │                                                     │
│             ├─────────────────────────────────────────────────────┤
│             │  OUTPUT CONSOLE                        [Terminal]  │
│             ├─────────────────────────────────────────────────────┤
│             │  > Ready to execute...                             │
│             │  > Click the play button to run                    │
│             │                                                     │
└─────────────┴─────────────────────────────────────────────────────┘
```

## Key Features

### 1. Layer Stack Panel

**Location**: Left sidebar

**Features**:
- 🎨 Color-coded layer cards (Green=Core, Blue=UI, Orange=Debug)
- 👁️ Visibility toggle (enable/disable layers)
- 🔒 Lock indicator for immutable layers
- 📍 Drag handles for reordering
- ✨ Material Design elevation and transitions
- 🎯 Click to select and edit layer code

**Layer Card States**:
```
┌──────────────┐
│ ≡  🟢 core   │  ← Drag handle, color indicator, name
│    👁  🔒    │  ← Eye (toggle), Lock (if immutable)
│ Active • Locked │  ← Status indicators
└──────────────┘
```

### 2. Code Editor

**Location**: Center panel

**Features**:
- ✏️ Monaco Editor (VS Code's editor)
- 🎨 Syntax highlighting
- 🔢 Line numbers
- 📝 Word wrap
- 🏷️ Layer name badge (color-coded)
- ⚡ Auto-save to layer state
- 🔍 Search and replace (Cmd/Ctrl+F)

**Header**:
```
┌────────────────────────────────────┐
│ Code Editor              [core]    │  ← Layer badge
├────────────────────────────────────┤
│  1  # Core layer                   │
│  2  fn factorial(n) = ...          │
```

### 3. Output Console

**Location**: Bottom panel

**Features**:
- 🖥️ Terminal-style dark theme
- 📊 Separate display of logs and results
- 🎨 Syntax highlighting for errors
- 📜 Auto-scroll to latest output
- 🧹 Clear formatting
- ⚡ Real-time output streaming

**Example Output**:
```
┌─────────────────────────────────────────┐
│ 🖥 Output Console                       │
├─────────────────────────────────────────┤
│ [11:30:45] [TRACE] Computing fib(5)    │
│ [11:30:45] [TRACE] Computing fib(4)    │
│ ...                                     │
│                                         │
│ Result: 120                             │
└─────────────────────────────────────────┘
```

### 4. Top Toolbar

**Features**:
- ▶️ Execute button (runs the Layer program)
- 💾 Save button (save to file)
- ❓ Help button (opens documentation)
- 🌙 Dark mode toggle (coming soon)

## User Interactions

### Selecting a Layer

1. Click on any layer card
2. Card highlights with blue border and elevation
3. Code editor updates with layer's code
4. Layer name badge updates color

**Visual Feedback**:
```
Before:                After:
┌─────────┐          ┌═════════┐  ← Thicker border
│  core   │          ║  core   ║  ← Higher elevation
└─────────┘          ║ 👁  🔒  ║
                     ╚═════════╝
```

### Toggling Layer Visibility

1. Click the eye icon 👁
2. Icon changes to 👁‍🗨 (disabled)
3. Layer card becomes semi-transparent
4. Layer is excluded from execution

**States**:
```
Enabled:            Disabled:
┌─────────┐         ┌─────────┐
│  debug  │         │  debug  │  ← 50% opacity
│  👁     │         │  👁‍🗨    │  ← Crossed eye
└─────────┘         └─────────┘
```

### Executing Code

1. Click ▶️ play button
2. Button shows loading spinner
3. Output console shows "Executing..."
4. Results appear in console
5. Debug logs appear in orange

## Color Scheme

### Layer Colors

- 🟢 **Core** (`#4caf50`): Green - Foundation, always enabled
- 🔵 **UI** (`#2196f3`): Blue - User interface layer
- 🟠 **Debug** (`#ff9800`): Orange - Debugging and logging
- 🟣 **Main** (`#9c27b0`): Purple - Entry point

### Theme Colors

**Light Mode** (default):
- Background: `#f5f5f5` (light gray)
- Cards: `#ffffff` (white)
- Primary: `#1976d2` (blue)
- Text: `#212121` (dark gray)

**Dark Mode** (coming soon):
- Background: `#121212`
- Cards: `#1e1e1e`
- Primary: `#90caf9`
- Text: `#ffffff`

## Material UI Components Used

- **Card**: Layer cards with elevation
- **AppBar**: Top navigation bar
- **Toolbar**: Button container
- **IconButton**: Action buttons
- **Chip**: Layer name badges
- **Typography**: Text styling
- **Box**: Layout container
- **Divider**: Section separators
- **Paper**: Console background

## Responsive Design

The editor adapts to different screen sizes:

- **Desktop** (1920x1080): Full 3-panel layout
- **Laptop** (1366x768): Optimized spacing
- **Tablet** (768px): Collapsible layer panel
- **Mobile** (375px): Stack layout (coming soon)

## Animation & Transitions

All transitions use Material UI's default:
- **Duration**: 200ms
- **Easing**: cubic-bezier(0.4, 0, 0.2, 1)
- **Properties**: elevation, transform, opacity

**Examples**:
- Layer selection: Elevation 1 → 8
- Layer hover: translateY(0) → translateY(-2px)
- Toggle: opacity 1 → 0.5

## Accessibility

- ✅ Keyboard navigation (Tab, Enter, Space)
- ✅ Screen reader support (ARIA labels)
- ✅ High contrast mode compatible
- ✅ Focus indicators
- ✅ Tooltip descriptions

## Performance

- ⚡ Lazy loading for Monaco Editor
- ⚡ Debounced code changes
- ⚡ Virtualized layer list (for many layers)
- ⚡ Memoized React components

---

*Note: Actual screenshots will be added once the app is running. This document describes the UI/UX design.*
