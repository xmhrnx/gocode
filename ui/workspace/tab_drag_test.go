package workspace

import "testing"

func TestDragPlacesTabByDistanceMoved(t *testing.T) {
	w := New(Factory{})
	for _, title := range []string{"a", "b", "c"} {
		w.Add(stubTab{title})
	}
	strip := NewTabStrip(w)
	for _, e := range w.entries {
		strip.chip(e.id).width = 100 // at 1 px per dp in tests
	}
	first := w.entries[0].id
	strip.startDrag(first)
	strip.moveDrag(first, 40) // not yet past b's center
	if got := titles(w); got[0] != "a" {
		t.Fatalf("moved too early: %v", got)
	}
	strip.moveDrag(first, 120) // past b's center
	if got := titles(w); got[0] != "b" || got[1] != "a" {
		t.Fatalf("after passing b: %v", got)
	}
	strip.moveDrag(first, 230) // past c's center too
	if got := titles(w); got[2] != "a" {
		t.Fatalf("after passing c: %v", got)
	}
	strip.moveDrag(first, -50) // dragged back before the start
	if got := titles(w); got[0] != "a" {
		t.Fatalf("after dragging back: %v", got)
	}
	strip.endDrag()
}
