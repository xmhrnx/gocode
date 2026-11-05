package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gocode/internal/llm"
)

// newServer replays a recorded SSE stream and captures the request.
func newServer(t *testing.T, status int, stream string) (*httptest.Server, *http.Request, map[string]any) {
	t.Helper()
	captured, body := &http.Request{}, map[string]any{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured = *r.Clone(context.Background())
		data, err := io.ReadAll(r.Body)
		if err == nil && len(data) > 0 && json.Unmarshal(data, &body) != nil {
			t.Errorf("bad body %s", data)
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, stream)
	}))
	t.Cleanup(server.Close)
	return server, captured, body
}

func sseEvent(name, data string) string { return "event: " + name + "\ndata: " + data + "\n\n" }

func collect(t *testing.T, client *Client, req llm.Request) ([]llm.Event, error) {
	t.Helper()
	var events []llm.Event
	err := client.Stream(context.Background(), req, func(e llm.Event) { events = append(events, e) })
	return events, err
}

var toolLoopStream = sseEvent("message_start", `{"type":"message_start","message":{"id":"msg_1","role":"assistant","content":[],"usage":{"input_tokens":12,"cache_creation_input_tokens":300,"cache_read_input_tokens":2000,"output_tokens":1}}}`) +
	sseEvent("content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"thinking","thinking":"","signature":""}}`) +
	sseEvent("content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"thinking_delta","thinking":"Check the file."}}`) +
	sseEvent("content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"signature_delta","signature":"sig-1"}}`) +
	sseEvent("content_block_stop", `{"type":"content_block_stop","index":0}`) +
	sseEvent("content_block_start", `{"type":"content_block_start","index":1,"content_block":{"type":"redacted_thinking","data":"opaque"}}`) +
	sseEvent("content_block_stop", `{"type":"content_block_stop","index":1}`) +
	": comment line\n\n" +
	sseEvent("ping", `{"type":"ping"}`) +
	sseEvent("content_block_start", `{"type":"content_block_start","index":2,"content_block":{"type":"text","text":""}}`) +
	sseEvent("content_block_delta", `{"type":"content_block_delta","index":2,"delta":{"type":"text_delta","text":"Reading "}}`) +
	sseEvent("content_block_delta", `{"type":"content_block_delta","index":2,"delta":{"type":"text_delta","text":"it."}}`) +
	sseEvent("content_block_stop", `{"type":"content_block_stop","index":2}`) +
	sseEvent("content_block_start", `{"type":"content_block_start","index":3,"content_block":{"type":"tool_use","id":"toolu_1","name":"read","input":{}}}`) +
	sseEvent("content_block_delta", `{"type":"content_block_delta","index":3,"delta":{"type":"input_json_delta","partial_json":""}}`) +
	sseEvent("content_block_delta", `{"type":"content_block_delta","index":3,"delta":{"type":"input_json_delta","partial_json":"{\"path\": \"a"}}`) +
	sseEvent("content_block_delta", `{"type":"content_block_delta","index":3,"delta":{"type":"input_json_delta","partial_json":".go\"}"}}`) +
	sseEvent("content_block_stop", `{"type":"content_block_stop","index":3}`) +
	sseEvent("content_block_start", `{"type":"content_block_start","index":4,"content_block":{"type":"tool_use","id":"toolu_2","name":"list","input":{}}}`) +
	sseEvent("content_block_stop", `{"type":"content_block_stop","index":4}`) +
	sseEvent("message_delta", `{"type":"message_delta","delta":{"stop_reason":"tool_use","stop_sequence":null},"usage":{"output_tokens":87}}`) +
	sseEvent("message_stop", `{"type":"message_stop"}`)

func TestStreamToolLoopReply(t *testing.T) {
	server, captured, _ := newServer(t, 200, toolLoopStream)
	events, err := collect(t, New(Options{BaseURL: server.URL, APIKey: "key"}), llm.Request{Model: "claude-opus-5-5"})
	if err != nil {
		t.Fatal(err)
	}
	var kinds []string
	for _, e := range events {
		kinds = append(kinds, kindName(e))
	}
	want := "usage reasoning:Check the file. sig:sig-1 redacted:opaque text:Reading  text:it. " +
		`call:toolu_1/read/{"path": "a.go"} call:toolu_2/list/{} usage stop:tool_use`
	if got := strings.Join(kinds, " "); got != want {
		t.Errorf("events\n got: %s\nwant: %s", got, want)
	}
	last := events[len(events)-2].Usage
	if *last != (llm.Usage{InputTokens: 12, OutputTokens: 87, CacheReadTokens: 2000, CacheWriteTokens: 300}) {
		t.Errorf("usage = %+v", *last)
	}
	if captured.URL.Path != "/messages" || captured.Header.Get("x-api-key") != "key" ||
		captured.Header.Get("anthropic-version") != "2026-10-21" || captured.Header.Get("Content-Type") != "application/json" {
		t.Errorf("request = %s %v", captured.URL.Path, captured.Header)
	}
}

func kindName(e llm.Event) string {
	switch {
	case e.Kind == llm.EventText:
		return "text:" + e.Text
	case e.Kind == llm.EventReasoning && e.Redacted:
		return "redacted:" + e.Signature
	case e.Kind == llm.EventReasoning && e.Signature != "":
		return "sig:" + e.Signature
	case e.Kind == llm.EventReasoning:
		return "reasoning:" + e.Text
	case e.Kind == llm.EventToolCall:
		return "call:" + e.ToolCall.ID + "/" + e.ToolCall.Name + "/" + string(e.ToolCall.Arguments)
	case e.Kind == llm.EventStop:
		return "stop:" + e.StopReason
	}
	return "usage"
}

func TestStreamErrors(t *testing.T) {
	overloaded := sseEvent("message_start", `{"type":"message_start","message":{"usage":{"input_tokens":1}}}`) +
		sseEvent("error", `{"type":"error","error":{"type":"overloaded_error","message":"Overloaded"}}`)
	server, _, _ := newServer(t, 200, overloaded)
	_, err := collect(t, New(Options{BaseURL: server.URL}), llm.Request{})
	var apiErr *llm.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 529 || apiErr.Message != "Overloaded" || !apiErr.Retryable() {
		t.Errorf("mid-stream error = %v", err)
	}

	rejected, _, _ := newServer(t, 400, `{"type":"error","error":{"type":"invalid_request_error","message":"temperature is not supported"}}`)
	if _, err := collect(t, New(Options{BaseURL: rejected.URL}), llm.Request{}); !errors.As(err, &apiErr) ||
		apiErr.StatusCode != 400 || apiErr.Message != "temperature is not supported" {
		t.Errorf("http error = %v", err)
	}

	cut, _, _ := newServer(t, 200, sseEvent("message_start", `{"type":"message_start","message":{"usage":{}}}`))
	if _, err := collect(t, New(Options{BaseURL: cut.URL}), llm.Request{}); !errors.Is(err, errIncomplete) {
		t.Errorf("truncated stream error = %v", err)
	}
}
