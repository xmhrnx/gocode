package tools

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// checkReadable explains why abs cannot be read as text, or returns "".
func checkReadable(abs string) string {
	info, err := os.Stat(abs)
	if os.IsNotExist(err) {
		return fmt.Sprintf("file not found: %s", abs)
	}
	if err != nil {
		return err.Error()
	}
	if info.IsDir() {
		return fmt.Sprintf("%s is a directory, not a file; use list_dir to see its contents", abs)
	}
	binary, err := fileLooksBinary(abs)
	if err != nil {
		return err.Error()
	}
	if binary {
		return fmt.Sprintf("%s looks like a binary file (it contains NUL bytes); read_file only shows text", abs)
	}
	return ""
}

// numberedLines formats lines [offset, offset+limit) in cat -n style and
// notes when more lines follow.
func numberedLines(r io.Reader, offset, limit int) (string, error) {
	reader := bufio.NewReader(r)
	var out strings.Builder
	lineNo := 0
	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			lineNo++
			if lineNo >= offset && lineNo < offset+limit {
				line = strings.TrimRight(line, "\r\n")
				fmt.Fprintf(&out, "%6d\t%s\n", lineNo, truncateLine(line, maxReadLineBytes))
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}
	return readSummary(out.String(), lineNo, offset, limit)
}

// readSummary adds the notes for empty files, out-of-range offsets and
// partial reads.
func readSummary(body string, total, offset, limit int) (string, error) {
	switch {
	case total == 0:
		return "(empty file)", nil
	case offset > total:
		return "", fmt.Errorf("offset %d is past the end of the file, which has %d lines", offset, total)
	case offset+limit-1 < total:
		last := offset + limit - 1
		return body + fmt.Sprintf("[showing lines %d-%d of %d; use offset=%d to read more]", offset, last, total, last+1), nil
	}
	return body, nil
}
