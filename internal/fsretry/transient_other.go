//go:build !windows

package fsretry

// transient reports whether err may clear by itself. Unix doesn't lock
// files against writes or renames, so no error does.
func transient(error) bool { return false }
