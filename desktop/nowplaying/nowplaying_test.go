package nowplaying

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNopMethods(t *testing.T) {
	c := New()

	// None of these should panic
	c.SetPlaying("test", time.Second)
	c.SetStation("test station")
	c.SetPaused()
	c.SetStopped()
	c.SetProgress(time.Minute)
	c.SetCommandHandler(nil)
	c.SetCommandHandler(func(Command, float64) {})
}

func TestCommandValues(t *testing.T) {
	tests := []struct {
		cmd  Command
		want int
	}{
		{CmdPlay, 0},
		{CmdPause, 1},
		{CmdTogglePlayPause, 2},
		{CmdNextTrack, 3},
		{CmdPreviousTrack, 4},
		{CmdChangePlaybackPosition, 5},
	}

	for _, tt := range tests {
		if int(tt.cmd) != tt.want {
			t.Errorf("got %d, want %d", tt.cmd, tt.want)
		}
	}
}
