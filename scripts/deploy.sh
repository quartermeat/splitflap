#!/usr/bin/env bash
set -e
cd "$(dirname "$0")/.."

WASM_EXEC="$(go env GOMODCACHE)/golang.org/toolchain@v0.0.1-go1.24.5.linux-amd64/lib/wasm/wasm_exec.js"

echo "Building WASM..."
GOOS=js GOARCH=wasm go build -o splitflap.wasm .

# Stash both files before branch switch so they survive
cp splitflap.wasm /tmp/splitflap.wasm
cp "$WASM_EXEC" /tmp/wasm_exec.js

echo "Deploying to gh-pages..."
git stash
git checkout gh-pages
cp /tmp/splitflap.wasm .
cp /tmp/wasm_exec.js .
git add splitflap.wasm wasm_exec.js
git commit -m "Deploy: $(git log master --oneline -1 | cut -d' ' -f2-)"
git push origin gh-pages
git checkout master
git stash pop 2>/dev/null || true

# Restore local copies for the dev server
cp /tmp/splitflap.wasm .
cp /tmp/wasm_exec.js .

echo "Done. https://quartermeat.github.io/splitflap"
