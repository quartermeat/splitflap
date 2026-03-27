#!/bin/bash
set -e

cd "$(dirname "$0")/.."

echo "Building splitflap.wasm..."
GOOS=js GOARCH=wasm go build -o splitflap.wasm .

echo "Copying wasm_exec.js..."
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .

echo "Done. Run: go run serve.go"
