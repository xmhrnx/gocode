package chat

import (
	"github.com/xmhrnx/giode"

	"gocode/internal/attach"
	"gocode/ui/components"
	"gocode/ui/theme"
)

// attachmentChip shows an attachment waiting in the composer, with a
// button that removes it.
type attachmentChip struct {
	attachment pendingAttachment
	thumbnail  giode.Element
	remove     *components.IconButton
}

func newAttachmentChip(c *composer, att pendingAttachment) *attachmentChip {
	chip := &attachmentChip{attachment: att}
	chip.remove = components.NewIconButton("x", "Remove", func() { c.remove(att.id) }).SetSize(12)
	if att.Thumbnail != nil {
		chip.thumbnail = giode.Image(att.Thumbnail, giode.Styles{Width: 28, Height: 28, Fit: giode.FitCover, BorderRadius: 5})
	}
	return chip
}

func (a *attachmentChip) Layout(gtx giode.Context) giode.Dimensions {
	preview := a.thumbnail
	if preview == nil {
		preview = giode.Box(giode.Styles{Width: 28, Height: 28, BorderRadius: 5, Background: theme.Hover,
			Align: giode.AlignCenter, Justify: giode.JustifyCenter}, theme.Icon(fileIcon(a.attachment.Kind), 14, theme.Muted))
	}
	return giode.Row(giode.Styles{
		Align: giode.AlignCenter, Gap: 8, Wrap: giode.WrapNoWrap,
		Padding:    giode.Inset{Left: 4, Top: 4, Bottom: 4, Right: 2},
		Background: theme.Raised, BorderRadius: 8, BorderWidth: 1, BorderColor: theme.SoftBorder,
	},
		preview,
		giode.Text(a.attachment.Name, giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Small,
			TextOverflow: giode.TextOverflowEllipsis, MaxWidth: 180}),
		a.remove,
	).Layout(gtx)
}

func fileIcon(kind attach.Kind) string {
	if kind == attach.KindImage {
		return "image"
	}
	return "file-text"
}
