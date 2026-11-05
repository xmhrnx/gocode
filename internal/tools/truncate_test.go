package tools

import "testing"

func TestTruncateHead(t *testing.T) {
	if got := truncateHead("short", 10); got != "short" {
		t.Errorf("got %q", got)
	}
	if got := truncateHead("0123456789", 4); got != "0123\n[truncated: 6 more bytes]" {
		t.Errorf("got %q", got)
	}
	// "é" is two bytes; the cut must not split it.
	if got := truncateHead("aé", 2); got != "a\n[truncated: 2 more bytes]" {
		t.Errorf("got %q", got)
	}
}

func TestHeadTailBuffer(t *testing.T) {
	b := newHeadTailBuffer(10)
	for _, chunk := range []string{"start", "-middle-", "middle-", "the end"} {
		if _, err := b.Write([]byte(chunk)); err != nil {
			t.Fatal(err)
		}
	}
	want := "start\n[truncated: 17 more bytes]\ne end"
	if got := b.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	small := newHeadTailBuffer(100)
	_, _ = small.Write([]byte("all of it"))
	if got := small.String(); got != "all of it" {
		t.Errorf("got %q", got)
	}
}
