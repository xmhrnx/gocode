package mcp

import (
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
)

const (
	// maxToolNameLength is the longest tool name most provider APIs accept.
	maxToolNameLength = 64
	toolNamePrefix    = "mcp__"
	// hashSuffixLength is the size of "_" plus eight hex digits.
	hashSuffixLength = 9
)

// toolName builds the agent-facing name of a server's tool,
// mcp__<server>__<tool>, restricted to the [a-zA-Z0-9_-] charset and 64
// characters that provider APIs require.
func toolName(server, tool string) string {
	name := toolNamePrefix + sanitizeName(server) + "__" + sanitizeName(tool)
	if len(name) <= maxToolNameLength {
		return name
	}
	// Cutting alone could make two long names equal; a hash of the
	// original names keeps them apart.
	sum := crc32.ChecksumIEEE([]byte(server + "\x00" + tool))
	return fmt.Sprintf("%s_%08x", name[:maxToolNameLength-hashSuffixLength], sum)
}

// sanitizeName replaces every character outside [a-zA-Z0-9_-] with "_".
// The result is ASCII, so it can be cut at any byte.
func sanitizeName(s string) string {
	return strings.Map(func(r rune) rune {
		if isNameRune(r) {
			return r
		}
		return '_'
	}, s)
}

// isNameRune reports whether r may appear in a tool name.
func isNameRune(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-'
}

// uniqueName returns base, or base with a numeric suffix if base is taken
// (sanitising can map different names to the same one), and marks the
// result as used.
func uniqueName(base string, used map[string]bool) string {
	name := base
	for n := 2; used[name]; n++ {
		suffix := "_" + strconv.Itoa(n)
		name = base[:min(len(base), maxToolNameLength-len(suffix))] + suffix
	}
	used[name] = true
	return name
}
