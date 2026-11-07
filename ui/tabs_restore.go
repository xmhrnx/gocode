package ui

import (
	"gocode/internal/tabstate"
	"gocode/ui/chat"
	"gocode/ui/core"
	"gocode/ui/settingsview"
)

// restoreTabs reopens the tabs open when GoCode last closed, when the
// setting is on. It reports whether any tab opened.
func (r *Root) restoreTabs() bool {
	if !r.services.Settings.General.RestoreTabs {
		return false
	}
	state, err := tabstate.Load(r.services.TabsPath)
	if err != nil {
		r.services.Error(err.Error())
		return false
	}
	return r.workspace.Restore(state, r.openSaved)
}

// openSaved makes the tab a saved entry describes, or nil when it can't.
func (r *Root) openSaved(saved tabstate.Tab) core.Tab {
	s := r.services
	switch saved.Kind {
	case tabstate.KindChat:
		if saved.SessionID == "" {
			return chat.New(s, saved.WorkDir, "")
		}
		sess, err := s.Sessions.Load(saved.SessionID)
		if err != nil {
			return nil
		}
		return chat.Open(s, sess)
	case tabstate.KindSettings:
		return settingsview.New(s, "")
	}
	return nil
}

// SaveTabs remembers the open tabs for the next start. The snapshot is
// taken here, on the UI goroutine; the file is written in the background.
func (r *Root) SaveTabs() {
	state, path, services := r.workspace.Snapshot(), r.services.TabsPath, r.services
	go func() {
		if err := tabstate.Save(path, state); err != nil {
			services.Error("Couldn't remember the open tabs: " + err.Error())
		}
	}()
}

// SaveTabsBeforeExit writes the open tabs and waits for the write, since
// the process ends right after.
func (r *Root) SaveTabsBeforeExit() error {
	return tabstate.Save(r.services.TabsPath, r.workspace.Snapshot())
}
