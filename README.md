# go-beats

> A terminal-based lofi music player with internet radio streaming and a built-in pomodoro timer.

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey)

```
   ██████╗  ██████╗       ██████╗ ███████╗ █████╗ ████████╗███████╗
  ██╔════╝ ██╔═══██╗      ██╔══██╗██╔════╝██╔══██╗╚══██╔══╝██╔════╝
  ██║  ███╗██║   ██║█████╗██████╔╝█████╗  ███████║   ██║   ███████╗
  ██║   ██║██║   ██║╚════╝██╔══██╗██╔══╝  ██╔══██║   ██║   ╚════██║
  ╚██████╔╝╚██████╔╝      ██████╔╝███████╗██║  ██║   ██║   ███████║
   ╚═════╝  ╚═════╝       ╚═════╝ ╚══════╝╚═╝  ╚═╝   ╚═╝   ╚══════╝
                    ☕ lofi beats to relax/study to
```

---

## Features

- **Local MP3 Playback** — Play your own `.mp3` files with progress bar, loop, and auto-advance
- **Internet Radio** — 11 curated lofi/ambient/synthwave stations, always streaming
- **Pomodoro Timer** — Built-in 25/5/15 focus timer that runs alongside your music
- **Beautiful TUI** — Animated audio visualizer, Rosé Pine color theme, full keyboard control
- **Auto-Reconnect** — Radio streams reconnect automatically on failure (3 retries with exponential backoff)
- **Graceful Fallback** — No local music? App automatically starts in radio mode

---

## Quick Start

### Prerequisites

- **Go 1.25+** (for `go install`)
- **[Task](https://taskfile.dev/)** (task runner, for development) — `brew install go-task`
- **ffmpeg** (optional, for generating sample tracks) — `brew install ffmpeg`

## Install

### macOS Desktop App (one command)

```bash
curl -fsSL https://raw.githubusercontent.com/rolniuq/go-beats/main/install.sh | bash
```

This downloads the latest release, installs **Go-Beats.app** to `/Applications`, and makes it available in Launchpad and Spotlight.

### Via Go (TUI version, requires Go 1.25+)

```bash
go install github.com/rolniuq/go-beats/cmd/go-beats@latest
```

### Via GitHub Releases

Download the latest binary from [Releases](https://github.com/rolniuq/go-beats/releases), extract, and run:

```bash
# TUI (terminal)
tar -xzf go-beats_*.tar.gz
./go-beats --radio

# Desktop app (macOS)
tar -xzf Go-Beats_*_darwin_arm64.app.tar.gz
cp -R Go-Beats.app /Applications/
open -a Go-Beats
```

### From Source

```bash
# Clone the repo
git clone git@github.com:rolniuq/go-beats.git
cd go-beats

# Install dependencies
task deps

# Generate sample tracks (optional)
task music-gen

# Build and run (TUI)
task run

# Build and install desktop app to /Applications
task desktop-install
```

### Or run directly with Go

```bash
# TUI
go run ./cmd/go-beats ./music

# Desktop (dev mode)
task desktop-dev
```

---

## Usage

### CLI Flags

| Flag | Description |
|------|-------------|
| `--radio` | Start directly in radio mode |
| `--station <index>` | Auto-play a specific station (implies `--radio`) |
| `--list-stations` | List all available radio stations and exit |
| `--version` | Print version and exit |
| `[path]` | Music directory path (default: `./music`) |

### Examples

```bash
# Play local music from a directory
./go-beats ~/Music/lofi

# Start in radio mode
./go-beats --radio

# Auto-play station #5 (SomaFM Groove Salad)
./go-beats --station 5

# List all stations
./go-beats --list-stations
```

---

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Space` | Play / Pause |
| `n` | Next track / station |
| `p` | Previous track / station |
| `Tab` | Switch between Local and Radio mode |
| `+` / `-` | Volume up / down |
| `l` | Toggle loop mode (local only) |
| `Enter` | Play selected track / station / retry connection |
| `↑` `↓` / `k` `j` | Navigate track / station list |
| `t` | Start / stop pomodoro timer |
| `T` | Pause pomodoro timer |
| `s` | Skip pomodoro phase |
| `?` | Toggle help |
| `q` / `Ctrl+C` | Quit |

---

## Radio Stations

11 curated stations for coding, studying, and chilling:

| # | Station | Genre |
|---|---------|-------|
| 0 | Lofi Girl | lofi hip-hop |
| 1 | Chillhop | chillhop |
| 2 | Box Lofi | lofi |
| 3 | LofiRadio24 | lofi hip-hop |
| 4 | Nightride FM | synthwave |
| 5 | Plaza One | vaporwave |
| 6 | SomaFM Groove Salad | ambient/downtempo |
| 7 | SomaFM DEF CON | electronic |
| 8 | SomaFM Drone Zone | ambient/drone |
| 9 | SomaFM Deep Space One | space ambient |
| 10 | SomaFM Lush | electronic/female vocal |

---

## Pomodoro Timer

Built-in focus timer using the classic Pomodoro Technique:

| Phase | Duration |
|-------|----------|
| Focus | 25 minutes |
| Short Break | 5 minutes |
| Long Break (every 4th) | 15 minutes |

Press `t` to start, `T` to pause, `s` to skip a phase. The timer runs alongside your music and auto-advances through phases.

---

## Available Tasks

Run `task` to see all commands:

| Command | Description |
|---------|-------------|
| `task run` | Build and run go-beats |
| `task dev` | Run with `go run` (no binary) |
| `task build` | Build binary for current platform |
| `task build-all` | Cross-compile for macOS (arm64 + amd64) and Linux |
| `task desktop-build` | Build Go-Beats desktop binary |
| `task desktop-app` | Package as macOS .app bundle |
| `task desktop-run` | Build and run desktop app |
| `task desktop-install` | Install Go-Beats.app to /Applications |
| `task desktop-dev` | Run desktop app in dev mode |
| `task check` | Run all quality checks (fmt + vet + test) |
| `task test` | Run all tests |
| `task test-cover` | Run tests with coverage report |
| `task fmt` | Format all Go source files |
| `task vet` | Run go vet |
| `task lint` | Run golangci-lint |
| `task deps` | Download and tidy dependencies |
| `task deps-update` | Update all dependencies to latest |
| `task music-gen` | Generate sample test tracks (requires ffmpeg) |
| `task install` | Install go-beats to `$GOPATH/bin` |
| `task clean` | Remove build artifacts |

---

## Project Structure

```
go-beats/
├── cmd/
│   ├── go-beats/
│   │   └── main.go              # TUI entry point — CLI flags, wiring
│   └── go-beats-desktop/
│       └── main.go              # Desktop (Wails) entry point
├── desktop/
│   ├── main.go                  # Wails app bootstrap
│   ├── app.go                   # Desktop backend (state, controls)
│   └── frontend/dist/           # HTML/CSS/JS frontend
├── internal/
│   ├── audio/
│   │   └── engine.go            # Local MP3 playback engine
│   ├── radio/
│   │   ├── player.go            # Internet radio HTTP stream player
│   │   ├── stations.go          # Curated station registry
│   │   └── stations_test.go     # Station tests
│   ├── pomodoro/
│   │   ├── timer.go             # Pomodoro timer logic
│   │   └── timer_test.go        # Timer tests
│   └── ui/
│       └── tui.go               # Bubbletea TUI (visualizer, controls, rendering)
├── build/darwin/                 # macOS .app bundle assets (Info.plist, icons)
├── music/                       # Your .mp3 files go here
├── install.sh                   # One-command installer for macOS
├── Taskfile.yml                 # Task runner config
├── CONTRIBUTING.md              # Git workflow & PR guide
└── go.mod
```

---

## Tech Stack

| Component | Library |
|-----------|---------|
| TUI Framework | [Bubbletea](https://github.com/charmbracelet/bubbletea) |
| TUI Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| Desktop GUI | [Wails v2](https://wails.io/) |
| Audio Playback | [Beep](https://github.com/gopxl/beep) |
| MP3 Decoding | [go-mp3](https://github.com/hajimehoshi/go-mp3) |
| Audio Output | [Oto](https://github.com/ebitengine/oto) |

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full guide, including:

- Ticket pickup rules
- Branch naming convention
- Commit message format
- PR template & review process

**TL;DR:**
1. Pick an open ticket → assign yourself
2. Branch from `main` → `feat/ticket-<N>-description`
3. Code → `task check` → push → create PR with `Closes #N`
4. Wait for tech lead review → merge

---

## License

MIT
