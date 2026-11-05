package shell

import (
	"strings"
	"unicode/utf8"
)

// outputFilter turns raw command output into display text: escape
// sequences are removed and carriage returns rewind to the start of the
// line, as in a terminal, so a progress bar redrawn a hundred times ends
// up as one line. Only completed lines are emitted, since a line can
// still be overwritten until its newline arrives.
type outputFilter struct {
	stripper  ansiStripper
	partial   []byte // incomplete UTF-8 sequence from the previous chunk
	line      []rune
	col       int
	pendingCR bool // a \r whose meaning depends on whether \n follows
}

// write consumes a chunk and returns the lines it completed.
func (f *outputFilter) write(chunk []byte) string {
	data := append(f.partial, chunk...)
	f.partial = nil
	var out strings.Builder
	for len(data) > 0 {
		if !utf8.FullRune(data) {
			f.partial = append([]byte(nil), data...)
			break
		}
		r, size := utf8.DecodeRune(data)
		data = data[size:]
		if f.stripper.keep(r) {
			f.put(r, &out)
		}
	}
	return out.String()
}

// flush returns the unfinished last line, for when the command ends.
func (f *outputFilter) flush() string {
	if len(f.partial) > 0 {
		f.line = append(f.line, []rune(string(f.partial))...)
		f.partial = nil
	}
	text := string(f.line)
	f.line, f.col, f.pendingCR = nil, 0, false
	return text
}

// put applies one visible rune to the current line.
func (f *outputFilter) put(r rune, out *strings.Builder) {
	if f.pendingCR {
		f.pendingCR = false
		if r != '\n' {
			f.col = 0
		}
	}
	switch r {
	case '\n':
		out.WriteString(string(f.line))
		out.WriteByte('\n')
		f.line, f.col = f.line[:0], 0
	case '\r':
		f.pendingCR = true
	case '\b':
		f.col = max(f.col-1, 0)
	default:
		if f.col < len(f.line) {
			f.line[f.col] = r
		} else {
			f.line = append(f.line, r)
		}
		f.col++
	}
}
