package tools

import (
	"strings"
	"testing"
)

func TestHTMLToText(t *testing.T) {
	tests := []struct {
		name, html, want string
	}{
		{"strips script and style", "<style>p{}</style><p>Hi</p><script>alert('<p>x</p>')</script>", "Hi"},
		{"blocks become lines", "<div>one</div><div>two<br>three</div>", "one\ntwo\nthree"},
		{"collapses whitespace", "<p>  a \n\n  b  <b>bold</b>  c </p>", "a b bold c"},
		{"entities", "<p>&lt;tag&gt; &amp; &quot;q&quot; &#169;</p>", "<tag> & \"q\" ©"},
		{"lists and headings", "<h2>Steps</h2><ul><li>first</li><li>second</li></ul>", "## Steps\n\n- first\n- second"},
		{"pre keeps layout", "<pre>a  b\n  c</pre>", "a  b\n  c"},
		{"blank lines squeezed", "<p>a</p><p></p><p></p><p>b</p>", "a\n\nb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := htmlToText(strings.NewReader(tt.html))
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
