package terminal

import "testing"

func TestHistoryStep(t *testing.T) {
	var h commandHistory
	h.add("ls")
	h.add("go test")
	if got, _ := h.step(true); got != "go test" {
		t.Fatalf("first up = %q", got)
	}
	if got, _ := h.step(true); got != "ls" {
		t.Fatalf("second up = %q", got)
	}
	if _, ok := h.step(true); ok {
		t.Fatal("stepping past the oldest should do nothing")
	}
	h.step(false)
	if got, ok := h.step(false); !ok || got != "" {
		t.Fatalf("down past newest = %q, %v", got, ok)
	}
}
