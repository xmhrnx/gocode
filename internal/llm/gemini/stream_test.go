package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gocode/internal/llm"
)

// newServer serves body with status and captures the request.
func newServer(t *testing.T, status int, body string) (*httptest.Server, *http.Request) {
	t.Helper()
	captured := &http.Request{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured = *r.Clone(context.Background())
		if _, err := io.ReadAll(r.Body); err != nil {
			t.Error(err)
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(server.Close)
	return server, captured
}

func data(payload string) string { return "data: " + payload + "\r\n\r\n" }

func collect(t *testing.T, server *httptest.Server, req llm.Request) ([]llm.Event, error) {
	t.Helper()
	var events []llm.Event
	err := New(Options{BaseURL: server.URL, APIKey: "gk"}).Stream(context.Background(), req, func(e llm.Event) { events = append(events, e) })
	return events, err
}

func TestStreamThoughtsTextAndFunctionCalls(t *testing.T) {
	stream := data(`{"candidates":[{"content":{"role":"model","parts":[{"text":"Plan the read.","thought":true}]}}],"usageMetadata":{"promptTokenCount":50,"thoughtsTokenCount":5}}`) +
		data(`{"candidates":[{"content":{"role":"model","parts":[{"text":"Reading"},{"text":" now.","thoughtSignature":"c2lnLXRleHQ="}]}}]}`) +
		data(`{"candidates":[{"content":{"role":"model","parts":[{"functionCall":{"id":"fc1","name":"read","args":{"path":"a.go"}},"thoughtSignature":"c2lnLWNhbGw="},{"functionCall":{"name":"list"}}]},"finishReason":"STOP"}],`+
			`"usageMetadata":{"promptTokenCount":50,"candidatesTokenCount":20,"thoughtsTokenCount":5,"cachedContentTokenCount":30}}`)
	server, captured := newServer(t, 200, stream)
	events, err := collect(t, server, llm.Request{Model: "models/gemini-4-argon"})
	if err != nil {
		t.Fatal(err)
	}
	if captured.URL.Path != "/models/gemini-4-argon:streamGenerateContent" || captured.URL.Query().Get("alt") != "sse" || captured.Header.Get("x-goog-api-key") != "gk" {
		t.Errorf("request = %s %v", captured.URL, captured.Header)
	}
	want := []llm.Event{
		{Kind: llm.EventReasoning, Text: "Plan the read."},
		{Kind: llm.EventUsage, Usage: &llm.Usage{InputTokens: 50, OutputTokens: 5}},
		{Kind: llm.EventText, Text: "Reading"},
		{Kind: llm.EventText, Text: " now.", Signature: "c2lnLXRleHQ="},
		{Kind: llm.EventToolCall, ToolCall: &llm.ToolCall{ID: "fc1", Name: "read", Arguments: json.RawMessage(`{"path":"a.go"}`), Signature: "c2lnLWNhbGw="}},
		{Kind: llm.EventToolCall, ToolCall: &llm.ToolCall{ID: "gocode_call_2", Name: "list", Arguments: json.RawMessage(`{}`)}},
		{Kind: llm.EventUsage, Usage: &llm.Usage{InputTokens: 20, OutputTokens: 25, CacheReadTokens: 30}},
		{Kind: llm.EventStop, StopReason: "tool_use"},
	}
	if got, expected := mustJSON(t, events), mustJSON(t, want); got != expected {
		t.Errorf("events\n got: %s\nwant: %s", got, expected)
	}
}

func TestStreamErrors(t *testing.T) {
	midStream, _ := newServer(t, 200, data(`{"candidates":[{"content":{"parts":[{"text":"Hi"}]}}]}`)+
		data(`{"error":{"code":503,"message":"The model is overloaded.","status":"UNAVAILABLE"}}`))
	_, err := collect(t, midStream, llm.Request{Model: "m"})
	var apiErr *llm.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 503 || apiErr.Message != "The model is overloaded." {
		t.Errorf("mid-stream error = %v", err)
	}

	badKey, _ := newServer(t, 400, `{"error":{"code":400,"message":"API key not valid.","status":"INVALID_ARGUMENT"}}`)
	if _, err := collect(t, badKey, llm.Request{Model: "m"}); !errors.As(err, &apiErr) || apiErr.Message != "API key not valid." {
		t.Errorf("http error = %v", err)
	}

	blocked, _ := newServer(t, 200, data(`{"promptFeedback":{"blockReason":"SAFETY"}}`))
	if _, err := collect(t, blocked, llm.Request{Model: "m"}); !errors.As(err, &apiErr) || apiErr.StatusCode != 400 {
		t.Errorf("blocked error = %v", err)
	}

	cut, _ := newServer(t, 200, data(`{"candidates":[{"content":{"parts":[{"text":"Hi"}]}}]}`))
	if _, err := collect(t, cut, llm.Request{Model: "m"}); !errors.Is(err, errIncomplete) {
		t.Errorf("truncated error = %v", err)
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}
