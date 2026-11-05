package gemini

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gocode/internal/llm"
)

func TestBuildRequestConvertsToolLoop(t *testing.T) {
	temperature := 0.5
	history := []llm.Message{
		{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart("look"), llm.ImagePart("image/png", []byte{1, 2, 3})}},
		{Role: llm.RoleAssistant, Parts: []llm.Part{
			{Kind: llm.PartReasoning, Text: "plan", Signature: "s1"},
			llm.TextPart("Reading."),
			{Kind: llm.PartToolCall, ToolCall: &llm.ToolCall{ID: "fc1", Name: "read", Arguments: json.RawMessage(`{"path":"a"}`), Signature: "s2"}},
			{Kind: llm.PartToolCall, ToolCall: &llm.ToolCall{ID: "gocode_call_2", Name: "list"}},
		}},
		{Role: llm.RoleUser, Parts: []llm.Part{
			llm.TextPart("thanks"),
			llm.ToolResultPart(llm.ToolResult{CallID: "fc1", Name: "read", Content: "body"}),
			llm.ToolResultPart(llm.ToolResult{CallID: "gocode_call_2", Name: "list", Content: "denied", IsError: true}),
		}},
	}
	got := mustJSON(t, buildRequest(llm.Request{
		Model: "m", System: "sys", Messages: history, MaxTokens: 64, Temperature: &temperature, ReasoningEffort: "medium",
		Tools: []llm.ToolSpec{{Name: "read", Description: "Read", Parameters: json.RawMessage(`{"type":"object"}`)}},
	}))
	want := `{"contents":[` +
		`{"role":"user","parts":[{"text":"look"},{"inlineData":{"mimeType":"image/png","data":"AQID"}}]},` +
		`{"role":"model","parts":[{"text":"plan","thought":true,"thoughtSignature":"s1"},{"text":"Reading."},` +
		`{"thoughtSignature":"s2","functionCall":{"id":"fc1","name":"read","args":{"path":"a"}}},{"functionCall":{"name":"list","args":{}}}]},` +
		`{"role":"user","parts":[{"functionResponse":{"id":"fc1","name":"read","response":{"content":"body"}}},` +
		`{"functionResponse":{"name":"list","response":{"error":"denied"}}},{"text":"thanks"}]}],` +
		`"systemInstruction":{"parts":[{"text":"sys"}]},` +
		`"tools":[{"functionDeclarations":[{"name":"read","description":"Read","parametersJsonSchema":{"type":"object"}}]}],` +
		`"generationConfig":{"maxOutputTokens":64,"temperature":0.5,"thinkingConfig":{"includeThoughts":true,"thinkingLevel":"MEDIUM"}}}`
	if got != want {
		t.Errorf("request JSON\n got: %s\nwant: %s", got, want)
	}
}

func TestListModelsFollowsPagesAndFilters(t *testing.T) {
	pages := map[string]string{
		"": `{"models":[{"name":"models/gemini-4-argon","displayName":"Gemini 4 Argon","inputTokenLimit":1048576,"supportedGenerationMethods":["generateContent","countTokens"]},` +
			`{"name":"models/text-embedding","supportedGenerationMethods":["embedContent"]}],"nextPageToken":"p2"}`,
		"p2": `{"models":[{"name":"models/gemini-3.8-flash","supportedGenerationMethods":["generateContent"]}]}`,
	}
	server := pagedServer(t, pages)
	models, err := New(Options{BaseURL: server.URL}).ListModels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 || models[0].ID != "gemini-3.8-flash" || models[1].Name != "Gemini 4 Argon" || models[1].ContextWindow != 1048576 {
		t.Errorf("models = %+v", models)
	}
}

// pagedServer answers GET /models with the page for the request's token.
func pagedServer(t *testing.T, pages map[string]string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" || r.URL.Query().Get("pageSize") != "1000" {
			t.Errorf("unexpected request %s", r.URL)
		}
		_, _ = io.WriteString(w, pages[r.URL.Query().Get("pageToken")])
	}))
	t.Cleanup(server.Close)
	return server
}
