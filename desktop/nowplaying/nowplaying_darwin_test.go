//go:build desktop

package nowplaying

import (
	"testing"
	"time"
)

func TestNewDarwin(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
}

func TestSetCommandHandler(t *testing.T) {
	c := New()
	called := false
	c.SetCommandHandler(func(cmd Command, val float64) {
		called = true
		if cmd != CmdPlay {
			t.Errorf("cmd = %d, want %d", cmd, CmdPlay)
		}
	})
	// Simulate a remote command by calling the exported function
	goNowPlayingCommand(0, 0)
	if !called {
		t.Error("handler was not called")
	}
}

func TestClearCommandHandler(t *testing.T) {
	c := New()
	c.SetCommandHandler(func(Command, float64) {})
	c.SetCommandHandler(nil)
	// Should not panic when handler is nil
	goNowPlayingCommand(1, 0)
}

func TestDarwinMethods(t *testing.T) {
	c := New()
	// Methods should not panic
	c.SetPlaying("test track", 3*time.Minute)
	c.SetStation("test station")
	c.SetPaused()
	c.SetProgress(90 * time.Second)
	c.SetStopped()
}
