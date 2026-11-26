package mobile

import (
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/lazylist"

	"gocode/internal/remote/wire"
	"gocode/ui/components"
	"gocode/ui/theme"
)

// chatScreen follows one chat: its latest rows, the run's progress, a
// permission request, and a field to send prompts.
type chatScreen struct {
	app      *App
	back     *components.IconButton
	stop     *giode.ButtonElement
	list     *lazylist.List
	field    *input.Input
	send     *components.IconButton
	decision *permissionCard
}

func newChatScreen(a *App) *chatScreen {
	c := &chatScreen{app: a, back: backButton(a)}
	c.stop = giode.Button("Stop", giode.Styles{BorderColor: theme.Border, Color: theme.Text, BorderRadius: 8,
		FontSize: theme.Sizes.Small, Padding: giode.SymmetricInset(12, 6)}).Variant(giode.VariantOutlined).OnClick(c.stopRun)
	c.list = giode.LazyList(c.count, func(i int) any { return i }, c.row,
		giode.Styles{ScrollBar: giode.ScrollBar{Width: 4, AutoHide: true, TrackColor: "transparent"}}).
		Version(c.rowVersion).Anchor(giode.LazyAnchorBottom)
	c.field = giode.Input.Text("Message the agent", giode.Styles{Color: theme.Text, Background: "transparent",
		BorderColor: "transparent", Padding: giode.NoInset}).Lines(1, 6).Submit(func(string) { c.sendPrompt() })
	c.field.SubmitOn(giode.SubmitNever) // Return adds a line on a phone keyboard; the button sends
	c.send = components.NewIconButton("arrow-up", "Send", c.sendPrompt).SetSize(18)
	c.decision = newPermissionCard(a)
	return c
}

func (c *chatScreen) view() wire.ChatView {
	if c.app.link == nil {
		return wire.ChatView{}
	}
	return c.app.link.view
}

func (c *chatScreen) count() int { return len(c.view().Rows) }

func (c *chatScreen) rowVersion(i int) any {
	rows := c.view().Rows
	if i >= len(rows) {
		return 0
	}
	h := fnv.New64a()
	r := rows[i]
	fmt.Fprint(h, r.Kind, r.State, r.Text)
	return h.Sum64()
}

func (c *chatScreen) row(i int) giode.Element {
	rows := c.view().Rows
	if i >= len(rows) {
		return giode.Box(giode.Styles{})
	}
	return giode.Box(giode.Styles{Width: -1, Padding: giode.SymmetricInset(14, 5)}, rowView(rows[i]))
}

func (c *chatScreen) sendPrompt() {
	text := strings.TrimSpace(c.field.Text())
	l := c.app.link
	if text == "" || l == nil || l.watching == "" {
		return
	}
	l.act(wire.TypeSend, wire.Send{Chat: l.watching, Text: text})
	c.field.SetText("")
}

func (c *chatScreen) stopRun() {
	if l := c.app.link; l != nil && l.watching != "" {
		l.act(wire.TypeStop, wire.ChatRef{Chat: l.watching})
	}
}

func (c *chatScreen) Layout(gtx giode.Context) giode.Dimensions {
	l := c.app.link
	if l == nil {
		return giode.Dimensions{}
	}
	view := l.view
	title := view.Title
	if title == "" {
		title = "Chat"
	}
	parts := []giode.Element{giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, c.list)}
	if l.state != linkOpen {
		parts = append(parts, giode.Text(linkLabel(l)+". "+l.err, giode.Styles{Color: theme.Warning, FontSize: theme.Sizes.Small,
			Padding: giode.SymmetricInset(14, 6)}))
	}
	if view.Permission != nil {
		parts = append(parts, c.decision)
	}
	if view.Busy {
		parts = append(parts, c.progress(view))
	}
	parts = append(parts, c.composer())
	return screen(c.back, title, view.Folder, nil, parts...).Layout(gtx)
}

// progress shows the run's status, time and speed, with Stop.
func (c *chatScreen) progress(v wire.ChatView) giode.Element {
	text := v.Status + " · " + workingTime(time.Duration(v.Seconds)*time.Second)
	if v.TokensPerSecond > 0 {
		text += fmt.Sprintf(" · %.0f tok/s", v.TokensPerSecond)
	}
	if v.Queued > 0 {
		text += fmt.Sprintf(" · %d queued", v.Queued)
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Padding: giode.SymmetricInset(14, 6)},
		giode.Spinner(giode.Styles{Width: 14, Height: 14, Color: theme.Primary}),
		giode.Text(text, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small, FlexGrow: 1}),
		c.stop)
}

func (c *chatScreen) composer() giode.Element {
	return giode.Row(giode.Styles{Width: -1, Padding: giode.Inset{Left: 10, Right: 10, Top: 6, Bottom: 12}},
		giode.Row(giode.Styles{Width: -1, Gap: 6, Align: giode.AlignEnd, Wrap: giode.WrapNoWrap, Padding: giode.Inset{Left: 14, Right: 6, Top: 8, Bottom: 8},
			BorderRadius: 14, Background: theme.Panel, BorderWidth: 1, BorderColor: theme.Border},
			giode.Box(giode.Styles{FlexGrow: 1, Padding: giode.Inset{Top: 4, Bottom: 4}}, c.field),
			c.send))
}

// workingTime shows seconds, then minutes and seconds, then hours and
// minutes, as the desktop does.
func workingTime(d time.Duration) string {
	s := int(d / time.Second)
	switch {
	case s < 60:
		return fmt.Sprintf("%ds", s)
	case s < 3600:
		return fmt.Sprintf("%dm %02ds", s/60, s%60)
	}
	return fmt.Sprintf("%dh %02dm", s/3600, s%3600/60)
}
