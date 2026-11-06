//go:build windows

package tools

import (
	"path/filepath"
	"regexp"
	"strings"
)

// unixDrivePath matches the drive paths of Git Bash and MSYS (/c/Users),
// Cygwin (/cygdrive/c/Users) and WSL (/mnt/c/Users).
var unixDrivePath = regexp.MustCompile(`^/(?:cygdrive/|mnt/)?([a-zA-Z])(/.*)?$`)

// nativePath turns a Unix-style drive path into a Windows one. The agent's
// shell is often Git Bash, which prints paths like /c/Users/me/project;
// Windows would read that as C:\c\Users\me\project.
func nativePath(path string) string {
	match := unixDrivePath.FindStringSubmatch(path)
	if match == nil {
		return path
	}
	return filepath.FromSlash(strings.ToUpper(match[1]) + ":/" + strings.TrimPrefix(match[2], "/"))
}
