package chat

import (
	"strings"
	"testing"
)

func TestStreamSegmentsCutAtParagraphsOutsideCode(t *testing.T) {
	text := "aaaa\n\nbbbb\n\n```\ncode\n\nmore code\n```\n\ntail"
	segments := streamSegments(text, 4)
	if strings.Join(segments, "") != text {
		t.Fatalf("segments lose text: %q", segments)
	}
	for _, segment := range segments[:len(segments)-1] {
		if strings.Count(segment, "```")%2 != 0 {
			t.Errorf("segment cuts a code block: %q", segment)
		}
	}
	if got := segments[len(segments)-1]; got != "tail" {
		t.Errorf("last segment = %q", got)
	}
}

func TestStreamSegmentsStayPutAsTextGrows(t *testing.T) {
	short := streamSegments("one\n\ntwo\n\nthr", 3)
	long := streamSegments("one\n\ntwo\n\nthree\n\nfour", 3)
	for i := 0; i < len(short)-1; i++ {
		if short[i] != long[i] {
			t.Fatalf("segment %d moved: %q vs %q", i, short[i], long[i])
		}
	}
}

func TestLiveTailStartsAtALineAndStaysBounded(t *testing.T) {
	text := strings.Repeat("line of text\n", 400)
	got := liveTail(text, 1000)
	if !strings.HasPrefix(got, "…\nline") || len(got) > 1000+10 {
		t.Fatalf("tail of %d bytes starts %q", len(got), got[:12])
	}
}

// Between steps the tail only grows at its end, so the lines before stay
// identical and keep their cached layout.
func TestLiveTailGrowsAtTheEndBetweenSteps(t *testing.T) {
	text := strings.Repeat("abcdefghi\n", 140)
	before := liveTail(text, 1000)
	after := liveTail(text+"more", 1000)
	if !strings.HasPrefix(after, before) {
		t.Fatalf("the shown text moved:\n%q\n%q", before[:30], after[:30])
	}
}
