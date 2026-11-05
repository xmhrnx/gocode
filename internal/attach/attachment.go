// Package attach turns files the user adds to a message into message
// parts: images are scaled down and re-encoded, text files are inlined.
package attach

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"net/http"
	"unicode/utf8"
)

// Kind is what an attachment is sent as.
type Kind int

// Attachment is a file ready to send.
type Attachment struct {
	Name      string
	Kind      Kind
	MediaType string
	Data      []byte
	// Thumbnail is a small preview of an image attachment.
	Thumbnail image.Image
}

// Limits bounds what can be attached.
type Limits struct {
	MaxBytes     int64
	MaxTextBytes int64
}

const (
	KindText Kind = iota
	KindImage
)

// Read reads one attachment from r.
func Read(name string, r io.Reader, limits Limits) (Attachment, error) {
	data, err := io.ReadAll(io.LimitReader(r, limits.MaxBytes+1))
	if err != nil {
		return Attachment{}, fmt.Errorf("read %s: %w", name, err)
	}
	if int64(len(data)) > limits.MaxBytes {
		return Attachment{}, fmt.Errorf("%s is larger than %d MB", name, limits.MaxBytes>>20)
	}
	mediaType := http.DetectContentType(data)
	if isImage(mediaType) {
		return readImage(name, data)
	}
	if !looksLikeText(data) {
		return Attachment{}, fmt.Errorf("%s isn't text or an image (%s)", name, mediaType)
	}
	if int64(len(data)) > limits.MaxTextBytes {
		return Attachment{}, fmt.Errorf("%s is larger than %d KB; ask the agent to read it instead", name, limits.MaxTextBytes>>10)
	}
	return Attachment{Name: name, Kind: KindText, MediaType: "text/plain", Data: data}, nil
}

func isImage(mediaType string) bool {
	switch mediaType {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
		return true
	}
	return false
}

func looksLikeText(data []byte) bool {
	sample := data[:min(len(data), 8192)]
	return !bytes.Contains(sample, []byte{0}) && utf8.Valid(trimPartialRune(sample))
}

// trimPartialRune drops a rune cut in half by sampling.
func trimPartialRune(b []byte) []byte {
	for i := 0; i < utf8.UTFMax && len(b) > 0 && !utf8.Valid(b); i++ {
		b = b[:len(b)-1]
	}
	return b
}
