package skills

import (
	"strconv"
	"strings"
)

// scalarValue decodes a value written after "key:" and its continuation.
func scalarValue(raw string, continuation []string) string {
	switch {
	case strings.HasPrefix(raw, ">"):
		return strings.TrimSpace(foldLines(continuation))
	case strings.HasPrefix(raw, "|"):
		return literalLines(continuation)
	case raw == "":
		return "" // a nested map or list, which no field we read uses
	}
	text := foldLines(append([]string{raw}, continuation...))
	if unquoted, ok := unquote(text); ok {
		return unquoted
	}
	plain, _, _ := strings.Cut(text, " #") // " #" starts a YAML comment
	return strings.TrimSpace(plain)
}

// foldLines joins lines with spaces; a blank line becomes a line break,
// as YAML folding does.
func foldLines(lines []string) string {
	var b strings.Builder
	for _, line := range lines {
		text := strings.TrimSpace(line)
		switch {
		case text == "":
			b.WriteString("\n")
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "\n"):
			b.WriteString(" " + text)
		default:
			b.WriteString(text)
		}
	}
	return b.String()
}

// literalLines keeps line breaks, removing the block's common indentation.
func literalLines(lines []string) string {
	indent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		width := len(line) - len(strings.TrimLeft(line, " \t"))
		if indent < 0 || width < indent {
			indent = width
		}
	}
	kept := make([]string, len(lines))
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			kept[i] = line[indent:] // every non-blank line is at least this indented
		}
	}
	return strings.TrimRight(strings.Join(kept, "\n"), "\n ")
}

// unquote decodes a single- or double-quoted scalar.
func unquote(text string) (string, bool) {
	if len(text) < 2 || text[0] != text[len(text)-1] {
		return "", false
	}
	inner := text[1 : len(text)-1]
	switch text[0] {
	case '\'':
		return strings.ReplaceAll(inner, "''", "'"), true
	case '"':
		if decoded, err := strconv.Unquote(text); err == nil {
			return decoded, true
		}
		return inner, true // a YAML-only escape; the raw text is close enough
	}
	return "", false
}
