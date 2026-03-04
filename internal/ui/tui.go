package ui

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/quynh-vo/go-beats/internal/audio"
	"github.com/quynh-vo/go-beats/internal/pomodoro"
	"github.com/quynh-vo/go-beats/internal/radio"
)

// ── Messages ────────────────────────────────────────────────────────────────

type tickMsg time.Time
type trackEndMsg struct{}

type Mode int

const (
	ModeLocal Mode = iota
	ModeRadio
)

// ── Key bindings ────────────────────────────────────────────────────────────

type keyMap struct {
	Play      key.Binding
	Next      key.Binding
	Prev      key.Binding
	Tab       key.Binding
	VolUp     key.Binding
	VolDown   key.Binding
	Loop      key.Binding
	Pomo      key.Binding
	PomoPause key.Binding
	PomoSkip  key.Binding
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Quit      key.Binding
	Help      key.Binding
}

var keys = keyMap{
	Play:      key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "play/pause")),
	Next:      key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "next track")),
	Prev:      key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "prev track")),
	Tab:       key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch mode")),
	VolUp:     key.NewBinding(key.WithKeys("=", "+"), key.WithHelp("+/=", "volume up")),
	VolDown:   key.NewBinding(key.WithKeys("-"), key.WithHelp("-", "volume down")),
	Loop:      key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "loop mode")),
	Pomo:      key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "start pomodoro")),
	PomoPause: key.NewBinding(key.WithKeys("T"), key.WithHelp("T", "pause pomodoro")),
	PomoSkip:  key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "skip phase")),
	Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "scroll up")),
	Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "scroll down")),
	Enter:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select track")),
	Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
}

// ── Styles ──────────────────────────────────────────────────────────────────

var (
	// Colors - lofi aesthetic palette
	colorPrimary   = lipgloss.Color("#c4a7e7") // lavender
	colorSecondary = lipgloss.Color("#9ccfd8") // foam
	colorAccent    = lipgloss.Color("#f6c177") // gold
	colorMuted     = lipgloss.Color("#6e6a86") // muted
	colorSurface   = lipgloss.Color("#1f1d2e") // surface
	colorText      = lipgloss.Color("#e0def4") // text
	colorLove      = lipgloss.Color("#eb6f92") // love/red
	colorGreen     = lipgloss.Color("#31748f") // pine

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			PaddingLeft(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 1)

	trackStyle = lipgloss.NewStyle().
			Foreground(colorText).
			PaddingLeft(2)

	activeTrackStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true).
				PaddingLeft(1)

	selectedTrackStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).
				PaddingLeft(1)

	progressStyle = lipgloss.NewStyle().
			Foreground(colorPrimary)

	pomoStyle = lipgloss.NewStyle().
			Foreground(colorLove).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			PaddingLeft(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(colorGreen)

	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)
)

// ── Model ───────────────────────────────────────────────────────────────────

type Model struct {
	engine      *audio.Engine
	radioPlayer *radio.Player
	mode        Mode
	pomo        *pomodoro.Timer
	width       int
	height      int

	cursor   int // cursor in track list
	showHelp bool

	// Visualizer
	vizBars []float64

	// Status message
	statusMsg string
	statusExp time.Time
}

func NewModel(engine *audio.Engine, radioPlayer *radio.Player) Model {
	pomo := pomodoro.NewTimer(pomodoro.DefaultConfig())

	m := Model{
		engine:      engine,
		radioPlayer: radioPlayer,
		mode:        ModeLocal,
		pomo:        pomo,
		vizBars:     make([]float64, 30),
	}

	// When a track ends, send a message
	engine.OnTrackEnd = func() {
		// This runs in the audio goroutine, need to handle carefully
	}

	return m
}

func (m *Model) SetMode(mode Mode) {
	if m.mode == mode {
		return
	}

	if mode == ModeRadio {
		if m.engine != nil {
			m.engine.Stop()
		}
	} else {
		if m.radioPlayer != nil {
			m.radioPlayer.Stop()
		}
	}

	m.cursor = 0
	m.mode = mode
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), tea.WindowSize())
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		m.pomo.Tick()
		m.updateVisualizer()

		// Check radio stream status
		if m.mode == ModeRadio && m.radioPlayer != nil {
			m.radioPlayer.CheckStream()
		}

		// Check if we need to auto-advance (track finished)
		if m.engine.IsPlaying() {
			pos := m.engine.GetPosition()
			dur := m.engine.GetDuration()
			if dur > 0 && pos >= dur-time.Millisecond*200 {
				if m.engine.IsLoop() {
					idx := m.engine.CurrentIndex()
					if idx >= 0 {
						m.engine.Play(idx)
					}
				} else {
					m.engine.Next()
				}
			}
		}

		return m, tickCmd()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		m.engine.Stop()
		if m.radioPlayer != nil {
			m.radioPlayer.Stop()
		}
		return m, tea.Quit

	case key.Matches(msg, keys.Help):
		m.showHelp = !m.showHelp
		return m, nil

	case key.Matches(msg, keys.Tab):
		if m.mode == ModeLocal {
			m.SetMode(ModeRadio)
			m.setStatus("📻 Radio mode")
		} else {
			m.SetMode(ModeLocal)
			m.setStatus("♪ Local mode")
		}
		return m, nil

	case key.Matches(msg, keys.Play):
		if m.mode == ModeRadio {
			if m.radioPlayer != nil {
				m.radioPlayer.Pause()
				if m.radioPlayer.IsPaused() {
					m.setStatus("⏸ Radio paused")
				} else {
					m.setStatus("▶ Radio playing")
				}
			}
			return m, nil
		}

		if m.engine.TrackCount() == 0 {
			return m, nil
		}
		if m.engine.CurrentIndex() < 0 {
			// Nothing playing yet, start first track
			m.engine.Play(0)
			m.cursor = 0
			m.setStatus("▶ Playing")
		} else {
			m.engine.Pause()
			if m.engine.IsPaused() {
				m.setStatus("⏸ Paused")
			} else {
				m.setStatus("▶ Playing")
			}
		}
		return m, nil

	case key.Matches(msg, keys.Next):
		if m.mode == ModeRadio && m.radioPlayer != nil {
			if err := m.radioPlayer.NextStation(); err == nil {
				m.cursor = m.radioPlayer.CurrentStationIndex()
				m.setStatus("⏭ Next station")
			}
			return m, nil
		}

		if err := m.engine.Next(); err == nil {
			m.cursor = m.engine.CurrentIndex()
			m.setStatus("⏭ Next track")
		}
		return m, nil

	case key.Matches(msg, keys.Prev):
		if m.mode == ModeRadio && m.radioPlayer != nil {
			if err := m.radioPlayer.PrevStation(); err == nil {
				m.cursor = m.radioPlayer.CurrentStationIndex()
				m.setStatus("⏮ Previous station")
			}
			return m, nil
		}

		if err := m.engine.Prev(); err == nil {
			m.cursor = m.engine.CurrentIndex()
			m.setStatus("⏮ Previous track")
		}
		return m, nil

	case key.Matches(msg, keys.VolUp):
		if m.mode == ModeRadio && m.radioPlayer != nil {
			m.radioPlayer.VolumeUp()
			m.setStatus(fmt.Sprintf("🔊 Volume: %d%%", m.radioPlayer.GetVolumePercent()))
			return m, nil
		}

		m.engine.VolumeUp()
		m.setStatus(fmt.Sprintf("🔊 Volume: %d%%", m.engine.GetVolumePercent()))
		return m, nil

	case key.Matches(msg, keys.VolDown):
		if m.mode == ModeRadio && m.radioPlayer != nil {
			m.radioPlayer.VolumeDown()
			m.setStatus(fmt.Sprintf("🔉 Volume: %d%%", m.radioPlayer.GetVolumePercent()))
			return m, nil
		}

		m.engine.VolumeDown()
		m.setStatus(fmt.Sprintf("🔉 Volume: %d%%", m.engine.GetVolumePercent()))
		return m, nil

	case key.Matches(msg, keys.Loop):
		looping := m.engine.ToggleLoop()
		if looping {
			m.setStatus("🔁 Loop ON")
		} else {
			m.setStatus("🔁 Loop OFF")
		}
		return m, nil

	case key.Matches(msg, keys.Pomo):
		if m.pomo.Phase() == pomodoro.PhaseStopped {
			m.pomo.Start()
			m.setStatus("🍅 Pomodoro started!")
			// Auto-play if not playing
			if !m.engine.IsPlaying() && m.engine.TrackCount() > 0 {
				if m.engine.CurrentIndex() < 0 {
					m.engine.Play(0)
				} else {
					m.engine.Pause() // unpause
				}
			}
		} else {
			m.pomo.Stop()
			m.setStatus("🍅 Pomodoro stopped")
		}
		return m, nil

	case key.Matches(msg, keys.PomoPause):
		m.pomo.Pause()
		if m.pomo.IsRunning() {
			m.setStatus("🍅 Pomodoro resumed")
		} else {
			m.setStatus("🍅 Pomodoro paused")
		}
		return m, nil

	case key.Matches(msg, keys.PomoSkip):
		m.pomo.Skip()
		m.setStatus(fmt.Sprintf("🍅 Skipped to %s", m.pomo.Phase()))
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		maxIdx := m.engine.TrackCount() - 1
		if m.mode == ModeRadio && m.radioPlayer != nil {
			maxIdx = m.radioPlayer.StationCount() - 1
		}
		if m.cursor < maxIdx {
			m.cursor++
		}
		return m, nil

	case key.Matches(msg, keys.Enter):
		if m.mode == ModeRadio && m.radioPlayer != nil {
			// Check for manual retry
			if m.radioPlayer.CanRetry() {
				if m.radioPlayer.Retry() {
					m.setStatus("🔄 Retrying...")
				}
				return m, nil
			}

			if m.cursor >= 0 && m.cursor < m.radioPlayer.StationCount() {
				if err := m.radioPlayer.Play(m.cursor); err != nil {
					m.setStatus("❌ " + err.Error())
				} else {
					m.setStatus("📻 Connecting...")
				}
			}
			return m, nil
		}

		if m.cursor >= 0 && m.cursor < m.engine.TrackCount() {
			m.engine.Play(m.cursor)
			m.setStatus("▶ Playing")
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) setStatus(msg string) {
	m.statusMsg = msg
	m.statusExp = time.Now().Add(3 * time.Second)
}

func (m *Model) updateVisualizer() {
	isPlaying := m.engine.IsPlaying()
	if m.mode == ModeRadio && m.radioPlayer != nil {
		isPlaying = m.radioPlayer.IsPlaying()
	}

	if isPlaying {
		for i := range m.vizBars {
			target := rand.Float64()*0.8 + 0.1
			m.vizBars[i] = m.vizBars[i]*0.6 + target*0.4
		}
	} else {
		for i := range m.vizBars {
			m.vizBars[i] *= 0.85
		}
	}
}

// ── View ────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var sections []string

	// Header
	sections = append(sections, m.renderHeader())

	// Visualizer
	sections = append(sections, m.renderVisualizer())

	// Now Playing
	sections = append(sections, m.renderNowPlaying())

	// Pomodoro Timer
	if m.pomo.Phase() != pomodoro.PhaseStopped {
		sections = append(sections, m.renderPomodoro())
	}

	// Track List / Station List
	if m.mode == ModeRadio {
		sections = append(sections, m.renderStationList())
	} else {
		sections = append(sections, m.renderTrackList())
	}

	// Status
	if time.Now().Before(m.statusExp) {
		sections = append(sections, statusStyle.Render("  "+m.statusMsg))
	}

	// Help
	sections = append(sections, m.renderHelp())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderHeader() string {
	logo := `
   ██████╗  ██████╗       ██████╗ ███████╗ █████╗ ████████╗███████╗
  ██╔════╝ ██╔═══██╗      ██╔══██╗██╔════╝██╔══██╗╚══██╔══╝██╔════╝
  ██║  ███╗██║   ██║█████╗██████╔╝█████╗  ███████║   ██║   ███████╗
  ██║   ██║██║   ██║╚════╝██╔══██╗██╔══╝  ██╔══██║   ██║   ╚════██║
  ╚██████╔╝╚██████╔╝      ██████╔╝███████╗██║  ██║   ██║   ███████║
   ╚═════╝  ╚═════╝       ╚═════╝ ╚══════╝╚═╝  ╚═╝   ╚═╝   ╚══════╝`

	subtitle := "  ☕ lofi beats to relax/study to"

	return lipgloss.NewStyle().
		Foreground(colorPrimary).
		Render(logo) + "\n" +
		mutedStyle.Render(subtitle)
}

func (m Model) renderVisualizer() string {
	if len(m.vizBars) == 0 {
		return ""
	}

	barChars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	maxBars := 30
	if m.width > 10 {
		maxBars = min(m.width-10, 50)
	}

	var viz strings.Builder
	viz.WriteString("  ")

	for i := 0; i < maxBars && i < len(m.vizBars); i++ {
		level := int(m.vizBars[i] * float64(len(barChars)-1))
		if level < 0 {
			level = 0
		}
		if level >= len(barChars) {
			level = len(barChars) - 1
		}

		// Color gradient based on height
		var color lipgloss.Color
		switch {
		case level >= 6:
			color = colorLove
		case level >= 4:
			color = colorAccent
		case level >= 2:
			color = colorSecondary
		default:
			color = colorPrimary
		}

		viz.WriteString(lipgloss.NewStyle().Foreground(color).Render(barChars[level]))
	}

	return "\n" + viz.String() + "\n"
}

func (m Model) renderNowPlaying() string {
	if m.mode == ModeRadio {
		return m.renderRadioNowPlaying()
	}

	track := m.engine.CurrentTrack()

	var trackName string
	if track != nil {
		trackName = track.Name
	} else {
		trackName = "No track selected"
	}

	var stateIcon string
	if m.engine.IsPlaying() {
		stateIcon = "▶"
	} else if m.engine.IsPaused() {
		stateIcon = "⏸"
	} else {
		stateIcon = "⏹"
	}

	loopIndicator := ""
	if m.engine.IsLoop() {
		loopIndicator = " 🔁"
	}

	pos := m.engine.GetPosition()
	dur := m.engine.GetDuration()
	progressBar := m.renderProgressBar(pos, dur, 40)

	volPct := m.engine.GetVolumePercent()
	volBar := m.renderVolumeBar(volPct, 15)

	nowPlaying := fmt.Sprintf(
		"  %s %s%s\n  %s\n  %s",
		lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(stateIcon),
		lipgloss.NewStyle().Foreground(colorText).Bold(true).Render(trackName),
		mutedStyle.Render(loopIndicator),
		progressBar,
		volBar,
	)

	return boxStyle.Width(min(m.width-4, 70)).Render(nowPlaying)
}

func (m Model) renderRadioNowPlaying() string {
	var stateIcon string
	var statusText string
	var stationName string

	if m.radioPlayer == nil {
		stateIcon = "⏹"
		statusText = "No radio player"
		stationName = "No station"
	} else if m.radioPlayer.IsConnecting() {
		stateIcon = "🔄"
		statusText = "Connecting..."
		stationName = m.radioPlayer.CurrentStation().Name
	} else if m.radioPlayer.IsReconnecting() {
		stateIcon = "🔄"
		retryCount := m.radioPlayer.ReconnectCount()
		maxRetries := m.radioPlayer.MaxRetries()
		statusText = fmt.Sprintf("Reconnecting... (attempt %d/%d)", retryCount+1, maxRetries)
		stationName = m.radioPlayer.CurrentStation().Name
	} else if m.radioPlayer.CanRetry() {
		stateIcon = "❌"
		err := m.radioPlayer.Error()
		statusText = fmt.Sprintf("Connection failed: %v (press Enter to retry)", err)
		stationName = m.radioPlayer.CurrentStation().Name
	} else if m.radioPlayer.IsPlaying() {
		stateIcon = "📻"
		statusText = "LIVE"
		stationName = m.radioPlayer.CurrentStation().Name
	} else if m.radioPlayer.IsPaused() {
		stateIcon = "⏸"
		statusText = "Paused"
		stationName = m.radioPlayer.CurrentStation().Name
	} else {
		stateIcon = "⏹"
		statusText = "Stopped"
		stationName = "No station"
	}

	liveIndicator := lipgloss.NewStyle().Foreground(colorLove).Bold(true).Render(" 📍 LIVE")

	volPct := m.radioPlayer.GetVolumePercent()
	volBar := m.renderVolumeBar(volPct, 15)

	nowPlaying := fmt.Sprintf(
		"  %s %s%s\n  %s\n  %s",
		lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(stateIcon),
		lipgloss.NewStyle().Foreground(colorText).Bold(true).Render(stationName),
		liveIndicator,
		statusStyle.Render("  "+statusText),
		volBar,
	)

	return boxStyle.Width(min(m.width-4, 70)).Render(nowPlaying)
}

func (m Model) renderProgressBar(pos, dur time.Duration, width int) string {
	var progress float64
	if dur > 0 {
		progress = float64(pos) / float64(dur)
		if progress > 1 {
			progress = 1
		}
	}

	filled := int(progress * float64(width))
	empty := width - filled

	bar := progressStyle.Render(strings.Repeat("━", filled)) +
		mutedStyle.Render(strings.Repeat("─", empty))

	posStr := formatDuration(pos)
	durStr := formatDuration(dur)

	return fmt.Sprintf("  %s %s %s",
		mutedStyle.Render(posStr),
		bar,
		mutedStyle.Render(durStr),
	)
}

func (m Model) renderVolumeBar(pct int, width int) string {
	filled := int(math.Round(float64(pct) / 100 * float64(width)))
	empty := width - filled

	bar := lipgloss.NewStyle().Foreground(colorSecondary).Render(strings.Repeat("█", filled)) +
		mutedStyle.Render(strings.Repeat("░", empty))

	return fmt.Sprintf("  🔊 %s %d%%", bar, pct)
}

func (m Model) renderPomodoro() string {
	remaining := m.pomo.Remaining()
	mins := int(remaining.Minutes())
	secs := int(remaining.Seconds()) % 60

	timeStr := fmt.Sprintf("%02d:%02d", mins, secs)

	// Progress bar for pomodoro
	progress := m.pomo.Progress()
	width := 30
	filled := int(progress * float64(width))
	empty := width - filled

	bar := pomoStyle.Render(strings.Repeat("█", filled)) +
		mutedStyle.Render(strings.Repeat("░", empty))

	phase := m.pomo.Phase().String()
	sessions := m.pomo.Sessions()

	pauseIndicator := ""
	if !m.pomo.IsRunning() {
		pauseIndicator = " [PAUSED]"
	}

	pomoContent := fmt.Sprintf(
		"  %s %s%s\n  %s %s\n  Sessions: %d",
		pomoStyle.Render(phase),
		lipgloss.NewStyle().Foreground(colorText).Bold(true).Render(timeStr),
		mutedStyle.Render(pauseIndicator),
		bar,
		mutedStyle.Render(fmt.Sprintf("%.0f%%", progress*100)),
		sessions,
	)

	return "\n" + boxStyle.
		BorderForeground(colorLove).
		Width(min(m.width-4, 70)).
		Render(pomoContent)
}

func (m Model) renderStationList() string {
	stations := m.radioPlayer.Stations()
	if len(stations) == 0 {
		return "\n" + mutedStyle.Render("  No radio stations available.")
	}

	header := titleStyle.Render("📻 Radio Stations")

	maxVisible := 8
	if m.height > 30 {
		maxVisible = min(m.height-25, 15)
	}

	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(stations) {
		end = len(stations)
	}

	var list strings.Builder
	for i := start; i < end; i++ {
		station := stations[i]
		currentIdx := m.radioPlayer.CurrentStationIndex()

		var line string
		if i == currentIdx && i == m.cursor {
			line = activeTrackStyle.Render(fmt.Sprintf("▸ 📻 %s", station.Name))
		} else if i == currentIdx {
			line = activeTrackStyle.Render(fmt.Sprintf("  📻 %s", station.Name))
		} else if i == m.cursor {
			line = selectedTrackStyle.Render(fmt.Sprintf("▸   %s", station.Name))
		} else {
			line = trackStyle.Render(fmt.Sprintf("    %s", station.Name))
		}

		if station.Genre != "" {
			line += mutedStyle.Render(fmt.Sprintf(" (%s)", station.Genre))
		}
		list.WriteString(line + "\n")
	}

	scrollInfo := ""
	if len(stations) > maxVisible {
		scrollInfo = mutedStyle.Render(fmt.Sprintf("  [%d/%d stations]", m.cursor+1, len(stations)))
	}

	return "\n" + header + "\n" + list.String() + scrollInfo
}

func (m Model) renderTrackList() string {
	tracks := m.engine.Tracks()
	if len(tracks) == 0 {
		return "\n" + mutedStyle.Render("  No tracks found. Add .mp3 files to the music/ directory.")
	}

	header := titleStyle.Render("♪ Track List")

	// Calculate visible window
	maxVisible := 8
	if m.height > 30 {
		maxVisible = min(m.height-25, 15)
	}

	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(tracks) {
		end = len(tracks)
	}

	var list strings.Builder
	for i := start; i < end; i++ {
		track := tracks[i]
		currentIdx := m.engine.CurrentIndex()

		var line string
		if i == currentIdx && i == m.cursor {
			line = activeTrackStyle.Render(fmt.Sprintf("▸ ♪ %s", track.Name))
		} else if i == currentIdx {
			line = activeTrackStyle.Render(fmt.Sprintf("  ♪ %s", track.Name))
		} else if i == m.cursor {
			line = selectedTrackStyle.Render(fmt.Sprintf("▸   %s", track.Name))
		} else {
			line = trackStyle.Render(fmt.Sprintf("    %s", track.Name))
		}
		list.WriteString(line + "\n")
	}

	// Scroll indicators
	scrollInfo := ""
	if len(tracks) > maxVisible {
		scrollInfo = mutedStyle.Render(fmt.Sprintf("  [%d/%d tracks]", m.cursor+1, len(tracks)))
	}

	return "\n" + header + "\n" + list.String() + scrollInfo
}

func (m Model) renderHelp() string {
	if m.showHelp {
		var helpItems []string
		if m.mode == ModeRadio {
			helpItems = []string{
				"space: play/pause", "n/p: next/prev station",
				"+/-: volume", "↑↓: navigate stations",
				"enter: play/retry", "tab: switch mode",
				"t: start/stop pomodoro", "T: pause pomodoro",
				"s: skip phase", "?: toggle help", "q: quit",
			}
		} else {
			helpItems = []string{
				"space: play/pause", "n: next", "p: prev",
				"+/-: volume", "l: loop", "↑↓: navigate",
				"enter: play track", "tab: switch mode",
				"t: start/stop pomodoro", "T: pause pomodoro",
				"s: skip phase", "?: toggle help", "q: quit",
			}
		}
		return "\n" + helpStyle.Render("  "+strings.Join(helpItems, " │ "))
	}

	return "\n" + helpStyle.Render("  Press ? for help")
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	mins := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
