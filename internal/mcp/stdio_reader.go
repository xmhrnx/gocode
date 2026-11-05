package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// readLoop reads the server's stdout until it closes, routing each message.
// When the stream ends it reaps the process and fails the pending calls.
func (t *stdioTransport) readLoop(stdout io.Reader) {
	defer close(t.done)
	reader := bufio.NewReader(stdout)
	for {
		line, readErr := reader.ReadBytes('\n')
		t.handleLine(line)
		if readErr != nil {
			err := t.exitError(readErr)
			t.fail(err)
			t.events.lost(err)
			return
		}
	}
}

// handleLine decodes and routes one line of server output.
func (t *stdioTransport) handleLine(line []byte) {
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return
	}
	var msg message
	if err := json.Unmarshal(line, &msg); err != nil {
		// Some servers log to stdout by mistake; a stray line must not end
		// the session.
		return
	}
	switch {
	case msg.isResponse():
		t.deliver(msg)
	case msg.isRequest():
		go t.answer(msg)
	case msg.Method != "":
		t.events.notify(msg.Method, msg.Params)
	}
}

// answer replies to a server-initiated request. It runs on its own
// goroutine so a full stdin pipe cannot stall the reader.
func (t *stdioTransport) answer(req message) {
	// A failed write means the server is gone; the reader notices that and
	// fails the pending calls, so there is nothing more to do here.
	_ = t.send(context.Background(), replyToServer(req))
}

// exitError reaps the process and explains its exit, with the end of its
// stderr, which usually says what went wrong.
func (t *stdioTransport) exitError(readErr error) error {
	waitErr := t.cmd.Wait()
	reason := "MCP server closed its output"
	if readErr != io.EOF {
		reason = fmt.Sprintf("read from MCP server: %v", readErr)
	} else if waitErr != nil {
		reason = fmt.Sprintf("MCP server exited: %v", waitErr)
	}
	if tail := t.stderr.String(); tail != "" {
		return fmt.Errorf("%s; stderr: %s", reason, tail)
	}
	return errors.New(reason)
}
