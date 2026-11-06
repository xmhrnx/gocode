// Package components holds GoCode's reusable UI pieces.
package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/tooltip"

	"gocode/ui/theme"
)

// IconButton is a square, borderless button showing an icon, with a
// tooltip naming what it does.
type IconButton struct {
	icon     string
	size     int
	active   bool
	disabled bool
	press    *giode.PressableElement
	tip      *tooltip.Tooltip
}

// NewIconButton returns a button showing icon, described by hint.
func NewIconButton(icon, hint string, onClick func()) *IconButton {
	b := &IconButton{icon: icon, size: 16}
	b.press = giode.Pressable(b.content).Label(hint).OnClick(func(giode.ClickEvent) {
		if !b.disabled {
			onClick()
		}
	})
	b.tip = giode.Tooltip(b.press, hint, giode.Styles{FontSize: 12})
	return b
}

// SetIcon replaces the icon.
func (b *IconButton) SetIcon(icon string) *IconButton { b.icon = icon; return b }

// SetHint replaces the tooltip.
func (b *IconButton) SetHint(hint string) *IconButton {
	b.tip.SetContent(hint)
	b.press.Label(hint)
	return b
}

// SetSize sets the icon size in dp.
func (b *IconButton) SetSize(size int) *IconButton { b.size = size; return b }

// SetActive highlights the button, for toggles.
func (b *IconButton) SetActive(active bool) *IconButton { b.active = active; return b }

// SetDisabled dims the button and ignores clicks.
func (b *IconButton) SetDisabled(disabled bool) *IconButton {
	b.disabled = disabled
	b.press.Disabled(disabled)
	return b
}

// Hovered reports whether the pointer is over the button.
func (b *IconButton) Hovered() bool { return b.press.Hovered() }

// Layout lays the button out.
func (b *IconButton) Layout(gtx giode.Context) giode.Dimensions { return b.tip.Layout(gtx) }

func (b *IconButton) content(state giode.PressState) giode.Element {
	color, background := theme.Muted, ""
	switch {
	case b.disabled:
		color = theme.Faint
	case state.Pressed:
		color, background = theme.Text, theme.Pressed
	case state.Hovered || b.active:
		color, background = theme.Text, theme.Hover
	}
	return giode.Box(giode.Styles{
		Padding: giode.UniformInset(6), BorderRadius: 7, Background: background,
		Align: giode.AlignCenter, Justify: giode.JustifyCenter,
	}, theme.Icon(b.icon, b.size, color))
}
