#!/bin/bash

# Build Layer language to WebAssembly

echo "Building Layer WASM module..."

# Build the WASM file
GOOS=js GOARCH=wasm go build -o layer.wasm ./cmd/wasm

if [ $? -eq 0 ]; then
    echo "✅ Successfully built layer.wasm"

    # Get the size
    SIZE=$(ls -lh layer.wasm | awk '{print $5}')
    echo "📦 WASM file size: $SIZE"

    # Copy Go's WASM JavaScript support file
    GOROOT=$(go env GOROOT)
    if [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
        cp "$GOROOT/misc/wasm/wasm_exec.js" .
        echo "✅ Copied wasm_exec.js"
    else
        echo "⚠️  Warning: Could not find wasm_exec.js"
    fi

    echo ""
    echo "🚀 WASM module ready!"
    echo ""
    echo "To use in the frontend:"
    echo "  1. Copy layer.wasm to visual-editor/frontend/public/"
    echo "  2. Copy wasm_exec.js to visual-editor/frontend/public/"
    echo "  3. Load in React using the WASM loader"
else
    echo "❌ Build failed"
    exit 1
fi
