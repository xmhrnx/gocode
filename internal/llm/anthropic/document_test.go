package anthropic

import (
	"encoding/json"
	"strings"
	"testing"

	"gocode/internal/llm"
)

func TestDocumentsGoAsDocumentBlocks(t *testing.T) {
	doc := llm.Document{Name: "a.pdf", MediaType: "application/pdf", Data: []byte("%PDF-1")}
	sent, _ := json.Marshal(convertMessages([]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.DocumentPart(doc)}}}))
	want := `{"type":"document","source":{"type":"base64","media_type":"application/pdf","data":"JVBERi0x"},"title":"a.pdf"}`
	if !strings.Contains(string(sent), want) {
		t.Fatalf("sent %s", sent)
	}
}
