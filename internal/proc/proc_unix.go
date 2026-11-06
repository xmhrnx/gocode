//go:build !windows

package proc

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

// StartGroup makes cmd lead a new process group, so KillGroup stops
// everything it spawned.
func StartGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// Hide does nothing on Unix, where a program started from a window has
// no window of its own.
func Hide(*exec.Cmd) {}

// KillGroup kills cmd's whole process group. It returns os.ErrProcessDone
// when the group is already gone, as exec.Cmd.Cancel expects.
func KillGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return os.ErrProcessDone
	}
	return err
}
