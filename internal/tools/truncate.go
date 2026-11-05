package tools

import (
	"fmt"
	"unicode/utf8"
)

// defaultMaxOutputBytes bounds tool output when Env.MaxOutputBytes is 0.
const defaultMaxOutputBytes = 30000

// outputLimit returns the output size bound for env.
func outputLimit(env Env) int {
	if env.MaxOutputBytes > 0 {
		return env.MaxOutputBytes
	}
	return defaultMaxOutputBytes
}

// truncateHead keeps the first max bytes of s, noting how much was cut.
func truncateHead(s string, max int) string {
	if len(s) <= max {
		return s
	}
	cut := runeBoundary(s, max)
	return s[:cut] + fmt.Sprintf("\n[truncated: %d more bytes]", len(s)-cut)
}

// runeBoundary moves i back to the start of the UTF-8 sequence it falls in,
// so cuts never split a character.
func runeBoundary(s string, i int) int {
	if i <= 0 {
		return 0
	}
	if i >= len(s) {
		return len(s)
	}
	for i > 0 && !utf8.RuneStart(s[i]) {
		i--
	}
	return i
}

// truncateLine cuts one line to max bytes for line-oriented output.
func truncateLine(line string, max int) string {
	if len(line) <= max {
		return line
	}
	return line[:runeBoundary(line, max)] + " … [line truncated]"
}
