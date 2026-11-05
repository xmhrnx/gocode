package shell

import (
	"path/filepath"
	"strings"
)

// wrapCommand returns the arguments that make shell run commandLine and
// then write its final working directory to pwdFile, out of band so the
// user's output stays untouched. PowerShell gets its own syntax; every
// other shell is assumed to be POSIX-compatible.
func wrapCommand(shell, commandLine, pwdFile string) []string {
	if isPowerShell(shell) {
		return []string{"-NoProfile", "-NonInteractive", "-Command", powerShellScript(commandLine, pwdFile)}
	}
	return []string{"-c", posixScript(commandLine, pwdFile)}
}

// posixScript records $PWD from an EXIT trap, so the directory is captured
// even when the command line calls exit, and the exit status is kept.
// ">|" overrides noclobber, since the temp file already exists.
func posixScript(commandLine, pwdFile string) string {
	// Git Bash on Windows reports /c/Users/...; "pwd -W" gives the C:/Users/...
	// form Windows programs understand, and fails harmlessly elsewhere.
	trap := `printf '%s' "$(pwd -W 2>/dev/null || pwd)" >| ` + posixQuote(pwdFile)
	return "trap " + posixQuote(trap) + " EXIT\n" + commandLine
}

// powerShellScript records the location in a finally block and passes on
// a native command's exit code, which PowerShell would otherwise drop.
func powerShellScript(commandLine, pwdFile string) string {
	return "try {\n" + commandLine + "\n} finally { [IO.File]::WriteAllText(" + powerShellQuote(pwdFile) +
		", (Get-Location).ProviderPath) }\nif ($LASTEXITCODE) { exit $LASTEXITCODE }"
}

// isPowerShell reports whether shell is Windows PowerShell or pwsh.
func isPowerShell(shell string) bool {
	name := strings.ToLower(strings.TrimSuffix(filepath.Base(shell), filepath.Ext(shell)))
	return name == "powershell" || name == "pwsh"
}

// posixQuote single-quotes s for a POSIX shell.
func posixQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// powerShellQuote single-quotes s for PowerShell, which escapes a quote by
// doubling it.
func powerShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
