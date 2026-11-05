package chat

import (
	"context"

	"gocode/internal/agent"
)

// pendingPermission is a tool call waiting for the user's approval.
type pendingPermission struct {
	request agent.PermissionRequest
	answer  chan agent.PermissionAnswer
}

// ask is the agent's Asker: it shows the request in the transcript and
// waits for the user. It runs on the run's goroutine.
func (c *ChatTab) ask(ctx context.Context, req agent.PermissionRequest) agent.PermissionAnswer {
	answer := make(chan agent.PermissionAnswer, 1)
	c.services.App.Post(func() {
		if c.run != nil {
			c.run.ask = &pendingPermission{request: req, answer: answer}
			c.changed()
		}
	})
	select {
	case a := <-answer:
		return a
	case <-ctx.Done():
		return agent.PermissionAnswer{Decision: agent.DecisionDeny}
	}
}

// answerPermission resolves the pending request.
func (c *ChatTab) answerPermission(decision agent.Decision, feedback string) {
	if c.run == nil || c.run.ask == nil {
		return
	}
	c.run.ask.answer <- agent.PermissionAnswer{Decision: decision, Feedback: feedback}
	c.run.ask = nil
	c.changed()
}

// pendingRequest returns the request awaiting an answer, or nil.
func (c *ChatTab) pendingRequest() *pendingPermission {
	if c.run == nil {
		return nil
	}
	return c.run.ask
}
