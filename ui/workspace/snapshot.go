package workspace

import (
	"gocode/internal/tabstate"
	"gocode/ui/core"
)

// Snapshot describes the open tabs for reopening them on the next start.
func (w *Workspace) Snapshot() tabstate.State {
	var state tabstate.State
	for _, e := range w.entries {
		restorable, ok := e.tab.(core.Restorable)
		if !ok {
			continue
		}
		if e.id == w.activeID {
			state.Active = len(state.Tabs)
		}
		state.Tabs = append(state.Tabs, restorable.Snapshot())
	}
	return state
}

// Restore opens the tabs in state, using open to make each, and shows the
// one that was active. It reports whether any tab opened.
func (w *Workspace) Restore(state tabstate.State, open func(tabstate.Tab) core.Tab) bool {
	activeID := 0
	for i, saved := range state.Tabs {
		tab := open(saved)
		if tab == nil {
			continue // its chat may have been deleted
		}
		w.Add(tab)
		if saved.Kind == tabstate.KindSettings {
			w.settingsID = w.nextID
		}
		if i == state.Active || activeID == 0 {
			activeID = w.nextID
		}
	}
	if activeID == 0 {
		return false
	}
	w.activate(activeID)
	return true
}
