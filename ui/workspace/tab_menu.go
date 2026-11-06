package workspace

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/menu"

	"gocode/ui/theme"
)

// tabMenu is the right-click menu of a tab.
type tabMenu struct {
	menu *menu.Menu
	// target is the tab the menu was opened on.
	target int
}

func newTabMenu(w *Workspace) *tabMenu {
	m := &tabMenu{}
	m.menu = giode.Menu(theme.MenuStyle()).
		Item("Close", func() { w.Close(m.target) }).
		Item("Close others", func() { w.CloseOthers(m.target) }).
		Separator().
		Item("Move left", func() { w.Move(m.target, -1) }).
		Item("Move right", func() { w.Move(m.target, 1) })
	return m
}

// openFor opens the menu for tab id where the click happened.
func (m *tabMenu) openFor(id int, e giode.ClickEvent) {
	m.target = id
	m.menu.OpenAtClick(e)
}
