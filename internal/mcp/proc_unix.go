//go:build !windows

package mcp

import (
	"errors"
	"os/exec"
	"syscall"
)

// startInOwnGroup makes cmd lead a new process group, so closing the server
// also stops whatever it spawned (npx and uvx launch the real server as a
// child).
func startInOwnGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killGroup kills cmd's whole process group. A group that is already gone
// is not an error.
func killGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}
