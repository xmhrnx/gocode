package tools

// myersDiff finds a shortest edit script with Myers' O((N+M)D) algorithm,
// recording each round's furthest-reaching paths for the backtrack.
func myersDiff(a, b []string) []lineOp {
	n, m := len(a), len(b)
	if n == 0 || m == 0 {
		return replaceOps(a, b)
	}
	offset := n + m
	v := make([]int, 2*offset+2)
	var trace [][]int
	for d := 0; d <= n+m && d <= maxEditDistance; d++ {
		done := false
		for k := -d; k <= d; k += 2 {
			x := furthestStart(v, offset, d, k)
			y := x - k
			for x < n && y < m && a[x] == b[y] {
				x, y = x+1, y+1
			}
			v[offset+k] = x
			done = done || (x >= n && y >= m)
		}
		trace = append(trace, append([]int(nil), v[offset-d:offset+d+1]...))
		if done {
			return backtrack(trace, a, b)
		}
	}
	return replaceOps(a, b)
}

// furthestStart picks where diagonal k's path starts in round d: one step
// down from diagonal k+1 (an insertion) or one step right from k-1 (a
// deletion), whichever reached further. v is indexed by diagonal+offset.
func furthestStart(v []int, offset, d, k int) int {
	at := func(diag int) int { return v[offset+diag] }
	if k == -d || (k != d && at(k-1) < at(k+1)) {
		return at(k + 1)
	}
	return at(k-1) + 1
}

// backtrack walks the trace from the end back to the start, emitting the
// script in reverse and then flipping it.
func backtrack(trace [][]int, a, b []string) []lineOp {
	x, y := len(a), len(b)
	var ops []lineOp
	for d := len(trace) - 1; d > 0; d-- {
		prev := trace[d-1]
		at := func(k int) int { return prev[k+d-1] }
		k := x - y
		inserted := k == -d || (k != d && at(k-1) < at(k+1))
		prevK := k - 1
		if inserted {
			prevK = k + 1
		}
		prevX := at(prevK)
		midX := prevX + 1
		if inserted {
			midX = prevX
		}
		for ; x > midX; x, y = x-1, y-1 {
			ops = append(ops, lineOp{opEqual, a[x-1]})
		}
		if inserted {
			ops = append(ops, lineOp{opInsert, b[y-1]})
		} else {
			ops = append(ops, lineOp{opDelete, a[x-1]})
		}
		x, y = prevX, prevX-prevK
	}
	for ; x > 0; x-- {
		ops = append(ops, lineOp{opEqual, a[x-1]})
	}
	for i, j := 0, len(ops)-1; i < j; i, j = i+1, j-1 {
		ops[i], ops[j] = ops[j], ops[i]
	}
	return ops
}
