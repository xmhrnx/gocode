package chat

import (
	"errors"
	"strings"

	"gocode/internal/agent"
	"gocode/internal/llm"
	"gocode/internal/remote/wire"
)

// RemoteSend sends a prompt from a phone, as if typed here.
func (c *ChatTab) RemoteSend(text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return errors.New("the message is empty")
	}
	return c.submit(llm.Message{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart(text)}})
}

// RemoteStop stops the chat's run from a phone.
func (c *ChatTab) RemoteStop() { c.stop() }

// RemoteDecide answers the chat's permission request from a phone.
func (c *ChatTab) RemoteDecide(d wire.Decide) error {
	if c.pendingRequest() == nil {
		return errors.New("nothing is waiting for an answer")
	}
	switch d.Decision {
	case wire.DecisionAllow:
		c.answerPermission(agent.DecisionAllow, "")
	case wire.DecisionAlways:
		c.answerPermission(agent.DecisionAllowAlways, "")
	case wire.DecisionDeny:
		c.answerPermission(agent.DecisionDeny, d.Feedback)
	default:
		return errors.New("unknown decision " + d.Decision)
	}
	return nil
}
