package core

import (
	"github.com/xmhrnx/giode"

	"gocode/internal/tabstate"
)

// Tab is one page in the window's tab strip.
type Tab interface {
	// Title is the label in the tab strip.
	Title() string
	// Icon is the name of the tab's icon.
	Icon() string
	// Busy reports work in progress, shown as a spinner on the tab.
	Busy() bool
	// View builds the tab's content.
	View() giode.Element
	// Close stops the tab's work; the tab is not shown again.
	Close()
}

// Restorable is implemented by tabs that can be reopened on the next start.
type Restorable interface {
	Snapshot() tabstate.Tab
}

// FileDropTarget is implemented by tabs that accept dropped files.
type FileDropTarget interface {
	DropFiles(paths []string)
}

// Focusable is implemented by tabs with a primary input to focus when
// they become active.
type Focusable interface {
	Focus()
}
