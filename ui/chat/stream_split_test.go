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

func TestReasoningSegmentsCutAtLinesAndStayPut(t *testing.T) {
	text := strings.Repeat("a line of thinking\n", 40)
	segments := reasoningSegments(text, 100)
	if strings.Join(segments, "") != text {
		t.Fatal("segments lost text")
	}
	for _, s := range segments[:len(segments)-1] {
		if !strings.HasSuffix(s, "\n") || len(s) < 100 {
			t.Fatalf("segment %q isn't cut after a line past the size", s)
		}
	}
	grown := reasoningSegments(text+"more thinking", 100)
	for i := range segments[:len(segments)-1] {
		if segments[i] != grown[i] {
			t.Fatalf("segment %d moved as the text grew", i)
		}
	}
}

func TestReasoningSegmentsCutLongLinesAtSpaces(t *testing.T) {
	text := strings.Repeat("word ", 100) // 500 bytes, no line break
	segments := reasoningSegments(text, 100)
	if strings.Join(segments, "") != text || len(segments) < 3 {
		t.Fatalf("got %d segments", len(segments))
	}
	if !strings.HasSuffix(segments[0], " ") || len(segments[0]) > 200 {
		t.Fatalf("first segment %q", segments[0])
	}
}
