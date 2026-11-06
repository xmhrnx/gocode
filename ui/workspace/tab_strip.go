package workspace

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/components"
)

// TabStrip shows the open tabs and the buttons that open new ones. It
// lives in the window's title bar.
type TabStrip struct {
	workspace   *Workspace
	chips       *giode.KeyedCache[int, *tabChip]
	newChat     *components.IconButton
	newTerminal *components.IconButton
	menu        *tabMenu
}

// NewTabStrip returns the tab strip for w.
func NewTabStrip(w *Workspace) *TabStrip {
	return &TabStrip{
		workspace:   w,
		chips:       giode.NewKeyedCache[int, *tabChip](),
		newChat:     components.NewIconButton("plus", "New chat  ("+giode.FormatShortcut("Mod+T")+")", func() { w.OpenChat("", "") }),
		newTerminal: components.NewIconButton("square-terminal", "New terminal  ("+giode.FormatShortcut("Mod+Shift+T")+")", w.OpenTerminal),
		menu:        newTabMenu(w),
	}
}

// Layout lays the strip out.
func (s *TabStrip) Layout(gtx giode.Context) giode.Dimensions {
	items := make([]giode.Element, 0, len(s.workspace.entries)+2)
	for _, e := range s.workspace.entries {
		id := e.id
		items = append(items, s.chip(id))
	}
	items = append(items, s.newChat, s.newTerminal, s.menu.menu)
	return giode.Row(giode.Styles{
		Align: giode.AlignEnd, Gap: 2, Wrap: giode.WrapNoWrap, Height: -1,
		Padding: giode.Inset{Left: 8},
	}, items...).Layout(gtx)
}

func (s *TabStrip) chip(id int) *tabChip {
	return s.chips.Get(id, func() *tabChip { return newTabChip(s, id) })
}

// dragOver moves the dragged tab past a neighbour once the pointer
// crosses the neighbour's middle, so the tabs rearrange live and a wide
// neighbour doesn't swap back and forth under the pointer.
func (s *TabStrip) dragOver(id, pointerX int) {
	from := s.workspace.indexOf(id)
	target := from
	for index, e := range s.workspace.entries {
		middle := s.chip(e.id).origin.X + s.chip(e.id).width/2
		switch {
		case index < from && pointerX < middle && target == from:
			target = index // the first tab passed on the left
		case index > from && pointerX > middle:
			target = index // the last tab passed on the right
		}
	}
	s.workspace.MoveTo(id, target)
}
