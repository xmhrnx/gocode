package chat

import "strings"

const (
	// maxLiveReasoning bounds the reasoning shown while it streams: drawing
	// only the tail keeps each frame cheap however long the reasoning gets.
	maxLiveReasoning = 1500
	// streamSegmentSize is about how much finished text one transcript row
	// holds while a reply streams.
	streamSegmentSize = 3000
)

// streamSegments cuts a streaming Markdown reply into rows: finished
// segments of about size bytes, each ending at a paragraph break outside a
// code block, then the rest. A finished segment never changes again, so
// its row is laid out once and replayed from cache; each frame only lays
// out the last segment, however long the reply grows. Cuts depend only on
// the text before them, so they stay put as the reply grows.
func streamSegments(text string, size int) []string {
	var segments []string
	segmentStart, inFence := 0, false
	for start := 0; start < len(text); {
		end := strings.IndexByte(text[start:], '\n')
		if end < 0 {
			break // the last line is still being written
		}
		line := text[start : start+end]
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
		}
		lineEnd := start + end + 1
		if strings.TrimSpace(line) == "" && !inFence && lineEnd-segmentStart >= size {
			segments = append(segments, text[segmentStart:lineEnd])
			segmentStart = lineEnd
		}
		start = lineEnd
	}
	return append(segments, text[segmentStart:])
}

// liveTail returns the end of text, at most about limit bytes, so a long
// stream shows its newest part. The start moves in steps of half the
// limit, not with every chunk: between steps the shown text only grows at
// its end, so the lines already shown keep their cached layout.
func liveTail(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	step := max(limit/2, 1)
	start := (len(text) - limit + step - 1) / step * step
	tail := text[start:]
	if newline := strings.IndexByte(tail, '\n'); newline >= 0 {
		tail = tail[newline+1:]
	}
	return "…\n" + tail
}
