package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/tooltip"

	"gocode/ui/theme"
)

// PillButton is a compact, borderless button with an icon, a label and a
// chevron, for choosers such as the model and mode pickers.
type PillButton struct {
	icon, label string
	press       *giode.PressableElement
	tip         *tooltip.Tooltip
}

// NewPillButton returns a pill showing icon, described by hint.
func NewPillButton(icon, hint string, onClick func()) *PillButton {
	b := &PillButton{icon: icon}
	b.press = giode.Pressable(b.content).Label(hint).OnClick(func(giode.ClickEvent) { onClick() })
	b.tip = giode.Tooltip(b.press, hint, giode.Styles{FontSize: 12})
	return b
}

// SetLabel replaces the label.
func (b *PillButton) SetLabel(label string) *PillButton { b.label = label; return b }

// SetIcon replaces the icon.
func (b *PillButton) SetIcon(icon string) *PillButton { b.icon = icon; return b }

// Layout lays the pill out.
func (b *PillButton) Layout(gtx giode.Context) giode.Dimensions { return b.tip.Layout(gtx) }

func (b *PillButton) content(state giode.PressState) giode.Element {
	color, background := theme.Muted, ""
	if state.Hovered || state.Pressed {
		color, background = theme.Text, theme.Hover
	}
	return giode.Row(giode.Styles{
		Align: giode.AlignCenter, Gap: 6, Wrap: giode.WrapNoWrap, Color: color,
		Padding: giode.SymmetricInset(9, 6), BorderRadius: 8, Background: background,
	},
		theme.Icon(b.icon, 14, color),
		giode.Text(b.label, giode.Styles{Color: color, FontSize: theme.Sizes.Small, FontWeight: giode.FontWeightMedium,
			TextOverflow: giode.TextOverflowEllipsis, MaxWidth: 260}),
		theme.Icon("chevron-down", 12, color),
	)
}
