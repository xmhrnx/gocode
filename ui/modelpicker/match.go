package modelpicker

import "strings"

// matches reports whether every word of query appears in text, ignoring
// case, so "son 5" finds "claude-sonnet-5-5".
func matches(query, text string) bool {
	text = strings.ToLower(text)
	for _, word := range strings.Fields(strings.ToLower(query)) {
		if !strings.Contains(text, word) {
			return false
		}
	}
	return true
}
