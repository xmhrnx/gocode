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

func TestLiveTailStartsAtALine(t *testing.T) {
	if got := liveTail("aaaa\nbbbb\ncccc", 7); got != "…\ncccc" {
		t.Fatalf("got %q", got)
	}
}
