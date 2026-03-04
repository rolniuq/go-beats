package pomodoro

import (
	"time"
)

// Phase represents the current pomodoro phase
type Phase int

const (
	PhaseWork Phase = iota
	PhaseShortBreak
	PhaseLongBreak
	PhaseStopped
)

func (p Phase) String() string {
	switch p {
	case PhaseWork:
		return "🎯 FOCUS"
	case PhaseShortBreak:
		return "☕ SHORT BREAK"
	case PhaseLongBreak:
		return "🌿 LONG BREAK"
	case PhaseStopped:
		return "⏹ STOPPED"
	default:
		return "UNKNOWN"
	}
}

// Config holds pomodoro timer settings
type Config struct {
	WorkDuration       time.Duration
	ShortBreakDuration time.Duration
	LongBreakDuration  time.Duration
	LongBreakInterval  int // number of work sessions before long break
}

// DefaultConfig returns the classic 25/5/15 pomodoro
func DefaultConfig() Config {
	return Config{
		WorkDuration:       25 * time.Minute,
		ShortBreakDuration: 5 * time.Minute,
		LongBreakDuration:  15 * time.Minute,
		LongBreakInterval:  4,
	}
}

// Timer is a pomodoro timer
type Timer struct {
	Config Config

	phase         Phase
	remaining     time.Duration
	totalDuration time.Duration
	running       bool
	sessions      int // completed work sessions

	lastTick time.Time

	// Callbacks
	OnPhaseEnd func(completed Phase, next Phase)
}

// NewTimer creates a new pomodoro timer
func NewTimer(cfg Config) *Timer {
	return &Timer{
		Config: cfg,
		phase:  PhaseStopped,
	}
}

// Start begins a work session
func (t *Timer) Start() {
	t.phase = PhaseWork
	t.totalDuration = t.Config.WorkDuration
	t.remaining = t.totalDuration
	t.running = true
	t.lastTick = time.Now()
}

// Tick updates the timer - call this on every frame
func (t *Timer) Tick() {
	if !t.running || t.phase == PhaseStopped {
		return
	}

	now := time.Now()
	elapsed := now.Sub(t.lastTick)
	t.lastTick = now

	t.remaining -= elapsed
	if t.remaining <= 0 {
		t.remaining = 0
		t.advancePhase()
	}
}

// advancePhase moves to the next phase
func (t *Timer) advancePhase() {
	completedPhase := t.phase

	switch t.phase {
	case PhaseWork:
		t.sessions++
		if t.sessions%t.Config.LongBreakInterval == 0 {
			t.phase = PhaseLongBreak
			t.totalDuration = t.Config.LongBreakDuration
			t.remaining = t.Config.LongBreakDuration
		} else {
			t.phase = PhaseShortBreak
			t.totalDuration = t.Config.ShortBreakDuration
			t.remaining = t.Config.ShortBreakDuration
		}
	case PhaseShortBreak, PhaseLongBreak:
		t.phase = PhaseWork
		t.totalDuration = t.Config.WorkDuration
		t.remaining = t.Config.WorkDuration
	}

	t.lastTick = time.Now()

	if t.OnPhaseEnd != nil {
		t.OnPhaseEnd(completedPhase, t.phase)
	}
}

// Pause toggles pause
func (t *Timer) Pause() {
	if t.phase == PhaseStopped {
		return
	}
	t.running = !t.running
	if t.running {
		t.lastTick = time.Now()
	}
}

// Stop stops the timer
func (t *Timer) Stop() {
	t.phase = PhaseStopped
	t.running = false
	t.remaining = 0
}

// Reset restarts the current phase
func (t *Timer) Reset() {
	if t.phase == PhaseStopped {
		return
	}
	t.remaining = t.totalDuration
	t.lastTick = time.Now()
}

// Skip skips to the next phase
func (t *Timer) Skip() {
	if t.phase == PhaseStopped {
		return
	}
	t.remaining = 0
	t.advancePhase()
}

// Phase returns current phase
func (t *Timer) Phase() Phase {
	return t.phase
}

// Remaining returns remaining time
func (t *Timer) Remaining() time.Duration {
	return t.remaining
}

// TotalDuration returns the total duration for the current phase
func (t *Timer) TotalDuration() time.Duration {
	return t.totalDuration
}

// IsRunning returns if the timer is running
func (t *Timer) IsRunning() bool {
	return t.running
}

// Sessions returns completed work sessions
func (t *Timer) Sessions() int {
	return t.sessions
}

// Progress returns progress as 0.0 to 1.0
func (t *Timer) Progress() float64 {
	if t.totalDuration == 0 {
		return 0
	}
	return 1.0 - float64(t.remaining)/float64(t.totalDuration)
}
