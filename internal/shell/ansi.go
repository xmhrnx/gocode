package shell

import "strings"

// escapeState is where an ansiStripper is within an escape sequence.
type escapeState int

// ansiStripper removes terminal escape sequences from a stream of runes.
// It keeps state between calls, so sequences split across output chunks
// are still removed.
type ansiStripper struct {
	state escapeState
}

const (
	stateText      escapeState = iota
	stateEscape                // after ESC
	stateCSI                   // ESC [ … final byte
	stateString                // OSC, DCS, APC, PM or SOS, ended by BEL or ESC \
	stateStringEsc             // ESC seen inside a string; "\" ends it
	stateCharset               // ESC ( and friends take one more character
)

// keep reports whether r is visible text, consuming it otherwise.
func (s *ansiStripper) keep(r rune) bool {
	switch s.state {
	case stateEscape:
		s.state = afterEscape(r)
		return false
	case stateCSI:
		// Parameter and intermediate bytes run until a final byte.
		if r >= 0x40 && r <= 0x7e {
			s.state = stateText
		}
		return false
	case stateString:
		switch r {
		case '\a':
			s.state = stateText
		case 0x1b:
			s.state = stateStringEsc
		}
		return false
	case stateStringEsc:
		s.state = stateString
		if r == '\\' {
			s.state = stateText
		}
		return false
	case stateCharset:
		s.state = stateText
		return false
	}
	if r == 0x1b {
		s.state = stateEscape
		return false
	}
	return !isHiddenControl(r)
}

// afterEscape picks the state for the character that follows ESC.
func afterEscape(r rune) escapeState {
	switch r {
	case '[':
		return stateCSI
	case ']', 'P', '_', '^', 'X':
		return stateString
	case '(', ')', '*', '+', '#', '%':
		return stateCharset
	}
	// Any other character completes a two-character sequence like ESC =.
	return stateText
}

// isHiddenControl reports control characters that are not text and are
// not handled by the line logic (\n, \r, \t and backspace are kept).
func isHiddenControl(r rune) bool {
	switch r {
	case '\n', '\r', '\t', '\b':
		return false
	}
	return r < 0x20 || r == 0x7f || (r >= 0x80 && r < 0xa0)
}

// StripANSI removes ANSI escape sequences (CSI such as colors and cursor
// movement, OSC such as window titles and hyperlinks, and two-character
// escapes) and other invisible control characters from s. Newlines,
// carriage returns, tabs and backspaces are kept.
func StripANSI(s string) string {
	var stripper ansiStripper
	var out strings.Builder
	for _, r := range s {
		if stripper.keep(r) {
			out.WriteRune(r)
		}
	}
	return out.String()
}
