package openai

import (
	"encoding/json"
	"strings"
	"testing"

	"gocode/internal/llm"
)

func TestWebSearchUsesTheProvidersTool(t *testing.T) {
	tools := []llm.ToolSpec{{Name: "web_search"}, {Name: "read_file"}}
	router := New(Options{BaseURL: "https://openrouter.ai/api/v1", WebSearchTool: "openrouter:web_search"})
	out, _ := json.Marshal(router.buildRequest(llm.Request{Model: "m", Tools: tools, WebSearch: true}))
	if !strings.Contains(string(out), `{"type":"openrouter:web_search"}`) || strings.Contains(string(out), `"name":"web_search"`) {
		t.Fatalf("openrouter: %s", out)
	}
	plain := New(Options{BaseURL: "https://api.example.com/v1"})
	out, _ = json.Marshal(plain.buildRequest(llm.Request{Model: "m", Tools: tools, WebSearch: true}))
	if strings.Contains(string(out), "openrouter") || !strings.Contains(string(out), `"name":"web_search"`) {
		t.Fatalf("a provider without search keeps our tool: %s", out)
	}
}
