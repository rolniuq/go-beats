package pomodoro

import (
	"testing"
	"time"
)

func TestPhaseString(t *testing.T) {
	tests := []struct {
		phase Phase
		want  string
	}{
		{PhaseWork, "🎯 FOCUS"},
		{PhaseShortBreak, "☕ SHORT BREAK"},
		{PhaseLongBreak, "🌿 LONG BREAK"},
		{PhaseStopped, "⏹ STOPPED"},
		{Phase(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.phase.String(); got != tt.want {
				t.Errorf("Phase.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.WorkDuration != 25*time.Minute {
		t.Errorf("WorkDuration = %v, want %v", cfg.WorkDuration, 25*time.Minute)
	}
	if cfg.ShortBreakDuration != 5*time.Minute {
		t.Errorf("ShortBreakDuration = %v, want %v", cfg.ShortBreakDuration, 5*time.Minute)
	}
	if cfg.LongBreakDuration != 15*time.Minute {
		t.Errorf("LongBreakDuration = %v, want %v", cfg.LongBreakDuration, 15*time.Minute)
	}
	if cfg.LongBreakInterval != 4 {
		t.Errorf("LongBreakInterval = %v, want %v", cfg.LongBreakInterval, 4)
	}
}

func TestNewTimer(t *testing.T) {
	cfg := DefaultConfig()
	timer := NewTimer(cfg)

	if timer.phase != PhaseStopped {
		t.Errorf("NewTimer().phase = %v, want %v", timer.phase, PhaseStopped)
	}
	if timer.running {
		t.Errorf("NewTimer().running = %v, want %v", timer.running, false)
	}
	if timer.sessions != 0 {
		t.Errorf("NewTimer().sessions = %v, want %v", timer.sessions, 0)
	}
}

func TestTimerStart(t *testing.T) {
	cfg := Config{
		WorkDuration:       1 * time.Minute,
		ShortBreakDuration: 30 * time.Second,
		LongBreakDuration:  2 * time.Minute,
		LongBreakInterval:  2,
	}
	timer := NewTimer(cfg)
	timer.Start()

	if timer.phase != PhaseWork {
		t.Errorf("After Start(), phase = %v, want %v", timer.phase, PhaseWork)
	}
	if !timer.running {
		t.Errorf("After Start(), running = %v, want %v", timer.running, true)
	}
	if timer.remaining != timer.totalDuration {
		t.Errorf("remaining = %v, want %v", timer.remaining, timer.totalDuration)
	}
}

func TestTimerTick(t *testing.T) {
	cfg := Config{
		WorkDuration:       1 * time.Minute,
		ShortBreakDuration: 30 * time.Second,
		LongBreakDuration:  2 * time.Minute,
		LongBreakInterval:  2,
	}
	timer := NewTimer(cfg)
	timer.Start()

	initialRemaining := timer.remaining
	timer.Tick()

	if timer.remaining >= initialRemaining {
		t.Errorf("After Tick(), remaining should decrease")
	}
}

func TestTimerTickNotRunning(t *testing.T) {
	cfg := DefaultConfig()
	timer := NewTimer(cfg)

	initialRemaining := timer.remaining
	timer.Tick()

	if timer.remaining != initialRemaining {
		t.Errorf("Tick() should not change remaining when not running")
	}
}

func TestTimerTickStoppedPhase(t *testing.T) {
	cfg := DefaultConfig()
	timer := NewTimer(cfg)
	timer.phase = PhaseStopped
	timer.running = true

	initialRemaining := timer.remaining
	timer.Tick()

	if timer.remaining != initialRemaining {
		t.Errorf("Tick() should not change remaining when phase is PhaseStopped")
	}
}

func TestTimerPause(t *testing.T) {
	cfg := DefaultConfig()
	timer := NewTimer(cfg)
	timer.Start()

	timer.Pause()

	if timer.running {
		t.Errorf("After Pause(), running = %v, want %v", timer.running, false)
	}

	timer.Pause()

	if !timer.running {
		t.Errorf("After second Pause(), running = %v, want %v", timer.running, true)
	}
}

func TestTimerPauseStopped(t *testing.T) {
	cfg := DefaultConfig()
	timer := NewTimer(cfg)
	timer.phase = PhaseStopped

	timer.Pause()

	if timer.running {
		t.Errorf("Pause() should not start timer when phase is PhaseStopped")
	}
}

func TestTimerStop(t *testing.T) {
	cfg := DefaultConfig()
	timer := NewTimer(cfg)
	timer.Start()

	timer.Stop()

	if timer.phase != PhaseStopped {
		t.Errorf("After Stop(), phase = %v, want %v", timer.phase, PhaseStopped)
	}
	if timer.running {
		t.Errorf("After Stop(), running = %v, want %v", timer.running, false)
	}
	if timer.remaining != 0 {
		t.Errorf("After Stop(), remaining = %v, want %v", timer.remaining, 0)
	}
}

func TestTimerReset(t *testing.T) {
	cfg := Config{
		WorkDuration:       1 * time.Minute,
		ShortBreakDuration: 30 * time.Second,
		LongBreakDuration:  2 * time.Minute,
		LongBreakInterval:  2,
	}
	timer := NewTimer(cfg)
	timer.Start()

	elapsed := 30 * time.Second
	timer.remaining -= elapsed
	timer.Reset()

	if timer.remaining != timer.totalDuration {
		t.Errorf("After Reset(), remaining = %v, want %v", timer.remaining, timer.totalDuration)
	}
}

func TestTimerResetStopped(t *testing.T) {
	cfg := DefaultConfig()
	timer := NewTimer(cfg)
	timer.phase = PhaseStopped

	initialRemaining := timer.remaining
	timer.Reset()

	if timer.remaining != initialRemaining {
		t.Errorf("Reset() should not change remaining when phase is PhaseStopped")
	}
}

func TestTimerSkip(t *testing.T) {
	cfg := Config{
		WorkDuration:       1 * time.Minute,
		ShortBreakDuration: 30 * time.Second,
		LongBreakDuration:  2 * time.Minute,
		LongBreakInterval:  2,
	}
	timer := NewTimer(cfg)
	timer.Start()

	timer.Skip()

	if timer.phase != PhaseShortBreak {
		t.Errorf("After Skip(), phase = %v, want %v", timer.phase, PhaseShortBreak)
	}
}

func TestTimerSkipStopped(t *testing.T) {
	cfg := DefaultConfig()
	timer := NewTimer(cfg)
	timer.phase = PhaseStopped

	initialPhase := timer.phase
	timer.Skip()

	if timer.phase != initialPhase {
		t.Errorf("Skip() should not change phase when phase is PhaseStopped")
	}
}

func TestSessionsCount(t *testing.T) {
	cfg := DefaultConfig()
	timer := NewTimer(cfg)

	if timer.Sessions() != 0 {
		t.Errorf("Initial sessions = %v, want %v", timer.Sessions(), 0)
	}

	timer.sessions = 3

	if timer.Sessions() != 3 {
		t.Errorf("Sessions() = %v, want %v", timer.Sessions(), 3)
	}
}

func TestProgress(t *testing.T) {
	cfg := Config{
		WorkDuration:       100 * time.Millisecond,
		ShortBreakDuration: 50 * time.Millisecond,
		LongBreakDuration:  100 * time.Millisecond,
		LongBreakInterval:  2,
	}
	timer := NewTimer(cfg)
	timer.Start()

	timer.remaining = 50 * time.Millisecond

	progress := timer.Progress()
	if progress != 0.5 {
		t.Errorf("Progress() = %v, want %v", progress, 0.5)
	}
}

func TestProgressZeroDuration(t *testing.T) {
	cfg := DefaultConfig()
	timer := NewTimer(cfg)

	progress := timer.Progress()
	if progress != 0 {
		t.Errorf("Progress() with zero duration = %v, want %v", progress, 0)
	}
}

func TestPhase(t *testing.T) {
	cfg := DefaultConfig()
	timer := NewTimer(cfg)

	if timer.Phase() != PhaseStopped {
		t.Errorf("Phase() = %v, want %v", timer.Phase(), PhaseStopped)
	}

	timer.phase = PhaseWork
	if timer.Phase() != PhaseWork {
		t.Errorf("Phase() = %v, want %v", timer.Phase(), PhaseWork)
	}
}

func TestRemaining(t *testing.T) {
	cfg := DefaultConfig()
	timer := NewTimer(cfg)

	if timer.Remaining() != 0 {
		t.Errorf("Remaining() = %v, want %v", timer.Remaining(), 0)
	}

	timer.remaining = 5 * time.Minute
	if timer.Remaining() != 5*time.Minute {
		t.Errorf("Remaining() = %v, want %v", timer.Remaining(), 5*time.Minute)
	}
}

func TestTotalDuration(t *testing.T) {
	cfg := Config{
		WorkDuration:       25 * time.Minute,
		ShortBreakDuration: 5 * time.Minute,
		LongBreakDuration:  15 * time.Minute,
		LongBreakInterval:  4,
	}
	timer := NewTimer(cfg)
	timer.Start()

	if timer.TotalDuration() != 25*time.Minute {
		t.Errorf("TotalDuration() = %v, want %v", timer.TotalDuration(), 25*time.Minute)
	}
}

func TestIsRunning(t *testing.T) {
	cfg := DefaultConfig()
	timer := NewTimer(cfg)

	if timer.IsRunning() {
		t.Errorf("IsRunning() = %v, want %v", timer.IsRunning(), false)
	}

	timer.running = true
	if !timer.IsRunning() {
		t.Errorf("IsRunning() = %v, want %v", timer.IsRunning(), true)
	}
}
