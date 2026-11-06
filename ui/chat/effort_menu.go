package chat

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/menu"

	"gocode/ui/theme"
)

// effortOption is one choice in the thinking effort menu.
type effortOption struct {
	value string
	label string
}

// effortMenu picks how hard reasoning models think in this chat.
type effortMenu struct {
	tab     *ChatTab
	menu    *menu.Menu
	trigger giode.Element
}

// effortLabel draws the menu's trigger: the current effort.
type effortLabel struct {
	menu *effortMenu
}

// effortOptions are the levels every provider understands. "" leaves the
// choice to the model.
var effortOptions = []effortOption{
	{"", "Default"},
	{"low", "Low"},
	{"medium", "Medium"},
	{"high", "High"},
}

func newEffortMenu(tab *ChatTab) *effortMenu {
	m := &effortMenu{tab: tab, menu: giode.Menu(theme.MenuStyle())}
	for _, option := range effortOptions {
		m.menu.RadioItem(option.label, func() bool { return tab.effort == option.value }, func() {
			tab.effort = option.value
			tab.composer.editor.Focus()
		})
	}
	button := giode.Box(giode.Styles{BorderRadius: 8, Cursor: giode.CursorPointer}, effortLabel{menu: m}).
		OnHover(giode.Styles{Background: theme.Hover})
	m.trigger = m.menu.Trigger(button)
	return m
}

func (m *effortMenu) Layout(gtx giode.Context) giode.Dimensions {
	return giode.Stack(giode.Styles{}, m.trigger, m.menu).Layout(gtx)
}

func (l effortLabel) Layout(gtx giode.Context) giode.Dimensions {
	label := "Thinking: default"
	for _, option := range effortOptions {
		if option.value != "" && option.value == l.menu.tab.effort {
			label = "Thinking: " + option.value
		}
	}
	return giode.Row(giode.Styles{Align: giode.AlignCenter, Gap: 6, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(9, 6)},
		theme.Icon("brain", 14, theme.Muted),
		giode.Text(label, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small, FontWeight: giode.FontWeightMedium}),
		theme.Icon("chevron-down", 12, theme.Muted),
	).Layout(gtx)
}
