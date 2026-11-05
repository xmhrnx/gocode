package mcp

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// maxSSELine bounds one line of an event stream; a tool result can be large.
const maxSSELine = 16 << 20

// readSSE reads a text/event-stream and calls onData with the data of each
// event, until onData returns false or the stream ends. onData must not
// keep the slice it is given.
func readSSE(r io.Reader, onData func(data []byte) bool) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64<<10), maxSSELine)
	var data bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if data.Len() > 0 && !onData(data.Bytes()) {
				return nil
			}
			data.Reset()
			continue
		}
		field, value, _ := strings.Cut(line, ":")
		if field != "data" {
			continue // comments (": ...") and the id, event and retry fields
		}
		if data.Len() > 0 {
			data.WriteByte('\n')
		}
		data.WriteString(strings.TrimPrefix(value, " "))
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	// Be lenient with a final event that lacks its closing blank line.
	if data.Len() > 0 {
		onData(data.Bytes())
	}
	return nil
}
