package notification

import (
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/generators"
	"github.com/gopxl/beep/v2/speaker"
)

const sampleRate = beep.SampleRate(44100)

// PlayFocusEnd plays a gentle notification sound when focus time ends.
// Two ascending tones: "time to take a break!"
func PlayFocusEnd() {
	playTones([]tone{
		{freq: 523.25, duration: 150 * time.Millisecond}, // C5
		{freq: 659.25, duration: 150 * time.Millisecond}, // E5
		{freq: 783.99, duration: 250 * time.Millisecond}, // G5
	})
}

// PlayBreakEnd plays a notification sound when break time ends.
// Two descending tones: "back to work!"
func PlayBreakEnd() {
	playTones([]tone{
		{freq: 783.99, duration: 150 * time.Millisecond}, // G5
		{freq: 659.25, duration: 150 * time.Millisecond}, // E5
		{freq: 523.25, duration: 250 * time.Millisecond}, // C5
	})
}

type tone struct {
	freq     float64
	duration time.Duration
}

func playTones(tones []tone) {
	var streamers []beep.Streamer

	for _, t := range tones {
		sine, err := generators.SineTone(sampleRate, t.freq)
		if err != nil {
			continue
		}
		samples := sampleRate.N(t.duration)
		streamers = append(streamers, beep.Take(samples, sine))
		// Small silence gap between tones
		streamers = append(streamers, generators.Silence(sampleRate.N(50*time.Millisecond)))
	}

	if len(streamers) > 0 {
		speaker.Play(beep.Seq(streamers...))
	}
}
