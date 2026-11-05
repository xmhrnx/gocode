package mcp

import "encoding/json"

// expect registers a caller waiting for the response with id.
func (t *stdioTransport) expect(id json.RawMessage) (chan message, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.err != nil {
		return nil, t.err
	}
	replies := make(chan message, 1)
	t.pending[string(id)] = replies
	return replies, nil
}

// forget drops the waiter for id once its call has returned.
func (t *stdioTransport) forget(id json.RawMessage) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.pending, string(id))
}

// deliver hands a response to the caller waiting for its id. Responses
// nobody waits for (the caller timed out) are dropped.
func (t *stdioTransport) deliver(resp message) {
	t.mu.Lock()
	replies, ok := t.pending[string(resp.ID)]
	delete(t.pending, string(resp.ID))
	t.mu.Unlock()
	if ok {
		replies <- resp
	}
}

// fail records why the connection ended and wakes every waiting caller.
func (t *stdioTransport) fail(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.err = err
	for id, replies := range t.pending {
		close(replies)
		delete(t.pending, id)
	}
}

// failure returns why the connection ended.
func (t *stdioTransport) failure() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.err
}
