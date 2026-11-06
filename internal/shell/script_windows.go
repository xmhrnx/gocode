//go:build windows

package shell

import (
	"os"
	"path/filepath"
)

// utf8BOM makes Windows PowerShell 5.1 read a script as UTF-8 rather than
// the legacy code page.
const utf8BOM = "\uFEFF"

// ScriptArgs returns the arguments that make program run script. On
// Windows a program receives one command line string that it splits by
// its own rules, and Git Bash's rules differ from the quoting Go applies,
// so quotes, backslashes and tabs in a script passed with -c can arrive
// changed. The script is therefore written to a file whose path is the
// only argument. cleanup removes the file once the program has exited.
func ScriptArgs(program, script string) (args []string, cleanup func(), err error) {
	extension, content := ".sh", script
	switch KindOf(program) {
	case KindPowerShell:
		extension, content = ".ps1", utf8BOM+script
	case KindCmd:
		extension, content = ".cmd", "@echo off\r\n"+script
	}
	path, err := writeScript(extension, content)
	if err != nil {
		return nil, nil, err
	}
	cleanup = func() { os.Remove(path) }
	switch KindOf(program) {
	case KindPowerShell:
		return []string{"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", path}, cleanup, nil
	case KindCmd:
		return []string{"/D", "/C", path}, cleanup, nil
	}
	// Git Bash understands C:/... paths; backslashes would be escapes.
	return []string{filepath.ToSlash(path)}, cleanup, nil
}

func writeScript(extension, content string) (string, error) {
	f, err := os.CreateTemp("", "gocode-*"+extension)
	if err != nil {
		return "", err
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), f.Close()
}
