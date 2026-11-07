package core

import (
	"errors"

	"github.com/xmhrnx/giode"

	"gocode/internal/chime"
)

// PlaySound plays sound if the settings allow it for that event.
func (s *Services) PlaySound(sound chime.Sound) {
	cfg := s.Settings.Sounds
	allowed := map[chime.Sound]bool{chime.Done: cfg.Done, chime.Attention: cfg.Permission, chime.Failure: cfg.Errors}
	if cfg.Enabled && allowed[sound] {
		s.PreviewSound(sound)
	}
}

// PreviewSound plays sound at the set volume, whatever the switches say.
func (s *Services) PreviewSound(sound chime.Sound) {
	wav := giode.EncodeWAV(chime.Samples(sound, float64(s.Settings.Sounds.Volume)/100), chime.SampleRate)
	go func() {
		err := s.App.PlaySound(wav)
		if err == nil {
			return
		}
		s.soundTrouble.Do(func() {
			if errors.Is(err, giode.ErrNotSupported) {
				s.Info("Sounds can't play here: no sound player found (on Linux, pw-play, paplay or aplay).")
				return
			}
			s.Error("Couldn't play a sound: " + err.Error())
		})
	}()
}
