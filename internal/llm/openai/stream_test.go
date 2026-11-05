package openai

import (
	"errors"
	"testing"

	"gocode/internal/llm"
)

func TestStreamTextReasoningAndUsage(t *testing.T) {
	server, captured := newSSEServer(t,
		`{"choices":[{"index":0,"delta":{"role":"assistant","content":null,"reasoning_content":"Let me "}}]}`,
		`{"choices":[{"index":0,"delta":{"reasoning":"think.","reasoning_details":[{"type":"reasoning.text","text":"think."}]}}]}`,
		`{"choices":[{"index":0,"delta":{"content":"Hello"}}]}`,
		`{"choices":[{"index":0,"delta":{"content":" world"},"finish_reason":"stop"}]}`,
		`{"choices":[],"usage":{"prompt_tokens":100,"completion_tokens":20,"prompt_tokens_details":{"cached_tokens":60}}}`,
		`[DONE]`)
	client := New(Options{BaseURL: server.URL + "/", APIKey: "sk-test", ExtraHeaders: map[string]string{"X-Title": "GoCode"}})
	events, err := streamAll(t, client, llm.Request{Model: "m", Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart("hi")}}}})
	if err != nil {
		t.Fatal(err)
	}
	if got := joined(events, llm.EventText); got != "Hello world" {
		t.Errorf("text = %q", got)
	}
	if got := joined(events, llm.EventReasoning); got != "Let me think." {
		t.Errorf("reasoning = %q", got)
	}
	usage := eventsOf(events, llm.EventUsage)
	if len(usage) != 1 || *usage[0].Usage != (llm.Usage{InputTokens: 40, OutputTokens: 20, CacheReadTokens: 60}) {
		t.Errorf("usage = %+v", usage)
	}
	if stop := eventsOf(events, llm.EventStop); len(stop) != 1 || stop[0].StopReason != "end_turn" {
		t.Errorf("stop = %+v", stop)
	}
	if captured.path != "/chat/completions" || captured.header.Get("Authorization") != "Bearer sk-test" || captured.header.Get("X-Title") != "GoCode" {
		t.Errorf("request path/headers wrong: %s %v", captured.path, captured.header)
	}
	if captured.body["stream"] != true || captured.body["stream_options"].(map[string]any)["include_usage"] != true {
		t.Errorf("stream options missing: %v", captured.body)
	}
}

func TestStreamToolCallsSplitAcrossChunks(t *testing.T) {
	server, _ := newSSEServer(t,
		`{"choices":[{"delta":{"tool_calls":[{"index":1,"id":"call_b","type":"function","function":{"name":"read","arguments":""}}]}}]}`,
		`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_a","type":"function","function":{"name":"list","arguments":""}}]}}]}`,
		`{"choices":[{"delta":{"tool_calls":[{"index":1,"function":{"arguments":"{\"path\":"}}]}}]}`,
		`{"choices":[{"delta":{"tool_calls":[{"index":1,"function":{"arguments":"\"a.go\"}"}}]}}]}`,
		`{"choices":[{"delta":{},"finish_reason":"tool_calls"}]}`,
		`[DONE]`)
	events, err := streamAll(t, New(Options{BaseURL: server.URL}), llm.Request{Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
	calls := eventsOf(events, llm.EventToolCall)
	if len(calls) != 2 {
		t.Fatalf("calls = %+v", calls)
	}
	first, second := calls[0].ToolCall, calls[1].ToolCall
	if first.ID != "call_a" || first.Name != "list" || string(first.Arguments) != "{}" {
		t.Errorf("first call = %+v", first)
	}
	if second.ID != "call_b" || second.Name != "read" || string(second.Arguments) != `{"path":"a.go"}` {
		t.Errorf("second call = %+v", second)
	}
	if stop := eventsOf(events, llm.EventStop); stop[0].StopReason != "tool_use" {
		t.Errorf("stop = %+v", stop)
	}
}

func TestStreamWholeCallsWithoutIndex(t *testing.T) {
	server, _ := newSSEServer(t,
		`{"choices":[{"delta":{"tool_calls":[{"id":"x","function":{"name":"a","arguments":{"n":1}}}]}}]}`,
		`{"choices":[{"delta":{"tool_calls":[{"id":"y","function":{"name":"b","arguments":"{\"n\":2}"}}]}}]}`,
		`{"choices":[{"delta":{},"finish_reason":"stop"}]}`)
	events, err := streamAll(t, New(Options{BaseURL: server.URL}), llm.Request{Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
	calls := eventsOf(events, llm.EventToolCall)
	if len(calls) != 2 || string(calls[0].ToolCall.Arguments) != `{"n":1}` || calls[1].ToolCall.Name != "b" {
		t.Fatalf("calls = %+v", calls)
	}
}

func TestStreamMistralContentArrayAndInlineThink(t *testing.T) {
	server, _ := newSSEServer(t,
		`{"choices":[{"delta":{"content":[{"type":"thinking","thinking":[{"type":"text","text":"plan"}]}]}}]}`,
		`{"choices":[{"delta":{"content":"<th"}}]}`,
		`{"choices":[{"delta":{"content":"ink>more</thi"}}]}`,
		`{"choices":[{"delta":{"content":"nk>\n\nDone <think> stays"}}]}`,
		`{"choices":[{"delta":{"content":[{"type":"text","text":"!"}]},"finish_reason":"stop"}]}`,
		`[DONE]`)
	events, err := streamAll(t, New(Options{BaseURL: server.URL}), llm.Request{Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
	if got := joined(events, llm.EventReasoning); got != "planmore" {
		t.Errorf("reasoning = %q", got)
	}
	if got := joined(events, llm.EventText); got != "Done <think> stays!" {
		t.Errorf("text = %q", got)
	}
}

func TestStreamErrors(t *testing.T) {
	server, _ := newSSEServer(t,
		`{"choices":[{"delta":{"content":"partial"}}]}`,
		`{"error":{"message":"upstream overloaded","code":502}}`)
	_, err := streamAll(t, New(Options{BaseURL: server.URL}), llm.Request{Model: "m"})
	var apiErr *llm.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 502 || apiErr.Message != "upstream overloaded" {
		t.Errorf("mid-stream error = %v", err)
	}

	truncated, _ := newSSEServer(t, `{"choices":[{"delta":{"content":"cut"}}]}`)
	if _, err := streamAll(t, New(Options{BaseURL: truncated.URL}), llm.Request{Model: "m"}); !errors.Is(err, errIncomplete) {
		t.Errorf("truncated stream error = %v", err)
	}
}
