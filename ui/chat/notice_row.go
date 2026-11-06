package chat

import (
	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

// noticeRow shows an error or status line that isn't sent to the model.
type noticeRow struct {
	rowBase
	notice notice
}

func newNoticeRow(n notice) *noticeRow { return &noticeRow{notice: n} }

func (r *noticeRow) version() int { return 0 }

func (r *noticeRow) Layout(gtx giode.Context) giode.Dimensions {
	icon, color, background := "info", theme.Muted, ""
	if r.notice.kind == noticeError {
		icon, color, background = "triangle-alert", theme.Danger, theme.DangerSoft
	}
	return giode.Box(giode.Styles{Width: -1, Padding: giode.SymmetricInset(0, 6)},
		giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignStart, Wrap: giode.WrapNoWrap,
			Padding: giode.SymmetricInset(12, 8), BorderRadius: 9, Background: background},
			theme.Icon(icon, 15, color),
			giode.Text(r.notice.text, giode.Styles{Color: color, FontSize: theme.Sizes.Small, FlexGrow: 1}),
		),
	).Layout(gtx)
}
