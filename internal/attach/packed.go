package attach

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
)

// xmlRules says how to turn a document's XML into text: which elements
// hold text (all, when textIn is nil) and what each element's end adds.
type xmlRules struct {
	textIn map[string]bool
	ends   map[string]string
}

// maxPartBytes bounds how much of one file in a document is read.
const maxPartBytes = 64 << 20

var (
	wordRules  = xmlRules{textIn: set("t"), ends: map[string]string{"p": "\n", "tab": "\t", "br": "\n", "cr": "\n"}}
	slideRules = xmlRules{textIn: set("t"), ends: map[string]string{"p": "\n", "br": "\n"}}
	odfRules   = xmlRules{ends: map[string]string{"p": "\n", "h": "\n", "tab": "\t", "line-break": "\n", "s": " ",
		"table-cell": "\t", "table-row": "\n"}}
)

// packers are the zipped document formats, by extension.
var packers = map[string]func(*zip.Reader) (string, error){
	".docx": func(z *zip.Reader) (string, error) { return xmlFileText(z, "word/document.xml", wordRules) },
	".pptx": slidesText,
	".xlsx": workbookText,
	".odt":  odfText,
	".ods":  odfText,
	".odp":  odfText,
	".epub": epubText,
}

// isPacked reports whether name is a document format read as text.
func isPacked(name string) bool {
	_, ok := packers[strings.ToLower(path.Ext(name))]
	return ok
}

// unpackText returns the text of a zipped document.
func unpackText(name string, data []byte) (string, error) {
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("not a valid %s file: %w", path.Ext(name), err)
	}
	text, err := packers[strings.ToLower(path.Ext(name))](z)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(text) == "" {
		return "", errors.New("it has no text")
	}
	return strings.TrimSpace(text), nil
}

func odfText(z *zip.Reader) (string, error) { return xmlFileText(z, "content.xml", odfRules) }

// slidesText returns each slide's text, in order.
func slidesText(z *zip.Reader) (string, error) {
	var b strings.Builder
	for i, name := range numbered(z, "ppt/slides/slide", ".xml") {
		text, err := xmlFileText(z, name, slideRules)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "## Slide %d\n%s\n\n", i+1, strings.TrimSpace(text))
	}
	return b.String(), nil
}

// numbered returns the files prefix<n>suffix in order of n.
func numbered(z *zip.Reader, prefix, suffix string) []string {
	type entry struct {
		n    int
		name string
	}
	var entries []entry
	for _, f := range z.File {
		digits, ok := strings.CutPrefix(f.Name, prefix)
		if !ok || !strings.HasSuffix(digits, suffix) {
			continue
		}
		if n, err := strconv.Atoi(strings.TrimSuffix(digits, suffix)); err == nil {
			entries = append(entries, entry{n, f.Name})
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].n < entries[j].n })
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.name
	}
	return names
}

// xmlFileText returns the text of one XML file in z by rules.
func xmlFileText(z *zip.Reader, name string, rules xmlRules) (string, error) {
	r, err := openZipped(z, name)
	if err != nil {
		return "", err
	}
	defer r.Close()
	return xmlText(r, rules)
}

func openZipped(z *zip.Reader, name string) (io.ReadCloser, error) {
	f, err := z.Open(name)
	if err != nil {
		return nil, fmt.Errorf("missing %s: %w", name, err)
	}
	return struct {
		io.Reader
		io.Closer
	}{io.LimitReader(f, maxPartBytes), f}, nil
}

// xmlText walks an XML document, keeping the text rules asks for.
func xmlText(r io.Reader, rules xmlRules) (string, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	var b strings.Builder
	depth := 0 // how many text-holding elements are open
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return b.String(), nil
		}
		if err != nil {
			return "", fmt.Errorf("unreadable XML: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if rules.textIn[t.Name.Local] {
				depth++
			}
		case xml.EndElement:
			if rules.textIn[t.Name.Local] {
				depth--
			}
			b.WriteString(rules.ends[t.Name.Local])
		case xml.CharData:
			if rules.textIn == nil || depth > 0 {
				b.Write(t)
			}
		}
	}
}

func set(names ...string) map[string]bool {
	m := make(map[string]bool, len(names))
	for _, n := range names {
		m[n] = true
	}
	return m
}
