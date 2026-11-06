package chat

import (
	"gocode/internal/agent"
	"gocode/internal/llm"
)

// subagentProgress is what a running task's sub-agent has done so far,
// shown inside the task's card.
type subagentProgress struct {
	steps    int
	activity []string // the latest tool calls, newest last
	status   string
	usage    llm.Usage
	revision int
}

// maxShownActivity is how many of a sub-agent's latest steps its card lists.
const maxShownActivity = 4

// handleSubagentEvent records an event from a task's sub-agent. Its
// messages stay out of the conversation: only the task's report reaches
// the main agent.
func (c *ChatTab) handleSubagentEvent(ev agent.Event) {
	progress := c.run.subagent(ev.Parent)
	switch ev.Kind {
	case agent.EventToolStarted:
		progress.steps++
		progress.activity = append(progress.activity, ev.Summary)
		if len(progress.activity) > maxShownActivity {
			progress.activity = progress.activity[1:]
		}
		progress.status = ""
	case agent.EventReasoning:
		progress.status = "Thinking…"
	case agent.EventText:
		progress.status = "Writing its report…"
	case agent.EventUsage:
		progress.usage = *ev.Usage
	case agent.EventMessage:
		if ev.Message.Role == llm.RoleAssistant {
			c.session.Usage = c.session.Usage.Add(progress.usage) // its tokens count toward the chat
			progress.usage = llm.Usage{}
		}
	}
	progress.revision++
	c.changed()
}

// subagent returns the progress of the sub-agent of task call callID.
func (r *activeRun) subagent(callID string) *subagentProgress {
	if r.subagents == nil {
		r.subagents = map[string]*subagentProgress{}
	}
	progress, ok := r.subagents[callID]
	if !ok {
		progress = &subagentProgress{}
		r.subagents[callID] = progress
	}
	return progress
}
