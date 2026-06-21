//go:build desktop

package desktop

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rolniuq/go-beats/desktop/nowplaying"
	"github.com/rolniuq/go-beats/internal/audio"
	"github.com/rolniuq/go-beats/internal/notification"
	"github.com/rolniuq/go-beats/internal/pomodoro"
	"github.com/rolniuq/go-beats/internal/radio"
	"github.com/rolniuq/go-beats/internal/util"
)

// TrackDTO is the JSON-safe track info
type TrackDTO struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Duration string `json:"duration"`
}

// StationDTO is the JSON-safe station info
type StationDTO struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Genre       string `json:"genre"`
	Description string `json:"description"`
}

// PlayerState contains the full state for the frontend
type PlayerState struct {
	// Mode
	Mode string `json:"mode"` // "local" or "radio"

	// Local playback
	IsPlaying    bool   `json:"isPlaying"`
	IsPaused     bool   `json:"isPaused"`
	CurrentTrack string `json:"currentTrack"`
	TrackIndex   int    `json:"trackIndex"`
	Position     string `json:"position"`
	Duration     string `json:"duration"`
	PositionMs   int64  `json:"positionMs"`
	DurationMs   int64  `json:"durationMs"`
	Volume       int    `json:"volume"`
	Loop         bool   `json:"loop"`
	TrackCount   int    `json:"trackCount"`

	// Radio
	RadioPlaying      bool   `json:"radioPlaying"`
	RadioPaused       bool   `json:"radioPaused"`
	RadioConnecting   bool   `json:"radioConnecting"`
	RadioReconnecting bool   `json:"radioReconnecting"`
	RadioCanRetry     bool   `json:"radioCanRetry"`
	RadioStation      string `json:"radioStation"`
	RadioStationIndex int    `json:"radioStationIndex"`
	RadioVolume       int    `json:"radioVolume"`
	RadioError        string `json:"radioError"`

	// Pomodoro
	PomodoroPhase     string  `json:"pomodoroPhase"`
	PomodoroRemaining string  `json:"pomodoroRemaining"`
	PomodoroProgress  float64 `json:"pomodoroProgress"`
	PomodoroRunning   bool    `json:"pomodoroRunning"`
	PomodoroSessions  int     `json:"pomodoroSessions"`
}

// App is the Wails application backend
type App struct {
	ctx         context.Context
	engine      *audio.Engine
	radioPlayer *radio.Player
	pomo        *pomodoro.Timer
	np          nowplaying.Controller
	mode        string // "local" or "radio"
	musicDir    string
}

// NewApp creates a new App
func NewApp() *App {
	return &App{
		mode: "local",
	}
}

// Startup is called when the app starts
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize audio engine (delay speaker init to avoid CoreAudio/Cocoa conflict)
	a.engine = audio.NewEngine()

	// Initialize radio player
	a.radioPlayer = radio.NewPlayer()

	// Initialize pomodoro timer
	a.pomo = pomodoro.NewTimer(pomodoro.DefaultConfig())

	// Initialize Now Playing controller
	a.np = nowplaying.New()
	a.np.SetCommandHandler(a.handleNowPlayingCommand)

	// Play notification sound when a pomodoro phase ends
	a.pomo.OnPhaseEnd = func(completed pomodoro.Phase, next pomodoro.Phase) {
		switch completed {
		case pomodoro.PhaseWork:
			go notification.PlayFocusEnd()
		case pomodoro.PhaseShortBreak, pomodoro.PhaseLongBreak:
			go notification.PlayBreakEnd()
		}
	}

	// Try to find music directory
	a.musicDir = "./music"
	homeDir, err := os.UserHomeDir()
	if err == nil {
		musicPath := filepath.Join(homeDir, "Music", "go-beats")
		if info, err := os.Stat(musicPath); err == nil && info.IsDir() {
			a.musicDir = musicPath
		}
	}

	// Initialize speaker and scan music directory in background
	// to avoid blocking the Wails startup on the main thread
	go func() {
		if err := a.engine.InitSpeaker(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing audio: %v\n", err)
		}

		// Try scanning the music directory
		if info, err := os.Stat(a.musicDir); err == nil && info.IsDir() {
			a.engine.ScanDirectory(a.musicDir)
		}
	}()

	// Start pomodoro ticker
	go a.pomodoroTicker()
}

// Shutdown is called when the app is closing
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

func (a *App) pomodoroTicker() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			if a.pomo != nil {
				a.pomo.Tick()
			}

			// Auto-advance local tracks
			if a.mode == "local" && a.engine != nil && a.engine.IsPlaying() {
				pos := a.engine.GetPosition()
				dur := a.engine.GetDuration()
				if dur > 0 && pos >= dur-200*time.Millisecond {
					if a.engine.IsLoop() {
						idx := a.engine.CurrentIndex()
						if idx >= 0 {
							a.engine.Play(idx)
						}
					} else {
						a.engine.Next()
					}
				}
			}

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

			// Check radio stream
			if a.mode == "radio" && a.radioPlayer != nil {
				a.radioPlayer.CheckStream()
			}
		}
	}
}

// ── State Queries ──────────────────────────────────────────────────────────

// GetState returns the full player state for the frontend
func (a *App) GetState() PlayerState {
	state := PlayerState{
		Mode: a.mode,
	}

	// Local player state
	if a.engine != nil {
		state.IsPlaying = a.engine.IsPlaying()
		state.IsPaused = a.engine.IsPaused()
		state.TrackIndex = a.engine.CurrentIndex()
		state.Volume = a.engine.GetVolumePercent()
		state.Loop = a.engine.IsLoop()
		state.TrackCount = a.engine.TrackCount()

		track := a.engine.CurrentTrack()
		if track != nil {
			state.CurrentTrack = track.Name
		}

		pos := a.engine.GetPosition()
		dur := a.engine.GetDuration()
		state.Position = util.FormatDuration(pos)
		state.Duration = util.FormatDuration(dur)
		state.PositionMs = pos.Milliseconds()
		state.DurationMs = dur.Milliseconds()
	}

	// Radio state
	if a.radioPlayer != nil {
		state.RadioPlaying = a.radioPlayer.IsPlaying()
		state.RadioPaused = a.radioPlayer.IsPaused()
		state.RadioConnecting = a.radioPlayer.IsConnecting()
		state.RadioReconnecting = a.radioPlayer.IsReconnecting()
		state.RadioCanRetry = a.radioPlayer.CanRetry()
		state.RadioStationIndex = a.radioPlayer.CurrentStationIndex()
		state.RadioVolume = a.radioPlayer.GetVolumePercent()

		station := a.radioPlayer.CurrentStation()
		if station != nil {
			state.RadioStation = station.Name
		}

		if err := a.radioPlayer.Error(); err != nil {
			state.RadioError = err.Error()
		}
	}

	// Pomodoro state
	if a.pomo != nil {
		phase := a.pomo.Phase()
		switch phase {
		case pomodoro.PhaseWork:
			state.PomodoroPhase = "focus"
		case pomodoro.PhaseShortBreak:
			state.PomodoroPhase = "short_break"
		case pomodoro.PhaseLongBreak:
			state.PomodoroPhase = "long_break"
		default:
			state.PomodoroPhase = "stopped"
		}
		state.PomodoroRemaining = util.FormatDuration(a.pomo.Remaining())
		state.PomodoroProgress = a.pomo.Progress()
		state.PomodoroRunning = a.pomo.IsRunning()
		state.PomodoroSessions = a.pomo.Sessions()
	}

	return state
}

// GetTracks returns all loaded tracks
func (a *App) GetTracks() []TrackDTO {
	if a.engine == nil {
		return nil
	}
	tracks := a.engine.Tracks()
	result := make([]TrackDTO, len(tracks))
	for i, t := range tracks {
		result[i] = TrackDTO{
			Name:     t.Name,
			Path:     t.Path,
			Duration: util.FormatDuration(t.Duration),
		}
	}
	return result
}

// GetStations returns all radio stations
func (a *App) GetStations() []StationDTO {
	if a.radioPlayer == nil {
		return nil
	}
	stations := a.radioPlayer.Stations()
	result := make([]StationDTO, len(stations))
	for i, s := range stations {
		result[i] = StationDTO{
			Name:        s.Name,
			URL:         s.URL,
			Genre:       s.Genre,
			Description: s.Description,
		}
	}
	return result
}

// ── Player Controls ────────────────────────────────────────────────────────

// PlayTrack plays a local track by index
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

// TogglePlay toggles play/pause
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

// NextTrack advances to next track/station
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

// PrevTrack goes to previous track/station
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

// VolumeUp increases volume
func (a *App) VolumeUp() {
	if a.mode == "radio" && a.radioPlayer != nil {
		a.radioPlayer.VolumeUp()
		return
	}
	if a.engine != nil {
		a.engine.VolumeUp()
	}
}

// VolumeDown decreases volume
func (a *App) VolumeDown() {
	if a.mode == "radio" && a.radioPlayer != nil {
		a.radioPlayer.VolumeDown()
		return
	}
	if a.engine != nil {
		a.engine.VolumeDown()
	}
}

// ToggleLoop toggles loop mode
func (a *App) ToggleLoop() bool {
	if a.engine != nil {
		return a.engine.ToggleLoop()
	}
	return false
}

// SetMode switches between local and radio mode
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

// PlayStation plays a radio station by index
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

// RetryRadio retries the failed radio connection
func (a *App) RetryRadio() bool {
	if a.radioPlayer != nil {
		return a.radioPlayer.Retry()
	}
	return false
}

// ── Pomodoro Controls ──────────────────────────────────────────────────────

// TogglePomodoro starts or stops the pomodoro timer
func (a *App) TogglePomodoro() {
	if a.pomo == nil {
		return
	}
	if a.pomo.Phase() == pomodoro.PhaseStopped {
		a.pomo.Start()
	} else {
		a.pomo.Stop()
	}
}

// PausePomodoro pauses or resumes the pomodoro timer
func (a *App) PausePomodoro() {
	if a.pomo != nil {
		a.pomo.Pause()
	}
}

// SkipPomodoro skips to the next pomodoro phase
func (a *App) SkipPomodoro() {
	if a.pomo != nil {
		a.pomo.Skip()
	}
}

// ── File Operations ────────────────────────────────────────────────────────

// OpenMusicDirectory opens a directory selection dialog (simplified: scans a given path)
func (a *App) LoadMusicDirectory(dir string) (int, error) {
	if a.engine == nil {
		return 0, fmt.Errorf("engine not initialized")
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return 0, fmt.Errorf("invalid path: %w", err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return 0, fmt.Errorf("directory not found: %w", err)
	}
	if !info.IsDir() {
		return 0, fmt.Errorf("not a directory: %s", absDir)
	}

	if err := a.engine.ScanDirectory(absDir); err != nil {
		return 0, err
	}

	a.musicDir = absDir
	return a.engine.TrackCount(), nil
}

// GetMusicDir returns the current music directory
func (a *App) GetMusicDir() string {
	return a.musicDir
}

// BrowseMusicFolder uses native file dialog to pick a folder
func (a *App) BrowseMusicFolder() (int, error) {
	// Wails runtime dialog is available via frontend JS; we'll handle it there
	// This is a fallback that scans common locations
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}

	candidates := []string{
		filepath.Join(homeDir, "Music", "go-beats"),
		filepath.Join(homeDir, "Music"),
		"./music",
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			if util.HasMP3Files(dir) {
				return a.LoadMusicDirectory(dir)
			}
		}
	}

	return 0, fmt.Errorf("no music directory with .mp3 files found")
}


