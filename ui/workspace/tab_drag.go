package workspace

import "github.com/xmhrnx/giode"

// tabDrag is a tab being dragged to a new place. Positions come from the
// tabs' widths when the drag started and from how far the pointer has
// moved since, so nothing depends on where the strip sits in the window.
type tabDrag struct {
	id    int
	start int
	// centers holds each tab's center, in dp from the strip's start, in
	// the order the tabs had when the drag started.
	centers []float32
}

// tabGapDp is the space between tabs in the strip.
const tabGapDp = 2

// startDrag records the tabs' layout as the drag of tab id begins.
func (s *TabStrip) startDrag(id int) {
	d := &tabDrag{id: id, start: s.workspace.indexOf(id)}
	var x float32
	for _, e := range s.workspace.entries {
		width := giode.PxToDp(s.chip(e.id).width)
		d.centers = append(d.centers, x+width/2)
		x += width + tabGapDp
	}
	s.drag = d
	s.workspace.activate(id)
}

// moveDrag places the dragged tab where its center now falls among the
// other tabs, deltaDp being how far the pointer moved since the press.
func (s *TabStrip) moveDrag(id int, deltaDp float32) {
	d := s.drag
	if d == nil || d.id != id || d.start < 0 {
		return
	}
	center := d.centers[d.start] + deltaDp
	target := 0
	for index, other := range d.centers {
		if index != d.start && other < center {
			target++
		}
	}
	s.workspace.MoveTo(id, target)
}

// endDrag forgets the drag.
func (s *TabStrip) endDrag() { s.drag = nil }
