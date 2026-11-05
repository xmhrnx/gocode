package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gocode/internal/settings"
)

// Store keeps sessions as one JSON file each in a folder.
type Store struct {
	Dir string
}

// Save writes the session.
func (s Store) Save(sess *Session) error {
	data, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("encode session: %w", err)
	}
	return settings.WriteFileAtomic(s.path(sess.ID), data)
}

// Load reads the session with the given ID.
func (s Store) Load(id string) (*Session, error) {
	data, err := os.ReadFile(s.path(id))
	if err != nil {
		return nil, fmt.Errorf("read session: %w", err)
	}
	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("parse session %s: %w", id, err)
	}
	if sess.ToolDisplays == nil {
		sess.ToolDisplays = map[string]string{}
	}
	return &sess, nil
}

// Delete removes the session.
func (s Store) Delete(id string) error {
	return os.Remove(s.path(id))
}

// List returns summaries of every stored session, newest first. Files
// that fail to parse are skipped, so one corrupt file can't hide the rest.
func (s Store) List() ([]Summary, error) {
	entries, err := os.ReadDir(s.Dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	var summaries []Summary
	for _, entry := range entries {
		id, ok := strings.CutSuffix(entry.Name(), ".json")
		if !ok {
			continue
		}
		if summary, err := s.readSummary(id); err == nil {
			summaries = append(summaries, summary)
		}
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].Updated.After(summaries[j].Updated) })
	return summaries, nil
}

// readSummary decodes only the summary fields, skipping the messages.
func (s Store) readSummary(id string) (Summary, error) {
	data, err := os.ReadFile(s.path(id))
	if err != nil {
		return Summary{}, err
	}
	var summary Summary
	err = json.Unmarshal(data, &summary)
	return summary, err
}

func (s Store) path(id string) string { return filepath.Join(s.Dir, id+".json") }
