package tools

import (
	"fmt"
	"strings"
)

// applyEdit replaces oldString with newString in content and returns the
// new content and the number of replacements. When the file uses CRLF
// line endings but the strings use LF, as models usually write them, the
// match is retried on LF-normalized text and the result is written back
// with CRLF.
func applyEdit(content, oldString, newString string, replaceAll bool) (string, int, error) {
	count := strings.Count(content, oldString)
	if count == 0 && strings.Contains(content, "\r\n") && !strings.Contains(oldString, "\r\n") {
		lf := strings.ReplaceAll(content, "\r\n", "\n")
		updated, n, err := replaceChecked(lf, oldString, toLF(newString), replaceAll)
		return strings.ReplaceAll(updated, "\n", "\r\n"), n, err
	}
	return replaceChecked(content, oldString, newString, replaceAll)
}

// replaceChecked performs the replacement, refusing ambiguous or missing
// matches with guidance for the model.
func replaceChecked(content, oldString, newString string, replaceAll bool) (string, int, error) {
	count := strings.Count(content, oldString)
	switch {
	case count == 0:
		return "", 0, fmt.Errorf("old_string was not found in the file. It must match exactly, including whitespace " +
			"and indentation; read the file again and copy the text without the line-number prefixes")
	case count > 1 && !replaceAll:
		return "", 0, fmt.Errorf("old_string matches %d places in the file; include more surrounding lines to make "+
			"it unique, or set replace_all to true to replace every occurrence", count)
	case replaceAll:
		return strings.ReplaceAll(content, oldString, newString), count, nil
	}
	return strings.Replace(content, oldString, newString, 1), 1, nil
}

// toLF converts CRLF line endings to LF.
func toLF(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }
