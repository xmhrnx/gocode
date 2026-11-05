package sse

import (
	"errors"
	"io"
)

// ErrStop may be returned by a ForEach handler to end the stream early
// without reporting an error, e.g. on a "[DONE]" sentinel.
var ErrStop = errors.New("sse: stop")

// ForEach calls handle for every event in r until the stream ends or handle
// returns an error. ErrStop ends the loop and ForEach returns nil.
func ForEach(r io.Reader, handle func(Event) error) error {
	reader := NewReader(r)
	for {
		event, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := handle(event); err != nil {
			if errors.Is(err, ErrStop) {
				return nil
			}
			return err
		}
	}
}
