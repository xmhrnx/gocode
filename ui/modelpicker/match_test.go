package modelpicker

import "testing"

func TestMatches(t *testing.T) {
	cases := []struct {
		query, text string
		want        bool
	}{
		{"", "anything", true},
		{"son 5", "Anthropic claude-sonnet-5-5", true},
		{"GPT", "openai gpt-6-astra", true},
		{"llama", "openai gpt-6-astra", false},
	}
	for _, c := range cases {
		if got := matches(c.query, c.text); got != c.want {
			t.Errorf("matches(%q, %q) = %v", c.query, c.text, got)
		}
	}
}
