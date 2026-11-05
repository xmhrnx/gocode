// Package sse reads Server-Sent Events streams, as used by every provider's
// streaming endpoint.
package sse

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

// Event is one dispatched server-sent event.
type Event struct {
	// Name is the "event:" field, or "" when the server sent none.
	Name string
	// Data is the "data:" lines joined with newlines.
	Data string
}

// Reader parses events from a stream. It uses bufio.Reader rather than
// bufio.Scanner so that very long data lines (large tool arguments, base64
// payloads) are never rejected.
type Reader struct {
	source  *bufio.Reader
	name    string
	data    []string
	hasData bool
}

// NewReader returns a Reader that parses events from r.
func NewReader(r io.Reader) *Reader {
	return &Reader{source: bufio.NewReaderSize(r, 64*1024)}
}

// Next returns the next event. It returns io.EOF once the stream ends with
// no further event; a final event without a trailing blank line is still
// delivered.
func (r *Reader) Next() (Event, error) {
	for {
		line, err := r.source.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return Event{}, err
		}
		atEOF := err != nil
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if r.hasData || r.name != "" {
				return r.dispatch(), nil
			}
			if atEOF {
				return Event{}, io.EOF
			}
			continue
		}
		r.addLine(line)
		if atEOF {
			if r.hasData || r.name != "" {
				return r.dispatch(), nil
			}
			return Event{}, io.EOF
		}
	}
}

// addLine records one non-empty line. Comment lines (starting with ':')
// and fields other than event and data are ignored.
func (r *Reader) addLine(line string) {
	if strings.HasPrefix(line, ":") {
		return
	}
	field, value, _ := strings.Cut(line, ":")
	value = strings.TrimPrefix(value, " ")
	switch field {
	case "event":
		r.name = value
	case "data":
		r.data = append(r.data, value)
		r.hasData = true
	}
}

func (r *Reader) dispatch() Event {
	event := Event{Name: r.name, Data: strings.Join(r.data, "\n")}
	r.name, r.data, r.hasData = "", r.data[:0], false
	return event
}
