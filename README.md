# splitflap

A community split-flap display running in the browser, built with Go + WebAssembly.

**Live:** https://quartermeat.github.io/splitflap

## What it is

A shared split-flap display board — the kind you see in airports and train stations. Anyone can type a message and send it to the public queue. Messages display for 30 seconds. The most-liked message holds the champion spot and shows by default when the queue is empty.

- **Community mode** (default) — shared board, public queue, live like counter, leaderboard
- **Solo mode** (`?solo`) — private board, just for you, no backend

## Tech

- **Go + Ebitengine** — game loop, rendering, synthesized audio (no assets)
- **WebAssembly** — runs in the browser on any platform
- **Supabase** — Postgres + real-time subscriptions for community state
- **GitHub Pages** — static hosting for the WASM app

## Controls

| Input | Action |
|-------|--------|
| Type | Fill the input box |
| Enter | Send to the board (community queue in community mode) |
| Shift+Enter | Insert a line break |
| Backspace | Delete last character |
| F2 | Screenshot |
| Tap ♥ | Like the current message |
| Tap SOLO → | Switch to solo mode |

## Local dev

```bash
# Build WASM
GOOS=js GOARCH=wasm go build -o splitflap.wasm .

# Copy wasm_exec.js (first time)
cp "$(go env GOMODCACHE)/golang.org/toolchain@v0.0.1-go1.24.5.linux-amd64/lib/wasm/wasm_exec.js" .

# Serve locally
go run serve.go
# → http://localhost:8080
```

## Deploy

```bash
./scripts/deploy.sh
```

Builds WASM, commits to `gh-pages`, restores local files automatically.
