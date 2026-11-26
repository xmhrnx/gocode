package chat

import (
	"fmt"
	"time"

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

// progress is how long the run has worked and, once the model has
// written something, how many tokens: "1m 05s · 2.4k tokens".
func (r *statusRow) progress(run *activeRun) string {
	text := workingTime(r.elapsed())
	if tokens := run.tokens(); tokens > 0 {
		text += " · " + tokenCount(tokens) + " tokens"
	}
	return text
}

// workingTime shows seconds, then minutes and seconds from a minute on,
// then hours and minutes from an hour on.
func workingTime(d time.Duration) string {
	s := int(d / time.Second)
	switch {
	case s < 60:
		return fmt.Sprintf("%ds", s)
	case s < 3600:
		return fmt.Sprintf("%dm %02ds", s/60, s%60)
	}
	return fmt.Sprintf("%dh %02dm", s/3600, s%3600/60)
}

// tokenCount shortens a token count: 950, 2.4k, 12k, 1.3M.
func tokenCount(n int) string {
	switch {
	case n < 1000:
		return fmt.Sprint(n)
	case n < 10_000:
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	case n < 1_000_000:
		return fmt.Sprintf("%dk", n/1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
}

func (r *statusRow) Layout(gtx giode.Context) giode.Dimensions {
	run := r.tab.run
	if run == nil || run.ask != nil {
		return giode.Dimensions{}
	}
	label := "Working"
	if run.status != "" {
		label = run.status
	}
	kit.RequestFrameAt(gtx, time.Now().Add(time.Second-r.elapsed()%time.Second))
	return giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Padding: giode.SymmetricInset(0, 10)},
		giode.Spinner(giode.Styles{Width: 14, Height: 14, Color: theme.Primary}),
		giode.Text(label, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small}),
		giode.Text(r.progress(run), giode.Styles{Color: theme.Faint, FontSize: theme.Sizes.Small}),
		giode.Text("Esc to stop", giode.Styles{Color: theme.Faint, FontSize: theme.Sizes.Tiny}),
	).Layout(gtx)
}
