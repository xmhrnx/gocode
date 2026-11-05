package tools

import (
	"bytes"
	"io"
	"os"
)

// binarySniffBytes is how much of a file is checked for NUL bytes, the same
// heuristic git and grep use to tell binary from text.
const binarySniffBytes = 8 << 10

// looksBinary reports whether data's first bytes contain a NUL.
func looksBinary(data []byte) bool {
	if len(data) > binarySniffBytes {
		data = data[:binarySniffBytes]
	}
	return bytes.IndexByte(data, 0) >= 0
}

// fileLooksBinary sniffs the start of the file at path.
func fileLooksBinary(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	head := make([]byte, binarySniffBytes)
	n, err := io.ReadFull(f, head)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return false, err
	}
	return looksBinary(head[:n]), nil
}
