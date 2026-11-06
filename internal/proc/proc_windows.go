//go:build windows

package proc

import (
	"os/exec"
	"strconv"
	"syscall"
)

// createNoWindow keeps a console program (Git Bash, PowerShell, cmd, git)
// started from GoCode, a windowed app, from opening a console window of
// its own, which would also take the focus.
const createNoWindow = 0x08000000

// StartGroup starts cmd in a new process group, without a console window,
// so KillGroup can stop it without affecting GoCode.
func StartGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | createNoWindow,
		HideWindow:    true,
	}
}

// Hide starts cmd without a console window.
func Hide(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNoWindow, HideWindow: true}
}

// KillGroup kills cmd and its descendants. Windows has no process-group
// signal, so taskkill walks the tree; if it fails, at least the direct
// child dies.
func KillGroup(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	tree := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	Hide(tree)
	if err := tree.Run(); err != nil {
		return cmd.Process.Kill()
	}
	return nil
}
