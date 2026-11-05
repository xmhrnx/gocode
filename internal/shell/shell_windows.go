//go:build windows

package shell

import (
	"os/exec"
	"strconv"
	"syscall"
)

// defaultShell returns Windows PowerShell, which every Windows has. $SHELL
// is ignored because, when set at all, it usually names an MSYS path that
// cannot be started directly.
func defaultShell() string { return "powershell.exe" }

// startInOwnGroup starts cmd in a new process group so cancelling it does
// not affect GoCode.
func startInOwnGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}

// killGroup kills cmd and its descendants with taskkill, falling back to
// killing just the direct child.
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
