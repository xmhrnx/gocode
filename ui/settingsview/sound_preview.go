package settingsview

import (
	"github.com/xmhrnx/giode"

	"gocode/internal/chime"
	"gocode/ui/core"
)

// soundPreview plays each sound at the set volume.
type soundPreview struct {
	buttons []giode.Element
}

func newSoundPreview(s *core.Services) *soundPreview {
	sounds := []struct {
		label string
		sound chime.Sound
	}{{"Done", chime.Done}, {"Permission", chime.Attention}, {"Error", chime.Failure}}
	p := &soundPreview{}
	for _, entry := range sounds {
		p.buttons = append(p.buttons, smallButton(entry.label, false).OnClick(func() { s.PreviewSound(entry.sound) }))
	}
	return p
}

func (p *soundPreview) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter}, p.buttons...).Layout(gtx)
}
