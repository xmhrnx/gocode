// Package shell runs the commands typed in the terminal tab as Warp-style
// blocks: each command line runs to completion in its own shell process,
// its output is streamed as plain text, and the working directory carries
// over from one block to the next.
package shell

import (
	"os"
	"sync"
	"time"
)

// Session is a sequence of command blocks sharing a working directory.
// It is safe for concurrent use, though blocks are meant to run one at a
// time.
type Session struct {
	shell string

	mu      sync.Mutex
	workDir string
}

// Result describes a finished block.
type Result struct {
	// ExitCode is the command's exit status; -1 if it was killed.
	ExitCode int
	// Duration is how long the command ran.
	Duration time.Duration
	// WorkDir is the session's working directory after the command.
	WorkDir string
}

// NewSession returns a session that runs commands with shell, starting in
// workDir. An empty shell picks the platform default: $SHELL or /bin/bash
// on Unix, PowerShell on Windows.
func NewSession(shell, workDir string) *Session {
	if shell == "" {
		shell = Default()
	}
	return &Session{shell: shell, workDir: workDir}
}

// WorkDir returns the directory the next command runs in.
func (s *Session) WorkDir() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.workDir
}

// setWorkDir adopts dir if it is still an existing directory; the command
// may have removed it, in which case the old directory is kept.
func (s *Session) setWorkDir(dir string) {
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workDir = dir
}
