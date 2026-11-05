package workspace

import "gocode/ui/core"

// Select shows the tab at position index (0-based), if there is one.
func (w *Workspace) Select(index int) {
	if index >= 0 && index < len(w.entries) {
		w.activate(w.entries[index].id)
	}
}

// Next shows the tab after the current one, wrapping around.
func (w *Workspace) Next() { w.step(1) }

// Previous shows the tab before the current one, wrapping around.
func (w *Workspace) Previous() { w.step(-1) }

func (w *Workspace) step(delta int) {
	if len(w.entries) == 0 {
		return
	}
	index := (w.indexOf(w.activeID) + delta + len(w.entries)) % len(w.entries)
	w.activate(w.entries[index].id)
}

// AnyBusy reports whether some tab has work in progress.
func (w *Workspace) AnyBusy() bool {
	for _, e := range w.entries {
		if e.tab.Busy() {
			return true
		}
	}
	return false
}

// DropFiles hands dropped files to the shown tab, if it takes files.
func (w *Workspace) DropFiles(paths []string) bool {
	target, ok := w.Active().(core.FileDropTarget)
	if ok {
		target.DropFiles(paths)
	}
	return ok
}

func (w *Workspace) indexOf(id int) int {
	for i, e := range w.entries {
		if e.id == id {
			return i
		}
	}
	return -1
}
