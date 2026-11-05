package mcp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"gocode/internal/settings"
)

// httpTransport speaks MCP's Streamable HTTP transport: every message is a
// POST, and a request's answer comes back as JSON or as an SSE stream.
type httpTransport struct {
	url     string
	headers map[string]string
	client  *http.Client
	events  serverEvents

	mu              sync.Mutex
	sessionID       string
	protocolVersion string
}

const (
	headerSessionID       = "Mcp-Session-Id"
	headerProtocolVersion = "MCP-Protocol-Version"
	// errorBodyLimit bounds how much of a failed response is quoted.
	errorBodyLimit = 512
)

// newHTTPTransport checks the server's URL; no connection is made until the
// first message.
func newHTTPTransport(server settings.MCPServer, events serverEvents) (transport, error) {
	parsed, err := url.Parse(server.URL)
	if err != nil {
		return nil, fmt.Errorf("MCP server URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("MCP server URL %q must start with http:// or https://", server.URL)
	}
	return &httpTransport{url: server.URL, headers: server.Headers, client: &http.Client{}, events: events}, nil
}

// roundTrip posts req and reads the response with its id from the reply.
func (t *httpTransport) roundTrip(ctx context.Context, req message) (message, error) {
	resp, err := t.do(ctx, http.MethodPost, req)
	if err != nil {
		return message{}, err
	}
	defer resp.Body.Close()
	return t.readReply(resp, req.ID)
}

// send posts a notification or reply; the server answers 202 Accepted.
func (t *httpTransport) send(ctx context.Context, msg message) error {
	resp, err := t.do(ctx, http.MethodPost, msg)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

// setProtocolVersion records the version sent on every later request.
func (t *httpTransport) setProtocolVersion(version string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.protocolVersion = version
}

// close ends the session on the server. It is best effort: the server
// expires idle sessions anyway, and some do not allow DELETE.
func (t *httpTransport) close() error {
	t.mu.Lock()
	session := t.sessionID
	t.mu.Unlock()
	if session == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), closeGrace)
	defer cancel()
	resp, err := t.do(ctx, http.MethodDelete, nil)
	if err != nil {
		return nil // best effort, as above
	}
	return resp.Body.Close()
}
