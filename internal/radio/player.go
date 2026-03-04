package radio

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

// Player handles streaming internet radio
type Player struct {
	mu sync.Mutex

	stations       []Station
	currentStation int

	httpResp *http.Response
	streamer beep.StreamSeekCloser
	ctrl     *beep.Ctrl
	volume   *effects.Volume
	closer   io.Closer

	playing     bool
	paused      bool
	volumeLevel float64

	// Connection state
	connecting bool
	err        error

	// Reconnection state
	reconnecting   bool
	reconnectCount int
	maxRetries     int
	retryChan      chan struct{}
	stopReconnect  chan struct{}
}

// NewPlayer creates a new radio player with default stations
func NewPlayer() *Player {
	return &Player{
		stations:       DefaultStations(),
		currentStation: -1,
		volumeLevel:    0,
		maxRetries:     3,
		retryChan:      make(chan struct{}, 1),
		stopReconnect:  make(chan struct{}, 1),
	}
}

// Stations returns available stations
func (p *Player) Stations() []Station {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := make([]Station, len(p.stations))
	copy(result, p.stations)
	return result
}

// StationCount returns number of stations
func (p *Player) StationCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.stations)
}

// CurrentStation returns the currently playing station
func (p *Player) CurrentStation() *Station {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.currentStation < 0 || p.currentStation >= len(p.stations) {
		return nil
	}
	s := p.stations[p.currentStation]
	return &s
}

// CurrentStationIndex returns the current station index
func (p *Player) CurrentStationIndex() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.currentStation
}

// Play starts playing a station by index
func (p *Player) Play(index int) error {
	p.mu.Lock()
	if index < 0 || index >= len(p.stations) {
		p.mu.Unlock()
		return fmt.Errorf("station index %d out of range", index)
	}
	station := p.stations[index]
	p.mu.Unlock()

	// Stop current stream and any reconnection attempts
	// Close stopReconnect first to signal the old goroutine to exit
	// Use recover to handle case where it's already closed
	func() {
		defer func() { recover() }()
		close(p.stopReconnect)
	}()
	p.stopReconnect = make(chan struct{})
	p.retryChan = make(chan struct{}, 1)

	p.Stop()

	p.mu.Lock()
	p.connecting = true
	p.err = nil
	p.currentStation = index
	p.reconnectCount = 0
	p.reconnecting = false
	p.mu.Unlock()

	// Connect in a goroutine to avoid blocking the UI
	go p.connectWithRetry(station)

	return nil
}

func (p *Player) connectWithRetry(station Station) {
	backoff := 2 * time.Second

	for {
		err := p.connectAndPlay(station)
		p.mu.Lock()
		p.connecting = false
		if err != nil {
			p.err = err
			p.playing = false
			p.reconnectCount++

			if p.reconnectCount < p.maxRetries {
				p.reconnecting = true
				p.mu.Unlock()

				// Wait for exponential backoff or manual retry
				select {
				case <-p.stopReconnect:
					p.mu.Lock()
					p.reconnecting = false
					p.mu.Unlock()
					return
				case <-time.After(backoff):
					backoff *= 2 // Exponential backoff: 2s, 4s, 8s
				}
				p.mu.Lock()
				p.reconnecting = false
				p.mu.Unlock()

				// Try again
				continue
			}
		} else {
			p.reconnectCount = 0
		}
		p.mu.Unlock()

		// Check for manual retry request
		if p.reconnectCount >= p.maxRetries {
			select {
			case <-p.retryChan:
				p.mu.Lock()
				p.reconnectCount = 0
				p.err = nil
				p.connecting = true
				p.mu.Unlock()
				backoff = 2 * time.Second
				continue
			case <-p.stopReconnect:
				return
			}
		}

		return
	}
}

func (p *Player) connectAndPlay(station Station) error {
	// Create HTTP request
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest("GET", station.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Some streams need specific headers
	req.Header.Set("User-Agent", "go-beats/1.0")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", station.Name, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("stream returned status %d", resp.StatusCode)
	}

	// Decode MP3 stream
	streamer, format, err := mp3.Decode(resp.Body)
	if err != nil {
		resp.Body.Close()
		return fmt.Errorf("failed to decode stream from %s: %w", station.Name, err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.httpResp = resp
	p.streamer = streamer
	p.closer = resp.Body

	// Resample if needed
	targetSR := beep.SampleRate(44100)
	var baseStreamer beep.Streamer
	if format.SampleRate != targetSR {
		baseStreamer = beep.Resample(4, format.SampleRate, targetSR, streamer)
	} else {
		baseStreamer = streamer
	}

	// Wrap with control and volume
	p.ctrl = &beep.Ctrl{Streamer: baseStreamer}
	p.volume = &effects.Volume{
		Streamer: p.ctrl,
		Base:     2,
		Volume:   p.volumeLevel,
	}

	p.playing = true
	p.paused = false

	speaker.Play(p.volume)

	return nil
}

// Stop stops the current stream
func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	speaker.Clear()

	if p.streamer != nil {
		p.streamer.Close()
		p.streamer = nil
	}
	if p.httpResp != nil {
		p.httpResp.Body.Close()
		p.httpResp = nil
	}

	p.ctrl = nil
	p.volume = nil
	p.playing = false
	p.paused = false
}

// Pause toggles pause on the stream
func (p *Player) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ctrl == nil {
		return
	}

	speaker.Lock()
	p.ctrl.Paused = !p.ctrl.Paused
	p.paused = p.ctrl.Paused
	speaker.Unlock()
}

// NextStation plays the next station
func (p *Player) NextStation() error {
	p.mu.Lock()
	if len(p.stations) == 0 {
		p.mu.Unlock()
		return fmt.Errorf("no stations available")
	}
	next := (p.currentStation + 1) % len(p.stations)
	p.mu.Unlock()
	return p.Play(next)
}

// PrevStation plays the previous station
func (p *Player) PrevStation() error {
	p.mu.Lock()
	if len(p.stations) == 0 {
		p.mu.Unlock()
		return fmt.Errorf("no stations available")
	}
	prev := p.currentStation - 1
	if prev < 0 {
		prev = len(p.stations) - 1
	}
	p.mu.Unlock()
	return p.Play(prev)
}

// VolumeUp increases volume
func (p *Player) VolumeUp() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.volumeLevel < 5 {
		p.volumeLevel += 0.5
	}
	if p.volume != nil {
		speaker.Lock()
		p.volume.Volume = p.volumeLevel
		speaker.Unlock()
	}
}

// VolumeDown decreases volume
func (p *Player) VolumeDown() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.volumeLevel > -7 {
		p.volumeLevel -= 0.5
	}
	if p.volume != nil {
		speaker.Lock()
		p.volume.Volume = p.volumeLevel
		speaker.Unlock()
	}
}

// GetVolumePercent returns volume as 0-100
func (p *Player) GetVolumePercent() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	pct := int(((p.volumeLevel + 7) / 12) * 100)
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return pct
}

// IsPlaying returns whether the radio is playing
func (p *Player) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playing && !p.paused
}

// IsPaused returns whether the radio is paused
func (p *Player) IsPaused() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.paused
}

// IsConnecting returns whether a connection is in progress
func (p *Player) IsConnecting() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.connecting
}

// IsReconnecting returns whether a reconnection is in progress
func (p *Player) IsReconnecting() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.reconnecting
}

// ReconnectCount returns the number of reconnection attempts
func (p *Player) ReconnectCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.reconnectCount
}

// CanRetry returns whether manual retry is available (after max retries reached)
func (p *Player) CanRetry() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.reconnectCount >= p.maxRetries && !p.reconnecting && p.err != nil
}

// Retry manually retries the last failed connection
func (p *Player) Retry() bool {
	p.mu.Lock()
	if p.reconnectCount < p.maxRetries || p.err == nil {
		p.mu.Unlock()
		return false
	}
	p.mu.Unlock()

	select {
	case p.retryChan <- struct{}{}:
		return true
	default:
		return false
	}
}

// Error returns the last error, if any
func (p *Player) Error() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}

// MaxRetries returns the maximum number of reconnection attempts
func (p *Player) MaxRetries() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.maxRetries
}

// CheckStream checks if the stream is still active and attempts reconnection if not
// This should be called periodically by the UI tick
func (p *Player) CheckStream() {
	p.mu.Lock()
	wasPlaying := p.playing && !p.paused
	isReconnecting := p.reconnecting
	currentStation := p.currentStation
	var station Station
	if currentStation >= 0 && currentStation < len(p.stations) {
		station = p.stations[currentStation]
	}
	stillPlaying := p.playing && !p.paused
	p.mu.Unlock()

	if wasPlaying && !stillPlaying && !isReconnecting && currentStation >= 0 {
		p.mu.Lock()
		p.reconnecting = true
		p.err = fmt.Errorf("stream disconnected")
		p.mu.Unlock()

		go p.connectWithRetry(station)
	}
}
