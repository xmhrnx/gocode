package tools

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// treeLister accumulates an indented directory tree.
type treeLister struct {
	ctx      context.Context
	rules    *ignoreRules
	maxDepth int
	lines    []string
	entries  int
	capped   bool
}

// list appends dir's children at the given level, recursing into
// subdirectories until maxDepth.
func (t *treeLister) list(dir string, level int) {
	indent := strings.Repeat("  ", level)
	children, err := os.ReadDir(dir)
	if err != nil {
		t.lines = append(t.lines, indent+"[unreadable: "+err.Error()+"]")
		return
	}
	sort.SliceStable(children, func(i, j int) bool {
		if children[i].IsDir() != children[j].IsDir() {
			return children[i].IsDir()
		}
		return children[i].Name() < children[j].Name()
	})
	for _, child := range children {
		if t.ctx.Err() != nil || t.capped {
			return
		}
		path := filepath.Join(dir, child.Name())
		if t.rules.skip(path, child.IsDir()) {
			continue
		}
		if t.entries == maxListEntries {
			t.capped = true
			return
		}
		t.entries++
		if !child.IsDir() {
			t.lines = append(t.lines, indent+child.Name())
			continue
		}
		t.lines = append(t.lines, indent+child.Name()+"/")
		if level < t.maxDepth {
			t.list(path, level+1)
		}
	}
}
