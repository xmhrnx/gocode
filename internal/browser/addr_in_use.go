//go:build !windows

package browser

import (
	"errors"
	"syscall"
)

// isAddrInUse reports whether listening failed because the port is taken.
func isAddrInUse(err error) bool {
	return errors.Is(err, syscall.EADDRINUSE)
}
