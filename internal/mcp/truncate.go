package mcp

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// truncateMiddle bounds s to about max bytes, keeping its head and tail:
// tool output often starts with context and ends with the conclusion.
func truncateMiddle(s string, max int) string {
	if len(s) <= max {
		return s
	}
	headEnd := runeBoundary(s, max/2)
	tailStart := runeBoundary(s, len(s)-(max-max/2))
	cut := tailStart - headEnd
	return fmt.Sprintf("%s\n[... %d bytes truncated ...]\n%s", s[:headEnd], cut, s[tailStart:])
}

// oneLine collapses whitespace runs to single spaces and cuts the result
// to max bytes, for summaries shown in a single row.
func oneLine(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= max {
		return s
	}
	return s[:runeBoundary(s, max)] + "…"
}

// runeBoundary moves i back to the start of the UTF-8 sequence it falls
// in, so cuts never split a character.
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
