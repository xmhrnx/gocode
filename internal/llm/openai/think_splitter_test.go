package openai

import "testing"

func splitAll(chunks ...string) (reasoning, text string) {
	var splitter thinkSplitter
	emit := func(isReasoning bool, piece string) {
		if isReasoning {
			reasoning += piece
		} else {
			text += piece
		}
	}
	for _, chunk := range chunks {
		splitter.feed(chunk, emit)
	}
	splitter.flush(emit)
	return reasoning, text
}

func TestThinkSplitter(t *testing.T) {
	cases := []struct {
		name            string
		chunks          []string
		reasoning, text string
	}{
		{"plain", []string{"a < b", " ok"}, "", "a < b ok"},
		{"tag split everywhere", []string{"\n<", "thi", "nk>x", "y</", "think", ">z"}, "xy", "z"},
		{"unclosed", []string{"<think>still going"}, "still going", ""},
		{"partial tag at end", []string{"<thin"}, "", "<thin"},
		{"late tag is text", []string{"hi <think>no</think>"}, "", "hi <think>no</think>"},
	}
	for _, c := range cases {
		reasoning, text := splitAll(c.chunks...)
		if reasoning != c.reasoning || text != c.text {
			t.Errorf("%s: got (%q, %q), want (%q, %q)", c.name, reasoning, text, c.reasoning, c.text)
		}
	}
}
