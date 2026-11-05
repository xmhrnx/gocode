package chat

import (
	"slices"
	"strings"

	"gocode/internal/llm"
)

// retry runs the last request again, dropping the replies to it.
func (c *ChatTab) retry() {
	if c.run != nil {
		return
	}
	last := c.lastUserMessage(len(c.session.Messages))
	if last < 0 {
		return
	}
	c.truncate(last + 1)
	c.startRun()
}

// editMessage puts a sent message back into the composer and removes it
// and everything after it, so the user can change it and send again.
func (c *ChatTab) editMessage(index int) {
	if c.run != nil || index >= len(c.session.Messages) {
		return
	}
	c.composer.restore(c.session.Messages[index])
	c.truncate(index)
	c.composer.editor.Focus()
}

// truncate keeps the first n messages and the notices among them.
func (c *ChatTab) truncate(n int) {
	c.session.Messages = slices.Clone(c.session.Messages[:n])
	c.notices = slices.DeleteFunc(c.notices, func(x notice) bool { return x.after >= n })
	c.save()
	c.changed()
}

// lastUserMessage returns the index of the last message the user typed
// before index, or -1.
func (c *ChatTab) lastUserMessage(before int) int {
	for i := before - 1; i >= 0; i-- {
		if msg := c.session.Messages[i]; msg.Role == llm.RoleUser && hasUserContent(msg) {
			return i
		}
	}
	return -1
}

// turnText joins the assistant's text in the turn containing message
// index, for the copy button.
func (c *ChatTab) turnText(index int) string {
	start := c.lastUserMessage(index + 1)
	var parts []string
	for i := start + 1; i < len(c.session.Messages); i++ {
		msg := c.session.Messages[i]
		if msg.Role == llm.RoleUser && hasUserContent(msg) {
			break
		}
		if text := strings.TrimSpace(msg.Text()); msg.Role == llm.RoleAssistant && text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n\n")
}
