#!/usr/bin/env bash
# Build the Layer interpreter to WebAssembly for the visual editor.
#
# Writes layer.wasm and the matching wasm_exec.js (the Go JS runtime glue,
# which must come from the same Go version that compiled the module) into
# visual-editor/frontend/public/.
set -euo pipefail

cd "$(dirname "$0")"
OUT_DIR="../visual-editor/frontend/public"
mkdir -p "$OUT_DIR"

echo "Building Layer WASM module with $(go version | awk '{print $3}')..."
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o "$OUT_DIR/layer.wasm" ./cmd/wasm

GOROOT="$(go env GOROOT)"
WASM_EXEC=""
for candidate in "$GOROOT/lib/wasm/wasm_exec.js" "$GOROOT/misc/wasm/wasm_exec.js"; do
    if [ -f "$candidate" ]; then
        WASM_EXEC="$candidate"
        break
    fi
done
if [ -z "$WASM_EXEC" ]; then
    echo "❌ Could not find wasm_exec.js under $GOROOT" >&2
    exit 1
fi
cp "$WASM_EXEC" "$OUT_DIR/wasm_exec.js"

echo "✅ $(ls -lh "$OUT_DIR/layer.wasm" | awk '{print $5}') → $OUT_DIR/layer.wasm"
echo "✅ wasm_exec.js copied from $WASM_EXEC"
