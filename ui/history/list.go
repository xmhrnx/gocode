package history

import (
	"strings"

	"github.com/xmhrnx/giode"

	"gocode/internal/session"
	"gocode/ui/theme"
)

func (d *Dialog) listView() giode.Element {
	if d.loading {
		return message("Loading…")
	}
	query := strings.ToLower(strings.TrimSpace(d.search.Text()))
	var rows []giode.Element
	for _, s := range d.summaries {
		if query != "" && !strings.Contains(strings.ToLower(s.Title+" "+s.WorkDir), query) {
			continue
		}
		summary := s
		rows = append(rows, d.rows.Get(s.ID, func() *row { return newRow(d, summary) }))
	}
	if len(rows) == 0 {
		return message(emptyText(query, d.summaries))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 2, Padding: giode.Inset{Right: 8}}, rows...)
}

func emptyText(query string, all []session.Summary) string {
	if len(all) == 0 {
		return "No chats yet. Chats are saved here as you work."
	}
	return "No chats match “" + query + "”."
}

func message(text string) giode.Element {
	return giode.Box(giode.Styles{Width: -1, Padding: giode.UniformInset(24), Justify: giode.JustifyCenter},
		giode.Text(text, giode.Styles{Color: theme.Muted}))
}
