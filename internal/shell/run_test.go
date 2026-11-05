//go:build !windows

package shell

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// runCollect runs a command and returns its result and collected output.
func runCollect(t *testing.T, s *Session, ctx context.Context, line string) (Result, string, error) {
	t.Helper()
	var out strings.Builder
	res, err := s.Run(ctx, line, func(chunk string) { out.WriteString(chunk) })
	return res, out.String(), err
}

func TestRunOutputAndExitCode(t *testing.T) {
	s := NewSession("/bin/sh", t.TempDir())
	res, out, err := runCollect(t, s, context.Background(), "printf '\\033[31mred\\033[0m\\n'; echo err >&2; exit 4")
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 4 || out != "red\nerr\n" || res.Duration <= 0 {
		t.Errorf("got exit %d, output %q, duration %v", res.ExitCode, out, res.Duration)
	}
	res, _, err = runCollect(t, s, context.Background(), "true")
	if err != nil || res.ExitCode != 0 {
		t.Errorf("got exit %d, err %v", res.ExitCode, err)
	}
}

func TestRunCarriageReturns(t *testing.T) {
	s := NewSession("/bin/sh", t.TempDir())
	_, out, err := runCollect(t, s, context.Background(), `printf 'step 1\rstep 2\rstep 3\ndone'`)
	if err != nil || out != "step 3\ndone" {
		t.Errorf("got %q, err %v", out, err)
	}
}

func TestRunCdPersists(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "sub dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	s := NewSession("/bin/sh", root)
	res, _, err := runCollect(t, s, context.Background(), "cd 'sub dir'")
	want := filepath.Join(root, "sub dir")
	if err != nil || res.WorkDir != want || s.WorkDir() != want {
		t.Fatalf("WorkDir %q (session %q), want %q, err %v", res.WorkDir, s.WorkDir(), want, err)
	}
	_, out, _ := runCollect(t, s, context.Background(), "pwd")
	if out != want+"\n" {
		t.Errorf("next block ran in %q", out)
	}
	// A cd followed by exit still counts, and a failed cd changes nothing.
	if res, _, _ := runCollect(t, s, context.Background(), "cd .. && exit 2"); res.WorkDir != root || res.ExitCode != 2 {
		t.Errorf("got %+v", res)
	}
	if res, _, _ := runCollect(t, s, context.Background(), "cd /does/not/exist"); res.WorkDir != root {
		t.Errorf("got %+v", res)
	}
}

func TestRunCancelKillsProcessGroup(t *testing.T) {
	s := NewSession("/bin/sh", t.TempDir())
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	start := time.Now()
	res, out, err := runCollect(t, s, ctx, "echo started; sleep 30 & sleep 30")
	if err == nil || res.ExitCode != -1 {
		t.Errorf("got exit %d, err %v", res.ExitCode, err)
	}
	if out != "started\n" {
		t.Errorf("output %q", out)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("took %v; process group not killed", elapsed)
	}
}

func TestRunBackgroundJobDoesNotBlock(t *testing.T) {
	s := NewSession("/bin/sh", t.TempDir())
	start := time.Now()
	res, out, err := runCollect(t, s, context.Background(), "(sleep 5; echo late) & echo now")
	if err != nil || res.ExitCode != 0 || out != "now\n" {
		t.Errorf("got exit %d, output %q, err %v", res.ExitCode, out, err)
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Errorf("block waited %v for a background job", elapsed)
	}
}

func TestRunSetsTermDumb(t *testing.T) {
	s := NewSession("/bin/sh", t.TempDir())
	if _, out, _ := runCollect(t, s, context.Background(), "echo $TERM"); out != "dumb\n" {
		t.Errorf("TERM = %q", out)
	}
}

func TestPosixWrapperQuotesPwdFile(t *testing.T) {
	dir := t.TempDir()
	pwdFile := filepath.Join(dir, "it's here")
	cmd := exec.Command("/bin/sh", wrapCommand("/bin/sh", "cd /", pwdFile)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	if got, err := os.ReadFile(pwdFile); err != nil || string(got) != "/" {
		t.Errorf("pwd file holds %q, err %v", got, err)
	}
}
