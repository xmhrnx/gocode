//go:build windows

package fsretry

import (
	"errors"
	"syscall"
)

// Windows error codes for a file another program holds.
const (
	errorAccessDenied     syscall.Errno = 5
	errorSharingViolation syscall.Errno = 32
	errorLockViolation    syscall.Errno = 33
)

// transient reports whether err may clear once another program lets go
// of the file. "Access is denied" is included because Windows also
// reports a file pending deletion or being scanned that way.
func transient(err error) bool {
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return false
	}
	return errno == errorAccessDenied || errno == errorSharingViolation || errno == errorLockViolation
}
