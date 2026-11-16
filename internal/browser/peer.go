package browser

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"golang.org/x/net/websocket"
)

// message is anything the extension sends.
type message struct {
	Type    string          `json:"type,omitempty"`
	Browser string          `json:"browser,omitempty"`
	ID      int64           `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// request is a command for the extension.
type request struct {
	ID     int64  `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
}

type reply struct {
	result json.RawMessage
	err    error
}

var errDisconnected = errors.New("the browser disconnected before answering")

// peer is one connected extension.
type peer struct {
	ws      *websocket.Conn
	browser string
	writeMu sync.Mutex

	mu      sync.Mutex
	pending map[int64]chan reply
	nextID  int64
	closed  bool
}

func (p *peer) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	done := make(chan reply, 1)
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, errDisconnected
	}
	p.nextID++
	id := p.nextID
	p.pending[id] = done
	p.mu.Unlock()
	defer p.forget(id)
	if err := p.send(request{ID: id, Method: method, Params: params}); err != nil {
		return nil, err
	}
	select {
	case r := <-done:
		return r.result, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *peer) send(v any) error {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	return websocket.JSON.Send(p.ws, v)
}

func (p *peer) forget(id int64) {
	p.mu.Lock()
	delete(p.pending, id)
	p.mu.Unlock()
}

// readLoop delivers replies until the connection closes.
func (p *peer) readLoop() {
	for {
		var msg message
		if err := websocket.JSON.Receive(p.ws, &msg); err != nil {
			return
		}
		if msg.Type == "ping" {
			_ = p.send(message{Type: "pong"})
			continue
		}
		p.mu.Lock()
		done := p.pending[msg.ID]
		p.mu.Unlock()
		if done == nil {
			continue
		}
		if msg.Error != "" {
			done <- reply{err: errors.New(msg.Error)}
		} else {
			done <- reply{result: msg.Result}
		}
	}
}

// close ends the connection and fails the calls waiting on it.
func (p *peer) close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	pending := p.pending
	p.pending = map[int64]chan reply{}
	p.mu.Unlock()
	_ = p.ws.Close()
	for _, done := range pending {
		done <- reply{err: errDisconnected}
	}
}
