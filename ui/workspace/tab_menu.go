package workspace

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/menu"

	"gocode/ui/theme"
)

// newTabMenu returns the right-click menu of tab id.
func newTabMenu(w *Workspace, id int) *menu.Menu {
	return giode.Menu(theme.MenuStyle()).
		Item("Close", func() { w.Close(id) }).
		Item("Close others", func() { w.CloseOthers(id) }).
		Separator().
		Item("Move left", func() { w.Move(id, -1) }).
		Item("Move right", func() { w.Move(id, 1) })
}
