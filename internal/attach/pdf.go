package attach

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

// readPDF keeps a PDF whole, for providers that read PDFs, with its text
// for those that don't. A PDF whose text can't be extracted (scanned
// pages) still attaches.
func readPDF(name string, data []byte, limits Limits) (Attachment, error) {
	text, err := pdfText(data)
	if err != nil {
		text = "" // the PDF itself still goes to providers that read PDFs
	}
	if int64(len(text)) > limits.MaxTextBytes {
		text = text[:limits.MaxTextBytes] + "\n[The rest of the text was cut; it is longer than the limit in Settings → Chat.]"
	}
	return Attachment{Name: name, Kind: KindDocument, MediaType: "application/pdf", Data: data, Text: text}, nil
}

// pdfText extracts a PDF's text, recovering from documents the parser
// can't handle.
func pdfText(data []byte) (text string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unreadable PDF: %v", r)
		}
	}()
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	plain, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}
	out, err := io.ReadAll(plain)
	return strings.TrimSpace(string(out)), err
}
