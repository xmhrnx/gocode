package chat

import "strings"

// maxLiveReasoning bounds the reasoning shown while it streams: drawing
// only the tail keeps each frame cheap however long the reasoning gets.
const maxLiveReasoning = 1500

// splitSettled splits streaming Markdown into a settled head, made of
// complete blocks that won't change, and the tail still being written.
// The head's source stays identical between frames, so its parse is
// served from cache and only the short tail is parsed again.
func splitSettled(text string) (settled, tail string) {
	cut, inFence, previousBlank := -1, false, false
	for start := 0; start < len(text); {
		end := strings.IndexByte(text[start:], '\n')
		if end < 0 {
			break // the last line is still being written
		}
		line := text[start : start+end]
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
		}
		blank := strings.TrimSpace(line) == ""
		if blank && !previousBlank && !inFence && start > 0 {
			cut = start - 1 // the newline ending the block before
		}
		previousBlank = blank
		start += end + 1
	}
	if cut <= 0 {
		return "", text
	}
	return text[:cut], strings.TrimLeft(text[cut:], "\n")
}

// liveTail returns the end of text, starting at a line boundary, so a
// long stream shows its newest part.
func liveTail(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	tail := text[len(text)-limit:]
	if newline := strings.IndexByte(tail, '\n'); newline >= 0 {
		tail = tail[newline+1:]
	}
	return "…\n" + tail
}
