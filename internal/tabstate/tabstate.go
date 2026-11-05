// Package tabstate remembers which tabs were open, so the next start can
// reopen them.
package tabstate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"gocode/internal/settings"
)

// Kind is what a tab shows.
type Kind string

// Tab is what is needed to reopen one tab.
type Tab struct {
	Kind Kind `json:"kind"`
	// SessionID names a chat's stored session; empty for a chat with no
	// messages yet.
	SessionID string `json:"session_id,omitempty"`
	WorkDir   string `json:"work_dir,omitempty"`
}

// State is the open tabs, in order, and which one was shown.
type State struct {
	Tabs   []Tab `json:"tabs"`
	Active int   `json:"active"`
}

const (
	KindChat     Kind = "chat"
	KindTerminal Kind = "terminal"
	KindSettings Kind = "settings"
)

// Load reads the state; a missing file is an empty state.
func Load(path string) (State, error) {
	var state State
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return state, fmt.Errorf("read open tabs: %w", err)
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, fmt.Errorf("parse open tabs: %w", err)
	}
	return state, nil
}

// Save writes the state.
func Save(path string, state State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encode open tabs: %w", err)
	}
	return settings.WriteFileAtomic(path, data)
}
