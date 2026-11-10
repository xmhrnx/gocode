package openai

import (
	"encoding/json"
	"strings"
	"testing"

	"gocode/internal/llm"
)

func TestDocumentsGoAsFilesOrText(t *testing.T) {
	doc := llm.Document{Name: "a.pdf", MediaType: "application/pdf", Data: []byte("%PDF-1"), Text: "the text"}
	history := []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.DocumentPart(doc), llm.TextPart("sum it up")}}}
	files, _ := json.Marshal(convertMessages("", history, "", mediaInput{files: true}))
	if !strings.Contains(string(files), `"type":"file","file":{"filename":"a.pdf","file_data":"data:application/pdf;base64,JVBERi0x"}`) {
		t.Fatalf("as a file: %s", files)
	}
	text, _ := json.Marshal(convertMessages("", history, "", mediaInput{}))
	if !strings.Contains(string(text), `the text`) || strings.Contains(string(text), "file_data") {
		t.Fatalf("as text: %s", text)
	}
}

func TestAudioGoesAsInputAudio(t *testing.T) {
	doc := llm.Document{Name: "a.wav", MediaType: "audio/wav", Data: []byte("RIFF"), Text: "Audio file, WAV."}
	history := []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.DocumentPart(doc)}}}
	audio, _ := json.Marshal(convertMessages("", history, "", mediaInput{audio: true}))
	if !strings.Contains(string(audio), `"type":"input_audio","input_audio":{"data":"UklGRg==","format":"wav"}`) {
		t.Fatalf("as audio: %s", audio)
	}
	described, _ := json.Marshal(convertMessages("", history, "", mediaInput{}))
	if !strings.Contains(string(described), "can't receive") {
		t.Fatalf("described: %s", described)
	}
}
