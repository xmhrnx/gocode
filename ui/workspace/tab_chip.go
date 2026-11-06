package workspace

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/components"
	"gocode/ui/theme"
)

// tabChip is one tab in the strip: icon, title and a close button.
type tabChip struct {
	workspace *Workspace
	id        int
	press     *giode.PressableElement
	closer    *components.IconButton
}

func newTabChip(w *Workspace, id int) *tabChip {
	c := &tabChip{workspace: w, id: id}
	c.closer = components.NewIconButton("x", "Close tab", func() { w.Close(id) }).SetSize(13)
	c.press = giode.Pressable(c.content).Label("Tab").OnClick(func(giode.ClickEvent) { w.activate(id) })
	return c
}

func (c *tabChip) Layout(gtx giode.Context) giode.Dimensions { return c.press.Layout(gtx) }

func (c *tabChip) content(state giode.PressState) giode.Element {
	index := c.workspace.indexOf(c.id)
	if index < 0 {
		return giode.Box(giode.Styles{})
	}
	tab := c.workspace.entries[index].tab
	c.press.Label(tab.Title())
	active := c.id == c.workspace.activeID
	background, color := "", theme.Muted
	switch {
	case active:
		background, color = theme.Canvas, theme.Text
	case state.Hovered:
		background, color = theme.Hover, theme.Text
	}
	closer := giode.Element(giode.Box(giode.Styles{Width: 25}))
	if active || state.Hovered || c.closer.Hovered() {
		closer = c.closer
	}
	return giode.Row(giode.Styles{
		Align: giode.AlignCenter, Gap: 7, Wrap: giode.WrapNoWrap,
		Padding: giode.Inset{Left: 12, Right: 3}, Height: 30, MaxWidth: 220,
		Background: background, BorderRadii: giode.TopCorners(9),
	}, c.icon(tab.Busy(), tab.Icon(), color),
		giode.Text(tab.Title(), giode.Styles{
			Color: color, FontSize: theme.Sizes.Small, FontWeight: giode.FontWeightMedium,
			TextOverflow: giode.TextOverflowEllipsis, MaxWidth: 150,
		}),
		closer,
	)
}

func (c *tabChip) icon(busy bool, name, color string) giode.Element {
	if busy {
		return giode.Spinner(giode.Styles{Width: 13, Height: 13, Color: theme.Primary})
	}
	return theme.Icon(name, 14, color)
}
