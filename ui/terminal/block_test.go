package terminal

import (
	"strings"
	"testing"
)

func TestAppendOutputKeepsTheLastLines(t *testing.T) {
	var b block
	for i := 0; i < 100; i++ {
		b.appendOutput("line\n", 10)
	}
	if got := strings.Count(b.output.String(), "\n"); got > 13 {
		t.Fatalf("kept %d lines, want about 10", got)
	}
	if b.dropped == 0 {
		t.Fatal("dropped lines not counted")
	}
}
