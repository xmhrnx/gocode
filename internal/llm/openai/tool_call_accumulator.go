package openai

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"gocode/internal/llm"
	"gocode/internal/llm/jsonstr"
)

// toolCallDelta is one streamed fragment of a tool call.
type toolCallDelta struct {
	Index    *int   `json:"index"`
	ID       string `json:"id"`
	Function struct {
		Name string `json:"name"`
		// Arguments is normally a string fragment, but some servers send
		// the whole arguments object.
		Arguments json.RawMessage `json:"arguments"`
	} `json:"function"`
}

type pendingCall struct {
	id, name  string
	arguments strings.Builder
}

// toolCallAccumulator joins fragments into complete calls. Fragments are
// matched by index; servers that omit it are matched by id, and anonymous
// fragments continue the most recent call.
type toolCallAccumulator struct {
	calls map[int]*pendingCall
	last  int
}

func (a *toolCallAccumulator) add(delta toolCallDelta) {
	call := a.slot(delta)
	if delta.ID != "" {
		call.id = delta.ID
	}
	if delta.Function.Name != "" {
		call.name = delta.Function.Name
	}
	call.arguments.WriteString(argumentText(delta.Function.Arguments))
}

func (a *toolCallAccumulator) slot(delta toolCallDelta) *pendingCall {
	if a.calls == nil {
		a.calls = map[int]*pendingCall{}
	}
	index := a.last
	if delta.Index != nil {
		index = *delta.Index
	}
	if delta.ID != "" && a.matchByID(index, delta) {
		index = a.indexForID(delta.ID)
	}
	a.last = index
	if _, ok := a.calls[index]; !ok {
		a.calls[index] = &pendingCall{}
	}
	return a.calls[index]
}

// matchByID reports whether a fragment carrying an id must be placed by
// that id: it has no index, or its index already holds a different call
// (some servers number every call 0).
func (a *toolCallAccumulator) matchByID(index int, delta toolCallDelta) bool {
	if delta.Index == nil {
		return true
	}
	existing, ok := a.calls[index]
	return ok && existing.id != "" && existing.id != delta.ID
}

func (a *toolCallAccumulator) indexForID(id string) int {
	next := 0
	for index, call := range a.calls {
		if call.id == id {
			return index
		}
		next = max(next, index+1)
	}
	return next
}

// complete returns the accumulated calls in index order.
func (a *toolCallAccumulator) complete() []llm.ToolCall {
	indexes := make([]int, 0, len(a.calls))
	for index := range a.calls {
		indexes = append(indexes, index)
	}
	slices.Sort(indexes)
	calls := make([]llm.ToolCall, 0, len(indexes))
	for _, index := range indexes {
		pending := a.calls[index]
		id := pending.id
		if id == "" {
			id = fmt.Sprintf("call_%d", index)
		}
		calls = append(calls, llm.ToolCall{ID: id, Name: pending.name, Arguments: validArguments(pending.arguments.String())})
	}
	return calls
}

func argumentText(raw json.RawMessage) string {
	if text, ok := jsonstr.Decode(raw); ok {
		return text
	}
	if string(raw) == "null" {
		return ""
	}
	return string(raw)
}

// validArguments guarantees valid JSON, since the arguments are stored in
// history as raw JSON. Malformed text is kept as a JSON string so the tool
// reports the problem to the model instead of the history breaking.
func validArguments(text string) json.RawMessage {
	text = strings.TrimSpace(text)
	if text == "" {
		return json.RawMessage("{}")
	}
	if json.Valid([]byte(text)) {
		return json.RawMessage(text)
	}
	quoted, err := json.Marshal(text)
	if err != nil {
		return json.RawMessage("{}")
	}
	return quoted
}
