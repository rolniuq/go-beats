# Design: macOS Now Playing Integration + Codebase Cleanup

## Overview

Two goals:
1. Add **macOS Now Playing** (Lock Screen / Control Center) to the Wails
   desktop app via `MPNowPlayingInfoCenter` + `MPRemoteCommandCenter`.
2. **Clean up the codebase** — remove dead code, deduplicate helpers, slim
   dependencies, enable unused features (shuffle).

---

## 1. macOS Now Playing

### What it does

When a track or radio station plays in the Desktop app, macOS shows it on the
Lock Screen and in Control Center's Now Playing widget. Users can play, pause,
skip, and see progress without bringing the app to front.

### Architecture

```
desktop/nowplaying/
├── nowplaying.go            — Interface + New()
├── nowplaying_darwin.go     — CGO bridge (build tag: desktop)
├── bridge.m                 — ObjC implementation (MediaPlayer framework)
├── bridge.h                 — ObjC header
└── nowplaying_others.go     — Noop fallback (no build tag)
```

### Interface

```go
type Controller interface {
    SetPlaying(title string, duration time.Duration)
    SetPaused()
    SetStopped()
    SetProgress(position time.Duration)
    SetCommandHandler(handler func(Command))
}
```

Where `Command` is:
```go
type Command int
const (
    CmdPlay     Command = iota
    CmdPause
    CmdTogglePlayPause
    CmdNextTrack
    CmdPreviousTrack
    CmdChangePlaybackPosition
)
```

### Integration points in `App` (`desktop/app.go`)

| Lifecycle         | What to add                                          |
|-------------------|------------------------------------------------------|
| `Startup()`       | Create `nowplaying.Controller`, wire `OnPhaseEnd`    |
| `GetState()`      | Already has all the data — no change                 |
| `PlayTrack()`     | After `engine.Play()` → `nowplaying.SetPlaying(...)` |
| `TogglePlay()`    | After play/pause → update nowplaying state           |
| `NextTrack()`     | After advance → update nowplaying                    |
| `PrevTrack()`     | Same                                                  |
| `pomodoroTicker()` | Add `nowplaying.SetProgress(position)` call           |
| `PlayStation()`   | SetPlaying with station name (no duration for radio) |
| `Shutdown()`      | nowplaying.SetStopped()                               |

### ObjC Implementation (`bridge.m`)

- Links `MediaPlayer.framework` via `#cgo LDFLAGS: -framework MediaPlayer`
- `MPNowPlayingInfoCenter.defaultCenter` → set `nowPlayingInfo` dictionary
  - `MPMediaItemPropertyTitle` — track or station name
  - `MPMediaItemPropertyArtist` — "go-beats"
  - `MPMediaItemPropertyAlbumTitle` — "Lofi Beats"
  - `MPMediaItemPropertyPlaybackDuration` — duration in seconds
  - `MPNowPlayingInfoPropertyElapsedPlaybackTime` — current position
  - `MPNowPlayingInfoPropertyPlaybackRate` — 1.0 playing, 0.0 paused
  - `MPMediaItemPropertyArtwork` — app icon (from `iconfile.icns`)
- `MPRemoteCommandCenter.sharedCommandCenter` → register handlers
  - `playCommand`, `pauseCommand`, `togglePlayPauseCommand`
  - `nextTrackCommand`, `previousTrackCommand`
  - `changePlaybackPositionCommand` (seeking)
- Handlers call Go callback via `cgo.Handle`

### Artwork

Use embedded `build/darwin/iconfile.icns` as default album art. Convert to
`NSImage` and set as `MPMediaItemPropertyArtwork`. The `iconfile.icns` is
already bundled in the `.app` — reference it from the bundle resource path.

---

## 2. Codebase Cleanup

### Priority order (do first = highest impact)

#### P1: Remove dead `fyne.io/systray` dependency

**Problem:** `go.mod` requires `fyne.io/systray v1.12.0` but **no Go source
imports it**. It was left over from an attempted system tray implementation
(`nativetray_darwin.go` is a no-op). This dep pulls `godbus/dbus/v5` and
other transitive junk.

**Fix:** `go mod tidy` removes it automatically. Keep `nativetray_darwin.go`
as-is (it only mentions systray in comments, no import).

**Size impact:** Removes ~5 transitive deps including `godbus/dbus/v5`.

#### P2: Remove dead `shuffle` field from `audio.Engine`

**Problem:** `audio.Engine.shuffle` (engine.go:41) is declared but **never
read or written** outside the struct definition. Dead field.

**Fix:** Delete the field. If shuffle is implemented later, add it back
properly with full wiring.

#### P3: Remove redundant `min()` helper from `tui.go`

**Problem:** Custom `min()` at tui.go:880 is unnecessary. Go 1.21+ has
a built-in `min` (project uses Go 1.25.3).

**Fix:** Delete the function, replace all `min(a, b)` calls with
built-in `min(a, b)`. No behavior change.

#### P4: Deduplicate `formatDuration()` — share via `internal/`

**Problem:** Identical `formatDuration()` function exists in:
- `internal/ui/tui.go:871-878`
- `desktop/app.go:498-505`

**Fix:** Move to `internal/util/format.go` (new package). Both import it.

#### P5: Deduplicate `hasMP3Files()` — share via `internal/`

**Problem:** Identical `hasMP3Files()` function in:
- `cmd/go-beats/main.go:142-158`
- `desktop/app.go:507-518`

**Fix:** Move to `internal/util/format.go` alongside `formatDuration`.

#### P6: Misc dead/duplicate code

| Location              | Issue                                | Fix               |
|-----------------------|--------------------------------------|-------------------|
| `main.go:25`          | `fmt.Println("🎵 Starting...")`      | Remove (go-beats starts silently) |
| `main.go:139`         | `fmt.Println("👋 Thanks...")`        | Remove (already clean exit)  |

---

## 3. Build & Testing

### Build verification (before commit)

```bash
task check              # fmt + vet + test
task desktop-build      # ensure Now Playing compiles
```

### Manual testing (desktop)

1. `task desktop-dev` (needs frontend dist — may need placeholder file)
2. Play a track → Lock Screen should show "go-beats - Ambient Chill"
3. Press play/pause on Control Center → app responds
4. Press skip → next track plays

### Risk

- `MPRemoteCommandCenter` handlers run on the main thread (Cocoa requirement).
  The Go callback uses `cgo.Handle` which is safe.
- Wails uses its own NSApplication. MPRemoteCommandCenter hooks into
  the shared command center, which works alongside Wails.
- If Wails changes its Cocoa lifecycle handling in a future version,
  the ObjC bridge may need updates.

---

## 4. Files Changed

| File                                | Action   | Reason                        |
|-------------------------------------|----------|-------------------------------|
| `desktop/nowplaying/nowplaying.go`  | CREATE   | Interface + constructor       |
| `desktop/nowplaying/nowplaying_darwin.go` | CREATE | CGO bridge for macOS    |
| `desktop/nowplaying/bridge.m`       | CREATE   | ObjC implementation           |
| `desktop/nowplaying/bridge.h`       | CREATE   | ObjC header                   |
| `desktop/nowplaying/nowplaying_others.go` | CREATE | Noop for non-macOS      |
| `desktop/app.go`                    | MODIFY   | Wire nowplaying into lifecycle|
| `internal/util/format.go`           | CREATE   | Shared formatDuration + hasMP3Files |
| `internal/ui/tui.go`                | MODIFY   | Use built-in min, import formatDuration |
| `cmd/go-beats/main.go`              | MODIFY   | Remove startup/exit prints, import hasMP3Files |
| `internal/audio/engine.go`          | MODIFY   | Remove unused shuffle field   |
| `go.mod` / `go.sum`                 | MODIFY   | `go mod tidy` removes systray |
