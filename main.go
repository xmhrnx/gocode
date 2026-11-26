// Command gocode is a native, lightweight coding agent. On a desktop it
// is the agent; on a phone (desktop_app.go and mobile_app.go choose by
// build target) it follows and steers the agent on a paired desktop.
package main

import (
	"log"
	"os"
	"runtime/debug"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// tuneMemory makes the garbage collector run when the heap grows by half
// instead of doubling. A streaming chat makes a lot of short-lived garbage
// (each frame lays out new text), and at Go's default the heap holds up to
// twice what is live; the extra collections cost little CPU. GOGC, when
// set, still wins.
func tuneMemory() {
	if os.Getenv("GOGC") == "" {
		debug.SetGCPercent(50)
	}
}
