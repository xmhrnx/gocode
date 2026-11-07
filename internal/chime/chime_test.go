package chime

import (
	"math"
	"testing"
)

func TestSoundsAreShortQuietAndClickFree(t *testing.T) {
	for _, sound := range []Sound{Done, Attention, Failure} {
		samples := Samples(sound, 1)
		if seconds := float64(len(samples)) / SampleRate; seconds < 0.15 || seconds > 0.6 {
			t.Errorf("sound %d lasts %.2fs", sound, seconds)
		}
		loudest := 0.0
		for _, s := range samples {
			loudest = math.Max(loudest, math.Abs(float64(s)))
		}
		if loudest > peak+0.01 || loudest < peak/3 {
			t.Errorf("sound %d peaks at %.2f", sound, loudest)
		}
		if math.Abs(float64(samples[0])) > 0.01 || math.Abs(float64(samples[len(samples)-1])) > 0.01 {
			t.Errorf("sound %d starts or ends abruptly", sound)
		}
	}
}

func TestVolumeScales(t *testing.T) {
	loud, quiet := Samples(Done, 1), Samples(Done, 0.25)
	for i := range loud {
		if math.Abs(float64(quiet[i])*4-float64(loud[i])) > 1e-4 {
			t.Fatalf("sample %d: %f vs %f", i, quiet[i], loud[i])
		}
	}
	for _, s := range Samples(Done, 0) {
		if s != 0 {
			t.Fatal("volume 0 must be silent")
		}
	}
}

func TestNoHighNotes(t *testing.T) {
	for sound, notes := range melodies {
		for _, n := range notes {
			if n.hz > 530 {
				t.Errorf("sound %d has a %.0f Hz note", sound, n.hz)
			}
		}
	}
}
