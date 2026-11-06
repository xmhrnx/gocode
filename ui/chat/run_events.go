package chat

import (
	"gocode/internal/agent"
	"gocode/internal/llm"
)

func (c *ChatTab) handleEvent(ev agent.Event) {
	run := c.run
	if run == nil {
		return
	}
	if ev.Parent != "" {
		c.handleSubagentEvent(ev)
		return
	}
	switch ev.Kind {
	case agent.EventText:
		run.draft.AppendText(llm.PartText, ev.Text)
	case agent.EventReasoning:
		run.draft.AppendText(llm.PartReasoning, ev.Text)
	case agent.EventToolStarted:
		run.running[ev.Call.ID] = true
	case agent.EventToolFinished:
		delete(run.running, ev.Call.ID)
		run.finished[ev.Call.ID] = llm.ToolResult{CallID: ev.Call.ID, Name: ev.Call.Name,
			Content: ev.Result.Output, IsError: ev.Result.IsError}
		if ev.Result.Display != "" {
			c.session.ToolDisplays[ev.Call.ID] = ev.Result.Display
		}
	case agent.EventMessage:
		c.commitMessage(*ev.Message)
	case agent.EventUsage:
		run.callUsage = *ev.Usage
		c.contextTokens = ev.Usage.InputTokens + ev.Usage.CacheReadTokens + ev.Usage.CacheWriteTokens
	case agent.EventRetry:
		run.status = ev.Text
	case agent.EventNotice:
		c.addNotice(noticeInfo, ev.Text)
	case agent.EventCompacted:
		c.session.Compacted(ev.Text)
		c.contextTokens = 0
		c.addNotice(noticeInfo, "Context compacted: earlier messages were summarized for the model. They stay visible here.")
	}
	if ev.Kind != agent.EventRetry {
		run.status = ""
	}
	c.changed()
}

func (c *ChatTab) commitMessage(msg llm.Message) {
	c.session.Messages = append(c.session.Messages, msg)
	if msg.Role == llm.RoleAssistant {
		c.session.Usage = c.session.Usage.Add(c.run.callUsage)
		c.run.callUsage = llm.Usage{}
		c.run.draft = llm.PartBuilder{}
	}
	c.save()
}
