package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gocode/internal/settings"
)

// fakeHTTPServer is a Streamable HTTP MCP server that answers initialize
// with plain JSON and later requests with SSE streams.
type fakeHTTPServer struct {
	mu      sync.Mutex
	deleted bool
}

const fakeSessionID = "sess-1"

func (s *fakeHTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		s.mu.Lock()
		s.deleted = r.Header.Get(headerSessionID) == fakeSessionID
		s.mu.Unlock()
		return
	}
	var msg message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if msg.Method == "initialize" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(headerSessionID, fakeSessionID)
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":{"protocolVersion":"2025-06-18"}}`, msg.ID)
		return
	}
	if r.Header.Get(headerSessionID) != fakeSessionID || r.Header.Get(headerProtocolVersion) != "2025-06-18" ||
		r.Header.Get("X-Token") != "secret" || !strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		http.Error(w, "missing headers", http.StatusBadRequest)
		return
	}
	if !msg.isRequest() {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	s.stream(w, msg)
}

// stream answers over SSE, preceded by a comment and a notification, and
// with the response split over two data lines.
func (s *fakeHTTPServer) stream(w http.ResponseWriter, req message) {
	result := `{"tools":[{"name":"remote"}]}`
	if req.Method == "tools/call" {
		result = `{"content":[],"structuredContent":{"ok":true}}`
	}
	w.Header().Set("Content-Type", "text/event-stream")
	fmt.Fprint(w, ": keep-alive\n\n")
	fmt.Fprint(w, "event: message\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"notifications/tools/list_changed\"}\n\n")
	fmt.Fprintf(w, "id: 7\ndata: {\"jsonrpc\":\"2.0\",\"id\":%s,\r\ndata: \"result\":%s}\n\n", req.ID, result)
}

func TestHTTPTransportJSONAndSSE(t *testing.T) {
	fake := &fakeHTTPServer{}
	server := httptest.NewServer(fake)
	defer server.Close()
	var notified atomic.Int32
	events := serverEvents{notify: func(method string, _ json.RawMessage) {
		if method == "notifications/tools/list_changed" {
			notified.Add(1)
		}
	}}
	config := settings.MCPServer{Name: "remote", Transport: "http", URL: server.URL, Headers: map[string]string{"X-Token": "secret"}}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := connect(ctx, config, events)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	list, err := c.listTools(ctx)
	if err != nil || len(list) != 1 || list[0].Name != "remote" {
		t.Fatalf("listTools = %+v, %v", list, err)
	}
	if notified.Load() != 1 {
		t.Errorf("saw %d list_changed notifications, want 1", notified.Load())
	}
	result, err := c.callTool(ctx, "remote", json.RawMessage(`{}`))
	if err != nil || result.text() != `{"ok":true}` {
		t.Fatalf("callTool = %q, %v; want the structured content", result.text(), err)
	}
	if err := c.close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if !fake.deleted {
		t.Error("session was not deleted on close")
	}
}

func TestHTTPTransportRejectsBadURL(t *testing.T) {
	_, err := openTransport(settings.MCPServer{Transport: "http", URL: "ftp://example.com"}, serverEvents{})
	if err == nil {
		t.Fatal("ftp URL accepted")
	}
}
