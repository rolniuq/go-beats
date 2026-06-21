# macOS Now Playing + Codebase Cleanup — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add macOS Lock Screen / Control Center Now Playing to the Wails desktop app, and clean up dead code, duplicates, and unnecessary dependencies.

**Architecture:** A new `desktop/nowplaying/` package provides a cross-platform `Controller` interface backed by ObjC on macOS (via CGO + MediaPlayer framework) and a noop on other platforms. The `App` struct in `desktop/app.go` calls it at every lifecycle point (play, pause, skip, tick). Six cleanup items (P1–P6) remove dead deps, dead fields, redundant helpers, and startup noise.

**Tech Stack:** Go 1.25.3, CGO (desktop build only), macOS MediaPlayer.framework, Wails v2.

## Global Constraints

- All cleanup items (P1–P6) from the spec are mandatory
- Now Playing is desktop-only (build tag `desktop`) — TUI unchanged
- Must not break `task check` (fmt + vet + test)
- Must not break `task desktop-build`
- Must not introduce any new third-party Go dependency
- Use built-in `min` (Go 1.21+) — no custom helpers
- All paths are relative to project root unless absolute

---

### Task 1: Create shared `internal/util` package

**Files:**
- Create: `internal/util/util.go`

**Interfaces:**
- Produces: `util.FormatDuration(time.Duration) string`
- Produces: `util.HasMP3Files(string) bool`

- [ ] **Step 1: Create `internal/util/util.go`**

```go
package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	mins := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}

func HasMP3Files(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".mp3") {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/util/`
Expected: no error

- [ ] **Step 3: Commit**

```bash
git add internal/util/util.go
git commit -m "refactor: create internal/util package with FormatDuration and HasMP3Files"
```

---

### Task 2: Use shared `util` functions everywhere + remove custom `min` + remove startup prints

**Files:**
- Modify: `internal/ui/tui.go`
- Modify: `cmd/go-beats/main.go`
- Modify: `desktop/app.go`

**Consumes:** `util.FormatDuration`, `util.HasMP3Files`

- [ ] **Step 1: Update `internal/ui/tui.go`**

Replace `formatDuration` function usage with `util.FormatDuration`. Delete the `formatDuration` function and the custom `min` function. Replace all `min(a, b)` calls with the built-in `min(a, b)`.

Add import: `"github.com/rolniuq/go-beats/internal/util"`

Change `renderProgressBar` (line ~684):
```go
posStr := util.FormatDuration(pos)
durStr := util.FormatDuration(dur)
```

Change all calls to `min(` to use the built-in (just remove the custom function declaration). The built-in `min` is already available in Go 1.25.

Delete `formatDuration` function (lines ~871-878):
```go
// DELETE this entire block
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	mins := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}
```

Delete `min` function (lines ~880-885):
```go
// DELETE this entire block
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
```

Now `fmt` import may be unused (check if `fmt` is still needed elsewhere in the file — it's used in many places, so keep it).

- [ ] **Step 2: Run test to verify TUI still compiles**

Run: `go build ./internal/ui/`
Expected: no errors

- [ ] **Step 3: Update `cmd/go-beats/main.go`**

Add import: `"github.com/rolniuq/go-beats/internal/util"`

Replace `hasMP3Files` function body with call to `util.HasMP3Files`:

Change line ~97:
```go
hasMP3, scanErr := hasMP3Files(absDir)
```
to:
```go
hasMP3, scanErr := util.HasMP3Files(absDir), nil
```

Wait, this doesn't work because `util.HasMP3Files` doesn't return an error. Let me re-check the original code:

```go
hasMP3, scanErr := hasMP3Files(absDir)
if scanErr != nil {
```

The original `hasMP3Files` returns `(bool, error)`. The new `util.HasMP3Files` returns `bool`. I need to adjust the surrounding code.

In main.go, change lines 96-115:
```go
} else {
    if err := engine.ScanDirectory(absDir); err != nil {
        if !startInRadio {
            hasMP3, scanErr := hasMP3Files(absDir)
            if scanErr != nil {
                fmt.Fprintf(os.Stderr, "Error scanning music: %v\n", err)
                os.Exit(1)
            }
            if !hasMP3 {
                fmt.Printf("No local tracks found in %s\n", absDir)
                fmt.Println("Starting in radio mode. Use --list-stations to browse stations.")
                startInRadio = true
            } else {
                fmt.Fprintf(os.Stderr, "Error scanning music: %v\n", err)
                os.Exit(1)
            }
        }
    } else {
```

to:
```go
} else {
    if err := engine.ScanDirectory(absDir); err != nil {
        if !startInRadio {
            if !util.HasMP3Files(absDir) {
                fmt.Printf("No local tracks found in %s\n", absDir)
                fmt.Println("Starting in radio mode. Use --list-stations to browse stations.")
                startInRadio = true
            } else {
                fmt.Fprintf(os.Stderr, "Error scanning music: %v\n", err)
                os.Exit(1)
            }
        }
    } else {
```

Delete the `hasMP3Files` function at the bottom of main.go (lines 142-158).

Also **remove startup/exit prints** (P6):
- Delete line 25: `fmt.Println("🎵 Starting go-beats...")`
- Delete line 139: `fmt.Println("\n👋 Thanks for chilling with go-beats!")`

- [ ] **Step 4: Run test to verify main.go compiles**

Run: `go build ./cmd/go-beats`
Expected: no errors

- [ ] **Step 5: Update `desktop/app.go`**

Add import: `"github.com/rolniuq/go-beats/internal/util"`

Replace `formatDuration` function usage with `util.FormatDuration`.

In `GetState()` (line ~209-211):
```go
state.Position = formatDuration(pos)
state.Duration = formatDuration(dur)
```
to:
```go
state.Position = util.FormatDuration(pos)
state.Duration = util.FormatDuration(dur)
```

In `GetTracks()` (line ~268):
```go
Duration: formatDuration(t.Duration),
```
to:
```go
Duration: util.FormatDuration(t.Duration),
```

Delete `formatDuration` function (lines ~498-505) — the function definition.

Replace `hasMP3Files` function (lines ~507-518) — change to call `util.HasMP3Files`:

Line ~487:
```go
if hasMP3Files(dir) {
```
to:
```go
if util.HasMP3Files(dir) {
```

Delete `hasMP3Files` function definition (lines 507-518).

- [ ] **Step 6: Run test to verify desktop compiles**

Run: `go vet ./desktop/...`
Expected: no errors

- [ ] **Step 7: Run full check**

Run: `task check`
Expected: PASS (fmt + vet + test)

- [ ] **Step 8: Commit**

```bash
git add internal/ui/tui.go cmd/go-beats/main.go desktop/app.go
git commit -m "refactor: deduplicate formatDuration and hasMP3Files into internal/util"
```

---

### Task 3: Remove dead `fyne.io/systray` dependency (P1)

**Files:**
- Modify: `go.mod` (will be handled by `go mod tidy`)

- [ ] **Step 1: Run `go mod tidy` to remove unused deps**

Run: `go mod tidy`
Expected: `fyne.io/systray` removed from `go.mod` and `go.sum`

- [ ] **Step 2: Verify TUI and Desktop both build**

Run: `go build ./cmd/go-beats`
Expected: no error

Run: `go vet ./...`
Expected: no errors about missed deps

- [ ] **Step 3: Verify `fyne.io/systray` is gone**

Run: `grep fyne.io/systray go.mod`
Expected: no output

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: remove unused fyne.io/systray dependency"
```

---

### Task 4: Remove unused `shuffle` field from `audio.Engine` (P2)

**Files:**
- Modify: `internal/audio/engine.go:41`

- [ ] **Step 1: Remove the `shuffle` field**

Line 41:
```go
	shuffle bool
```
Delete this line.

- [ ] **Step 2: Verify compile**

Run: `go build ./internal/audio/`
Expected: no error

- [ ] **Step 3: Run tests**

Run: `go test ./internal/audio/...`
Expected: PASS (no audio tests exist, but verify no build breakage)

- [ ] **Step 4: Commit**

```bash
git add internal/audio/engine.go
git commit -m "refactor: remove unused shuffle field from audio.Engine"
```

---

### Task 5: Create `desktop/nowplaying` package (noop + macOS implementation)

**Files:**
- Create: `desktop/nowplaying/nowplaying.go`
- Create: `desktop/nowplaying/nowplaying_others.go`
- Create: `desktop/nowplaying/nowplaying_darwin.go`
- Create: `desktop/nowplaying/bridge.h`
- Create: `desktop/nowplaying/bridge.m`

**Interfaces:**
- Produces: `nowplaying.Controller` interface
- Produces: `nowplaying.New() Controller`
- Produces: `nowplaying.Command` enum
- Consumes: CGO + macOS MediaPlayer.framework

- [ ] **Step 1: Create `desktop/nowplaying/nowplaying.go`** — interface + constructor

```go
// Package nowplaying provides macOS Now Playing (Lock Screen / Control Center)
// integration for the go-beats desktop app.
package nowplaying

import "time"

// Command represents a remote control action from the system.
type Command int

const (
	CmdPlay Command = iota
	CmdPause
	CmdTogglePlayPause
	CmdNextTrack
	CmdPreviousTrack
	CmdChangePlaybackPosition
)

// Controller manages the Now Playing info on macOS.
type Controller interface {
	// SetPlaying updates the Now Playing info for a currently playing track.
	SetPlaying(title string, duration time.Duration)
	// SetStation sets Now Playing info for a radio station (no known duration).
	SetStation(name string)
	// SetPaused marks playback as paused (shows pause state on Lock Screen).
	SetPaused()
	// SetStopped removes Now Playing info.
	SetStopped()
	// SetProgress updates the elapsed playback time.
	SetProgress(position time.Duration)
	// SetCommandHandler registers a callback for remote commands.
	SetCommandHandler(handler func(Command, float64))
}

// New creates a platform-appropriate Controller.
func New() Controller {
	return &noopController{}
}
```

- [ ] **Step 2: Create `desktop/nowplaying/nowplaying_others.go`** — noop fallback

```go
//go:build !desktop

package nowplaying

import "time"

type noopController struct{}

func (n *noopController) SetPlaying(title string, duration time.Duration) {}
func (n *noopController) SetStation(name string)                         {}
func (n *noopController) SetPaused()                                     {}
func (n *noopController) SetStopped()                                    {}
func (n *noopController) SetProgress(position time.Duration)             {}
func (n *noopController) SetCommandHandler(handler func(Command, float64)) {}
```

- [ ] **Step 3: Create `desktop/nowplaying/bridge.h`** — CGO header for ObjC bridge

```objc
#import <MediaPlayer/MediaPlayer.h>
#import <Cocoa/Cocoa.h>

void nowplaying_setPlaying(const char* title, double duration);
void nowplaying_setStation(const char* name);
void nowplaying_setPaused(void);
void nowplaying_setStopped(void);
void nowplaying_setProgress(double position);
void nowplaying_setCommandHandler(void);
void nowplaying_clearCommandHandler(void);
```

- [ ] **Step 4: Create `desktop/nowplaying/bridge.m`** — ObjC implementation

```objc
#import "bridge.h"

// Declare the Go callback (exported from Go)
extern void goNowPlayingCommand(int command, double value);

void nowplaying_setPlaying(const char* title, double duration) {
    MPNowPlayingInfoCenter* center = [MPNowPlayingInfoCenter defaultCenter];
    
    NSString* titleStr = [NSString stringWithUTF8String:title];
    
    // Get app icon for artwork
    NSImage* appIcon = [NSImage imageNamed:@"AppIcon"];
    if (appIcon == nil) {
        appIcon = [NSImage imageNamed:@"NSApplicationIcon"];
    }
    MPMediaItemArtwork* artwork = nil;
    if (appIcon) {
        artwork = [[MPMediaItemArtwork alloc] initWithBoundsSize:appIcon.size
                                                   requestHandler:^NSImage* (CGSize size) {
            return appIcon;
        }];
    }
    
    NSMutableDictionary* info = [NSMutableDictionary dictionary];
    info[MPMediaItemPropertyTitle] = titleStr;
    info[MPMediaItemPropertyArtist] = @"go-beats";
    info[MPMediaItemPropertyAlbumTitle] = @"Lofi Beats";
    info[MPNowPlayingInfoPropertyPlaybackRate] = @(1.0);
    if (duration > 0) {
        info[MPMediaItemPropertyPlaybackDuration] = @(duration);
    }
    if (artwork) {
        info[MPMediaItemPropertyArtwork] = artwork;
    }
    
    center.nowPlayingInfo = info;
}

void nowplaying_setStation(const char* name) {
    // Radio has no known duration, so we still set it but without the duration key
    MPNowPlayingInfoCenter* center = [MPNowPlayingInfoCenter defaultCenter];
    
    NSString* nameStr = [NSString stringWithUTF8String:name];
    
    NSImage* appIcon = [NSImage imageNamed:@"AppIcon"];
    if (appIcon == nil) {
        appIcon = [NSImage imageNamed:@"NSApplicationIcon"];
    }
    MPMediaItemArtwork* artwork = nil;
    if (appIcon) {
        artwork = [[MPMediaItemArtwork alloc] initWithBoundsSize:appIcon.size
                                                   requestHandler:^NSImage* (CGSize size) {
            return appIcon;
        }];
    }
    
    NSMutableDictionary* info = [NSMutableDictionary dictionary];
    info[MPMediaItemPropertyTitle] = nameStr;
    info[MPMediaItemPropertyArtist] = @"go-beats";
    info[MPMediaItemPropertyAlbumTitle] = @"Radio";
    info[MPNowPlayingInfoPropertyPlaybackRate] = @(1.0);
    info[MPMediaItemPropertyPlaybackDuration] = @(-1);
    if (artwork) {
        info[MPMediaItemPropertyArtwork] = artwork;
    }
    
    center.nowPlayingInfo = info;
}

void nowplaying_setPaused() {
    MPNowPlayingInfoCenter* center = [MPNowPlayingInfoCenter defaultCenter];
    NSMutableDictionary* info = [center.nowPlayingInfo mutableCopy];
    if (info) {
        info[MPNowPlayingInfoPropertyPlaybackRate] = @(0.0);
        center.nowPlayingInfo = info;
    }
}

void nowplaying_setStopped() {
    [MPNowPlayingInfoCenter defaultCenter].nowPlayingInfo = nil;
}

void nowplaying_setProgress(double position) {
    MPNowPlayingInfoCenter* center = [MPNowPlayingInfoCenter defaultCenter];
    NSMutableDictionary* info = [center.nowPlayingInfo mutableCopy];
    if (info) {
        info[MPNowPlayingInfoPropertyElapsedPlaybackTime] = @(position);
        center.nowPlayingInfo = info;
    }
}

static void handleCommand(int command, double value) {
    goNowPlayingCommand(command, value);
}

void nowplaying_setCommandHandler() {
    MPRemoteCommandCenter* cmdCenter = [MPRemoteCommandCenter sharedCommandCenter];
    
    [cmdCenter.playCommand addTargetUsingBlock:^(MPRemoteCommandEvent* _Nonnull event) {
        handleCommand(0, 0);  // CmdPlay
    }];
    [cmdCenter.pauseCommand addTargetUsingBlock:^(MPRemoteCommandEvent* _Nonnull event) {
        handleCommand(1, 0);  // CmdPause
    }];
    [cmdCenter.togglePlayPauseCommand addTargetUsingBlock:^(MPRemoteCommandEvent* _Nonnull event) {
        handleCommand(2, 0);  // CmdTogglePlayPause
    }];
    [cmdCenter.nextTrackCommand addTargetUsingBlock:^(MPRemoteCommandEvent* _Nonnull event) {
        handleCommand(3, 0);  // CmdNextTrack
    }];
    [cmdCenter.previousTrackCommand addTargetUsingBlock:^(MPRemoteCommandEvent* _Nonnull event) {
        handleCommand(4, 0);  // CmdPreviousTrack
    }];
    [cmdCenter.changePlaybackPositionCommand addTargetUsingBlock:^(MPRemoteCommandEvent* _Nonnull event) {
        MPChangePlaybackPositionCommandEvent* posEvent = (MPChangePlaybackPositionCommandEvent*)event;
        handleCommand(5, posEvent.positionTime);  // CmdChangePlaybackPosition
    }];
}

void nowplaying_clearCommandHandler() {
    MPRemoteCommandCenter* cmdCenter = [MPRemoteCommandCenter sharedCommandCenter];
    [cmdCenter.playCommand removeTarget:nil];
    [cmdCenter.pauseCommand removeTarget:nil];
    [cmdCenter.togglePlayPauseCommand removeTarget:nil];
    [cmdCenter.nextTrackCommand removeTarget:nil];
    [cmdCenter.previousTrackCommand removeTarget:nil];
    [cmdCenter.changePlaybackPositionCommand removeTarget:nil];
}
```

- [ ] **Step 5: Create `desktop/nowplaying/nowplaying_darwin.go`** — CGO bridge

Uses a global handler variable to avoid `cgo.Handle` lifetime issues. The exported
`goNowPlayingCommand` is called directly from ObjC.

```go
//go:build desktop

package nowplaying

/*
#cgo LDFLAGS: -framework MediaPlayer
#include "bridge.h"
#include <stdlib.h>
*/
import "C"
import (
	"time"
	"unsafe"
)

var globalHandler func(Command, float64)

type darwinController struct{}

func New() Controller {
	return &darwinController{}
}

func (d *darwinController) SetPlaying(title string, duration time.Duration) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.nowplaying_setPlaying(cTitle, C.double(duration.Seconds()))
}

func (d *darwinController) SetStation(name string) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	C.nowplaying_setStation(cName)
}

func (d *darwinController) SetPaused() {
	C.nowplaying_setPaused()
}

func (d *darwinController) SetStopped() {
	C.nowplaying_setStopped()
}

func (d *darwinController) SetProgress(position time.Duration) {
	C.nowplaying_setProgress(C.double(position.Seconds()))
}

func (d *darwinController) SetCommandHandler(handler func(Command, float64)) {
	globalHandler = handler
	if handler != nil {
		C.nowplaying_setCommandHandler()
	} else {
		C.nowplaying_clearCommandHandler()
	}
}

//export goNowPlayingCommand
func goNowPlayingCommand(cmd C.int, value C.double) {
	if globalHandler != nil {
		globalHandler(Command(cmd), float64(value))
	}
}
```

- [ ] **Step 6: Verify macOS compilation**

Run: `go vet ./desktop/nowplaying/...`
(Will need the build tag: `go vet -tags desktop ./desktop/nowplaying/...`)
Expected: no errors

- [ ] **Step 7: Commit**

```bash
git add desktop/nowplaying/
git commit -m "feat(desktop): add nowplaying package with macOS MediaPlayer bridge"
```

---

### Task 6: Wire `nowplaying.Controller` into `App` lifecycle

**Files:**
- Modify: `desktop/app.go`

**Consumes:** `nowplaying.Controller`, `nowplaying.CmdPlay`, `nowplaying.CmdPause`, etc.
**Produces:** Full Now Playing integration in Desktop app

- [ ] **Step 1: Add import and field to `App` struct**

Add to imports at top of `app.go`:
```go
"time"

"github.com/rolniuq/go-beats/desktop/nowplaying"
```

Line 73, add field to `App` struct:
```go
	np nowplaying.Controller
```

- [ ] **Step 2: Create `nowplaying.Controller` in `Startup`**

After `a.pomo = pomodoro.NewTimer(pomodoro.DefaultConfig())` (line ~99), add:
```go
	// Initialize Now Playing controller
	a.np = nowplaying.New()
	a.np.SetCommandHandler(a.handleNowPlayingCommand)
```

Add the handler method (anywhere in the file):
```go
func (a *App) handleNowPlayingCommand(cmd nowplaying.Command, value float64) {
	switch cmd {
	case nowplaying.CmdTogglePlayPause, nowplaying.CmdPlay, nowplaying.CmdPause:
		a.TogglePlay()
	case nowplaying.CmdNextTrack:
		a.NextTrack()
	case nowplaying.CmdPreviousTrack:
		a.PrevTrack()
	}
}
```

- [ ] **Step 3: Update `PlayTrack` to call `np.SetPlaying`**

After line 299 (`return a.engine.Play(index)`), change to:
```go
func (a *App) PlayTrack(index int) error {
	if a.engine == nil {
		return fmt.Errorf("engine not initialized")
	}
	err := a.engine.Play(index)
	if err == nil {
		track := a.engine.CurrentTrack()
		if track != nil {
			dur := a.engine.GetDuration()
			a.np.SetPlaying(track.Name, dur)
		}
	}
	return err
}
```

- [ ] **Step 4: Update `TogglePlay` for Now Playing**

After `a.engine.Pause()` line and the radio branch, add Now Playing updates:

```go
func (a *App) TogglePlay() {
	if a.mode == "radio" {
		if a.radioPlayer != nil {
			a.radioPlayer.Pause()
			if a.radioPlayer.IsPaused() {
				a.np.SetPaused()
			} else {
				station := a.radioPlayer.CurrentStation()
				if station != nil {
					a.np.SetStation(station.Name)
				}
			}
		}
		return
	}

	if a.engine == nil {
		return
	}
	if a.engine.CurrentIndex() < 0 {
		if a.engine.TrackCount() > 0 {
			a.engine.Play(0)
			track := a.engine.CurrentTrack()
			if track != nil {
				dur := a.engine.GetDuration()
				a.np.SetPlaying(track.Name, dur)
			}
		}
	} else {
		a.engine.Pause()
		if a.engine.IsPaused() {
			a.np.SetPaused()
		} else {
			track := a.engine.CurrentTrack()
			if track != nil {
				dur := a.engine.GetDuration()
				a.np.SetPlaying(track.Name, dur)
			}
		}
	}
}
```

- [ ] **Step 5: Update `NextTrack` and `PrevTrack` for Now Playing**

After each track/station change, update Now Playing:

```go
func (a *App) NextTrack() {
	if a.mode == "radio" && a.radioPlayer != nil {
		a.radioPlayer.NextStation()
		station := a.radioPlayer.CurrentStation()
		if station != nil {
			a.np.SetStation(station.Name)
		}
		return
	}
	if a.engine != nil {
		a.engine.Next()
		track := a.engine.CurrentTrack()
		if track != nil {
			dur := a.engine.GetDuration()
			a.np.SetPlaying(track.Name, dur)
		}
	}
}

func (a *App) PrevTrack() {
	if a.mode == "radio" && a.radioPlayer != nil {
		a.radioPlayer.PrevStation()
		station := a.radioPlayer.CurrentStation()
		if station != nil {
			a.np.SetStation(station.Name)
		}
		return
	}
	if a.engine != nil {
		a.engine.Prev()
		track := a.engine.CurrentTrack()
		if track != nil {
			dur := a.engine.GetDuration()
			a.np.SetPlaying(track.Name, dur)
		}
	}
}
```

- [ ] **Step 6: Update `PlayStation` for Now Playing**

```go
func (a *App) PlayStation(index int) error {
	if a.radioPlayer == nil {
		return fmt.Errorf("radio player not initialized")
	}
	err := a.radioPlayer.Play(index)
	if err == nil {
		station := a.radioPlayer.CurrentStation()
		if station != nil {
			a.np.SetStation(station.Name)
		}
	}
	return err
}
```

- [ ] **Step 7: Add `SetProgress` to `pomodoroTicker`**

In `pomodoroTicker`, add Now Playing progress updates. After the auto-advance and radio check blocks (before the `case` for ticker), add:

```go
			// Update Now Playing progress
			if a.mode == "local" && a.engine != nil {
				if a.engine.IsPlaying() {
					pos := a.engine.GetPosition()
					a.np.SetProgress(pos)
				}
			} else if a.mode == "radio" && a.radioPlayer != nil {
				if a.radioPlayer.IsPlaying() {
					a.np.SetPaused() // Radio: just mark as playing, no progress
				}
			}
```

- [ ] **Step 8: Update `SetMode` for Now Playing**

```go
func (a *App) SetMode(mode string) {
	if a.mode == mode {
		return
	}
	if mode == "radio" {
		if a.engine != nil {
			a.engine.Stop()
		}
		a.np.SetStopped()
	} else {
		if a.radioPlayer != nil {
			a.radioPlayer.Stop()
		}
		a.np.SetStopped()
	}
	a.mode = mode
}
```

- [ ] **Step 9: Update `Shutdown` to clean up**

```go
func (a *App) Shutdown(ctx context.Context) {
	if a.engine != nil {
		a.engine.Stop()
	}
	if a.radioPlayer != nil {
		a.radioPlayer.Stop()
	}
	if a.np != nil {
		a.np.SetCommandHandler(nil)
		a.np.SetStopped()
	}
}
```

- [ ] **Step 10: Build verification**

Run: `task check`
Expected: PASS

- [ ] **Step 11: Commit**

```bash
git add desktop/app.go
git commit -m "feat(desktop): wire Now Playing into App lifecycle"
```

---

### Task 7: Final verification

- [ ] **Step 1: Run full check**

Run: `task check`
Expected: PASS

- [ ] **Step 2: Verify desktop build still works**

Run: `CGO_LDFLAGS="-framework UniformTypeIdentifiers" go build -tags desktop -o /dev/null ./cmd/go-beats-desktop 2>&1`
Expected: no errors

- [ ] **Step 3: Verify no residual issues**

Run: `grep -r "formatDuration\|hasMP3Files\|func min(" --include="*.go" cmd/ internal/ desktop/`
Expected: only function *calls* to `util.FormatDuration` and `util.HasMP3Files`, no definitions of `formatDuration`, `hasMP3Files`, or custom `min`

Run: `grep "fyne.io/systray" go.mod`
Expected: no match

Run: `grep "shuffle bool" internal/audio/engine.go`
Expected: no match

Run: `go mod tidy && go vet ./...`
Expected: PASS
