package shell

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Completion is the result of completing the last word of a command line.
type Completion struct {
	// Line is the command line with its last word completed as far as
	// the matches agree.
	Line string
	// Matches are the names the last word could become, shown when
	// there is more than one.
	Matches []string
}

// maxMatches caps how many matches are listed.
const maxMatches = 200

var builtins = []string{"alias", "bg", "cd", "echo", "exit", "export", "fg", "history", "jobs", "pwd", "set",
	"source", "type", "unalias", "unset"}

// Complete completes the last word of line: a command name when it is
// the first word, a file or folder relative to workDir otherwise.
func Complete(line, workDir string) Completion {
	start := strings.LastIndexAny(line, " \t") + 1
	if start >= 2 && line[start-2] == '\\' {
		start = lastUnescapedSpace(line) + 1 // a word with escaped spaces
	}
	head, word := line[:start], line[start:]
	firstWord := strings.TrimSpace(head) == ""

	var names []string
	var dirPart string
	if firstWord && !strings.ContainsAny(word, `/\`) && word != "" {
		names = commandsWithPrefix(word)
	} else {
		dirPart, names = pathsWithPrefix(unescape(word), workDir)
	}
	if len(names) == 0 {
		return Completion{Line: line}
	}
	sort.Strings(names)
	names = dedupe(names)

	completed := escape(dirPart + commonPrefix(names))
	if len(names) == 1 && !strings.HasSuffix(completed, "/") {
		completed += " " // a finished word: ready for the next one
	}
	result := Completion{Line: head + completed}
	if len(names) > 1 {
		result.Matches = names[:min(len(names), maxMatches)]
	}
	return result
}

// pathsWithPrefix lists the entries of the word's folder that start with
// its last part. Folders end in "/". It returns the folder as typed.
func pathsWithPrefix(word, workDir string) (string, []string) {
	cut := strings.LastIndexAny(word, `/`+string(filepath.Separator)) + 1
	dirPart, base := word[:cut], word[cut:]
	dir := expandHome(dirPart)
	if dir == "" {
		dir = "."
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(workDir, dir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return dirPart, nil
	}
	var names []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(base, ".") {
			continue
		}
		if !hasPrefix(name, base) {
			continue
		}
		if isDir(dir, entry) {
			name += "/"
		}
		names = append(names, name)
	}
	return dirPart, names
}

// commandsWithPrefix lists the executables on PATH and the shell's
// builtins that start with prefix.
func commandsWithPrefix(prefix string) []string {
	var names []string
	for _, name := range builtins {
		if hasPrefix(name, prefix) {
			names = append(names, name)
		}
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if name, ok := commandName(dir, entry); ok && hasPrefix(name, prefix) {
				names = append(names, name)
			}
		}
	}
	return names
}

// commandName reports the name an entry is run by, if it is a program.
func commandName(dir string, entry os.DirEntry) (string, bool) {
	if entry.IsDir() {
		return "", false
	}
	name := entry.Name()
	if runtime.GOOS == "windows" {
		ext := strings.ToLower(filepath.Ext(name))
		for _, runnable := range strings.Split(strings.ToLower(os.Getenv("PATHEXT")), ";") {
			if runnable != "" && ext == runnable {
				return strings.TrimSuffix(name, filepath.Ext(name)), true
			}
		}
		return "", false
	}
	info, err := os.Stat(filepath.Join(dir, name)) // follows links
	return name, err == nil && !info.IsDir() && info.Mode()&0o111 != 0
}

func isDir(dir string, entry os.DirEntry) bool {
	if entry.IsDir() {
		return true
	}
	if entry.Type()&os.ModeSymlink == 0 {
		return false
	}
	info, err := os.Stat(filepath.Join(dir, entry.Name()))
	return err == nil && info.IsDir()
}

func expandHome(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return home + path[1:]
}

// hasPrefix matches names case-insensitively on Windows, whose file
// names are.
func hasPrefix(name, prefix string) bool {
	if runtime.GOOS == "windows" {
		return strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix))
	}
	return strings.HasPrefix(name, prefix)
}

func commonPrefix(names []string) string {
	prefix := names[0]
	for _, name := range names[1:] {
		for !hasPrefix(name, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}

func dedupe(sorted []string) []string {
	out := sorted[:0]
	for i, name := range sorted {
		if i == 0 || name != sorted[i-1] {
			out = append(out, name)
		}
	}
	return out
}

// escape protects spaces in a completed path from the shell; Windows
// shells take the path as typed.
func escape(path string) string {
	if runtime.GOOS == "windows" {
		return path
	}
	return strings.ReplaceAll(path, " ", `\ `)
}

func unescape(word string) string {
	return strings.ReplaceAll(word, `\ `, " ")
}

func lastUnescapedSpace(line string) int {
	for i := len(line) - 1; i >= 0; i-- {
		if line[i] == ' ' && (i == 0 || line[i-1] != '\\') {
			return i
		}
	}
	return -1
}
