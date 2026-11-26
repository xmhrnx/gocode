package remoteui

import (
	"github.com/xmhrnx/giode"

	"gocode/internal/remote/client"
	"gocode/internal/remote/wire"
	"gocode/ui/components"
	"gocode/ui/theme"
)

// chatsScreen is a desktop's open chats.
type chatsScreen struct {
	app    *App
	back   *components.IconButton
	rows   map[string]*giode.PressableElement
	scroll giode.Element
}

func newChatsScreen(a *App) *chatsScreen {
	c := &chatsScreen{app: a, back: backButton(a), rows: map[string]*giode.PressableElement{}}
	c.scroll = giode.Scroll(c.list, giode.Styles{Width: -1, Height: -1})
	return c
}

func (c *chatsScreen) Layout(gtx giode.Context) giode.Dimensions {
	l := c.app.link
	if l == nil {
		return giode.Dimensions{}
	}
	if l.watching != "" {
		giode.AfterLayout(func() { l.watch("") }) // back from a chat
	}
	return screen(c.back, l.desktop.Name, linkLabel(l), nil, c.body(l)).Layout(gtx)
}

func (c *chatsScreen) body(l *link) giode.Element {
	switch {
	case l.state == linkConnecting:
		return note("loader-circle", "Connecting…", "Looking for "+l.desktop.Name+" on this network"+relayNote(l.desktop)+".")
	case l.state == linkFailed:
		return note("circle-x", "Can't reach "+l.desktop.Name, l.err+" Trying again…")
	case len(l.chats.Chats) == 0:
		return note("message-square", "No chats", "Chats open on the desktop show here once something was sent in them.")
	}
	return giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, c.scroll)
}

func (c *chatsScreen) list() giode.Element {
	l := c.app.link
	if l == nil {
		return giode.Box(giode.Styles{})
	}
	var rows []giode.Element
	for _, chat := range l.chats.Chats {
		press, ok := c.rows[chat.ID]
		if !ok {
			id := chat.ID
			press = giode.Pressable(func(state giode.PressState) giode.Element { return c.row(id, state) }).
				Label("Open chat").OnClick(func(giode.ClickEvent) {
				if c.app.link != nil {
					c.app.link.watch(id)
					c.app.router.Push("/chat")
				}
			})
			c.rows[chat.ID] = press
		}
		rows = append(rows, press)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 8, Padding: giode.SymmetricInset(12, 4)}, rows...)
}

func (c *chatsScreen) row(id string, state giode.PressState) giode.Element {
	var chat wire.ChatSummary
	if l := c.app.link; l != nil {
		for _, ch := range l.chats.Chats {
			if ch.ID == id {
				chat = ch
			}
		}
	}
	background := theme.Panel
	if state.Pressed || state.Hovered {
		background = theme.Raised
	}
	badge := giode.Element(theme.Icon("chevron-right", 16, theme.Faint))
	switch {
	case chat.Waiting:
		badge = giode.Text("Needs you", giode.Styles{Color: theme.Warning, FontSize: theme.Sizes.Small, FontWeight: giode.FontWeightMedium})
	case chat.Busy:
		badge = giode.Spinner(giode.Styles{Width: 14, Height: 14, Color: theme.Primary})
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Padding: giode.SymmetricInset(14, 12),
		BorderRadius: 12, Background: background, BorderWidth: 1, BorderColor: theme.SoftBorder},
		giode.Stack(giode.Styles{Gap: 2, FlexGrow: 1},
			giode.Text(chat.Title, giode.Styles{Color: theme.Text, FontWeight: giode.FontWeightMedium, MaxLines: 2}),
			giode.Text(chat.Folder, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small, MaxLines: 1})),
		badge)
}

// linkLabel says how the desktop is connected.
func linkLabel(l *link) string {
	switch l.state {
	case linkOpen:
		if l.route == client.RouteRelay {
			return "Connected through the relay"
		}
		return "Connected on your Wi-Fi"
	case linkFailed:
		return "Not connected"
	}
	return "Connecting…"
}

func relayNote(d client.Desktop) string {
	if d.Relay != "" {
		return ", then through the relay"
	}
	return ""
}
