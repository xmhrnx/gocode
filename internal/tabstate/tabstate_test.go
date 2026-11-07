package tabstate

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tabs.json")
	state := State{Tabs: []Tab{{Kind: KindChat, SessionID: "s1"}, {Kind: KindSettings}}, Active: 1}
	if err := Save(path, state); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil || !reflect.DeepEqual(got, state) {
		t.Fatalf("got %+v, %v", got, err)
	}
	if empty, err := Load(filepath.Join(t.TempDir(), "none.json")); err != nil || len(empty.Tabs) != 0 {
		t.Fatalf("missing file: %+v, %v", empty, err)
	}
}
