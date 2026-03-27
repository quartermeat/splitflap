#!/usr/bin/env bash
set -e
cd "$(dirname "$0")/.."

WASM_EXEC="$(go env GOMODCACHE)/golang.org/toolchain@v0.0.1-go1.24.5.linux-amd64/lib/wasm/wasm_exec.js"

echo "Building WASM..."
GOOS=js GOARCH=wasm go build -o splitflap.wasm .

echo "Copying to gh-pages..."
git stash
git checkout gh-pages
cp splitflap.wasm wasm_exec.js 2>/dev/null || cp "$WASM_EXEC" wasm_exec.js
git add splitflap.wasm wasm_exec.js
git commit -m "Deploy: $(git log master --oneline -1 | cut -d' ' -f2-)"
git push origin gh-pages
git checkout master
git stash pop 2>/dev/null || true

echo "Done. https://quartermeat.github.io/splitflap"
