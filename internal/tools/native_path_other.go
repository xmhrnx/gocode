//go:build !windows

package tools

// nativePath returns path unchanged: on Unix, /c/Users is a real path.
func nativePath(path string) string { return path }
