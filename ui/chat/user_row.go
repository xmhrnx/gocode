package chat

import (
	"github.com/xmhrnx/giode"

	"gocode/internal/attach"
	"gocode/internal/llm"
	"gocode/ui/components"
	"gocode/ui/theme"
)

// userRow shows a message the user sent, with copy and edit actions.
type userRow struct {
	rowBase
	copy   giode.Element
	edit   *components.IconButton
	images []giode.Element
}

func newUserRow(tab *ChatTab, it item) *userRow {
	r := &userRow{rowBase: rowBase{tab: tab, it: it}}
	r.copy = giode.CopyButton(func() string { return tab.message(r.it).Text() }, giode.Styles{Color: theme.Muted})
	r.edit = components.NewIconButton("pencil", "Edit and resend", func() { tab.editMessage(r.it.message) }).SetSize(14)
	for _, part := range tab.message(it).Parts {
		if part.Kind == llm.PartImage {
			r.images = append(r.images, imagePreview(part.Image))
		}
	}
	return r
}

func (r *userRow) version() int { return 0 }

func (r *userRow) Layout(gtx giode.Context) giode.Dimensions {
	msg := r.tab.message(r.it)
	var content []giode.Element
	if len(r.images) > 0 {
		content = append(content, giode.Row(giode.Styles{Gap: 6}, r.images...))
	}
	for _, part := range msg.Parts {
		if part.Kind != llm.PartText {
			continue
		}
		if name, isFile := attach.FileBlockName(part.Text); isFile {
			content = append(content, fileChip(name))
			continue
		}
		content = append(content, giode.Text(part.Text, theme.Body()))
	}
	bubble := giode.Stack(giode.Styles{
		Gap: 8, Padding: giode.SymmetricInset(14, 10), BorderRadius: 14,
		Background: theme.Raised, BorderWidth: 1, BorderColor: theme.SoftBorder, MaxWidth: 640,
	}, content...)
	return giode.Stack(giode.Styles{Width: -1, Align: giode.AlignEnd, Gap: 2, Padding: giode.Inset{Top: 18, Bottom: 6}},
		bubble,
		giode.Row(giode.Styles{Gap: 2, Align: giode.AlignCenter}, r.edit, r.copy),
	).Layout(gtx)
}

func fileChip(name string) giode.Element {
	return giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter, Padding: giode.SymmetricInset(8, 4),
		BorderRadius: 6, Background: theme.Hover},
		theme.Icon("file-text", 13, theme.Muted),
		giode.Text(name, giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Small}))
}

func imagePreview(img *llm.Image) giode.Element {
	preview, err := attach.Preview(img.Data, 320)
	if err != nil {
		return fileChip("image (" + img.MediaType + ")")
	}
	return giode.Image(preview, giode.Styles{MaxWidth: 220, MaxHeight: 160, BorderRadius: 8, Fit: giode.FitContain})
}
