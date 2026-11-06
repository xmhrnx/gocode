package workspace

import (
	"testing"

	"github.com/xmhrnx/giode"
)

// stubTab is a minimal tab for workspace tests.
type stubTab struct{ title string }

func (s stubTab) Title() string     { return s.title }
func (stubTab) Icon() string        { return "" }
func (stubTab) Busy() bool          { return false }
func (stubTab) View() giode.Element { return giode.Box(giode.Styles{}) }
func (stubTab) Close()              {}

func titles(w *Workspace) []string {
	var out []string
	for _, e := range w.entries {
		out = append(out, e.tab.Title())
	}
	return out
}

func TestMoveActiveSwapsWithNeighbourAndStopsAtEnds(t *testing.T) {
	w := New(Factory{})
	for _, title := range []string{"a", "b", "c"} {
		w.Add(stubTab{title})
	}
	w.MoveActive(-1) // "c" is active
	if got := titles(w); got[1] != "c" || got[2] != "b" {
		t.Fatalf("after moving left: %v", got)
	}
	w.MoveActive(-1)
	w.MoveActive(-1) // already first
	if got := titles(w); got[0] != "c" {
		t.Fatalf("after moving to the start: %v", got)
	}
	if w.Active().Title() != "c" {
		t.Fatal("the moved tab should stay shown")
	}
}

func TestMoveToShiftsTheTabsBetween(t *testing.T) {
	w := New(Factory{})
	for _, title := range []string{"a", "b", "c", "d"} {
		w.Add(stubTab{title})
	}
	w.MoveTo(w.entries[0].id, 2)
	if got := titles(w); got[0] != "b" || got[1] != "c" || got[2] != "a" || got[3] != "d" {
		t.Fatalf("got %v", got)
	}
}
