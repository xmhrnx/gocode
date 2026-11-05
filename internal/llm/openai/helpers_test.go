package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gocode/internal/llm"
)

// capturedRequest is what the fake server received.
type capturedRequest struct {
	header http.Header
	path   string
	body   map[string]any
}

// newSSEServer replays payloads as SSE data events, recording the request.
func newSSEServer(t *testing.T, payloads ...string) (*httptest.Server, *capturedRequest) {
	t.Helper()
	captured := &capturedRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.header, captured.path = r.Header.Clone(), r.URL.Path
		data, err := io.ReadAll(r.Body)
		if err != nil || json.Unmarshal(data, &captured.body) != nil {
			t.Errorf("bad request body %q: %v", data, err)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		for _, payload := range payloads {
			_, _ = io.WriteString(w, "data: "+payload+"\n\n")
		}
	}))
	t.Cleanup(server.Close)
	return server, captured
}

func streamAll(t *testing.T, client *Client, req llm.Request) ([]llm.Event, error) {
	t.Helper()
	var events []llm.Event
	err := client.Stream(context.Background(), req, func(e llm.Event) { events = append(events, e) })
	return events, err
}

// joined concatenates the text of all events of one kind.
func joined(events []llm.Event, kind llm.EventKind) string {
	var text strings.Builder
	for _, event := range events {
		if event.Kind == kind {
			text.WriteString(event.Text)
		}
	}
	return text.String()
}

func eventsOf(events []llm.Event, kind llm.EventKind) []llm.Event {
	var matching []llm.Event
	for _, event := range events {
		if event.Kind == kind {
			matching = append(matching, event)
		}
	}
	return matching
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
