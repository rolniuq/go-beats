# go-beats — Agent Guide

## Overview

A terminal-based lofi music player with internet radio streaming and a built-in
Pomodoro timer. Written in Go. Also has a macOS desktop app (Wails v2) whose
backend is complete but frontend is **empty** — the `desktop/frontend/dist/`
directory has no HTML/CSS/JS yet.

**Author:** rolniuq · **License:** MIT · **Version:** v0.1.0

## Tech Stack

| Concern          | Library                                                             |
| ---------------- | ------------------------------------------------------------------- |
| TUI Framework    | [Bubbletea](https://github.com/charmbracelet/bubbletea) (v1.3.10)   |
| TUI Styling      | [Lip Gloss](https://github.com/charmbracelet/lipgloss) (v1.1.0)     |
| Desktop GUI      | [Wails v2](https://wails.io/) (v2.11.0)                             |
| Audio Playback   | [Beep](https://github.com/gopxl/beep) (v2.1.1)                     |
| MP3 Decoding     | go-mp3 (hajimehoshi)                                                |
| Audio Output     | Oto v3 (ebitengine)                                                 |
| Task Runner      | [Task](https://taskfile.dev/) (Taskfile.yml)                        |

**Go version:** 1.26.1

## Architecture

### Internal Packages

```
internal/
├── audio/engine.go        — Local MP3 playback (scan, play, pause, volume, loop, auto-advance)
├── radio/
│   ├── stations.go        — 11 curated station definitions
│   ├── stations_test.go   — Station data integrity tests
│   └── player.go          — HTTP radio stream player (connect, reconnect, retry)
├── pomodoro/
│   ├── timer.go           — 25/5/15 pomodoro with phase transitions, callbacks
│   └── timer_test.go      — 14 test functions
├── notification/
│   └── sound.go           — Synthesized chime tones (C-E-G / G-E-C via Beep generators)
└── ui/
    └── tui.go             — ~885 line Bubbletea model (visualizer, all views, key handling)
```

### Entry Points

- **`cmd/go-beats/main.go`** — TUI app: CLI flags, audio init, runs Bubbletea program
- **`cmd/go-beats-desktop/main.go`** — Desktop (Wails): build tag `desktop`, thin wrapper

### Key Design Details

**Two independent audio engines:** Local MP3 (`audio.Engine`) and radio
(`radio.Player`) each manage their own `beep.Ctrl`, `effects.Volume`, and
speaker state. They share the global `speaker` package. Only one should be
active at a time — the `Model.SetMode()` method stops the other before
switching. **Do NOT run both simultaneously or you get double-audio.**

**Auto-advance (local):** Polled every 100ms in the TUI tick. When position
≥ duration-200ms, calls `Next()` (or `Play(index)` if loop is on).

**Auto-advance (radio):** After 3 failed reconnection attempts (2s/4s/8s
exponential backoff), moves to next station. After all stations fail, stops.

**Visualizer:** Fake animated bars — random smoothed values, no FFT/real
audio analysis. This is by design (term limitations).

**Pomodoro notifications:** `OnPhaseEnd` callback plays chimes via
`notification.PlayFocusEnd()` / `PlayBreakEnd()`. Runs in goroutine.

**Desktop app backend (`desktop/app.go`):** Contains complete `PlayerState`
DTO, `GetState()`, `GetTracks()`, `GetStations()`, all player/pomodoro
controls bound via Wails `Bind`. Frontend needs to be built — this is the
biggest unfinished surface.

## Desktop App State

**Backend is done.** `desktop/app.go` has:
- Full `PlayerState` JSON struct (mode, local player, radio player, pomodoro)
- `GetState()` returns everything frontend needs
- All controls: `PlayTrack`, `TogglePlay`, `NextTrack`, `PrevTrack`,
  `VolumeUp/Down`, `ToggleLoop`, `SetMode`, `PlayStation`, `RetryRadio`,
  `TogglePomodoro`, `PausePomodoro`, `SkipPomodoro`, `LoadMusicDirectory`,
  `BrowseMusicFolder`

**Frontend is empty.** `desktop/frontend/dist/` is an empty directory
(embedded via `//go:embed frontend/dist`). Needs HTML/JS/CSS that calls
`window.runtime.Call("GetState")` etc. No framework chosen — ready for
the frontend implementation.

**System tray is no-op.** `desktop/nativetray_darwin.go` exists but does
nothing due to CGO symbol conflicts between Wails and `fyne.io/systray`.
The app uses `HideWindowOnClose` with Dock icon reopening instead.

## Build System

Run `task` to see all commands. Key ones:

```
task dev             — go run ./cmd/go-beats ./music (fastest iteration)
task run             — build + run binary
task check           — fmt + vet + test (run before every push)
task test            — go test -v ./...
task lint            — golangci-lint (needs brew install golangci-lint)
task desktop-dev     — go run with Wails dev mode (needs frontend/dist assets)
task desktop-build   — CGO build with desktop tag
task desktop-app     — package as macOS .app bundle
task desktop-install — copy to /Applications
task music-gen       — generate fake MP3s for testing (needs ffmpeg)
```

CLI flags: `--radio`, `--station <N>`, `--list-stations`, `--version`, `[path]`

## Platform Notes

- **macOS:** Primary target. TUI + desktop both supported.
- **Linux:** TUI only. Needs `libasound2-dev` for ALSA.
- **Windows:** Not supported (no plans).
- **Desktop builds** need `CGO_LDFLAGS="-framework UniformTypeIdentifiers"`
  and build tag `desktop`.

## Rules & Conventions

### Code Patterns
- Follow existing code style (no comments, use built-in `min`/`max`, same libs)
- No new third-party dependencies without discussion
- Bubbletea models get their own file in `internal/ui/`
- Desktop-only code uses `//go:build desktop` tag
- Shared helpers go in `internal/util/` (no duplication)
- Prefer Go 1.21+ builtins over custom helpers

### Before Every Push
1. Run `task check` — fmt + vet + test must pass
2. Run `go clean -cache` if build behavior seems stale
3. Verify no build artifacts are staged (`dist/`, binaries, `coverage.out`)

### Repository Hygiene
- Do NOT commit: binaries, `dist/` contents, `coverage.*`, `.a` archives, IDE config
- Desktop frontend build output (`desktop/frontend/dist/*`) is committed as
  embed placeholder — only commit real assets after review
- The `go-beats-desktop` binary is CGO-linked; always rebuild on macOS SDK updates

## Git Workflow

- `main` is protected — all changes via feature branches → PR → squash merge
- Branch naming: `<type>/ticket-<N>-<description>`
- Commit format: `<type>(<scope>): <message>`
- Run `task check` + `task clean` before every push

## Known Limitations & Gotchas

1. **No shuffle mode** — `audio.Engine` has a `shuffle bool` field but it's
   never exposed in the UI or used.
2. **Volume is logarithmic** — mapped from -7..5 dB to 0-100% linearly.
3. **TUI uses `tea.WithAltScreen()`** — switches to alternate screen buffer.
4. **Radio stations may break** — URLs are public streams that change
   occasionally. Tests validate URL format but not liveness.
5. **Desktop frontend is TBD** — the biggest gap in the project.
6. **No session persistence** — volume, loop, last station/track not saved.
7. **No cover art** — audio engine doesn't parse ID3/metadata beyond filename.
