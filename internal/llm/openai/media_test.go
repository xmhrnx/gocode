package openai

import (
	"testing"

	"gocode/internal/llm"
)

func TestStreamImagesBecomeMediaParts(t *testing.T) {
	server, _ := newSSEServer(t,
		`{"choices":[{"index":0,"delta":{"content":"Here it is."}}]}`,
		`{"choices":[{"index":0,"delta":{"images":[{"type":"image_url","image_url":{"url":"data:image/png;base64,iVBORw0K"}}]}}]}`,
		`{"choices":[{"index":0,"delta":{"content":[{"type":"image_url","image_url":{"url":"https://example.com/a.png"}}]},"finish_reason":"stop"}]}`,
		`[DONE]`)
	client := New(Options{BaseURL: server.URL + "/", APIKey: "k"})
	events, err := streamAll(t, client, llm.Request{Model: "m", Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart("draw")}}}})
	if err != nil {
		t.Fatal(err)
	}
	parts := eventsOf(events, llm.EventPart)
	if len(parts) != 1 || parts[0].Part.Kind != llm.PartMedia {
		t.Fatalf("parts = %+v", parts)
	}
	if m := parts[0].Part.Media; m.Name != "image-1.png" || m.MediaType != "image/png" || len(m.Data) != 6 {
		t.Errorf("media = %s %s %d bytes", m.Name, m.MediaType, len(m.Data))
	}
	if got := joined(events, llm.EventText); got != "Here it is.\n\n[Image](https://example.com/a.png)\n\n" {
		t.Errorf("text = %q", got)
	}
}
