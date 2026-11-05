package settingsview

import (
	"slices"
	"strings"
)

func contains(list []string, value string) bool { return slices.Contains(list, value) }

func remove(list []string, value string) []string {
	return slices.DeleteFunc(slices.Clone(list), func(v string) bool { return v == value })
}

// firstSentence returns the first sentence of a description.
func firstSentence(text string) string {
	if i := strings.Index(text, ". "); i >= 0 {
		return text[:i+1]
	}
	return strings.TrimSpace(text)
}
