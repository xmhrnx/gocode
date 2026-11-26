package browser

import (
	"errors"
	"syscall"
)

// wsaeaddrinuse is Winsock's "address already in use".
const wsaeaddrinuse = syscall.Errno(10048)

// isAddrInUse reports whether listening failed because the port is taken.
func isAddrInUse(err error) bool {
	return errors.Is(err, wsaeaddrinuse) || errors.Is(err, syscall.EADDRINUSE)
}
