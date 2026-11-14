package chat

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"gocode/ui/theme"
)

// scrollDownButton floats above the bottom of the transcript while the
// user has scrolled up, and takes them back to the latest message.
type scrollDownButton struct {
	tab   *ChatTab
	press *giode.PressableElement
	none  giode.Element
}

func newScrollDownButton(tab *ChatTab) *scrollDownButton {
	b := &scrollDownButton{tab: tab, none: giode.Box(giode.Styles{})}
	b.press = giode.Pressable(b.content).Label("Scroll to the latest message").OnClick(func(giode.ClickEvent) {
		tab.transcript.list.ScrollToBottom()
	})
	return b
}

// element returns the button, or nothing while the transcript is at the
// end or empty.
func (b *scrollDownButton) element() giode.Element {
	t := b.tab.transcript
	if len(t.items) == 0 || t.list.Following() {
		return b.none
	}
	return giode.Row(giode.Styles{Position: giode.PositionAbsolute, Left: giode.Flush, Right: giode.Flush,
		Bottom: properties.Bottom(12), Justify: giode.JustifyCenter}, b.press)
}

func (b *scrollDownButton) content(state giode.PressState) giode.Element {
	background := theme.Panel
	if state.Hovered || state.Pressed {
		background = theme.Raised
	}
	return giode.Box(giode.Styles{
		Padding: giode.UniformInset(8), BorderRadius: 999, Background: background,
		BorderWidth: 1, BorderColor: theme.Border, BoxShadow: giode.BoxShadow{Y: 4, Blur: 14, Color: "#00000040"},
	}, theme.Icon("arrow-down", 16, theme.Text))
}
