package chat

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

// sendButton sends the message, or stops the run while one is going.
type sendButton struct {
	tab   *ChatTab
	press *giode.PressableElement
}

func newSendButton(tab *ChatTab) *sendButton {
	b := &sendButton{tab: tab}
	b.press = giode.Pressable(b.content).Label("Send").OnClick(func(giode.ClickEvent) {
		if tab.run != nil && tab.composer.isEmpty() {
			tab.stop()
		} else {
			tab.send()
		}
	})
	return b
}

func (b *sendButton) Layout(gtx giode.Context) giode.Dimensions {
	label := "Send"
	if b.tab.run != nil && b.tab.composer.isEmpty() {
		label = "Stop"
	}
	b.press.Label(label)
	return b.press.Layout(gtx)
}

func (b *sendButton) content(state giode.PressState) giode.Element {
	icon, background, color := "arrow-up", theme.Primary, theme.OnPrimary
	switch {
	case b.tab.run != nil && b.tab.composer.isEmpty():
		icon, background, color = "square", theme.Text, theme.Panel
	case b.tab.composer != nil && b.tab.composer.isEmpty():
		background, color = theme.Hover, theme.Faint
	}
	if state.Hovered && b.tab.run == nil {
		background = giode.Mix(background, "theme.text", 0.88)
	}
	return giode.Box(giode.Styles{
		Width: 32, Height: 32, BorderRadius: 16, Background: background,
		Align: giode.AlignCenter, Justify: giode.JustifyCenter, Margin: giode.Inset{Left: 4},
	}, theme.Icon(icon, 16, color))
}
