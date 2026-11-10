package attach

import (
	"archive/zip"
	"io"
	"path"
	"strings"

	"golang.org/x/net/html"
)

// containerXML points to an EPUB's package document.
type containerXML struct {
	Rootfiles []struct {
		Path string `xml:"full-path,attr"`
	} `xml:"rootfiles>rootfile"`
}

// packageXML lists an EPUB's files and their reading order.
type packageXML struct {
	Items []struct {
		ID   string `xml:"id,attr"`
		Href string `xml:"href,attr"`
	} `xml:"manifest>item"`
	Spine []struct {
		IDRef string `xml:"idref,attr"`
	} `xml:"spine>itemref"`
}

// blockElements end a line in an EPUB chapter's text.
var blockElements = set("p", "div", "br", "li", "tr", "h1", "h2", "h3", "h4", "h5", "h6", "blockquote", "pre", "section")

// epubText returns an EPUB's chapters in reading order.
func epubText(z *zip.Reader) (string, error) {
	var container containerXML
	if err := decodeZipped(z, "META-INF/container.xml", &container); err != nil {
		return "", err
	}
	if len(container.Rootfiles) == 0 {
		return "", errMissing("META-INF/container.xml rootfile")
	}
	opf := container.Rootfiles[0].Path
	var pkg packageXML
	if err := decodeZipped(z, opf, &pkg); err != nil {
		return "", err
	}
	hrefs := map[string]string{}
	for _, item := range pkg.Items {
		hrefs[item.ID] = item.Href
	}
	var b strings.Builder
	for _, ref := range pkg.Spine {
		href, ok := hrefs[ref.IDRef]
		if !ok {
			continue
		}
		r, err := openZipped(z, path.Join(path.Dir(opf), href))
		if err != nil {
			continue // a chapter missing from the file is skipped
		}
		htmlText(r, &b)
		r.Close()
		b.WriteString("\n")
	}
	return b.String(), nil
}

// htmlText writes the text of an XHTML chapter, a line per block.
func htmlText(r io.Reader, b *strings.Builder) {
	tokens := html.NewTokenizer(r)
	skip := 0 // inside <script>, <style> or <head>
	for {
		switch tokens.Next() {
		case html.ErrorToken:
			return
		case html.TextToken:
			if skip == 0 {
				b.WriteString(strings.Join(strings.Fields(string(tokens.Text())), " "))
				b.WriteString(" ")
			}
		case html.StartTagToken:
			if name, _ := tokens.TagName(); isSkipped(string(name)) {
				skip++
			}
		case html.EndTagToken, html.SelfClosingTagToken:
			name, _ := tokens.TagName()
			switch {
			case isSkipped(string(name)):
				skip = max(0, skip-1)
			case blockElements[string(name)]:
				b.WriteString("\n")
			}
		}
	}
}

func isSkipped(tag string) bool { return tag == "script" || tag == "style" || tag == "head" }

type errMissing string

func (e errMissing) Error() string { return "missing " + string(e) }
