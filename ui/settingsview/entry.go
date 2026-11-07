package settingsview

import (
	"strings"

	"github.com/xmhrnx/giode"
)

// entry is one setting: a title and description beside (or above) the
// control that changes it.
type entry struct {
	category    string
	title       string
	description string
	keywords    string
	control     giode.Element
	// wide places the control below the text, full width.
	wide bool
}

// category is a page in the settings navigation.
type category struct {
	name string
	icon string
}

// matches reports whether the entry contains every word of query.
func (e entry) matches(query string) bool {
	haystack := strings.ToLower(e.category + " " + e.title + " " + e.description + " " + e.keywords)
	for _, word := range strings.Fields(strings.ToLower(query)) {
		if !strings.Contains(haystack, word) {
			return false
		}
	}
	return true
}

var categories = []category{
	{"General", "sliders-horizontal"},
	{"Appearance", "palette"},
	{"Chat", "message-square"},
	{"Agent", "bot"},
	{"Permissions", "shield-check"},
	{"Prompts", "scroll-text"},
	{"Providers", "plug"},
	{"Models", "cpu"},
	{"Skills", "sparkles"},
	{"MCP servers", "server"},
	{"Web search", "search"},
	{"Sounds", "volume-2"},
	{"Network", "globe"},
	{"Shortcuts", "keyboard"},
	{"About", "info"},
}
