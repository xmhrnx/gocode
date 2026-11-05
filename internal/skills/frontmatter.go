package skills

import (
	"errors"
	"fmt"
	"strings"
)

const frontmatterDelimiter = "---"

var errNoFrontmatter = errors.New("SKILL.md must start with a --- frontmatter block")

// parseFrontmatter splits content into its top-level frontmatter fields and
// the Markdown body. It understands the subset of YAML skills use: plain,
// quoted and multi-line scalars plus > (folded) and | (literal) blocks.
// Nested maps and lists are skipped because no field we read uses them.
func parseFrontmatter(content string) (map[string]string, string, error) {
	header, body, err := splitFrontmatter(content)
	if err != nil {
		return nil, "", err
	}
	fields, err := parseFields(header)
	if err != nil {
		return nil, "", err
	}
	return fields, body, nil
}

// splitFrontmatter returns the lines between the --- delimiters and the
// text after the closing one.
func splitFrontmatter(content string) ([]string, string, error) {
	content = strings.TrimPrefix(content, "\ufeff")
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	if strings.TrimSpace(lines[0]) != frontmatterDelimiter {
		return nil, "", errNoFrontmatter
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], " \t") == frontmatterDelimiter {
			return lines[1:i], strings.Join(lines[i+1:], "\n"), nil
		}
	}
	return nil, "", errors.New("frontmatter is not closed with ---")
}

// parseFields reads each top-level "key: value" line together with the
// indented lines that continue it.
func parseFields(lines []string) (map[string]string, error) {
	fields := map[string]string{}
	for i := 0; i < len(lines); {
		line := lines[i]
		i++
		if isBlankOrComment(line) || isIndented(line) {
			continue
		}
		key, raw, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("frontmatter line %q is not \"key: value\"", line)
		}
		end := continuationEnd(lines, i)
		fields[strings.TrimSpace(key)] = scalarValue(strings.TrimSpace(raw), lines[i:end])
		i = end
	}
	return fields, nil
}

// continuationEnd returns the index after the indented (or blank) lines
// that start at start, leaving trailing blank lines out.
func continuationEnd(lines []string, start int) int {
	end := start
	for i := start; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		if !isIndented(lines[i]) {
			break
		}
		end = i + 1
	}
	return end
}

func isBlankOrComment(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "" || strings.HasPrefix(trimmed, "#")
}

func isIndented(line string) bool {
	return strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")
}
