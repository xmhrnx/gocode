package proc

import (
	"context"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func TestKillGroupStopsChildren(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", "sleep 30 & wait")
	StartGroup(cmd)
	cmd.Cancel = func() error { return KillGroup(cmd) }
	cmd.WaitDelay = time.Second
	start := time.Now()
	_ = cmd.Run()
	if time.Since(start) > 3*time.Second {
		t.Fatal("the group outlived its cancellation")
	}
}
