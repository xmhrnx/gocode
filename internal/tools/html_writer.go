package tools

import (
	"regexp"
	"strings"

	"golang.org/x/net/html/atom"
)

// textWriter builds the plain text of a page. It tracks whether the output
// ends at a line start or a space so collapsed whitespace never doubles up
// or indents lines.
type textWriter struct {
	out        strings.Builder
	newlines   int // line breaks at the end of the output
	afterSpace bool
}

var (
	whitespaceRun = regexp.MustCompile(`\s+`)
	blankLineRun  = regexp.MustCompile(`\n{3,}`)
)

// openTag starts an element: blocks begin a line, headings and list items
// get Markdown-style markers, table cells a separating space.
func (w *textWriter) openTag(tag atom.Atom) {
	w.breakLines(blockElements[tag])
	switch {
	case tag == atom.Br:
		w.write("\n")
	case headingLevels[tag] > 0:
		w.write(strings.Repeat("#", headingLevels[tag]) + " ")
	case tag == atom.Li:
		w.write("- ")
	case tag == atom.Td || tag == atom.Th:
		w.text(" ", false)
	}
}

// closeTag ends an element; blocks end their line.
func (w *textWriter) closeTag(tag atom.Atom) {
	w.breakLines(blockElements[tag])
}

// breakLines ensures the output ends with at least n line breaks.
func (w *textWriter) breakLines(n int) {
	for w.newlines < n {
		w.write("\n")
	}
}

// text writes a text node, collapsing whitespace unless preformatted.
func (w *textWriter) text(s string, pre bool) {
	if !pre {
		s = whitespaceRun.ReplaceAllString(s, " ")
		if w.newlines > 0 || w.afterSpace {
			s = strings.TrimPrefix(s, " ")
		}
	}
	w.write(s)
}

func (w *textWriter) write(s string) {
	if s == "" {
		return
	}
	w.out.WriteString(s)
	trimmed := strings.TrimRight(s, "\n")
	if trimmed == "" {
		w.newlines += len(s)
	} else {
		w.newlines = len(s) - len(trimmed)
	}
	w.afterSpace = strings.HasSuffix(s, " ")
}

// finish trims line ends and squeezes runs of blank lines into one.
func (w *textWriter) finish() string {
	lines := strings.Split(w.out.String(), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	text := blankLineRun.ReplaceAllString(strings.Join(lines, "\n"), "\n\n")
	return strings.TrimSpace(text)
}
