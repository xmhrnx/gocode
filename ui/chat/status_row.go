package chat

import (
	"fmt"
	"time"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/kit"

	"gocode/ui/theme"
)

// statusRow shows that the agent is working, for how long, and any retry
// in progress.
type statusRow struct {
	rowBase
}

func newStatusRow(tab *ChatTab) *statusRow { return &statusRow{rowBase{tab: tab}} }

func (r *statusRow) version() int { return int(r.elapsed() / time.Second) }

func (r *statusRow) elapsed() time.Duration {
	if r.tab.run == nil {
		return 0
	}
	return time.Since(r.tab.run.started)
}

func (r *statusRow) Layout(gtx layout.Context) layout.Dimensions {
	run := r.tab.run
	if run == nil || run.ask != nil {
		return layout.Dimensions{}
	}
	label := "Working"
	if run.status != "" {
		label = run.status
	}
	kit.RequestFrameAt(gtx, time.Now().Add(time.Second-r.elapsed()%time.Second))
	return giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Padding: giode.SymmetricInset(0, 10)},
		giode.Spinner(giode.Styles{Width: 14, Height: 14, Color: theme.Primary}),
		giode.Text(label, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small}),
		giode.Text(fmt.Sprintf("%ds", int(r.elapsed().Seconds())), giode.Styles{Color: theme.Faint, FontSize: theme.Sizes.Small}),
		giode.Text("Esc to stop", giode.Styles{Color: theme.Faint, FontSize: theme.Sizes.Tiny}),
	).Layout(gtx)
}
