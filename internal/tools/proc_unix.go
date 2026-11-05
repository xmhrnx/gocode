//go:build !windows

package tools

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

// startInOwnGroup makes cmd lead a new process group, so a timeout can kill
// everything it spawned, not just the shell.
func startInOwnGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killGroup kills cmd's whole process group.
func killGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return os.ErrProcessDone
	}
	return err
}
