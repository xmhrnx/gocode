// Package session holds conversations and stores them on disk.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gocode/internal/llm"
)

// Session is one conversation with the agent.
type Session struct {
	ID       string        `json:"id"`
	Title    string        `json:"title"`
	WorkDir  string        `json:"work_dir"`
	Model    string        `json:"model"` // "provider/model"
	Created  time.Time     `json:"created"`
	Updated  time.Time     `json:"updated"`
	Messages []llm.Message `json:"messages"`
	// ToolDisplays holds what the UI shows for a tool call (a diff),
	// by call ID, when it differs from what the model saw.
	ToolDisplays map[string]string `json:"tool_displays,omitempty"`
	Usage        llm.Usage         `json:"usage"`
	// ContextStart and ContextSummary record the last compaction: the
	// model sees the summary and the messages from ContextStart on, while
	// the transcript keeps everything.
	ContextStart   int    `json:"context_start,omitempty"`
	ContextSummary string `json:"context_summary,omitempty"`
}

// Summary is what a session list shows.
type Summary struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	WorkDir string    `json:"work_dir"`
	Updated time.Time `json:"updated"`
}

// New returns an empty session.
func New(workDir, model string) *Session {
	now := time.Now()
	return &Session{
		ID: newID(now), WorkDir: workDir, Model: model,
		Created: now, Updated: now, ToolDisplays: map[string]string{},
	}
}

// Summary returns the session's list entry.
func (s *Session) Summary() Summary {
	return Summary{ID: s.ID, Title: s.Title, WorkDir: s.WorkDir, Updated: s.Updated}
}

// newID returns a sortable unique ID: the time, then random bytes.
func newID(now time.Time) string {
	random := make([]byte, 4)
	_, _ = rand.Read(random) // crypto/rand never fails on supported platforms
	return now.UTC().Format("20060102-150405") + "-" + hex.EncodeToString(random)
}
