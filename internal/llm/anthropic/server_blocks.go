package anthropic

import (
	"encoding/json"
	"strings"

	"gocode/internal/llm"
)

// providerName marks the parts this adapter makes, so it sends back only
// its own.
const providerName = "anthropic"

// serverToolUse is what a server_tool_use block says about the call.
type serverToolUse struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Input struct {
		Query string `json:"query"`
		URL   string `json:"url"`
	} `json:"input"`
}

// emitServerBlock sends a finished server tool block as a provider part,
// with the input that streamed in put into the block.
func (s *streamState) emitServerBlock(block *pendingServerBlock) {
	raw := block.raw
	if input := strings.TrimSpace(block.input.String()); input != "" {
		if !json.Valid([]byte(input)) {
			return // cut short; it can't be sent back
		}
		raw = withField(raw, "input", json.RawMessage(input))
	}
	part := llm.Part{Kind: llm.PartProvider, Provider: providerName, Raw: raw, Text: describeServerBlock(raw)}
	s.emit(llm.Event{Kind: llm.EventPart, Part: &part})
}

// withField returns the JSON object raw with field set to value.
func withField(raw json.RawMessage, field string, value json.RawMessage) json.RawMessage {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		return raw
	}
	object[field] = value
	out, err := json.Marshal(object)
	if err != nil {
		return raw
	}
	return out
}

// describeServerBlock says what a server tool call did, for the
// transcript; result blocks and other tools get nothing.
func describeServerBlock(raw json.RawMessage) string {
	var use serverToolUse
	if json.Unmarshal(raw, &use) != nil || use.Type != "server_tool_use" {
		return ""
	}
	switch {
	case use.Name == "web_search" && use.Input.Query != "":
		return "Searched the web: " + use.Input.Query
	case use.Name == "web_fetch" && use.Input.URL != "":
		return "Read " + use.Input.URL
	}
	return ""
}

// webSearchTool is the server tool for model: the version with dynamic
// filtering where the model has it, the basic one otherwise.
func webSearchTool(model string) map[string]string {
	version := "web_search_20250305"
	for _, family := range []string{"opus-4-6", "opus-4-7", "opus-4-8", "opus-5", "sonnet-4-6", "sonnet-5", "fable", "mythos"} {
		if strings.Contains(model, family) {
			version = "web_search_20260209"
			break
		}
	}
	return map[string]string{"type": version, "name": "web_search"}
}
