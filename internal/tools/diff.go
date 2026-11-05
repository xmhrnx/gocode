package tools

// opKind says what a diff operation does to a line.
type opKind int

// lineOp is one line of an edit script turning the old text into the new.
type lineOp struct {
	kind opKind
	text string
}

const (
	opEqual opKind = iota
	opDelete
	opInsert
)

// maxEditDistance bounds the Myers search, whose trace grows with the
// square of the number of differences. Beyond it the differing middle is
// reported as wholly replaced, which is still a correct, if larger, diff.
const maxEditDistance = 1000

// diffLines returns an edit script turning a into b. Common leading and
// trailing lines are matched first, since edits are usually local and
// this keeps the Myers search small.
func diffLines(a, b []string) []lineOp {
	prefix := 0
	for prefix < len(a) && prefix < len(b) && a[prefix] == b[prefix] {
		prefix++
	}
	suffix := 0
	for suffix < len(a)-prefix && suffix < len(b)-prefix &&
		a[len(a)-1-suffix] == b[len(b)-1-suffix] {
		suffix++
	}
	ops := equalOps(a[:prefix])
	ops = append(ops, myersDiff(a[prefix:len(a)-suffix], b[prefix:len(b)-suffix])...)
	return append(ops, equalOps(a[len(a)-suffix:])...)
}

// equalOps marks every line as unchanged.
func equalOps(lines []string) []lineOp {
	ops := make([]lineOp, len(lines))
	for i, line := range lines {
		ops[i] = lineOp{opEqual, line}
	}
	return ops
}

// replaceOps deletes all of a and inserts all of b.
func replaceOps(a, b []string) []lineOp {
	ops := make([]lineOp, 0, len(a)+len(b))
	for _, line := range a {
		ops = append(ops, lineOp{opDelete, line})
	}
	for _, line := range b {
		ops = append(ops, lineOp{opInsert, line})
	}
	return ops
}
