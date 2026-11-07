package ui

import (
	"fmt"

	"github.com/xmhrnx/giode"
)

// registerShortcuts registers the window-wide keyboard shortcuts.
func (r *Root) registerShortcuts() {
	w := r.workspace
	list := []giode.KeyShortcut{
		giode.Shortcut("Mod+T", func() { w.OpenChat("", "") }),
		giode.Shortcut("Mod+W", w.CloseActive),
		giode.Shortcut("Mod+,", func() { w.OpenSettings("") }),
		giode.Shortcut("Mod+H", r.history.Open),
		giode.Shortcut("Ctrl+Tab", w.Next),
		giode.Shortcut("Ctrl+Shift+Tab", w.Previous),
		giode.Shortcut("Mod+Shift+PageUp", func() { w.MoveActive(-1) }),
		giode.Shortcut("Mod+Shift+PageDown", func() { w.MoveActive(1) }),
	}
	for i := 1; i <= 9; i++ {
		index := i - 1
		list = append(list, giode.Shortcut(fmt.Sprintf("Mod+%d", i), func() { w.Select(index) }))
	}
	giode.Shortcuts(list...)
}
