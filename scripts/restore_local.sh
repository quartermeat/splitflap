#!/usr/bin/env bash
# Restores the local WASM files needed to run the dev server.
# Run this after any git branch switch.
set -e
cd "$(dirname "$0")/.."

WASM_EXEC="$(go env GOMODCACHE)/golang.org/toolchain@v0.0.1-go1.24.5.linux-amd64/lib/wasm/wasm_exec.js"

if [ ! -f splitflap.wasm ]; then
    echo "Building splitflap.wasm..."
    GOOS=js GOARCH=wasm go build -o splitflap.wasm .
fi

if [ ! -f wasm_exec.js ]; then
    echo "Restoring wasm_exec.js..."
    cp "$WASM_EXEC" wasm_exec.js
fi

echo "Ready. Run: go run serve.go"
