package tools

import (
	"fmt"
	"regexp"
	"strings"
)

// compileGlob turns a slash-separated glob into an anchored regexp.
// Supported: "*" (any run within one path segment), "?" (one character
// other than "/"), "**" (any number of segments, including none, when it
// forms a whole segment), "[...]" classes (with "!" or "^" negation) and
// "{a,b}" alternatives.
func compileGlob(pattern string) (*regexp.Regexp, error) {
	var re strings.Builder
	re.WriteString("^")
	inBraces := false
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		switch {
		case strings.HasPrefix(pattern[i:], "**/") && (i == 0 || pattern[i-1] == '/'):
			re.WriteString("(?:.*/)?")
			i += 2
		case strings.HasPrefix(pattern[i:], "**"):
			re.WriteString(".*")
			i++
		case c == '*':
			re.WriteString("[^/]*")
		case c == '?':
			re.WriteString("[^/]")
		case c == '[':
			end := strings.IndexByte(pattern[i+1:], ']')
			if end < 0 {
				return nil, fmt.Errorf("glob %q has an unclosed '['", pattern)
			}
			class := pattern[i+1 : i+1+end]
			if strings.HasPrefix(class, "!") {
				class = "^" + class[1:]
			}
			re.WriteString("[" + strings.ReplaceAll(class, `\`, `\\`) + "]")
			i += end + 1
		case c == '{' && !inBraces:
			re.WriteString("(?:")
			inBraces = true
		case c == '}' && inBraces:
			re.WriteString(")")
			inBraces = false
		case c == ',' && inBraces:
			re.WriteString("|")
		default:
			re.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	if inBraces {
		return nil, fmt.Errorf("glob %q has an unclosed '{'", pattern)
	}
	re.WriteString("$")
	return regexp.Compile(re.String())
}
