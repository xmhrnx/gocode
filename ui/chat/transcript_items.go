package chat

import (
	"fmt"

	"gocode/internal/llm"
)

// itemKind is what a transcript row shows.
type itemKind int

// item is one row of the transcript.
type item struct {
	kind    itemKind
	key     string
	message int // index into the session's messages, or the draft's index
	part    int
	draft   bool
	notice  notice
	// turnEnd marks the assistant's last text before the next user
	// message, where the copy and retry actions go.
	turnEnd bool
	version int
	// text is a segment of a streaming reply or of reasoning.
	text string
	// group is the key of the part a reasoning segment belongs to, and
	// segment its place in it.
	group   string
	segment int
}

const (
	itemUser itemKind = iota
	itemText
	itemReasoning
	itemTool
	itemNotice
	itemPermission
	itemStatus
	itemSearch
)

// buildItems lays the conversation out as rows: user messages, assistant
// parts (tool results fold into their calls), notices, and the live run.
func (c *ChatTab) buildItems() []item {
	var items []item
	c.results = indexToolResults(c.session.Messages)
	notices := c.notices
	for i, msg := range c.session.Messages {
		items, notices = appendNotices(items, notices, i)
		items = c.appendMessageItems(items, msg, i, false)
	}
	items, _ = appendNotices(items, notices, len(c.session.Messages))
	if run := c.run; run != nil {
		items = c.appendMessageItems(items, run.draftMessage(), len(c.session.Messages), true)
		if run.ask != nil {
			items = append(items, item{kind: itemPermission, key: "permission", version: c.revision})
		}
		items = append(items, item{kind: itemStatus, key: "status", version: c.revision})
	}
	markTurnEnds(items)
	return items
}

func (c *ChatTab) appendMessageItems(items []item, msg llm.Message, index int, draft bool) []item {
	if msg.Role == llm.RoleUser {
		if hasUserContent(msg) {
			items = append(items, item{kind: itemUser, key: fmt.Sprint("m", index), message: index})
		}
		return items
	}
	for p, part := range msg.Parts {
		it := item{key: fmt.Sprintf("m%dp%d", index, p), message: index, part: p, draft: draft}
		switch part.Kind {
		case llm.PartText:
			if draft {
				items = appendStreamSegments(items, it, part.Text)
				continue
			}
			it.kind = itemText
		case llm.PartReasoning:
			if part.Text == "" {
				continue // encrypted or omitted reasoning has nothing to show
			}
			items = appendReasoningSegments(items, it, part.Text)
			continue
		case llm.PartProvider:
			if part.Text == "" {
				continue // search results and the like are for the model
			}
			it.kind = itemSearch
		case llm.PartToolCall:
			it.kind = itemTool
			it.version = c.toolVersion(part.ToolCall.ID)
		default:
			continue
		}
		if draft {
			it.version = c.revision
		}
		items = append(items, it)
	}
	return items
}

func appendNotices(items []item, notices []notice, before int) ([]item, []notice) {
	for len(notices) > 0 && notices[0].after <= before {
		n := notices[0]
		items = append(items, item{kind: itemNotice, key: fmt.Sprint("n", n.id), notice: n})
		notices = notices[1:]
	}
	return items, notices
}

func markTurnEnds(items []item) {
	lastText := -1
	for i, it := range items {
		switch it.kind {
		case itemText:
			lastText = i
		case itemUser:
			if lastText >= 0 {
				items[lastText].turnEnd = true
			}
			lastText = -1
		}
	}
	if lastText >= 0 && !items[lastText].draft {
		items[lastText].turnEnd = true
	}
}

// hasUserContent reports whether a user message holds more than tool
// results, which are shown with their calls instead.
func hasUserContent(msg llm.Message) bool {
	for _, part := range msg.Parts {
		if part.Kind == llm.PartText || part.Kind == llm.PartImage || part.Kind == llm.PartDocument {
			return true
		}
	}
	return false
}

// appendReasoningSegments adds reasoning as one row per segment, while
// it streams and after: the text stays where it was drawn, and only the
// growing segment is laid out again.
func appendReasoningSegments(items []item, base item, text string) []item {
	for n, segment := range reasoningSegments(text, streamSegmentSize) {
		it := base
		it.kind, it.text, it.group, it.segment = itemReasoning, segment, base.key, n
		it.key = fmt.Sprintf("%sr%d", base.key, n)
		it.version = len(segment) * 2
		if n == 0 && base.draft {
			it.version++ // the header says "Thinking…" until the reply is done
		}
		items = append(items, it)
	}
	return items
}

// appendStreamSegments adds a streaming reply as one row per segment. A
// segment's version is its length, so only the growing one is redrawn.
func appendStreamSegments(items []item, base item, text string) []item {
	for n, segment := range streamSegments(text, streamSegmentSize) {
		it := base
		it.kind, it.text, it.version = itemText, segment, len(segment)
		it.key = fmt.Sprintf("%ss%d", base.key, n)
		items = append(items, it)
	}
	return items
}
