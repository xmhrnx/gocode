package chat

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/menu"

	"gocode/internal/agent"
	"gocode/ui/theme"
)

// modeOption is one choice in the permission mode menu.
type modeOption struct {
	mode  agent.Mode
	label string
	icon  string
	tools bool
}

// modeMenu picks how much the agent may do without asking.
type modeMenu struct {
	tab     *ChatTab
	menu    *menu.Menu
	trigger giode.Element
}

// modeLabel draws the menu's trigger: the current mode's icon and name.
type modeLabel struct {
	menu *modeMenu
}

var modeOptions = []modeOption{
	{agent.ModeAsk, "Ask before changes", "shield", true},
	{agent.ModeAutoEdit, "Auto-accept edits", "file-pen-line", true},
	{agent.ModeFullAccess, "Full access", "zap", true},
	{agent.ModeReadOnly, "Read-only (plan)", "eye", true},
	{agent.ModeAsk, "Chat only, no tools", "message-square", false},
}

func newModeMenu(tab *ChatTab) *modeMenu {
	m := &modeMenu{tab: tab, menu: giode.Menu(theme.MenuStyle())}
	for _, option := range modeOptions {
		m.menu.RadioItem(option.label, func() bool { return m.current() == option }, func() { m.choose(option) })
	}
	m.trigger = m.menu.Trigger(modeLabel{menu: m})
	return m
}

func (m *modeMenu) current() modeOption {
	for _, option := range modeOptions {
		if option.tools == m.tab.useTools && (!option.tools || option.mode == m.tab.perms.Mode()) {
			return option
		}
	}
	return modeOptions[0]
}

func (m *modeMenu) choose(option modeOption) {
	m.tab.useTools = option.tools
	m.tab.perms.SetMode(option.mode)
	m.tab.composer.editor.Focus()
}

func (m *modeMenu) Layout(gtx layout.Context) layout.Dimensions {
	return giode.Stack(giode.Styles{}, m.trigger, m.menu).Layout(gtx)
}

func (l modeLabel) Layout(gtx layout.Context) layout.Dimensions {
	option := l.menu.current()
	return giode.Row(giode.Styles{Align: giode.AlignCenter, Gap: 6, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(9, 6), Cursor: giode.CursorPointer},
		theme.Icon(option.icon, 14, theme.Muted),
		giode.Text(option.label, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small, FontWeight: giode.FontWeightMedium}),
		theme.Icon("chevron-down", 12, theme.Muted),
	).Layout(gtx)
}
