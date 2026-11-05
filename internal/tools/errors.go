package tools

import "errors"

var (
	errIsDirectory = errors.New("is a directory; give a file path")
	errBinaryFile  = errors.New("looks like a binary file; text tools cannot change it")
)
