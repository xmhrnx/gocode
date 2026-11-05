package openai

import (
	"encoding/json"
	"testing"
)

// Servers that stream byte-level tokens can end a fragment inside a
// multi-byte character; the halves must rejoin, not become U+FFFD.
func TestToolArgumentsRejoinSplitCharacters(t *testing.T) {
	arrow := "→"
	fragments := []string{
		`"{\"content\":\"a` + arrow[:2] + `"`,
		`"` + arrow[2:] + `\\tb\"}"`,
	}
	var acc toolCallAccumulator
	index := 0
	for i, fragment := range fragments {
		delta := toolCallDelta{Index: &index}
		if i == 0 {
			delta.ID, delta.Function.Name = "call_1", "write_file"
		}
		delta.Function.Arguments = json.RawMessage(fragment)
		acc.add(delta)
	}
	var args struct{ Content string }
	if err := json.Unmarshal(acc.complete()[0].Arguments, &args); err != nil {
		t.Fatal(err)
	}
	if args.Content != "a→\tb" {
		t.Fatalf("content = %q", args.Content)
	}
}
