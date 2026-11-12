package chat

import (
	"fmt"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"gocode/internal/attach"
	"gocode/internal/llm"
	"gocode/ui/components"
	"gocode/ui/theme"
)

// mediaRow shows an image, audio clip or video the model made: images
// in the chat, the others as a file, each with a button to save it.
type mediaRow struct {
	rowBase
	media   *llm.Media
	preview giode.Element
	save    *components.IconButton
}

// mediaPreviewSide is the longest side, in pixels, of an image's preview,
// and mediaShownSide the longest side it is shown at, in dp.
const (
	mediaPreviewSide = 960
	mediaShownSide   = 480
)

func newMediaRow(tab *ChatTab, it item) *mediaRow {
	r := &mediaRow{rowBase: rowBase{tab: tab, it: it}}
	part, _ := tab.part(it)
	r.media = part.Media
	if r.media == nil {
		r.media = &llm.Media{Name: "file"}
	}
	r.save = components.NewIconButton("download", "Save "+r.media.Name, func() { tab.saveMedia(*r.media) }).SetSize(14)
	if r.media.IsImage() {
		if img, err := attach.Preview(r.media.Data, mediaPreviewSide); err == nil {
			w, h := fitInside(img.Bounds().Dx(), img.Bounds().Dy(), mediaShownSide)
			r.preview = giode.Image(img, giode.Styles{Width: properties.Width(w), Height: properties.Height(h), BorderRadius: 10, Fit: giode.FitContain})
		}
	}
	return r
}

func (r *mediaRow) version() int { return 0 }

func (r *mediaRow) Layout(gtx giode.Context) giode.Dimensions {
	content := r.preview
	if content == nil {
		content = r.fileCard()
	}
	return giode.Stack(giode.Styles{Gap: 2, Padding: giode.SymmetricInset(0, 6), Align: giode.AlignStart},
		content,
		giode.Row(giode.Styles{Gap: 2, Align: giode.AlignCenter, Margin: giode.Inset{Left: -6}}, r.save),
	).Layout(gtx)
}

// fileCard shows media the chat can't draw: its kind, name and size.
func (r *mediaRow) fileCard() giode.Element {
	return giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Padding: giode.SymmetricInset(12, 10),
		BorderRadius: 10, Background: theme.Raised, BorderWidth: 1, BorderColor: theme.SoftBorder},
		theme.Icon(fileIcon(attach.KindDocument, r.media.MediaType), 18, theme.Muted),
		giode.Stack(giode.Styles{Gap: 2},
			giode.Text(r.media.Name, giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Small}),
			giode.Text(byteSize(len(r.media.Data)), giode.Styles{Color: theme.Faint, FontSize: theme.Sizes.Tiny}),
		),
	)
}

// fitInside scales a w by h picture down, keeping its shape, until
// neither side is longer than side; smaller pictures keep their size.
func fitInside(w, h, side int) (int, int) {
	if w <= 0 || h <= 0 {
		return side, side
	}
	if longest := max(w, h); longest > side {
		return w * side / longest, h * side / longest
	}
	return w, h
}

// byteSize shows a file size in KB or MB.
func byteSize(n int) string {
	if n < 1<<20 {
		return fmt.Sprintf("%d KB", max(1, (n+512)/1024))
	}
	return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
}
