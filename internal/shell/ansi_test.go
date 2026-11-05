package shell

import "testing"

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{"plain", "hello\tworld\n", "hello\tworld\n"},
		{"sgr colors", "\x1b[1;31merror\x1b[0m: bad", "error: bad"},
		{"cursor movement", "a\x1b[2Kb\x1b[10;20Hc\x1b[?25l", "abc"},
		{"osc title bel", "\x1b]0;my title\x07prompt", "prompt"},
		{"osc hyperlink st", "\x1b]8;;http://x.y\x1b\\link\x1b]8;;\x1b\\", "link"},
		{"two-char escape", "\x1b=keypad\x1b>", "keypad"},
		{"charset", "\x1b(Btext", "text"},
		{"other controls", "bell\x07 nul\x00 del\x7f", "bell nul del"},
		{"keeps cr and backspace", "a\rb\bc", "a\rb\bc"},
		{"unicode", "\x1b[32m✓ héllo\x1b[m", "✓ héllo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripANSI(tt.in); got != tt.want {
				t.Errorf("StripANSI(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
