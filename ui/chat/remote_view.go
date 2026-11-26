package chat

import (
	"strings"
	"time"

	"gocode/internal/llm"
	"gocode/internal/remote/wire"
)

const (
	// remoteRows is how many of the latest rows a phone gets.
	remoteRows = 80
	// remoteText bounds one row's text for a phone.
	remoteText = 20000
)

// RemoteID names the chat for phones: its session's ID.
func (c *ChatTab) RemoteID() string { return c.session.ID }

// RemoteSummary is the chat as the phone's list shows it.
func (c *ChatTab) RemoteSummary() wire.ChatSummary {
	return wire.ChatSummary{ID: c.session.ID, Title: c.Title(), Folder: c.folderLabel(),
		Busy: c.run != nil, Waiting: c.pendingRequest() != nil}
}

// RemoteView is the chat as a phone shows it: the latest rows and the
// run's progress.
func (c *ChatTab) RemoteView() wire.ChatView {
	view := wire.ChatView{ChatSummary: c.RemoteSummary(), Queued: len(c.queue.texts())}
	rows := c.remoteRows()
	if len(rows) > remoteRows {
		view.Earlier = len(rows) - remoteRows
		rows = rows[len(rows)-remoteRows:]
	}
	view.Rows = rows
	if run := c.run; run != nil {
		view.Status = run.status
		if view.Status == "" {
			view.Status = "Working"
		}
		view.Seconds = int(time.Since(run.started) / time.Second)
		view.TokensPerSecond = run.tokensPerSecond()
		if ask := run.ask; ask != nil {
			req := ask.request
			detail := req.Preview
			if detail == "" {
				detail = string(req.Call.Arguments)
			}
			view.Permission = &wire.Permission{Tool: req.Call.Name, Summary: req.Summary, Detail: clipText(detail)}
		}
	}
	return view
}

// remoteRows lays the conversation and notices out as rows.
func (c *ChatTab) remoteRows() []wire.Row {
	var rows []wire.Row
	notices := c.notices
	addNotices := func(before int) {
		for len(notices) > 0 && notices[0].after <= before {
			rows = append(rows, wire.Row{Kind: wire.RowNotice, Text: notices[0].text})
			notices = notices[1:]
		}
	}
	for i, msg := range c.session.Messages {
		addNotices(i)
		rows = c.appendRemoteRows(rows, msg)
	}
	addNotices(len(c.session.Messages))
	if c.run != nil {
		rows = c.appendRemoteRows(rows, c.run.draftMessage())
	}
	return rows
}

func (c *ChatTab) appendRemoteRows(rows []wire.Row, msg llm.Message) []wire.Row {
	if msg.Role == llm.RoleUser {
		if hasUserContent(msg) {
			rows = append(rows, wire.Row{Kind: wire.RowUser, Text: clipText(userText(msg))})
		}
		return rows
	}
	for _, part := range msg.Parts {
		switch {
		case part.Kind == llm.PartText && strings.TrimSpace(part.Text) != "":
			rows = append(rows, wire.Row{Kind: wire.RowText, Text: clipText(part.Text)})
		case part.Kind == llm.PartReasoning && part.Text != "":
			rows = append(rows, wire.Row{Kind: wire.RowReasoning, Text: clipText(part.Text)})
		case part.Kind == llm.PartToolCall && part.ToolCall != nil:
			rows = append(rows, c.remoteToolRow(*part.ToolCall))
		case part.Kind == llm.PartMedia && part.Media != nil:
			rows = append(rows, wire.Row{Kind: wire.RowMedia, Text: part.Media.Name})
		case part.Kind == llm.PartProvider && part.Text != "":
			rows = append(rows, wire.Row{Kind: wire.RowNotice, Text: part.Text})
		}
	}
	return rows
}

func (c *ChatTab) remoteToolRow(call llm.ToolCall) wire.Row {
	row := wire.Row{Kind: wire.RowTool, Tool: call.Name, Text: c.toolSummary(call.Name, call.Arguments), State: wire.StateRunning}
	if result := c.toolResult(call.ID); result != nil {
		row.State = wire.StateDone
		if result.IsError {
			row.State = wire.StateFailed
		}
	} else if !c.toolRunning(call.ID) && c.run == nil {
		row.State = wire.StateFailed // never finished: the run stopped
	}
	return row
}

// userText is a user message's text, with attachments by name.
func userText(msg llm.Message) string {
	var parts []string
	for _, part := range msg.Parts {
		switch {
		case part.Kind == llm.PartText:
			parts = append(parts, part.Text)
		case part.Kind == llm.PartDocument && part.Document != nil:
			parts = append(parts, "📎 "+part.Document.Name)
		case part.Kind == llm.PartImage:
			parts = append(parts, "📎 image")
		}
	}
	return strings.Join(parts, "\n")
}

// clipText keeps a phone's copy of a long text to its start.
func clipText(text string) string {
	if len(text) <= remoteText {
		return text
	}
	cut := remoteText
	for cut > 0 && !utf8Start(text[cut]) {
		cut--
	}
	return text[:cut] + "\n…"
}

func utf8Start(b byte) bool { return b&0xC0 != 0x80 }
