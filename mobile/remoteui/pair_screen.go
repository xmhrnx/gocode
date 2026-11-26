package remoteui

import (
	"context"
	"strings"
	"time"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"gocode/internal/remote/client"
	"gocode/internal/remote/pair"
	"gocode/internal/remote/secure"
	"gocode/ui/components"
	"gocode/ui/theme"
)

// pairScreen takes the code a desktop shows and pairs with it.
type pairScreen struct {
	app    *App
	back   *components.IconButton
	code   *input.Input
	submit *giode.ButtonElement
	busy   bool
	err    string
}

// pairTimeout bounds getting the offer and the first connection.
const pairTimeout = 45 * time.Second

func newPairScreen(a *App) *pairScreen {
	p := &pairScreen{app: a, back: backButton(a)}
	p.code = giode.Input.Text("Code, like 7-crossover-clockwork", giode.Styles{Background: theme.Panel, BorderColor: theme.Border,
		BorderRadius: 10, Color: theme.Text, FontFamily: "JetBrains Mono", Padding: giode.SymmetricInset(14, 12)}).
		Submit(func(string) { p.start() })
	p.submit = wideButton("Pair", p.start)
	return p
}

// reset clears the screen for a new pairing.
func (p *pairScreen) reset() {
	p.code.SetText("")
	p.busy, p.err = false, ""
}

func (p *pairScreen) start() {
	code := strings.TrimSpace(p.code.Text())
	if p.busy || code == "" || !p.app.ready {
		return
	}
	p.busy, p.err = true, ""
	key, name := p.app.store.key, phoneName()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), pairTimeout)
		defer cancel()
		desk, err := pairWith(ctx, code, key, name)
		p.app.post(func() {
			p.busy = false
			if err != nil {
				p.err = err.Error()
				return
			}
			p.app.store.put(desk)
			p.app.toast("Paired with " + desk.Name + ".")
			p.app.router.Back()
			p.app.openDesktop(desk)
		})
	}()
}

// pairWith gets the desktop's offer under code and connects once with its
// token, which pairs this phone.
func pairWith(ctx context.Context, code string, key secure.KeyPair, name string) (client.Desktop, error) {
	offer, err := pair.Receive(ctx, "", code)
	if err != nil {
		return client.Desktop{}, err
	}
	desk := client.FromOffer(offer)
	s, err := client.Connect(ctx, key, name, &desk, offer.Token)
	if err != nil {
		return client.Desktop{}, err
	}
	s.Close()
	return desk, nil
}

func (p *pairScreen) Layout(gtx giode.Context) giode.Dimensions {
	status := giode.Element(giode.Box(giode.Styles{}))
	switch {
	case p.busy:
		status = giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
			giode.Spinner(giode.Styles{Width: 14, Height: 14, Color: theme.Primary}),
			giode.Text("Pairing…", theme.Caption()))
	case p.err != "":
		status = giode.Text(p.err, giode.Styles{Color: theme.Danger, FontSize: theme.Sizes.Small})
	}
	return screen(p.back, "Pair with a desktop", "", nil,
		giode.Stack(giode.Styles{Width: -1, Gap: 14, Padding: giode.SymmetricInset(16, 12)},
			giode.Text("In GoCode on your computer, open Settings → Phone and choose “Pair a phone”. Type the code it shows. "+
				"For pairing, the phone must be on the same Wi-Fi as the computer.", giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small}),
			giode.Box(giode.Styles{Width: -1}, p.code),
			giode.Box(giode.Styles{Width: -1}, p.submit),
			status,
		),
	).Layout(gtx)
}
