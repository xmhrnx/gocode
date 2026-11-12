package gemini

import (
	"testing"

	"gocode/internal/llm"
)

func TestStreamInlineDataBecomesMedia(t *testing.T) {
	body := data(`{"candidates":[{"content":{"role":"model","parts":[`+
		`{"inlineData":{"mimeType":"image/jpeg","data":"/9j/"},"thoughtSignature":"sig"},`+
		`{"inlineData":{"mimeType":"audio/L16;codec=pcm;rate=16000","data":"AAABAA=="}}]},"finishReason":"STOP"}]}`)
	server, _ := newServer(t, 200, body)
	events, err := collect(t, server, llm.Request{Model: "m", Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart("hi")}}}})
	if err != nil {
		t.Fatal(err)
	}
	var media []llm.Part
	for _, e := range events {
		if e.Kind == llm.EventPart {
			media = append(media, *e.Part)
		}
	}
	if len(media) != 2 {
		t.Fatalf("media = %+v", media)
	}
	if m := media[0]; m.Media.Name != "image-1.jpg" || m.Signature != "sig" {
		t.Errorf("image = %s %q", m.Media.Name, m.Signature)
	}
	if m := media[1].Media; m.Name != "audio-2.wav" || m.MediaType != "audio/wav" || len(m.Data) != 44+4 || string(m.Data[:4]) != "RIFF" {
		t.Errorf("audio = %s %s %d bytes", m.Name, m.MediaType, len(m.Data))
	}

	// The image goes back with its signature; the audio stays here.
	parts := modelParts(llm.Message{Role: llm.RoleAssistant, Parts: media})
	if len(parts) != 1 || parts[0].InlineData == nil || parts[0].ThoughtSignature != "sig" {
		t.Errorf("sent back = %+v", parts)
	}
}
