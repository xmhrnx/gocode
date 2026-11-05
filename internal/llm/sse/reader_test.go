package sse

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func readAll(t *testing.T, stream string) []Event {
	t.Helper()
	reader := NewReader(strings.NewReader(stream))
	var events []Event
	for {
		event, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return events
		}
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		events = append(events, event)
	}
}

func TestReaderParsesNamedMultiLineAndCommentedEvents(t *testing.T) {
	stream := ": keep-alive\n\n" +
		"event: message_start\ndata: {\"a\":1}\n\n" +
		"data: line one\r\ndata: line two\r\n\r\n" +
		"id: 7\nretry: 100\ndata:no-space\n\n" +
		"data: [DONE]\n\n"
	events := readAll(t, stream)
	want := []Event{
		{Name: "message_start", Data: `{"a":1}`},
		{Data: "line one\nline two"},
		{Data: "no-space"},
		{Data: "[DONE]"},
	}
	if len(events) != len(want) {
		t.Fatalf("got %d events %+v, want %d", len(events), events, len(want))
	}
	for i := range want {
		if events[i] != want[i] {
			t.Errorf("event %d = %+v, want %+v", i, events[i], want[i])
		}
	}
}

func TestReaderHandlesLargeLinesAndMissingTrailingBlankLine(t *testing.T) {
	big := strings.Repeat("x", 300*1024)
	events := readAll(t, "data: "+big+"\n\ndata: tail")
	if len(events) != 2 || events[0].Data != big || events[1].Data != "tail" {
		t.Fatalf("unexpected events: %d", len(events))
	}
}

func TestReaderIgnoresTrailingComment(t *testing.T) {
	events := readAll(t, "data: one\n\n: bye")
	if len(events) != 1 || events[0].Data != "one" {
		t.Fatalf("unexpected events: %+v", events)
	}
}

func TestForEachStopsOnErrStop(t *testing.T) {
	var seen []string
	err := ForEach(strings.NewReader("data: a\n\ndata: [DONE]\n\ndata: b\n\n"), func(e Event) error {
		if e.Data == "[DONE]" {
			return ErrStop
		}
		seen = append(seen, e.Data)
		return nil
	})
	if err != nil || len(seen) != 1 || seen[0] != "a" {
		t.Fatalf("ForEach = %v, seen %v", err, seen)
	}
}
