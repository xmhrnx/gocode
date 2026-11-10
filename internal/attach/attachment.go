// Package attach turns files the user adds to a message into message
// parts: images are scaled down and re-encoded, PDFs are kept as
// documents with their text extracted, office and e-book files are
// turned into text, and text files are inlined.
package attach

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"net/http"
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
	// Text is a document's extracted text, or what an audio or video
	// file is.
	Text string
	// Path is where the file is on disk, when known, so the agent can
	// work with it through its tools.
	Path string
}

// Limits bounds what can be attached.
type Limits struct {
	MaxBytes     int64
	MaxTextBytes int64
}

const (
	KindText Kind = iota
	KindImage
	// KindDocument is a PDF, audio or video file, sent whole where the
	// provider takes it and as its text or a description elsewhere.
	KindDocument
)

// Read reads one attachment from r: an image, a PDF, an office or
// e-book document (as its text), or a text file.
func Read(name string, r io.Reader, limits Limits) (Attachment, error) {
	data, err := io.ReadAll(io.LimitReader(r, limits.MaxBytes+1))
	if err != nil {
		return Attachment{}, fmt.Errorf("read %s: %w", name, err)
	}
	if int64(len(data)) > limits.MaxBytes {
		return Attachment{}, fmt.Errorf("%s is larger than %d MB", name, limits.MaxBytes>>20)
	}
	switch {
	case isImage(data):
		return readImage(name, data)
	case bytes.HasPrefix(data, []byte("%PDF-")):
		return readPDF(name, data, limits)
	case isMedia(name, data):
		media, _ := mediaType(name, data)
		return readMedia(name, media, data), nil
	case isPacked(name):
		text, err := unpackText(name, data)
		if err != nil {
			return Attachment{}, fmt.Errorf("read %s: %w", name, err)
		}
		return textAttachment(name, []byte(text), limits)
	}
	text, ok := decodeText(data)
	if !ok {
		return Attachment{}, fmt.Errorf("%s can't be attached (%s): attach text, code, images, PDFs, "+
			"Word, Excel, PowerPoint, OpenDocument, EPUB, audio or video files", name, http.DetectContentType(data))
	}
	return textAttachment(name, text, limits)
}

func isMedia(name string, data []byte) bool {
	_, ok := mediaType(name, data)
	return ok
}

// textAttachment inlines text, if it isn't too long to.
func textAttachment(name string, text []byte, limits Limits) (Attachment, error) {
	if int64(len(text)) > limits.MaxTextBytes {
		return Attachment{}, fmt.Errorf("%s has more than %d KB of text; ask the agent to read it instead", name, limits.MaxTextBytes>>10)
	}
	return Attachment{Name: name, Kind: KindText, MediaType: "text/plain", Data: text}, nil
}
