package chat

import "strings"

const (
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

// reasoningSegments cuts reasoning into rows of about size bytes. A cut
// goes after the first line break past size bytes or, in a line longer
// than twice that, after the last space before it. Like streamSegments,
// cuts depend only on the text before them, so finished rows never
// change while the reasoning grows.
func reasoningSegments(text string, size int) []string {
	var segments []string
	start := 0
	for len(text)-start > size {
		window := text[start:min(len(text), start+2*size)]
		cut := -1
		if nl := strings.IndexByte(window[size-1:], '\n'); nl >= 0 {
			cut = size + nl
		} else if len(window) == 2*size {
			cut = 2 * size
			if space := strings.LastIndexByte(window[size:], ' '); space >= 0 {
				cut = size + space + 1
			}
		}
		if cut < 0 {
			break // the long line is still being written
		}
		segments = append(segments, text[start:start+cut])
		start += cut
	}
	return append(segments, text[start:])
}
