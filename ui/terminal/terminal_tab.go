// Package terminal is the command tab: Warp-style blocks, one per
// command, with the output kept together and actions on each block.
package terminal

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/lazylist"

	"gocode/internal/shell"
	"gocode/internal/tabstate"
	"gocode/ui/components"
	"gocode/ui/core"
	"gocode/ui/theme"
)

// Tab is one terminal.
type Tab struct {
	services *core.Services
	session  *shell.Session
	blocks   []*block
	nextID   int
	input    *input.Input
	history  commandHistory
	list     *lazylist.List
	views    *giode.KeyedCache[int, *blockView]
	// matches are the completions offered by the last Tab, shown until
	// the command changes.
	matches []string
}

// New returns a terminal starting in workDir, or in the default folder
// when it is empty.
func New(services *core.Services, workDir string) *Tab {
	t := &Tab{services: services, views: giode.NewKeyedCache[int, *blockView]()}
	cfg := services.Settings
	if workDir == "" {
		workDir = startDir(cfg.General.DefaultWorkDir)
	}
	t.session = shell.NewSession(cfg.Terminal.Shell, workDir)
	t.input = components.NewMessageField("Type a command…", true).
		SubmitOn(giode.SubmitEnterUnlessShift).Submit(func(string) { t.submit() }).
		OnChange(func(string) { t.matches = nil })
	t.list = giode.LazyList(
		func() int { return len(t.blocks) },
		func(i int) any { return t.blocks[i].id },
		t.row,
		giode.Styles{ScrollBar: giode.ScrollBar{Width: 7, AutoHide: true, TrackColor: "transparent", Inset: 3}},
	).Version(func(i int) any {
		return [2]int{t.blocks[i].revision, services.SettingsRevision()}
	}).Anchor(giode.LazyAnchorBottom)
	return t
}

// Title implements core.Tab.
func (t *Tab) Title() string { return "Terminal · " + baseName(t.session.WorkDir()) }

// Snapshot implements core.Restorable.
func (t *Tab) Snapshot() tabstate.Tab {
	return tabstate.Tab{Kind: tabstate.KindTerminal, WorkDir: t.session.WorkDir()}
}

// Icon implements core.Tab.
func (t *Tab) Icon() string { return "terminal" }

// Busy implements core.Tab.
func (t *Tab) Busy() bool { return t.running() != nil }

// Close implements core.Tab.
func (t *Tab) Close() { t.stop() }

// Focus implements core.Focusable.
func (t *Tab) Focus() { t.input.Focus() }

// View implements core.Tab.
func (t *Tab) View() giode.Element {
	giode.Shortcuts(giode.Shortcut("Esc", t.stop), giode.Shortcut("Mod+L", t.clear),
		giode.Shortcut("Up", func() { t.recall(true) }), giode.Shortcut("Down", func() { t.recall(false) }),
		giode.Shortcut("Tab", t.complete))
	content := giode.Element(welcome())
	if len(t.blocks) > 0 {
		content = giode.SelectionArea(t, t.list) // select across blocks
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Background: theme.Canvas},
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, content),
		t.promptBar(),
	)
}

func (t *Tab) row(i int) giode.Element {
	b := t.blocks[i]
	return t.views.Get(b.id, func() *blockView { return newBlockView(t, b) })
}

func (t *Tab) running() *block {
	if n := len(t.blocks); n > 0 && t.blocks[n-1].running {
		return t.blocks[n-1]
	}
	return nil
}
