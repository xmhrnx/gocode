package mcp

import (
	"strings"
	"sync"
)

// tailBuffer is an io.Writer that keeps only the last max bytes written. It
// holds a server's stderr, whose end usually explains why it failed.
type tailBuffer struct {
	mu   sync.Mutex
	max  int
	data []byte
}

// stderrTailBytes is how much of a server's stderr is kept for errors.
const stderrTailBytes = 4096

// newTailBuffer returns a buffer keeping the last max bytes.
func newTailBuffer(max int) *tailBuffer {
	return &tailBuffer{max: max}
}

// Write appends p, dropping the oldest bytes beyond the limit.
func (b *tailBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = append(b.data, p...)
	if excess := len(b.data) - b.max; excess > 0 {
		kept := copy(b.data, b.data[excess:])
		b.data = b.data[:kept]
	}
	return len(p), nil
}

// String returns the kept text, trimmed. The cut may split a character, so
// invalid UTF-8 at the start is dropped.
func (b *tailBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return strings.TrimSpace(strings.ToValidUTF8(string(b.data), ""))
}
