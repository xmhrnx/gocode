package chat

import (
	"sync"

	"gocode/internal/llm"
)

// messageQueue holds messages the user sent while the agent was working,
// until the run hands them to the model. The run drains it from its own
// goroutine, so it is locked.
type messageQueue struct {
	mu       sync.Mutex
	messages []llm.Message
}

// push adds a message.
func (q *messageQueue) push(msg llm.Message) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.messages = append(q.messages, msg)
}

// drain returns the waiting messages and empties the queue.
func (q *messageQueue) drain() []llm.Message {
	q.mu.Lock()
	defer q.mu.Unlock()
	messages := q.messages
	q.messages = nil
	return messages
}

// texts returns the waiting messages' text, for display.
func (q *messageQueue) texts() []string {
	q.mu.Lock()
	defer q.mu.Unlock()
	texts := make([]string, len(q.messages))
	for i, msg := range q.messages {
		texts[i] = msg.Text()
	}
	return texts
}
