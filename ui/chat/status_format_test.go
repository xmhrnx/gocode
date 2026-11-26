package chat

import (
	"testing"
	"time"
)

func TestWorkingTime(t *testing.T) {
	for d, want := range map[time.Duration]string{
		9 * time.Second:                 "9s",
		65 * time.Second:                "1m 05s",
		59*time.Minute + 59*time.Second: "59m 59s",
		2*time.Hour + 3*time.Minute:     "2h 03m",
	} {
		if got := workingTime(d); got != want {
			t.Errorf("workingTime(%v) = %q, want %q", d, got, want)
		}
	}
}

func TestTokenCount(t *testing.T) {
	for n, want := range map[int]string{950: "950", 2412: "2.4k", 12345: "12k", 1_340_000: "1.3M"} {
		if got := tokenCount(n); got != want {
			t.Errorf("tokenCount(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestCleanTitle(t *testing.T) {
	for reply, want := range map[string]string{
		"\"Fix the parser\"\n":           "Fix the parser",
		"Title: Add dark mode.":          "Add dark mode",
		"  **Refactor tokenizer**  \nok": "Refactor tokenizer",
	} {
		if got := cleanTitle(reply); got != want {
			t.Errorf("cleanTitle(%q) = %q, want %q", reply, got, want)
		}
	}
}
