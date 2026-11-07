package chat

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/lazylist"
)

// rowView is a transcript row. Its version changes when its own state
// (an expanded section) changes, so the cached drawing is replaced.
type rowView interface {
	giode.Element
	version() int
	// setItem gives the row its current item: a draft row becomes a
	// committed one, and a text row can become the end of a turn.
	setItem(it item)
}

// rowBase holds what every row needs.
type rowBase struct {
	tab *ChatTab
	it  item
}

// rowVersion identifies what a cached row was drawn from.
type rowVersion struct {
	content, settings, local int
}

// transcript is the scrolling conversation. Only rows near the viewport
// are built, and rows that didn't change replay their last drawing.
type transcript struct {
	tab     *ChatTab
	list    *lazylist.List
	items   []item
	builtAt int
	rows    *giode.KeyedCache[string, rowView]
	empty   giode.Element
}

func (r *rowBase) setItem(it item) { r.it = it }

func newTranscript(tab *ChatTab) *transcript {
	t := &transcript{tab: tab, builtAt: -1, rows: giode.NewKeyedCache[string, rowView]().Retain(600)}
	t.list = giode.LazyList(
		func() int { return len(t.items) },
		func(i int) any { return t.items[i].key },
		t.row,
		giode.Styles{ScrollBar: giode.ScrollBar{Width: 7, AutoHide: true, TrackColor: "transparent", Inset: 3}},
	).Version(func(i int) any {
		it := t.items[i]
		return rowVersion{content: it.version, settings: tab.services.SettingsRevision(), local: t.rowFor(it).version()}
	}).Anchor(giode.LazyAnchorBottom)
	t.empty = newEmptyState(tab)
	return t
}

// element returns the transcript, rebuilding its rows when the chat
// changed.
func (t *transcript) element() giode.Element {
	if t.builtAt != t.tab.revision {
		t.items = t.tab.buildItems()
		t.builtAt = t.tab.revision
	}
	if len(t.items) == 0 {
		return t.empty
	}
	// One selection area over the whole conversation, so a selection can
	// run across messages, as on a web page.
	return giode.SelectionArea("transcript/"+t.tab.session.ID, t.list)
}

func (t *transcript) row(i int) giode.Element {
	it := t.items[i]
	view := t.rowFor(it)
	view.setItem(it)
	return giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyCenter, Padding: giode.Inset{Left: 20, Right: 20}},
		giode.Box(giode.Styles{Width: -1, MaxWidth: chatWidth(t.tab)}, view))
}

// rowFor returns the row component for an item, creating it on first use.
func (t *transcript) rowFor(it item) rowView {
	return t.rows.Get(it.key+kindSuffix(it.kind), func() rowView { return t.newRow(it) })
}

func (t *transcript) newRow(it item) rowView {
	switch it.kind {
	case itemUser:
		return newUserRow(t.tab, it)
	case itemText:
		return newTextRow(t.tab, it)
	case itemReasoning:
		return newReasoningRow(t.tab, it)
	case itemTool:
		return newToolRow(t.tab, it)
	case itemPermission:
		return newPermissionRow(t.tab)
	case itemStatus:
		return newStatusRow(t.tab)
	case itemSearch:
		return newSearchRow(t.tab, it)
	default:
		return newNoticeRow(it.notice)
	}
}

// kindSuffix keeps rows of different kinds apart when a draft part with
// the same key changes kind.
func kindSuffix(kind itemKind) string {
	return string(rune('a' + int(kind)))
}

// scrollToBottom follows the conversation again.
func (t *transcript) scrollToBottom() { t.list.ScrollToBottom() }
