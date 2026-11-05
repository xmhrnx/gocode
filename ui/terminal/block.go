package terminal

import (
	"context"
	"strings"
	"time"
)

// block is one command and its output.
type block struct {
	id       int
	command  string
	workDir  string
	output   strings.Builder
	lines    int
	dropped  int
	exitCode int
	started  time.Time
	duration time.Duration
	running  bool
	cancel   context.CancelFunc
	revision int
}

// appendOutput adds a chunk, keeping at most maxLines lines so a chatty
// command can't grow memory without bound.
func (b *block) appendOutput(chunk string, maxLines int) {
	b.output.WriteString(chunk)
	b.lines += strings.Count(chunk, "\n")
	if b.lines > maxLines+maxLines/4 {
		b.trimTo(maxLines)
	}
	b.revision++
}

func (b *block) trimTo(maxLines int) {
	text := b.output.String()
	excess := b.lines - maxLines
	cut := 0
	for i := 0; i < excess; i++ {
		next := strings.IndexByte(text[cut:], '\n')
		if next < 0 {
			break
		}
		cut += next + 1
	}
	b.output.Reset()
	b.output.WriteString(text[cut:])
	b.lines -= excess
	b.dropped += excess
}

// text returns the output with trailing blank lines removed.
func (b *block) text() string { return strings.TrimRight(b.output.String(), "\n") }
