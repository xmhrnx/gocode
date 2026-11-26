package chat

import (
	"context"
	"time"

	"gocode/internal/llm"
)

// nameThenRun handles the first message of a chat without a folder: the
// model names the chat, the folder is made with that name, and then the
// agent starts. Stopping meanwhile gives the message back.
func (c *ChatTab) nameThenRun(msg llm.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	c.run = newActiveRun(cancel)
	c.run.status = "Naming the chat"
	c.changed()
	ask, app := c.titleRequest(msg.Text()), c.services.App
	go func() {
		title := ""
		if ask != nil {
			askCtx, done := context.WithTimeout(ctx, titleTimeout)
			title = ask(askCtx)
			done()
		}
		app.Post(func() { c.afterNaming(ctx, msg, title) })
	}()
}

// afterNaming makes the folder and starts the agent, or gives the
// message back when the user stopped or the folder can't be made.
func (c *ChatTab) afterNaming(ctx context.Context, msg llm.Message, title string) {
	started := c.run.started
	c.run = nil
	if title == "" {
		title = fallbackTitle(msg.Text())
	}
	if ctx.Err() != nil || !c.ensureFolder(title) {
		c.giveBack(msg)
		return
	}
	c.session.Title = title
	c.startRun()
	if c.run != nil {
		c.run.started = started // the time includes naming
	}
}

// giveBack takes back the chat's last message, which was never answered,
// and puts it and anything queued since into the composer.
func (c *ChatTab) giveBack(msg llm.Message) {
	if n := len(c.session.Messages); n > 0 {
		c.session.Messages = c.session.Messages[:n-1]
	}
	c.composer.restoreQueued(append([]llm.Message{msg}, c.queue.drain()...))
	c.changed()
}

// newActiveRun returns the state of a run starting now.
func newActiveRun(cancel context.CancelFunc) *activeRun {
	return &activeRun{cancel: cancel, started: time.Now(), running: map[string]bool{}, finished: map[string]llm.ToolResult{}}
}
