package gemini

import (
	"encoding/json"
	"strings"
	"testing"

	"gocode/internal/llm"
)

func TestSearchPartsAreKeptAndSentBack(t *testing.T) {
	stream := data(`{"candidates":[{"content":{"role":"model","parts":[` +
		`{"toolCall":{"toolType":"GOOGLE_SEARCH_WEB","args":{"queries":["go 1.26"]},"id":"t1"},"thoughtSignature":"s1"},` +
		`{"toolResponse":{"toolType":"GOOGLE_SEARCH_WEB","response":{"search_suggestions":"x"},"id":"t1"}},` +
		`{"text":"Go 1.26 is out."}]},"finishReason":"STOP"}]}`)
	server, _ := newServer(t, 200, stream)
	events, err := collect(t, server, llm.Request{Model: "gemini-3-flash"})
	if err != nil {
		t.Fatal(err)
	}
	var parts []llm.Part
	for _, e := range events {
		if e.Kind == llm.EventPart {
			parts = append(parts, *e.Part)
		}
	}
	if len(parts) != 2 || parts[0].Text != "Searched the web: go 1.26" || parts[1].Text != "" {
		t.Fatalf("parts %+v", parts)
	}
	sent, _ := json.Marshal(convertContents([]llm.Message{{Role: llm.RoleAssistant, Parts: parts}}))
	if !strings.Contains(string(sent), `"thoughtSignature":"s1"`) || !strings.Contains(string(sent), `"toolResponse"`) {
		t.Fatalf("not sent back unchanged: %s", sent)
	}
}

func TestSearchToolNeedsGemini3BesideFunctions(t *testing.T) {
	tools := []llm.ToolSpec{{Name: "read_file"}, {Name: "web_search"}}
	body := buildRequest(llm.Request{Model: "gemini-3-pro", Tools: tools, WebSearch: true})
	out, _ := json.Marshal(body)
	if !strings.Contains(string(out), `"googleSearch":{}`) || strings.Contains(string(out), `"name":"web_search"`) ||
		!strings.Contains(string(out), `"includeServerSideToolInvocations":true`) {
		t.Fatalf("gemini 3: %s", out)
	}
	body = buildRequest(llm.Request{Model: "gemini-2.5-pro", Tools: tools, WebSearch: true})
	out, _ = json.Marshal(body)
	if strings.Contains(string(out), "googleSearch") || !strings.Contains(string(out), `"name":"web_search"`) {
		t.Fatalf("gemini 2.5 keeps our tool: %s", out)
	}
}
