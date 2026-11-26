package settingsview

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/components"
	"gocode/ui/core"
	"gocode/ui/theme"
)

// pairedPhones lists the paired phones, each with a remove button.
type pairedPhones struct {
	services *core.Services
	remove   map[string]*components.IconButton
}

func newPairedPhones(s *core.Services) *pairedPhones {
	return &pairedPhones{services: s, remove: map[string]*components.IconButton{}}
}

func (p *pairedPhones) Layout(gtx giode.Context) giode.Dimensions {
	h := p.services.Phone
	if h == nil || len(h.Devices().List()) == 0 {
		return giode.Text("No phones paired yet.", theme.Caption()).Layout(gtx)
	}
	var rows []giode.Element
	for _, dev := range h.Devices().List() {
		button, ok := p.remove[dev.Key]
		if !ok {
			key := dev.Key
			button = components.NewIconButton("trash-2", "Unpair "+dev.Name, func() {
				if err := h.Unpair(key); err != nil {
					p.services.Error("Couldn't unpair the phone: " + err.Error())
				}
			}).SetSize(14)
			p.remove[key] = button
		}
		seen := "Paired " + dev.Paired.Format("Jan 2, 2006")
		if !dev.LastSeen.IsZero() {
			seen += " · last connected " + dev.LastSeen.Format("Jan 2, 15:04")
		}
		rows = append(rows, giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter},
			theme.Icon("smartphone", 16, theme.Muted),
			giode.Stack(giode.Styles{Gap: 2, FlexGrow: 1},
				giode.Text(dev.Name, giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Small}),
				giode.Text(seen, theme.Caption())),
			button))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10}, rows...).Layout(gtx)
}
