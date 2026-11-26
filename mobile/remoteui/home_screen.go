package remoteui

import (
	"github.com/xmhrnx/giode"

	"gocode/internal/remote/client"
	"gocode/ui/components"
	"gocode/ui/theme"
)

// homeScreen lists the paired desktops.
type homeScreen struct {
	app    *App
	pair   *giode.ButtonElement
	rows   map[string]*desktopRow
	scroll giode.Element
}

func newHomeScreen(a *App) *homeScreen {
	h := &homeScreen{app: a, rows: map[string]*desktopRow{}}
	h.pair = wideButton("Pair with a desktop", func() { a.pairing.reset(); a.router.Push("/pair") })
	h.scroll = giode.Scroll(h.list, giode.Styles{Width: -1, Height: -1})
	return h
}

func (h *homeScreen) Layout(gtx giode.Context) giode.Dimensions {
	if h.app.link != nil {
		giode.AfterLayout(h.app.closeDesktop) // back from a desktop
	}
	body := giode.Element(giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, h.scroll))
	switch {
	case h.app.loadErr != "":
		body = note("triangle-alert", "Can't start", h.app.loadErr)
	case len(h.app.store.desktops) == 0:
		body = note("monitor", "No desktops yet",
			"In GoCode on your computer, open Settings → Phone, turn it on and choose “Pair a phone”. Then pair here with the code it shows.")
	}
	return screen(nil, "GoCode", "Your desktops", nil,
		body,
		giode.Row(giode.Styles{Width: -1, Padding: giode.SymmetricInset(16, 14)}, giode.Box(giode.Styles{Width: -1}, h.pair)),
	).Layout(gtx)
}

func (h *homeScreen) list() giode.Element {
	var rows []giode.Element
	for _, d := range h.app.store.desktops {
		row, ok := h.rows[d.Key]
		if !ok {
			row = newDesktopRow(h.app, d.Key)
			h.rows[d.Key] = row
		}
		rows = append(rows, row)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 8, Padding: giode.SymmetricInset(12, 4)}, rows...)
}

// desktopRow is one paired desktop: tap to open it.
type desktopRow struct {
	app    *App
	key    string
	press  *giode.PressableElement
	forget *components.IconButton
}

func newDesktopRow(a *App, key string) *desktopRow {
	r := &desktopRow{app: a, key: key}
	r.press = giode.Pressable(r.content).Label("Open").OnClick(func(giode.ClickEvent) {
		if d := a.store.find(key); d != nil && a.ready {
			a.openDesktop(*d)
		}
	})
	r.forget = components.NewIconButton("trash-2", "Forget this desktop", func() { a.store.remove(key) }).SetSize(16)
	return r
}

func (r *desktopRow) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 4, Wrap: giode.WrapNoWrap},
		giode.Box(giode.Styles{FlexGrow: 1}, r.press), r.forget).Layout(gtx)
}

func (r *desktopRow) content(state giode.PressState) giode.Element {
	d := r.app.store.find(r.key)
	if d == nil {
		return giode.Box(giode.Styles{})
	}
	background := theme.Panel
	if state.Pressed || state.Hovered {
		background = theme.Raised
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Padding: giode.SymmetricInset(14, 14),
		BorderRadius: 12, Background: background, BorderWidth: 1, BorderColor: theme.SoftBorder},
		theme.Icon("monitor", 20, theme.Primary),
		giode.Stack(giode.Styles{Gap: 2, FlexGrow: 1},
			giode.Text(d.Name, giode.Styles{Color: theme.Text, FontWeight: giode.FontWeightMedium}),
			giode.Text(reachLabel(*d), theme.Caption())),
		theme.Icon("chevron-right", 16, theme.Faint))
}

// reachLabel says how a desktop can be reached.
func reachLabel(d client.Desktop) string {
	if d.Relay != "" {
		return "On your Wi-Fi, or through the relay while it works"
	}
	return "On your Wi-Fi"
}
