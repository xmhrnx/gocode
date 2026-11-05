package history

import (
	"fmt"
	"time"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"

	"gocode/internal/session"
	"gocode/ui/components"
	"gocode/ui/theme"
)

// row is one past chat: its title, folder and age, with a delete button.
type row struct {
	summary session.Summary
	press   *giode.PressableElement
	remove  *components.IconButton
}

func newRow(d *Dialog, s session.Summary) *row {
	r := &row{summary: s}
	r.remove = components.NewIconButton("trash-2", "Delete chat", func() { d.remove(s.ID) }).SetSize(14)
	r.press = giode.Pressable(r.content).Label(s.Title).OnClick(func(giode.ClickEvent) { d.choose(s.ID) })
	return r
}

func (r *row) Layout(gtx layout.Context) layout.Dimensions { return r.press.Layout(gtx) }

func (r *row) content(state giode.PressState) giode.Element {
	background := ""
	if state.Hovered {
		background = theme.Hover
	}
	title := r.summary.Title
	if title == "" {
		title = "Untitled chat"
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(12, 8), BorderRadius: 9, Background: background},
		theme.Icon("message-square", 15, theme.Muted),
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2},
			giode.Text(title, giode.Styles{Color: theme.Text, TextOverflow: giode.TextOverflowEllipsis}),
			giode.Text(r.summary.WorkDir, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Tiny,
				TextOverflow: giode.TextOverflowEllipsis}),
		),
		giode.Text(age(r.summary.Updated), giode.Styles{Color: theme.Faint, FontSize: theme.Sizes.Small}),
		r.remove,
	)
}

// age describes how long ago t was, coarsely.
func age(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
	return t.Format("Jan 2, 2006")
}
