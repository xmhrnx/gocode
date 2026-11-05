package shell

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Run runs one command line in the session's working directory. Output
// (stdout and stderr interleaved, escape sequences removed, carriage
// returns applied) is passed to onOutput in chunks as it arrives, on the
// goroutine that called Run. Afterwards the session adopts the shell's
// final working directory, so "cd" persists across blocks. A non-zero
// exit is reported in Result.ExitCode, not as an error; cancelling ctx
// kills the command's whole process group and returns ctx's error.
func (s *Session) Run(ctx context.Context, commandLine string, onOutput func(chunk string)) (Result, error) {
	start := time.Now()
	pwdFile, err := newPwdFile()
	if err != nil {
		return Result{ExitCode: -1, WorkDir: s.WorkDir()}, err
	}
	defer os.Remove(pwdFile)

	cmd := exec.CommandContext(ctx, s.shell, wrapCommand(s.shell, commandLine, pwdFile)...)
	cmd.Dir = s.WorkDir()
	cmd.Env = append(os.Environ(), "TERM=dumb")
	startInOwnGroup(cmd)
	cmd.Cancel = func() error { return killGroup(cmd) }
	waitErr := runStreaming(cmd, onOutput)

	if dir, err := os.ReadFile(pwdFile); err == nil && len(dir) > 0 {
		s.setWorkDir(strings.TrimSpace(string(dir)))
	}
	result := Result{ExitCode: exitCode(cmd, waitErr), Duration: time.Since(start), WorkDir: s.WorkDir()}
	if ctx.Err() != nil {
		return result, ctx.Err()
	}
	var exitErr *exec.ExitError
	if waitErr != nil && !errors.As(waitErr, &exitErr) {
		return result, waitErr
	}
	return result, nil
}

// newPwdFile creates the temp file the shell writes its final directory to.
func newPwdFile() (string, error) {
	f, err := os.CreateTemp("", "gocode-pwd-*")
	if err != nil {
		return "", err
	}
	name := f.Name()
	if err := f.Close(); err != nil {
		os.Remove(name)
		return "", err
	}
	return name, nil
}

// exitCode reads the exit status, or -1 if the process did not exit
// normally (it never started, or was killed).
func exitCode(cmd *exec.Cmd, waitErr error) int {
	if cmd.ProcessState != nil {
		return cmd.ProcessState.ExitCode()
	}
	if waitErr != nil {
		return -1
	}
	return 0
}
