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
	// width is the chip's width in pixels, for placing a dragged tab.
	width int
}

func newTabChip(s *TabStrip, id int) *tabChip {
	w := s.workspace
	c := &tabChip{workspace: w, id: id}
	c.closer = components.NewIconButton("x", "Close tab", func() { w.Close(id) }).SetSize(13)
	c.press = giode.Pressable(c.content).Label("Tab").
		OnClick(func(giode.ClickEvent) { w.activate(id) }).
		OnSecondaryClick(func(e giode.ClickEvent) { s.menu.openFor(id, e) }).
		OnLongPress(func(e giode.ClickEvent) { s.menu.openFor(id, e) }).
		OnDragStart(func(giode.DragEvent) { s.startDrag(id) }).
		OnDragMove(func(e giode.DragEvent) { s.moveDrag(id, e.DeltaDp.X) }).
		OnDragEnd(func(giode.DragEvent) { s.endDrag() })
	return c
}

func (c *tabChip) Layout(gtx giode.Context) giode.Dimensions {
	dims := c.press.Layout(gtx)
	c.width = dims.Size.X
	return dims
}

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
	case state.Dragging:
		background, color = theme.Raised, theme.Text
	case active:
		background, color = theme.Canvas, theme.Text
	case state.Hovered:
		background, color = theme.Hover, theme.Text
	}
	// The close button is always there, so a click on it can't miss
	// while the pointer moves from the tab onto it; it's faint until the
	// tab is active or hovered.
	c.closer.SetQuiet(!active && !state.Hovered)
	return giode.Row(giode.Styles{
		Align: giode.AlignCenter, Gap: 7, Wrap: giode.WrapNoWrap,
		Padding: giode.Inset{Left: 12, Right: 3}, Height: 30, MaxWidth: 220,
		Background: background, BorderRadii: giode.TopCorners(9),
	}, c.icon(tab.Busy(), tab.Icon(), color),
		// The box caps the title's width, so a long title ends in "…"
		// instead of pushing the close button out of the tab.
		giode.Box(giode.Styles{MaxWidth: 150}, giode.Text(tab.Title(), giode.Styles{
			Color: color, FontSize: theme.Sizes.Small, FontWeight: giode.FontWeightMedium,
			TextOverflow: giode.TextOverflowEllipsis,
		})),
		c.closer,
	)
}

func (c *tabChip) icon(busy bool, name, color string) giode.Element {
	if busy {
		return giode.Spinner(giode.Styles{Width: 13, Height: 13, Color: theme.Primary})
	}
	return theme.Icon(name, 14, color)
}
