package settingsview

import (
	"strings"
	"unicode"
)

// splitCommandLine splits a command line into words the way a shell would
// for simple cases: spaces separate words, and single or double quotes
// keep spaces inside a word.
func splitCommandLine(line string) []string {
	var words []string
	var word strings.Builder
	quote, inWord := rune(0), false
	for _, r := range line {
		switch {
		case quote != 0 && r == quote:
			quote = 0
		case quote != 0:
			word.WriteRune(r)
		case r == '"' || r == '\'':
			quote, inWord = r, true
		case unicode.IsSpace(r):
			if inWord {
				words, inWord = append(words, word.String()), false
				word.Reset()
			}
		default:
			word.WriteRune(r)
			inWord = true
		}
	}
	if inWord {
		words = append(words, word.String())
	}
	return words
}

// parsePairs reads "KEY=value" or "Key: value" lines into a map.
func parsePairs(text, separator string) map[string]string {
	pairs := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		key, value, ok := strings.Cut(line, separator)
		if key = strings.TrimSpace(key); ok && key != "" {
			pairs[key] = strings.TrimSpace(value)
		}
	}
	if len(pairs) == 0 {
		return nil
	}
	return pairs
}
