package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"gocode/internal/settings"
	"gocode/ui/theme"
)

// NewMessageField returns a borderless, growing text field for composing
// messages and commands; monospace shows its text in the code font.
func NewMessageField(placeholder string, monospace bool) *input.Input {
	return giode.Input.Text(placeholder, MessageFieldStyle(monospace)).Lines(1, 14)
}

// MessageFieldStyle returns the field's style for the current text sizes.
func MessageFieldStyle(monospace bool) giode.Styles {
	style := giode.Styles{
		Color: theme.Text, Background: "transparent", BorderColor: "transparent",
		FontSize: theme.Sizes.Base, Padding: giode.NoInset,
	}
	if monospace {
		style.FontSize, style.FontFamily = theme.Sizes.Code, "JetBrains Mono"
	}
	return style
}

// SubmitMode returns the field behavior for a send key. Ctrl+Enter is not
// a field mode: the field leaves it alone, so a window shortcut sends.
func SubmitMode(key settings.SendKey) giode.SubmitMode {
	if key == settings.SendCtrlEnter {
		return giode.SubmitNever
	}
	return giode.SubmitEnterUnlessShift
}
