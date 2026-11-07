package chat

import "gocode/internal/tabstate"

// Title implements core.Tab.
func (c *ChatTab) Title() string {
	if c.session.Title != "" {
		return c.session.Title
	}
	return "New chat"
}

// Empty reports whether nothing was sent yet, so the tab can be reused.
func (c *ChatTab) Empty() bool {
	return len(c.session.Messages) == 0 && c.run == nil && c.composer.isEmpty()
}

// Snapshot implements core.Restorable.
func (c *ChatTab) Snapshot() tabstate.Tab {
	tab := tabstate.Tab{Kind: tabstate.KindChat}
	switch {
	case len(c.session.Messages) > 0:
		tab.SessionID, tab.WorkDir = c.session.ID, c.session.WorkDir
	case c.openedFolder:
		tab.WorkDir, tab.Opened = c.session.WorkDir, true
	}
	return tab
}

// Icon implements core.Tab.
func (c *ChatTab) Icon() string { return "message-square" }

// Busy implements core.Tab.
func (c *ChatTab) Busy() bool { return c.run != nil }

// Close implements core.Tab.
func (c *ChatTab) Close() {
	c.stop()
	c.unsubscribe()
}

// Focus implements core.Focusable.
func (c *ChatTab) Focus() { c.composer.editor.Focus() }

// DropFiles implements core.FileDropTarget.
func (c *ChatTab) DropFiles(paths []string) { c.composer.attachPaths(paths) }
