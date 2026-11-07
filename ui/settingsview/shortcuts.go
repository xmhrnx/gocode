package settingsview

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

var shortcutList = []struct{ keys, action string }{
	{"Mod+T", "New chat"},
	{"Mod+W", "Close tab"},
	{"Ctrl+Tab", "Next tab"},
	{"Ctrl+Shift+Tab", "Previous tab"},
	{"Mod+Shift+PageUp", "Move tab left"},
	{"Mod+Shift+PageDown", "Move tab right"},
	{"Mod+1", "Go to tab 1 (through 9)"},
	{"Mod+K", "Choose a model"},
	{"Mod+H", "Chat history"},
	{"Mod+,", "Settings"},
	{"Esc", "Stop the agent"},
}

func shortcutEntries() []entry {
	entries := make([]entry, 0, len(shortcutList))
	for _, sc := range shortcutList {
		entries = append(entries, entry{title: sc.action, keywords: "keyboard shortcut hotkey " + sc.keys,
			control: keyCap(giode.FormatShortcut(sc.keys))})
	}
	return entries
}

func keyCap(label string) giode.Element {
	return giode.Box(giode.Styles{Padding: giode.SymmetricInset(8, 3), BorderRadius: 6, Background: theme.Hover,
		BorderWidth: 1, BorderColor: theme.SoftBorder},
		giode.Text(label, giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Small, FontFamily: "JetBrains Mono"}))
}
