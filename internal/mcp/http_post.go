package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// do sends one HTTP request carrying msg (none when nil) with the MCP
// headers, failing on a non-2xx status.
func (t *httpTransport) do(ctx context.Context, method string, msg any) (*http.Response, error) {
	var body io.Reader
	if msg != nil {
		encoded, err := json.Marshal(msg)
		if err != nil {
			return nil, fmt.Errorf("encode message: %w", err)
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, t.url, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	t.setHeaders(req.Header)
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reach MCP server: %w", err)
	}
	if err := statusError(resp); err != nil {
		return nil, errors.Join(err, resp.Body.Close())
	}
	t.rememberSession(resp.Header.Get(headerSessionID))
	return resp, nil
}

// setHeaders adds the protocol headers and the user's own.
func (t *httpTransport) setHeaders(header http.Header) {
	header.Set("Content-Type", "application/json")
	header.Set("Accept", "application/json, text/event-stream")
	for key, value := range t.headers {
		header.Set(key, value)
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sessionID != "" {
		header.Set(headerSessionID, t.sessionID)
	}
	if t.protocolVersion != "" {
		header.Set(headerProtocolVersion, t.protocolVersion)
	}
}

// rememberSession stores the session id the server assigned, if any.
func (t *httpTransport) rememberSession(id string) {
	if id == "" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sessionID = id
}

// statusError explains a non-2xx response, quoting the start of its body.
func statusError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	snippet, err := io.ReadAll(io.LimitReader(resp.Body, errorBodyLimit))
	if err != nil {
		return fmt.Errorf("MCP server returned %s", resp.Status)
	}
	return fmt.Errorf("MCP server returned %s: %s", resp.Status, strings.TrimSpace(string(snippet)))
}
