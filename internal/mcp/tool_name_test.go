package mcp

import (
	"regexp"
	"strings"
	"testing"
)

var validToolName = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

func TestToolNameSanitises(t *testing.T) {
	tests := []struct{ server, tool, want string }{
		{"github", "create_issue", "mcp__github__create_issue"},
		{"my server", "read.file", "mcp__my_server__read_file"},
		{"bücher", "list-all", "mcp__b_cher__list-all"},
	}
	for _, test := range tests {
		if got := toolName(test.server, test.tool); got != test.want {
			t.Errorf("toolName(%q, %q) = %q, want %q", test.server, test.tool, got, test.want)
		}
	}
}

func TestToolNameTruncatesLongNamesDistinctly(t *testing.T) {
	long := strings.Repeat("x", 80)
	first, second := toolName("server", long+"a"), toolName("server", long+"b")
	for _, name := range []string{first, second} {
		if !validToolName.MatchString(name) {
			t.Errorf("invalid name %q (%d chars)", name, len(name))
		}
	}
	if first == second {
		t.Errorf("long names collide: %q", first)
	}
}

func TestUniqueNameAddsSuffix(t *testing.T) {
	used := make(map[string]bool)
	base := toolName("s", "a.b")
	if got := uniqueName(base, used); got != base {
		t.Errorf("first use = %q", got)
	}
	if got := uniqueName(toolName("s", "a_b"), used); got != base+"_2" {
		t.Errorf("collision = %q, want %q", got, base+"_2")
	}
	long := toolName("server", strings.Repeat("y", 80))
	uniqueName(long, used)
	if got := uniqueName(long, used); len(got) > maxToolNameLength || got == long {
		t.Errorf("long collision = %q", got)
	}
}
