//go:build !windows

package shell

// ScriptArgs returns the arguments that make program run script. On Unix
// the script travels as an argument: execve passes arguments through
// byte for byte. cleanup must be called once the program has exited.
func ScriptArgs(program, script string) (args []string, cleanup func(), err error) {
	if KindOf(program) == KindPowerShell {
		return []string{"-NoProfile", "-NonInteractive", "-Command", script}, func() {}, nil
	}
	return []string{"-c", script}, func() {}, nil
}
