package settings

import (
	"path/filepath"
	"testing"
)

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	got, err := Load(filepath.Join(t.TempDir(), "none.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Chat.SendKey != SendEnter || got.Agent.CompactAtPercent == 0 {
		t.Fatalf("defaults not applied: %+v", got.Chat)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "s.json")
	s := Default()
	s.Chat.SendKey = SendCtrlEnter
	s.Models.MarkUsed("openai/gpt-5")
	if err := Save(path, s); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Chat.SendKey != SendCtrlEnter || got.Models.Recent[0] != "openai/gpt-5" {
		t.Fatalf("round trip lost data: %+v %+v", got.Chat, got.Models)
	}
}

func TestMarkUsedMovesToFrontWithoutDuplicates(t *testing.T) {
	var m Models
	for _, id := range []string{"a", "b", "a"} {
		m.MarkUsed(id)
	}
	if len(m.Recent) != 2 || m.Recent[0] != "a" || m.Recent[1] != "b" {
		t.Fatalf("recent = %v", m.Recent)
	}
}

func TestSetHidden(t *testing.T) {
	var m Models
	m.SetHidden("x", true)
	m.SetHidden("x", true)
	if !m.IsHidden("x") || len(m.Hidden) != 1 {
		t.Fatalf("hidden = %v", m.Hidden)
	}
	m.SetHidden("x", false)
	if m.IsHidden("x") {
		t.Fatal("still hidden")
	}
}

func TestMigrateDropsOldDefaultLimits(t *testing.T) {
	s := Settings{Agent: Agent{MaxSteps: 60, MaxOutputTokens: 16000}}
	s.migrate()
	if s.Agent.MaxSteps != 0 || s.Agent.MaxOutputTokens != 0 || !s.Agent.AutoCompact || s.Version != currentVersion {
		t.Fatalf("agent = %+v", s.Agent)
	}
	kept := Settings{Agent: Agent{MaxSteps: 200, MaxOutputTokens: 8000}}
	kept.migrate()
	if kept.Agent.MaxSteps != 200 || kept.Agent.MaxOutputTokens != 8000 {
		t.Fatalf("user values lost: %+v", kept.Agent)
	}
}
