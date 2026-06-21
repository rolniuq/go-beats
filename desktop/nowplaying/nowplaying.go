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
