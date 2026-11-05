package skills

import (
	"strings"
	"testing"
)

func TestParseSkill(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantName    string
		wantDesc    string
		wantBody    string
		wantErrText string
	}{
		{
			name:     "plain values",
			content:  "---\nname: pdf-tools\ndescription: Work with PDFs.\nlicense: MIT\n---\n# PDF\nBody\n",
			wantName: "pdf-tools", wantDesc: "Work with PDFs.", wantBody: "# PDF\nBody\n",
		},
		{
			name:     "double quoted with escape and colon",
			content:  "---\nname: \"x1\"\ndescription: \"Use when: \\\"PDF\\\" files\"\n---\n",
			wantName: "x1", wantDesc: `Use when: "PDF" files`,
		},
		{
			name:     "single quoted",
			content:  "---\nname: 'x'\ndescription: 'It''s handy'\n---\n",
			wantName: "x", wantDesc: "It's handy",
		},
		{
			name:     "folded block",
			content:  "---\nname: x\ndescription: >\n  First line\n  continues here.\n\n  New paragraph.\nlicense: MIT\n---\nB",
			wantName: "x", wantDesc: "First line continues here.\nNew paragraph.", wantBody: "B",
		},
		{
			name:     "literal block with chomping indicator",
			content:  "---\nname: x\ndescription: |-\n  Line one\n    indented\n  Line three\n---\n",
			wantName: "x", wantDesc: "Line one\n  indented\nLine three",
		},
		{
			name:     "plain multi-line, comment and nested metadata",
			content:  "---\n# comment\nname: x # trailing comment\nmetadata:\n  version: 1\n  tags: [a]\ndescription: Spans\n  two lines\nallowed-tools:\n  - bash\n---\n",
			wantName: "x", wantDesc: "Spans two lines",
		},
		{
			name:     "CRLF and BOM",
			content:  "\ufeff---\r\nname: x\r\ndescription: d\r\n---\r\nBody\r\n",
			wantName: "x", wantDesc: "d", wantBody: "Body\n",
		},
		{name: "no frontmatter", content: "# Just markdown\n", wantErrText: "frontmatter"},
		{name: "unclosed", content: "---\nname: x\ndescription: d\n", wantErrText: "not closed"},
		{name: "missing name", content: "---\ndescription: d\n---\n", wantErrText: "missing the required name"},
		{name: "missing description", content: "---\nname: x\n---\n", wantErrText: "missing the required description"},
		{name: "uppercase name", content: "---\nname: PDF\ndescription: d\n---\n", wantErrText: "lowercase"},
		{name: "path in name", content: "---\nname: ../x\ndescription: d\n---\n", wantErrText: "lowercase"},
		{name: "long name", content: "---\nname: " + strings.Repeat("a", 65) + "\ndescription: d\n---\n", wantErrText: "longer than 64"},
		{name: "long description", content: "---\nname: x\ndescription: " + strings.Repeat("d", 1025) + "\n---\n", wantErrText: "longer than 1024"},
		{name: "not key value", content: "---\nname x\n---\n", wantErrText: "not \"key: value\""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skill, body, err := parseSkill(tt.content)
			if tt.wantErrText != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrText) {
					t.Fatalf("error = %v, want one containing %q", err, tt.wantErrText)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if skill.Name != tt.wantName || skill.Description != tt.wantDesc {
				t.Errorf("got name %q description %q, want %q %q", skill.Name, skill.Description, tt.wantName, tt.wantDesc)
			}
			if body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}
