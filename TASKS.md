# 📋 GO-BEATS — Sprint Tickets

> **Project:** go-beats — Terminal lofi music player with Pomodoro timer
> **Goal:** Add internet radio streaming support (lofi/chill stations)
> **Branch convention:** `feat/<ticket-id>-short-description`

---

## 🏗 Architecture Overview

```
internal/
├── audio/       # Local file playback engine (DONE ✅)
├── radio/       # NEW — Internet radio streaming
│   ├── stations.go   # Station registry (DONE ✅ — by tech lead)
│   └── player.go     # Stream player (DONE ✅ — by tech lead)
├── pomodoro/    # Pomodoro timer (DONE ✅)
└── ui/          # Bubbletea TUI
    └── tui.go   # Main TUI — needs radio mode
cmd/
└── main.go      # Entry point — needs --radio flag
```

---

## 🎫 TICKET-1: Integrate Radio Player into TUI (Frontend)

**Assignee:** Claude
**Priority:** HIGH
**Branch:** `feat/ticket-1-radio-tui`
**Status:** ✅ COMPLETED
**Depends on:** TICKET-2 (but can stub radio player for now)

### Description
Add a **radio mode** tab to the TUI so users can switch between local files and internet radio stations.

### Acceptance Criteria
- [ ] Add a `Tab` key binding to toggle between **Local** and **Radio** mode
- [ ] In Radio mode, show the station list instead of the track list
- [ ] Station list shows: station name, genre, description
- [ ] Currently playing station highlighted with `📻` icon
- [ ] Cursor navigation (↑/↓/j/k) and Enter to select a station
- [ ] `n/p` keys switch to next/prev station in radio mode
- [ ] Space bar pauses/resumes the radio stream
- [ ] Volume controls (+/-) work for radio too
- [ ] Show connection status: "Connecting...", "Playing", "Error: ..."
- [ ] Progress bar area shows `📻 LIVE` instead of time progress when in radio mode
- [ ] Pomodoro timer still works alongside radio mode
- [ ] Help section (`?`) updated with radio keybindings

### Files to Modify
- `internal/ui/tui.go` — Main changes here

### Key Implementation Notes
```go
// Add to Model struct:
type Mode int
const (
    ModeLocal Mode = iota
    ModeRadio
)

// Add fields:
mode         Mode
radioPlayer  *radio.Player
radioCursor  int

// New key binding:
Tab: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch mode"))
```

### How to Test
1. `task build && ./go-beats`
2. Press `Tab` → should switch to Radio view
3. Navigate stations with `j/k`, press `Enter` → should connect & play
4. Press `Space` → should pause/resume
5. Press `Tab` again → back to Local mode, local playback unaffected

---

## 🎫 TICKET-2: Wire Radio Player into Main & Handle Speaker Sharing

**Assignee:** OpenCode (Dev 2)
**Priority:** HIGH
**Branch:** `feat/ticket-2-radio-engine`
**Status:** ✅ DONE (OpenCode / Dev 2)
**Depends on:** None (radio package already scaffolded)

### Description
Wire the radio player into `main.go`, handle the `--radio` CLI flag, and ensure the audio speaker is properly shared between local player and radio player (only one should play at a time).

### Acceptance Criteria
- [x] Add `--radio` flag to start directly in radio mode
- [x] Add `--station <index>` flag to auto-play a specific station on startup
- [x] Add `--list-stations` flag to print available stations and exit
- [x] `NewModel()` in TUI now accepts both `*audio.Engine` and `*radio.Player`
- [x] When switching from local → radio, local playback stops
- [x] When switching from radio → local, radio stream stops
- [x] Speaker is initialized once, shared by both players
- [x] Graceful shutdown: stop both radio stream and local playback on quit
- [x] If no `music/` directory exists and no `--radio` flag, default to radio mode instead of exiting with error

### Files to Modify
- `cmd/main.go` — CLI flags and startup logic
- `internal/ui/tui.go` — Update `NewModel()` signature
- `internal/radio/player.go` — Minor tweaks if needed

### Key Implementation Notes
```go
// cmd/main.go — add flag parsing:
import "flag"

radioMode := flag.Bool("radio", false, "Start in radio mode")
stationIdx := flag.Int("station", -1, "Auto-play station index")
listStations := flag.Bool("list-stations", false, "List available stations")
flag.Parse()

// NewModel signature change:
func NewModel(engine *audio.Engine, radioPlayer *radio.Player) Model
```

### How to Test
```bash
# List stations
./go-beats --list-stations

# Start in radio mode
./go-beats --radio

# Auto-play station 0 (Lofi Girl)
./go-beats --radio --station 0

# Normal mode (local files)
./go-beats

# No music dir, should fallback to radio
rm -rf music/ && ./go-beats
```

---

## 🎫 TICKET-3: Add Unit Tests for Radio & Pomodoro

**Assignee:** Either dev (pick up when done with above)
**Priority:** MEDIUM
**Branch:** `feat/ticket-3-tests`

### Description
Add unit tests for the pomodoro timer and radio station registry.

### Acceptance Criteria
- [ ] `internal/pomodoro/timer_test.go` — Test phase transitions, tick, pause, skip, sessions count
- [ ] `internal/radio/stations_test.go` — Test default stations are valid (non-empty URLs, names)
- [ ] Tests pass with `task test`
- [ ] At least 70% coverage on pomodoro package

### How to Test
```bash
task test
task test-cover
```

---

## 🎫 TICKET-4: Error Handling & Stream Reconnection

**Assignee:** Either dev (pick up when done with above)
**Priority:** MEDIUM
**Branch:** `feat/ticket-4-reconnect`

### Description
Internet radio streams can drop. Add auto-reconnect logic.

### Acceptance Criteria
- [ ] If a stream disconnects, show "Reconnecting..." in the TUI
- [ ] Auto-retry up to 3 times with exponential backoff (2s, 4s, 8s)
- [ ] After 3 failures, show error and stop trying
- [ ] User can press `Enter` to manually retry after failure
- [ ] No crash on network loss — handle all error paths gracefully

### Files to Modify
- `internal/radio/player.go`
- `internal/ui/tui.go`

---

## 📌 Notes for Devs

### Quick Start
```bash
# Install deps
task deps

# Generate test music (if no mp3s)
task music-gen

# Build & run
task run

# Run checks
task check
```

### Git Workflow
1. Branch from `main`: `git checkout -b feat/ticket-X-description`
2. Make small commits with clear messages
3. Run `task check` before pushing
4. Open PR, tag the other dev for review

### Key Dependencies
| Package | Purpose |
|---|---|
| `github.com/charmbracelet/bubbletea` | TUI framework |
| `github.com/charmbracelet/lipgloss` | TUI styling |
| `github.com/charmbracelet/bubbles` | TUI components |
| `github.com/gopxl/beep/v2` | Audio playback |

### Radio Streams Info
- All streams are **public MP3/Icecast** streams — no API keys needed
- Station list is in `internal/radio/stations.go`
- Some streams may go offline — that's expected, handle gracefully
