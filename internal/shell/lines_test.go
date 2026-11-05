package shell

import "testing"

// filterChunks feeds chunks through a filter and returns everything it
// emitted, including the final flush.
func filterChunks(chunks ...string) string {
	var f outputFilter
	var out string
	for _, chunk := range chunks {
		out += f.write([]byte(chunk))
	}
	return out + f.flush()
}

func TestOutputFilter(t *testing.T) {
	tests := []struct {
		name   string
		chunks []string
		want   string
	}{
		{"plain lines", []string{"a\nb\n"}, "a\nb\n"},
		{"progress bar", []string{"10%\r50%\r100%\n"}, "100%\n"},
		{"overwrite keeps tail", []string{"abcdef\rXY\n"}, "XYcdef\n"},
		{"crlf is a newline", []string{"one\r\ntwo\r\n"}, "one\ntwo\n"},
		{"crlf split across chunks", []string{"one\r", "\ntwo"}, "one\ntwo"},
		{"cr split across chunks", []string{"old\r", "new\n"}, "new\n"},
		{"backspace", []string{"ab\bc\n"}, "ac\n"},
		{"escape split across chunks", []string{"\x1b[3", "1mred\x1b[0m\n"}, "red\n"},
		{"utf8 split across chunks", []string{"h\xc3", "\xa9llo\n"}, "héllo\n"},
		{"unfinished line flushed", []string{"no newline"}, "no newline"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterChunks(tt.chunks...); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOutputFilterHoldsPartialLine(t *testing.T) {
	var f outputFilter
	if got := f.write([]byte("done\nloading 5%")); got != "done\n" {
		t.Errorf("got %q; the partial line may still be overwritten", got)
	}
	if got := f.write([]byte("\rloading 100%\n")); got != "loading 100%\n" {
		t.Errorf("got %q", got)
	}
}
