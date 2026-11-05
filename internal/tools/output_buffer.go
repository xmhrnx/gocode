package tools

import (
	"fmt"
	"strings"
	"sync"
)

// headTailBuffer is an io.Writer that keeps only the first and last limit/2
// bytes written, so a command producing gigabytes of output cannot exhaust
// memory while its start and its (usually most telling) end survive.
type headTailBuffer struct {
	mu      sync.Mutex
	limit   int
	head    []byte
	tail    []byte
	dropped int
}

func newHeadTailBuffer(limit int) *headTailBuffer {
	return &headTailBuffer{limit: limit}
}

// Write records p; it never fails.
func (b *headTailBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := len(p)
	if room := b.limit/2 - len(b.head); room > 0 {
		take := min(room, len(p))
		b.head = append(b.head, p[:take]...)
		p = p[take:]
	}
	b.tail = append(b.tail, p...)
	if excess := len(b.tail) - (b.limit - b.limit/2); excess > 0 {
		b.dropped += excess
		b.tail = append(b.tail[:0], b.tail[excess:]...)
	}
	return n, nil
}

// String returns the kept output with a note where bytes were dropped.
func (b *headTailBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.dropped == 0 {
		return string(b.head) + string(b.tail)
	}
	// The cuts may split a multi-byte character; drop the broken bytes.
	head := strings.ToValidUTF8(string(b.head), "")
	tail := strings.ToValidUTF8(string(b.tail), "")
	return head + fmt.Sprintf("\n[truncated: %d more bytes]\n", b.dropped) + tail
}
