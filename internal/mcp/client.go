package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"gocode/internal/settings"
)

// client is a connection to one MCP server that has completed the
// initialize handshake. It is safe for concurrent use.
type client struct {
	transport transport
	nextID    atomic.Int64
}

// cancelledParams tells the server to stop work on an abandoned request.
type cancelledParams struct {
	RequestID json.RawMessage `json:"requestId"`
	Reason    string          `json:"reason"`
}

// cancelNoticeTimeout bounds sending a cancellation notice, which is a
// courtesy to the server and must not hold up the caller.
const cancelNoticeTimeout = 2 * time.Second

// connect opens a transport to server and performs the MCP handshake.
// events receives what the server does on its own initiative.
func connect(ctx context.Context, server settings.MCPServer, events serverEvents) (*client, error) {
	conn, err := openTransport(server, events)
	if err != nil {
		return nil, err
	}
	c := &client{transport: conn}
	if err := c.initialize(ctx); err != nil {
		return nil, joinClose(err, conn)
	}
	return c, nil
}

// request sends method with params and decodes the result into result
// (skipped when nil). A JSON-RPC error comes back as *RPCError.
func (c *client) request(ctx context.Context, method string, params, result any) error {
	id := json.RawMessage(strconv.FormatInt(c.nextID.Add(1), 10))
	req, err := newMessage(id, method, params)
	if err != nil {
		return err
	}
	resp, err := c.transport.roundTrip(ctx, req)
	if err != nil {
		c.cancelIfAbandoned(ctx, id, method)
		return fmt.Errorf("%s: %w", method, err)
	}
	raw, err := resultOf(resp)
	if err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	if result == nil {
		return nil
	}
	if err := json.Unmarshal(raw, result); err != nil {
		return fmt.Errorf("decode %s result: %w", method, err)
	}
	return nil
}

// notify sends a notification, which has no response.
func (c *client) notify(ctx context.Context, method string, params any) error {
	msg, err := newMessage(nil, method, params)
	if err != nil {
		return err
	}
	if err := c.transport.send(ctx, msg); err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	return nil
}

// cancelIfAbandoned tells the server to stop a request the caller gave up
// on, so a slow tool does not keep running. The protocol forbids
// cancelling initialize.
func (c *client) cancelIfAbandoned(ctx context.Context, id json.RawMessage, method string) {
	if ctx.Err() == nil || method == methodInitialize {
		return
	}
	notice, cancel := context.WithTimeout(context.Background(), cancelNoticeTimeout)
	defer cancel()
	// Best effort: if the notice fails the server merely finishes the work.
	_ = c.notify(notice, "notifications/cancelled", cancelledParams{RequestID: id, Reason: ctx.Err().Error()})
}

// close ends the connection.
func (c *client) close() error {
	return c.transport.close()
}

// joinClose closes conn after err and reports both failures.
func joinClose(err error, conn transport) error {
	if closeErr := conn.close(); closeErr != nil {
		return fmt.Errorf("%w (closing: %v)", err, closeErr)
	}
	return err
}
