package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"gocode/internal/settings"
)

// transport carries JSON-RPC messages to one server and back.
type transport interface {
	// roundTrip sends a request and waits for the response with its id.
	roundTrip(ctx context.Context, req message) (message, error)
	// send delivers a notification or a reply; nothing comes back.
	send(ctx context.Context, msg message) error
	// setProtocolVersion records the version agreed in initialize, which
	// HTTP repeats on every later request.
	setProtocolVersion(version string)
	// close ends the connection and frees its resources.
	close() error
}

// serverEvents receives what a server does on its own initiative. The
// callbacks run on the transport's reading goroutine, so they must not block.
type serverEvents struct {
	// notify receives the server's notifications.
	notify func(method string, params json.RawMessage)
	// lost is called once when a connection that can tell (stdio) ends,
	// with the reason.
	lost func(err error)
}

// Transport names accepted in settings.MCPServer.Transport.
const (
	transportStdio = "stdio"
	transportHTTP  = "http"
)

// withDefaults fills unset callbacks with no-ops, so transports can call
// them unconditionally.
func (e serverEvents) withDefaults() serverEvents {
	if e.notify == nil {
		e.notify = func(string, json.RawMessage) {}
	}
	if e.lost == nil {
		e.lost = func(error) {}
	}
	return e
}

// openTransport connects to server over the transport its config names.
// An empty transport means stdio when a command is set and HTTP otherwise.
func openTransport(server settings.MCPServer, events serverEvents) (transport, error) {
	events = events.withDefaults()
	switch transportKind(server) {
	case transportStdio:
		return startStdio(server, events)
	case transportHTTP:
		return newHTTPTransport(server, events)
	default:
		return nil, fmt.Errorf("unknown MCP transport %q", server.Transport)
	}
}

// transportKind resolves the transport of server, defaulting by its fields.
func transportKind(server settings.MCPServer) string {
	if server.Transport != "" {
		return server.Transport
	}
	if server.Command != "" {
		return transportStdio
	}
	return transportHTTP
}
