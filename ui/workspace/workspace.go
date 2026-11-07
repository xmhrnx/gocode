// Package workspace manages the window's tabs.
package workspace

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/core"
)

// entry is an open tab with an ID that stays stable as tabs move.
type entry struct {
	id  int
	tab core.Tab
}

// Factory creates the kinds of tabs the workspace can open.
type Factory struct {
	Chat     func(workDir, prompt string) core.Tab
	Settings func(category string) core.Tab
}

// Workspace is the ordered set of open tabs and which one is shown.
type Workspace struct {
	factory  Factory
	entries  []entry
	activeID int
	nextID   int
	// settingsID is the settings tab's ID, or 0 when it is closed.
	settingsID int
	// onChange runs after tabs open, close or switch.
	onChange func()
}

// New returns an empty workspace.
func New(factory Factory) *Workspace { return &Workspace{factory: factory} }

// OpenChat opens a new chat tab in workDir (empty: the default folder),
// starting with prompt in its composer.
func (w *Workspace) OpenChat(workDir, prompt string) { w.Add(w.factory.Chat(workDir, prompt)) }

// OpenSettings shows the settings tab, opening it on first use.
func (w *Workspace) OpenSettings(category string) {
	if index := w.indexOf(w.settingsID); index >= 0 {
		w.activate(w.settingsID)
		if nav, ok := w.entries[index].tab.(interface{ ShowCategory(string) }); ok && category != "" {
			nav.ShowCategory(category)
		}
		return
	}
	w.Add(w.factory.Settings(category))
	w.settingsID = w.nextID
}

// Add opens tab and shows it.
func (w *Workspace) Add(tab core.Tab) {
	w.nextID++
	w.entries = append(w.entries, entry{id: w.nextID, tab: tab})
	w.activate(w.nextID)
}

// Open shows tab in place of the active tab when that one is an unused
// chat, and in a new tab otherwise.
func (w *Workspace) Open(tab core.Tab) {
	index := w.indexOf(w.activeID)
	if unused, ok := w.Active().(interface{ Empty() bool }); ok && unused.Empty() {
		w.entries[index].tab.Close()
		w.entries[index].tab = tab
		w.activate(w.activeID)
		return
	}
	w.Add(tab)
}

// Close closes the tab; the last tab closed opens a fresh chat, so the
// window is never empty.
func (w *Workspace) Close(id int) {
	giode.AfterLayout(func() {
		index := w.indexOf(id)
		if index < 0 {
			return
		}
		w.entries[index].tab.Close()
		w.entries = append(w.entries[:index], w.entries[index+1:]...)
		if len(w.entries) == 0 {
			w.OpenChat("", "")
			return
		}
		if id == w.activeID {
			w.activate(w.entries[min(index, len(w.entries)-1)].id)
		} else if w.onChange != nil {
			w.onChange()
		}
	})
}

// CloseActive closes the shown tab.
func (w *Workspace) CloseActive() { w.Close(w.activeID) }

// Active returns the shown tab, or nil.
func (w *Workspace) Active() core.Tab {
	if index := w.indexOf(w.activeID); index >= 0 {
		return w.entries[index].tab
	}
	return nil
}

// OnChange sets a function to run after tabs open, close or switch.
func (w *Workspace) OnChange(fn func()) { w.onChange = fn }

func (w *Workspace) activate(id int) {
	w.activeID = id
	if w.onChange != nil {
		w.onChange()
	}
	if focusable, ok := w.Active().(core.Focusable); ok {
		focusable.Focus()
	}
}
