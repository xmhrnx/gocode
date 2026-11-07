package terminal

import (
	"fmt"

	"github.com/xmhrnx/giode"

	"gocode/ui/theme"
)

// shownMatches is how many completions are listed before "and N more".
const shownMatches = 60

// completionList shows the names a Tab could complete to, or nothing
// when there is no choice to make.
func completionList(matches []string) giode.Element {
	if len(matches) == 0 {
		return giode.Box(giode.Styles{})
	}
	style := giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small, FontFamily: "JetBrains Mono"}
	items := make([]giode.Element, 0, min(len(matches), shownMatches)+1)
	for _, name := range matches[:min(len(matches), shownMatches)] {
		items = append(items, giode.Text(name, style))
	}
	if extra := len(matches) - shownMatches; extra > 0 {
		items = append(items, giode.Text(fmt.Sprintf("and %d more", extra),
			giode.Styles{Color: theme.Faint, FontSize: theme.Sizes.Small}))
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 16, Wrap: giode.WrapWrap}, items...)
}
