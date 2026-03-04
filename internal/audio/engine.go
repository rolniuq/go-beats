package audio

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

// TrackInfo holds metadata about a music track
type TrackInfo struct {
	Name     string
	Path     string
	Duration time.Duration
}

// Engine is the core audio player
type Engine struct {
	mu sync.Mutex

	tracks       []TrackInfo
	currentIndex int

	streamer  beep.StreamSeekCloser
	format    beep.Format
	ctrl      *beep.Ctrl
	volume    *effects.Volume
	resampler beep.Streamer

	playing bool
	paused  bool
	loop    bool
	shuffle bool

	// Volume in dB, 0 = normal, -5 = quiet, 5 = loud
	volumeLevel float64

	// Callback when track ends naturally
	OnTrackEnd func()
}

// NewEngine creates a new audio engine
func NewEngine() *Engine {
	return &Engine{
		currentIndex: -1,
		volumeLevel:  0,
	}
}

// ScanDirectory scans a directory for .mp3 files
func (e *Engine) ScanDirectory(dir string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.tracks = nil

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read music directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".mp3" {
			name := strings.TrimSuffix(entry.Name(), ext)
			e.tracks = append(e.tracks, TrackInfo{
				Name: name,
				Path: filepath.Join(dir, entry.Name()),
			})
		}
	}

	if len(e.tracks) == 0 {
		return fmt.Errorf("no .mp3 files found in %s", dir)
	}

	return nil
}

// InitSpeaker initializes the speaker with a sample rate
func (e *Engine) InitSpeaker() error {
	// Initialize with a common sample rate; we'll resample tracks to match
	sr := beep.SampleRate(44100)
	return speaker.Init(sr, sr.N(time.Second/10))
}

// Play starts playing the track at the given index
func (e *Engine) Play(index int) error {
	e.mu.Lock()
	if index < 0 || index >= len(e.tracks) {
		e.mu.Unlock()
		return fmt.Errorf("track index %d out of range", index)
	}
	e.mu.Unlock()

	// Stop current track if any
	e.Stop()

	e.mu.Lock()
	defer e.mu.Unlock()

	track := e.tracks[index]

	f, err := os.Open(track.Path)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", track.Path, err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		f.Close()
		return fmt.Errorf("failed to decode %s: %w", track.Path, err)
	}

	e.streamer = streamer
	e.format = format
	e.currentIndex = index

	// Calculate duration
	e.tracks[index].Duration = format.SampleRate.D(streamer.Len())

	// Resample to 44100 if needed
	targetSR := beep.SampleRate(44100)
	var baseStreamer beep.Streamer
	if format.SampleRate != targetSR {
		baseStreamer = beep.Resample(4, format.SampleRate, targetSR, streamer)
	} else {
		baseStreamer = streamer
	}

	// Wrap with control and volume
	e.ctrl = &beep.Ctrl{Streamer: beep.Seq(baseStreamer, beep.Callback(func() {
		if e.OnTrackEnd != nil {
			e.OnTrackEnd()
		}
	}))}
	e.volume = &effects.Volume{
		Streamer: e.ctrl,
		Base:     2,
		Volume:   e.volumeLevel,
	}

	e.playing = true
	e.paused = false

	speaker.Play(e.volume)

	return nil
}

// Pause toggles pause
func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.ctrl == nil {
		return
	}

	speaker.Lock()
	e.ctrl.Paused = !e.ctrl.Paused
	e.paused = e.ctrl.Paused
	speaker.Unlock()
}

// Stop stops the current track
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.streamer != nil {
		speaker.Clear()
		e.streamer.Close()
		e.streamer = nil
		e.ctrl = nil
		e.volume = nil
	}
	e.playing = false
	e.paused = false
}

// VolumeUp increases volume
func (e *Engine) VolumeUp() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.volumeLevel = math.Min(e.volumeLevel+0.5, 5)
	if e.volume != nil {
		speaker.Lock()
		e.volume.Volume = e.volumeLevel
		speaker.Unlock()
	}
}

// VolumeDown decreases volume
func (e *Engine) VolumeDown() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.volumeLevel = math.Max(e.volumeLevel-0.5, -7)
	if e.volume != nil {
		speaker.Lock()
		e.volume.Volume = e.volumeLevel
		speaker.Unlock()
	}
}

// Next plays the next track
func (e *Engine) Next() error {
	e.mu.Lock()
	if len(e.tracks) == 0 {
		e.mu.Unlock()
		return fmt.Errorf("no tracks loaded")
	}
	next := (e.currentIndex + 1) % len(e.tracks)
	e.mu.Unlock()
	return e.Play(next)
}

// Prev plays the previous track
func (e *Engine) Prev() error {
	e.mu.Lock()
	if len(e.tracks) == 0 {
		e.mu.Unlock()
		return fmt.Errorf("no tracks loaded")
	}
	prev := e.currentIndex - 1
	if prev < 0 {
		prev = len(e.tracks) - 1
	}
	e.mu.Unlock()
	return e.Play(prev)
}

// IsPlaying returns whether audio is actively playing
func (e *Engine) IsPlaying() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.playing && !e.paused
}

// IsPaused returns whether audio is paused
func (e *Engine) IsPaused() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.paused
}

// GetPosition returns current playback position
func (e *Engine) GetPosition() time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.streamer == nil {
		return 0
	}

	speaker.Lock()
	pos := e.format.SampleRate.D(e.streamer.Position())
	speaker.Unlock()
	return pos
}

// GetDuration returns the duration of the current track
func (e *Engine) GetDuration() time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.currentIndex < 0 || e.currentIndex >= len(e.tracks) {
		return 0
	}
	return e.tracks[e.currentIndex].Duration
}

// CurrentTrack returns the current track info
func (e *Engine) CurrentTrack() *TrackInfo {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.currentIndex < 0 || e.currentIndex >= len(e.tracks) {
		return nil
	}
	t := e.tracks[e.currentIndex]
	return &t
}

// CurrentIndex returns current track index
func (e *Engine) CurrentIndex() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.currentIndex
}

// Tracks returns all loaded tracks
func (e *Engine) Tracks() []TrackInfo {
	e.mu.Lock()
	defer e.mu.Unlock()
	result := make([]TrackInfo, len(e.tracks))
	copy(result, e.tracks)
	return result
}

// TrackCount returns number of loaded tracks
func (e *Engine) TrackCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.tracks)
}

// GetVolumePercent returns volume as a percentage (0-100)
func (e *Engine) GetVolumePercent() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	// Map from -7..5 to 0..100
	pct := int(((e.volumeLevel + 7) / 12) * 100)
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return pct
}

// ToggleLoop toggles loop mode
func (e *Engine) ToggleLoop() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.loop = !e.loop
	return e.loop
}

// IsLoop returns loop state
func (e *Engine) IsLoop() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.loop
}
