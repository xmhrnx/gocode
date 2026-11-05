package chat

import "testing"

func TestSplitSettled(t *testing.T) {
	cases := []struct{ text, settled, tail string }{
		{"one paragraph", "", "one paragraph"},
		{"first\n\nsecond", "first", "second"},
		{"intro\n\n```go\nx\n\ny", "intro", "```go\nx\n\ny"},
		{"a\n\n```\ncode\n```\n\nb", "a\n\n```\ncode\n```", "b"},
	}
	for _, c := range cases {
		settled, tail := splitSettled(c.text)
		if settled != c.settled || tail != c.tail {
			t.Errorf("splitSettled(%q) = %q, %q; want %q, %q", c.text, settled, tail, c.settled, c.tail)
		}
	}
}

func TestLiveTailStartsAtALine(t *testing.T) {
	if got := liveTail("aaaa\nbbbb\ncccc", 7); got != "…\ncccc" {
		t.Fatalf("got %q", got)
	}
}
