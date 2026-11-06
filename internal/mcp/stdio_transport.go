package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gocode/internal/proc"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// stdioTransport runs a server as a child process and exchanges
// newline-delimited JSON-RPC over its stdin and stdout. A reader goroutine
// routes responses to the waiting callers by id, so calls can overlap.
type stdioTransport struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stderr *tailBuffer
	events serverEvents

	writeMu sync.Mutex

	mu      sync.Mutex
	pending map[string]chan message
	// err is set once the server is gone; it fails every later call.
	err error

	// done is closed when the reader has exited and the process is reaped.
	done      chan struct{}
	closeOnce sync.Once
	closeErr  error
}

const (
	// closeGrace is how long a server may take to exit after its stdin
	// closes before it is killed.
	closeGrace = 500 * time.Millisecond
	// pipeDrainDelay bounds the wait for stderr after the server exits, in
	// case a grandchild still holds the pipe open.
	pipeDrainDelay = time.Second
)

// roundTrip sends req and waits for the response carrying its id.
func (t *stdioTransport) roundTrip(ctx context.Context, req message) (message, error) {
	replies, err := t.expect(req.ID)
	if err != nil {
		return message{}, err
	}
	defer t.forget(req.ID)
	if err := t.send(ctx, req); err != nil {
		return message{}, err
	}
	select {
	case resp, ok := <-replies:
		if !ok {
			return message{}, t.failure()
		}
		return resp, nil
	case <-ctx.Done():
		return message{}, ctx.Err()
	}
}

// send writes msg as one line. Writes are serialised so lines never mix.
func (t *stdioTransport) send(ctx context.Context, msg message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	line, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("encode message: %w", err)
	}
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	if _, err := t.stdin.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("write to MCP server: %w", err)
	}
	return nil
}

// setProtocolVersion does nothing: stdio carries no per-message headers.
func (t *stdioTransport) setProtocolVersion(string) {}

// close asks the server to exit by closing its stdin, then kills its
// process group, which also stops anything it spawned.
func (t *stdioTransport) close() error {
	t.closeOnce.Do(func() {
		// Closing fails only if the pipe is already closed because the
		// server exited, which is the state we want anyway.
		_ = t.stdin.Close()
		select {
		case <-t.done:
		case <-time.After(closeGrace):
		}
		if err := proc.KillGroup(t.cmd); !errors.Is(err, os.ErrProcessDone) {
			t.closeErr = err
		}
		<-t.done
	})
	return t.closeErr
}
