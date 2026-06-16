# Teer

**Terminal Workspace Manager** — a desktop app for developers to organize and run multiple CLI sessions grouped by project workspace.

> Built with [Wails v3](https://v3.wails.io/) (Go + Svelte + TypeScript)

![Platform](https://img.shields.io/badge/platform-Linux-blue) ![Go](https://img.shields.io/badge/Go-1.25+-00ADD8) ![Status](https://img.shields.io/badge/status-alpha-orange)

---

## What is Teer?

Developers constantly juggle multiple terminal sessions: dev servers, watchers, log tailers, database clients, SSH sessions, build tools. These sessions scatter across tabs and windows with no project context.

**Teer** solves this with a **workspace model** — group your terminals by project, persist their definitions across restarts, and bring everything back in one click.

Think of it as a terminal emulator (like Tabby or Wave) combined with VS Code's workspace concept, built specifically for managing running CLI processes.

---

## Features

- **Workspaces** — create named workspaces (with color labels) for each project
- **Multiple terminals per workspace** — displayed as tabs, all full PTY (supports vim, htop, top, etc.)
- **Persistent layout** — workspace and session definitions survive app restarts
- **Keyboard-driven** — command palette + shortcuts for fast navigation
- **Lightweight** — uses OS WebView (no bundled Chromium), target idle RAM < 150 MB
- **Full ANSI support** — powered by xterm.js with fit, search, and web-links addons

> **Note:** The binary and Go module are named `teer` (lowercase). "Teer" is the display name of the application.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Desktop framework | Wails v3 (alpha.82) |
| Backend | Go 1.25+ |
| PTY / shell | `creack/pty` (Linux/macOS) |
| Frontend | Svelte 5 + TypeScript |
| Terminal renderer | xterm.js + fit/search/web-links addons |
| Build tool | Vite |
| Config storage | JSON at `~/.config/teer/` |
| FE ↔ BE comms | Wails RPC bindings + Events (streaming I/O) |

---

## Project Structure

```
teer/
├── main.go                          # Wails entry point, service registration
├── internal/
│   ├── domain/
│   │   ├── entities.go              # Workspace, Session types
│   │   └── ports.go                 # Interface definitions
│   ├── infra/
│   │   ├── config/store.go          # JSON config persistence (~/.config/teer/)
│   │   └── terminal/pty.go          # PTY abstraction (Unix/Windows)
│   └── service/
│       ├── session.go               # Session lifecycle: spawn PTY, stream I/O, resize
│       ├── workspace.go             # Workspace CRUD + persistence
│       ├── dialog.go                # Native dialog service
│       └── updater.go               # App update notifications
├── frontend/
│   └── src/
│       ├── application/             # App-level state and initialization
│       ├── domain/                  # Frontend domain models
│       ├── infrastructure/          # Wails binding wrappers
│       └── presentation/            # Svelte UI components
├── docs/PRD.md                      # Product Requirements Document
├── Taskfile.yml                     # Build and dev tasks
└── build/                           # Platform-specific build configs
```

---

## Getting Started

### Prerequisites

- Go 1.25+
- Node.js 18+
- [Wails v3 CLI](https://v3.wails.io/getting-started/installation)
- Linux (primary target) — macOS and Windows support planned

### Development

```bash
task dev
# or
wails3 dev -config ./build/config.yml
```

Hot-reload active for both frontend (Vite) and backend (Go).

### Build

```bash
task build
```

Produces a binary in `bin/`.

### Other tasks

```bash
task run            # run production build
task package        # package for distribution
task build:server   # headless HTTP server mode (no GUI)
task run:server     # run server mode
task build:docker   # build Docker image for server mode
task run:docker     # build and run Docker image
```

---

## Architecture

### Terminal I/O Flow

```
[xterm.js in Svelte]  --(keystroke)-->  WriteToSession(id, data)
        ^                                          |
        |                                          v
   Events.On(output)  <--(stream)--  Go: read PTY  -->  shell (bash/zsh)
```

Each session runs as a goroutine reading from its PTY, emitting `session:<id>:output` events to the frontend. Resize and write are separate Wails bindings.

### Data Model

```
Application
└── Workspace  (name, color, default cwd, env vars)
    └── Session[]  (name, shell, cwd, env, autoStart)
        └── Runtime state: PTY, status (running/exited), PID
```

Config is stored as JSON at `~/.config/teer/` with `0600` permissions. Terminal scrollback is not persisted in v1.

> Sessions are not preserved after Teer closes — PTY processes are killed on quit. For long-lived sessions (e.g. SSH), use a host-side `tmux` via `startupCommand`.

---

## UI Layout

```
┌───────────┬──────────────────────────────────────────────┐
│           │  [Tab: server] [Tab: worker] [Tab: db] [ + ]  │
│ Workspace │ ─────────────────────────────────────────────│
│  Sidebar  │                                               │
│           │                                               │
│ • Codemi  │            xterm.js terminal area             │
│ • Prod ●  │            (splittable into panes — P1)       │
│ • Sandbox │                                               │
│           │                                               │
│  [+ new]  │                                               │
└───────────┴──────────────────────────────────────────────┘
```

- **Left**: workspace list (dot indicator = active sessions running)
- **Top-right**: tab bar for terminals in active workspace
- **Center**: xterm.js terminal area

---

## Roadmap

| Milestone | Status | Scope |
|-----------|--------|-------|
| M0 — Scaffolding | ✅ Done | Wails v3 + Svelte TS setup |
| M1 — Single terminal | In progress | PTY integration, xterm.js, full I/O |
| M2 — Multi-terminal + tabs | Planned | Multiple sessions, tab bar, rename/close |
| M3 — Workspaces | Planned | Sidebar, CRUD, persistence |
| M4 — Automation + layout | Planned | autoStart, split pane |
| M5 — Polish | Planned | Shortcuts, command palette, themes |
| M6 — Cross-platform | Planned | Windows (ConPTY), packaging |

---

## Non-Goals (v1)

- Not a remote multiplexer (no tmux/screen on the server side)
- Not an IDE or code editor
- No cloud sync or multi-device support
- No real-time collaboration or shared sessions
- No plugin marketplace

---

## License

MIT
