//go:build windows

package tools

import (
	"os/exec"
	"strconv"
	"syscall"
)

// startInOwnGroup starts cmd in a new process group so it can be killed
// without affecting GoCode.
func startInOwnGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}

// killGroup kills cmd and its descendants. Windows has no process-group
// signal, so taskkill walks the tree; if it fails, at least the direct
// child dies.
func killGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	tree := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	if err := tree.Run(); err != nil {
		return cmd.Process.Kill()
	}
	return nil
}
