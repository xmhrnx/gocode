package tools

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

// grepSearch accumulates matches across files.
type grepSearch struct {
	re          *regexp.Regexp
	include     *regexp.Regexp
	includeBase bool // include has no "/", so it matches file names
	context     int
	filesOnly   bool
	out         strings.Builder
	matches     int
	capped      bool
}

const (
	maxGrepMatches   = 300
	maxGrepLineBytes = 500
	// maxGrepFileBytes skips huge files, which are almost always generated
	// data rather than source worth searching.
	maxGrepFileBytes = 20 << 20
)

// wants reports whether the include filter admits a file, given its path
// relative to the search base.
func (g *grepSearch) wants(rel string) bool {
	if g.include == nil {
		return true
	}
	if g.includeBase {
		rel = path.Base(rel)
	}
	return g.include.MatchString(rel)
}

// searchFile records the matches in one file. Binary, oversized and
// unreadable files are skipped silently, as grep does.
func (g *grepSearch) searchFile(abs, display string) {
	info, err := os.Stat(abs)
	if err != nil || info.Size() > maxGrepFileBytes {
		return
	}
	data, err := os.ReadFile(abs)
	if err != nil || looksBinary(data) {
		return
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	hits := g.matchingLines(lines)
	if len(hits) == 0 {
		return
	}
	if g.filesOnly {
		g.out.WriteString(display + "\n")
		g.matches++
		g.capped = g.matches >= maxGrepMatches
		return
	}
	g.writeHits(display, lines, hits)
}

// matchingLines returns the indexes of matching lines, stopping at the
// overall match cap.
func (g *grepSearch) matchingLines(lines []string) []int {
	var hits []int
	for i, line := range lines {
		if !g.re.MatchString(line) {
			continue
		}
		if g.matches+len(hits) == maxGrepMatches {
			g.capped = true
			break
		}
		hits = append(hits, i)
	}
	return hits
}

// writeHits prints matches as "path:line: text" and context lines as
// "path-line- text", with "--" between groups that are not adjacent.
func (g *grepSearch) writeHits(display string, lines []string, hits []int) {
	isHit := make(map[int]bool, len(hits))
	for _, h := range hits {
		isHit[h] = true
	}
	next := -1 // first line not yet printed
	for _, h := range hits {
		start := max(h-g.context, next, 0)
		if g.context > 0 && g.out.Len() > 0 && start != next {
			g.out.WriteString("--\n")
		}
		end := min(h+g.context, len(lines)-1)
		for i := start; i <= end; i++ {
			sep := "-"
			if isHit[i] {
				sep = ":"
			}
			fmt.Fprintf(&g.out, "%s%s%d%s %s\n", display, sep, i+1, sep, truncateLine(strings.TrimSuffix(lines[i], "\r"), maxGrepLineBytes))
		}
		next = end + 1
	}
	g.matches += len(hits)
}

// result returns the formatted matches with a note when capped.
func (g *grepSearch) result() string {
	if g.out.Len() == 0 {
		return "No matches found."
	}
	text := strings.TrimSuffix(g.out.String(), "\n")
	if g.capped {
		text += fmt.Sprintf("\n[stopped after %d matches; narrow the pattern, path or include filter]", maxGrepMatches)
	}
	return text
}
