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

    # Keep the frontend's embedded copy in sync
    FRONTEND_PUBLIC="../visual-editor/frontend/public"
    if [ -d "$FRONTEND_PUBLIC" ]; then
        cp layer.wasm "$FRONTEND_PUBLIC/"
        [ -f wasm_exec.js ] && cp wasm_exec.js "$FRONTEND_PUBLIC/"
        echo "✅ Copied layer.wasm (and wasm_exec.js) to $FRONTEND_PUBLIC"
    fi

    echo ""
    echo "🚀 WASM module ready!"
else
    echo "❌ Build failed"
    exit 1
fi
