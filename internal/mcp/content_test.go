package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCallToolResultText(t *testing.T) {
	tests := []struct{ name, result, want string }{
		{"joins text", `{"content":[{"type":"text","text":"a"},{"type":"text","text":"b"}]}`, "a\nb"},
		{"audio placeholder", `{"content":[{"type":"audio","mimeType":"audio/wav","data":"AAAA"}]}`, "[audio: audio/wav, 3 B]"},
		{"resource text", `{"content":[{"type":"resource","resource":{"uri":"file:///a","text":"body"}}]}`, "body"},
		{"resource blob", `{"content":[{"type":"resource","resource":{"uri":"file:///a","mimeType":"application/pdf","blob":"AAAA"}}]}`, "[resource: file:///a, application/pdf, 3 B]"},
		{"resource link", `{"content":[{"type":"resource_link","uri":"file:///b","name":"b"}]}`, "[resource link: file:///b]"},
		{"structured fallback", `{"content":[],"structuredContent":{"n":1}}`, `{"n":1}`},
		{"text wins over structured", `{"content":[{"type":"text","text":"t"}],"structuredContent":{"n":1}}`, "t"},
	}
	for _, test := range tests {
		var result callToolResult
		if err := json.Unmarshal([]byte(test.result), &result); err != nil {
			t.Fatalf("%s: %v", test.name, err)
		}
		if got := result.text(); got != test.want {
			t.Errorf("%s: text = %q, want %q", test.name, got, test.want)
		}
	}
}

func TestTruncateMiddleKeepsHeadAndTail(t *testing.T) {
	s := "HEAD" + strings.Repeat("-", 100) + "TAIL"
	got := truncateMiddle(s, 20)
	if !strings.HasPrefix(got, "HEAD") || !strings.HasSuffix(got, "TAIL") || !strings.Contains(got, "truncated") {
		t.Errorf("truncateMiddle = %q", got)
	}
	if truncateMiddle("short", 20) != "short" {
		t.Error("short text changed")
	}
}
