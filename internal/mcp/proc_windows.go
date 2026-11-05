//go:build windows

package mcp

import (
	"errors"
	"os"
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
// child dies. A process that already exited is not an error.
func killGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	tree := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	if tree.Run() == nil {
		return nil
	}
	err := cmd.Process.Kill()
	if errors.Is(err, os.ErrProcessDone) {
		return nil
	}
	return err
}
