package tools

import (
	"fmt"
	"strings"
)

// hunk is a run of ops [start, end) plus how many lines of each file come
// before it.
type hunk struct {
	start, end           int
	oldBefore, newBefore int
}

// diffContextLines is how many unchanged lines surround each change.
const diffContextLines = 3

// unifiedDiff renders the change from oldText to newText in unified
// format. oldName and newName are the header labels, e.g. "a/main.go" or
// "/dev/null". It returns "" when the texts are equal.
func unifiedDiff(oldName, newName, oldText, newText string) string {
	ops := diffLines(splitLines(oldText), splitLines(newText))
	hunks := groupHunks(ops, diffContextLines)
	if len(hunks) == 0 {
		return ""
	}
	var out strings.Builder
	fmt.Fprintf(&out, "--- %s\n+++ %s\n", oldName, newName)
	for _, h := range hunks {
		h.write(&out, ops)
	}
	return out.String()
}

// splitLines splits text into lines that keep their "\n", so a missing
// final newline shows up as a difference.
func splitLines(text string) []string {
	lines := strings.SplitAfter(text, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// groupHunks gathers changes into hunks, merging changes separated by no
// more than 2*context unchanged lines so their contexts never overlap.
func groupHunks(ops []lineOp, context int) []hunk {
	var hunks []hunk
	oldLine, newLine, prevEnd := 0, 0, 0
	for i := 0; i < len(ops); i++ {
		if ops[i].kind == opEqual {
			oldLine, newLine = oldLine+1, newLine+1
			continue
		}
		back := min(context, i-prevEnd)
		h := hunk{start: i - back, oldBefore: oldLine - back, newBefore: newLine - back}
		h.end = min(lastChangeOfGroup(ops, i, context)+1+context, len(ops))
		for _, op := range ops[i:h.end] {
			oldLine, newLine = advance(op, oldLine, newLine)
		}
		hunks = append(hunks, h)
		prevEnd = h.end
		i = h.end - 1
	}
	return hunks
}

// lastChangeOfGroup returns the index of the last change that belongs in
// the same hunk as the change at first.
func lastChangeOfGroup(ops []lineOp, first, context int) int {
	last := first
	for j := first + 1; j < len(ops) && j-last-1 <= 2*context; j++ {
		if ops[j].kind != opEqual {
			last = j
		}
	}
	return last
}

// advance moves the old and new line counters past op.
func advance(op lineOp, oldLine, newLine int) (int, int) {
	switch op.kind {
	case opDelete:
		return oldLine + 1, newLine
	case opInsert:
		return oldLine, newLine + 1
	default:
		return oldLine + 1, newLine + 1
	}
}

// write renders the hunk's header and lines.
func (h hunk) write(out *strings.Builder, ops []lineOp) {
	oldCount, newCount := 0, 0
	for _, op := range ops[h.start:h.end] {
		oldCount, newCount = advance(op, oldCount, newCount)
	}
	fmt.Fprintf(out, "@@ -%s +%s @@\n", hunkRange(h.oldBefore, oldCount), hunkRange(h.newBefore, newCount))
	for _, op := range ops[h.start:h.end] {
		out.WriteString(opPrefix[op.kind])
		out.WriteString(op.text)
		if !strings.HasSuffix(op.text, "\n") {
			out.WriteString("\n\\ No newline at end of file\n")
		}
	}
}

// opPrefix is the unified-diff marker for each kind of line.
var opPrefix = map[opKind]string{opEqual: " ", opDelete: "-", opInsert: "+"}

// hunkRange formats "start,count"; an empty range names the line before it,
// as diff(1) does.
func hunkRange(before, count int) string {
	if count == 0 {
		return fmt.Sprintf("%d,0", before)
	}
	return fmt.Sprintf("%d,%d", before+1, count)
}
