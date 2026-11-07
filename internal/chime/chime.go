// Package chime makes GoCode's sounds: short, soft tones in the low and
// middle range, like an old computer's, generated rather than recorded.
package chime

import "math"

// Sound is one of the app's sounds.
type Sound int

// note is one tone of a sound.
type note struct {
	hz       float64
	duration float64 // seconds
	gap      float64 // silence after it, seconds
}

// The sounds.
const (
	// Done: a run finished.
	Done Sound = iota
	// Attention: the agent asks for permission.
	Attention
	// Failure: a run stopped with an error.
	Failure
)

// SampleRate is the rate Samples generates at.
const SampleRate = 44100

// melodies are each sound's notes: nothing above C5 (523 Hz), so none
// of them is shrill.
var melodies = map[Sound][]note{
	Done:      {{hz: 392, duration: 0.11, gap: 0.02}, {hz: 523.25, duration: 0.22}},
	Attention: {{hz: 440, duration: 0.09, gap: 0.07}, {hz: 440, duration: 0.12}},
	Failure:   {{hz: 311.13, duration: 0.12, gap: 0.03}, {hz: 233.08, duration: 0.24}},
}

const (
	// peak is the loudest a sound gets at full volume.
	peak = 0.3
	// attack and release soften each note's start and end, so there is
	// no click.
	attack  = 0.008
	release = 0.05
	// overtone adds a quiet octave above, which makes a plain sine sound
	// like a computer's beep rather than a tuning fork.
	overtone = 0.22
)

// Samples returns sound at volume (0 to 1) as samples from -1 to 1 at
// SampleRate.
func Samples(sound Sound, volume float64) []float32 {
	volume = math.Max(0, math.Min(1, volume))
	var out []float32
	for _, n := range melodies[sound] {
		out = append(out, tone(n, volume)...)
		out = append(out, make([]float32, int(n.gap*SampleRate))...)
	}
	return out
}

func tone(n note, volume float64) []float32 {
	count := int(n.duration * SampleRate)
	samples := make([]float32, count)
	for i := range samples {
		t := float64(i) / SampleRate
		wave := math.Sin(2*math.Pi*n.hz*t) + overtone*math.Sin(4*math.Pi*n.hz*t)
		samples[i] = float32(peak * volume * envelope(t, n.duration) * wave / (1 + overtone))
	}
	return samples
}

// envelope rises over attack, decays gently, and falls to silence over
// release at the end.
func envelope(t, duration float64) float64 {
	level := math.Exp(-3 * t / duration)
	if t < attack {
		level *= t / attack
	}
	if left := duration - t; left < release {
		level *= left / release
	}
	return level
}
