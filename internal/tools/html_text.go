package tools

import (
	"io"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// skippedElements hold no readable text.
var skippedElements = map[atom.Atom]bool{
	atom.Script: true, atom.Style: true, atom.Noscript: true, atom.Template: true,
	atom.Svg: true, atom.Iframe: true, atom.Object: true,
}

// blockElements start and end on their own lines; the value is how many
// line breaks separate them from their surroundings (2 leaves a blank line
// around paragraph-like blocks).
var blockElements = map[atom.Atom]int{
	atom.Address: 1, atom.Article: 1, atom.Aside: 1, atom.Dd: 1, atom.Div: 1,
	atom.Dt: 1, atom.Figcaption: 1, atom.Footer: 1, atom.Form: 1, atom.Header: 1,
	atom.Hr: 1, atom.Li: 1, atom.Main: 1, atom.Nav: 1, atom.Section: 1,
	atom.Title: 1, atom.Tr: 1,
	atom.Blockquote: 2, atom.Dl: 2, atom.Figure: 2, atom.H1: 2, atom.H2: 2,
	atom.H3: 2, atom.H4: 2, atom.H5: 2, atom.H6: 2, atom.Ol: 2, atom.P: 2,
	atom.Pre: 2, atom.Table: 2, atom.Ul: 2,
}

// headingLevels mark headings in Markdown style, which models read well.
var headingLevels = map[atom.Atom]int{
	atom.H1: 1, atom.H2: 2, atom.H3: 3, atom.H4: 4, atom.H5: 5, atom.H6: 6,
}

// htmlToText extracts the readable text of an HTML document: scripts and
// styles are dropped, block elements become line breaks, entities are
// decoded and whitespace is collapsed outside <pre>.
func htmlToText(r io.Reader) (string, error) {
	z := html.NewTokenizer(r)
	var w textWriter
	skipDepth, preDepth := 0, 0
	for {
		tokenType := z.Next()
		if tokenType == html.ErrorToken {
			if z.Err() == io.EOF {
				return w.finish(), nil
			}
			return "", z.Err()
		}
		name, _ := z.TagName()
		tag := atom.Lookup(name)
		switch tokenType {
		case html.TextToken:
			if skipDepth == 0 {
				w.text(string(z.Text()), preDepth > 0)
			}
		case html.StartTagToken:
			skipDepth += boolInt(skippedElements[tag])
			preDepth += boolInt(tag == atom.Pre)
			w.openTag(tag)
		case html.SelfClosingTagToken:
			w.openTag(tag)
		case html.EndTagToken:
			skipDepth = max(skipDepth-boolInt(skippedElements[tag]), 0)
			preDepth = max(preDepth-boolInt(tag == atom.Pre), 0)
			w.closeTag(tag)
		}
	}
}

// boolInt is 1 for true and 0 for false, for depth counters.
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
