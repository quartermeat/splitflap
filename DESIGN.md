# splitflap — Design Document

## What it is

A community split-flap display running in the browser. Users submit messages to a shared queue. Each message gets 30 seconds on the board. The most-liked message holds the champion spot. Anyone can like what's currently showing. Everything updates in real-time.

**Live:** https://quartermeat.github.io/splitflap

---

## Modes

### Community (default — `/?`)
- Shared board visible to everyone
- Messages submitted to a 30-second public queue
- Like button on the current display (unlimited, vibe counter)
- Leaderboard shows all-time top messages by likes
- Champion = highest liked message, shown when queue is empty
- Real-time: all clients flip together, likes tick up live

### Solo (`/?solo`)
- Private board, no backend
- Current behavior: type → Enter → board flips
- Accessible via the SOLO → link in the bottom corner

---

## Display Logic

```
Queue (FIFO by submitted_at)
  └─ Active message (display_until = now+30s)
       └─ When expired → next from queue
            └─ If queue empty → show Champion
                 └─ If no champion → blank board

Champion election:
  When a message's 30s window ends:
    if message.likes > current_champion.likes → new champion
```

Vote atrophy is not yet implemented — planned: halve like scores daily.

---

## Tech Stack

| Layer | Tech |
|-------|------|
| Runtime | Go + Ebitengine v2.9.9 |
| Compilation | GOOS=js GOARCH=wasm |
| Hosting | GitHub Pages (`gh-pages` branch) |
| Backend | Supabase (Postgres + real-time) |
| Real-time | Supabase postgres_changes subscription |
| Audio | PCM synthesis in Go (no asset files) |

---

## Architecture

### Entry Points
- `main.go` — desktop binary + WASM entry
- `serve.go` (`//go:build ignore`) — local dev server on :8080

### Key Files

| File | Role |
|------|------|
| `game.go` | Ebitengine game loop — Update/Draw, mode switching, click handling |
| `board.go` | 6×22 tile grid, staggered cascade, word wrap |
| `tile.go` | Per-tile split-flap animation (idle → flip → bounce), concurrent sound |
| `ui.go` | Input box, cursor, mobile keyboard hit detection |
| `community.go` | Message struct, CommunityState, leaderboard + like bar drawing |
| `community_js.go` | JS bridge for Supabase (build tag: `js && wasm`) |
| `community_stub.go` | Non-WASM stubs for community JS functions |
| `js.go` | Mobile keyboard bridge, JS interop (build tag: `js && wasm`) |
| `js_stub.go` | Non-WASM stubs for JS functions |
| `sound.go` | Synthesized clack + flicker audio via Ebitengine audio context |
| `draw.go` | Font loading, tile rendering helpers, prompt + version overlay |
| `index.html` | Canvas setup, Supabase JS SDK, mobile keyboard bridge, community JS |

### Platform Split Pattern
- `*_js.go` + build tag `js && wasm` — browser/WASM only
- `*_stub.go` + build tag `!js` — desktop stubs
- Everything else compiles everywhere

---

## Supabase Schema

```sql
messages (
  id           uuid primary key default gen_random_uuid(),
  text         text not null,
  likes        integer default 0,
  submitted_at timestamptz default now(),
  display_until timestamptz,    -- null = in queue, set = active/displayed
  is_champion  boolean default false
)
```

**RPC:** `increment_likes(message_id uuid)` — atomic likes increment, callable by anon role.

**Real-time:** `supabase_realtime` publication includes `messages` table. JS subscribes via `postgres_changes` and forwards events to Go via `splitflapOnCommunityEvent(jsonStr)`.

---

## Mobile / WASM Layout Rules

- Canvas is `position: fixed` — keyboard cannot push it
- Board is **top-aligned** (offsetY = 16px fixed) so keyboard opening over bottom half doesn't obscure the board
- Layout reserves space below board: input bar + like bar + leaderboard + mode link
- `boardAreaH = screenH - 16 - belowBoard` — board scales to fill this budget at 97% margin

---

## Split-Flap Animation

Each tile cycles through a charset: ` ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.,!'?:/-`

Flip phases per tile:
1. **Idle** — show current char, no animation
2. **Flip** — intermediate chars flash quickly (`flipTicksFast = 2`), final char takes longer (`flipTicksFinal = 6`)
3. **Bounce** — brief settle animation after landing (`bounceSettleTicks = 8`)

Board cascade: 2-tick stagger between columns (left to right).

Audio: `go PlayFlicker()` per intermediate char, `go PlayClack()` on final landing. Pitch varies ±8% per play.

---

## Deploy Workflow

```bash
./scripts/deploy.sh
```

1. Builds WASM (`GOOS=js GOARCH=wasm`)
2. Stashes both binary files to `/tmp`
3. Checks out `gh-pages`, copies files, commits, pushes
4. Returns to `master`, restores local copies

**wasm_exec.js** location:
`$(go env GOMODCACHE)/golang.org/toolchain@v0.0.1-go1.24.5.linux-amd64/lib/wasm/wasm_exec.js`

---

## Planned / Not Yet Built

- Vote atrophy (halve likes daily)
- Queue position indicator ("your message is #3 in queue")
- Message moderation / spam throttle
- PWA manifest for Add to Home Screen
- Play Store TWA wrapper
