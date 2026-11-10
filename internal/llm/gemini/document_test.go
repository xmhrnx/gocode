package gemini

import (
	"encoding/json"
	"strings"
	"testing"

	"gocode/internal/llm"
)

func TestDocumentsGoInline(t *testing.T) {
	doc := llm.Document{Name: "a.pdf", MediaType: "application/pdf", Data: []byte("%PDF-1")}
	sent, _ := json.Marshal(convertContents([]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.DocumentPart(doc)}}}))
	if !strings.Contains(string(sent), `"inlineData":{"mimeType":"application/pdf","data":"JVBERi0x"}`) {
		t.Fatalf("sent %s", sent)
	}
}
