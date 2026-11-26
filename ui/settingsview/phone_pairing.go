package settingsview

import (
	"context"
	"errors"
	"time"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/kit"

	"gocode/internal/remote/host"
	"gocode/ui/core"
	"gocode/ui/theme"
)

// pairingControl starts pairing and shows the code while a phone may
// use it.
type pairingControl struct {
	services *core.Services
	start    *giode.ButtonElement
	cancel   *giode.ButtonElement
	copy     giode.Element
	// code is shown while pairing; busy while the code is being made.
	code    string
	expires time.Time
	busy    bool
	stop    context.CancelFunc
}

// pairingWindow mirrors how long the host accepts a code.
const pairingWindow = 10 * time.Minute

func newPairingControl(s *core.Services) *pairingControl {
	p := &pairingControl{services: s}
	p.start = giode.Button("Pair a phone", giode.Styles{Background: theme.Primary, Color: theme.OnPrimary, BorderRadius: 8,
		FontSize: theme.Sizes.Small}).OnClick(p.begin)
	p.cancel = giode.Button("Cancel", giode.Styles{BorderColor: theme.Border, Color: theme.Text, BorderRadius: 8,
		FontSize: theme.Sizes.Small}).Variant(giode.VariantOutlined).OnClick(p.end)
	p.copy = giode.CopyButton(func() string { return p.code }, giode.Styles{Color: theme.Muted})
	return p
}

func (p *pairingControl) begin() {
	h := p.services.Phone
	if h == nil || p.busy {
		return
	}
	if !p.services.Settings.Phone.Enabled {
		p.services.Error("Turn on “Let the GoCode phone app connect” first.")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	p.busy, p.stop = true, cancel
	app := p.services.App
	go func() {
		code, result, err := h.Pair(ctx, "")
		app.Post(func() {
			p.busy = false
			if err != nil {
				p.services.Error("Couldn't start pairing: " + err.Error())
				return
			}
			p.code, p.expires = code, time.Now().Add(pairingWindow)
		})
		if err != nil {
			return
		}
		r := <-result
		app.Post(func() { p.finish(r) })
	}()
}

func (p *pairingControl) finish(r host.PairResult) {
	p.code = ""
	switch {
	case r.Err == nil:
		p.services.Info("Paired " + r.Device + ".")
	case errors.Is(r.Err, context.Canceled):
	default:
		p.services.Error("Pairing didn't finish: " + r.Err.Error())
	}
}

func (p *pairingControl) end() {
	if p.services.Phone != nil {
		p.services.Phone.CancelPairing()
	}
	if p.stop != nil {
		p.stop()
	}
	p.code = ""
}

func (p *pairingControl) Layout(gtx giode.Context) giode.Dimensions {
	switch {
	case p.busy:
		return giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
			giode.Spinner(giode.Styles{Width: 14, Height: 14, Color: theme.Primary}),
			giode.Text("Making a code…", theme.Caption())).Layout(gtx)
	case p.code == "":
		return giode.Row(giode.Styles{}, p.start).Layout(gtx)
	}
	left := time.Until(p.expires).Round(time.Second)
	kit.RequestFrameAt(gtx, time.Now().Add(time.Second))
	return giode.Stack(giode.Styles{Gap: 8},
		giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter},
			giode.Text(p.code, giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Heading, FontWeight: giode.FontWeightSemiBold,
				FontFamily: "JetBrains Mono"}),
			p.copy),
		giode.Text("Type this code in the GoCode phone app, on the same Wi-Fi as this computer. It works for "+
			left.String()+" more.", theme.Caption()),
		giode.Row(giode.Styles{}, p.cancel),
	).Layout(gtx)
}
