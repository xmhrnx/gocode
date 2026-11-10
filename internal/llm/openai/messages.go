package openai

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"gocode/internal/llm"
)

// chatMessage is one entry of the request's messages array.
type chatMessage struct {
	Role string `json:"role"`
	// Content is a string, a []contentPart, or nil (null) for an assistant
	// message that only calls tools.
	Content    any            `json:"content"`
	ToolCalls  []chatToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	// reasoningField and reasoning echo reasoning back under a
	// provider-specific field name, added by MarshalJSON.
	reasoningField string
	reasoning      string
}

type contentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
	File     *filePart `json:"file,omitempty"`
	Audio    *audioIn  `json:"input_audio,omitempty"`
}

// audioIn is an attached WAV or MP3, for servers that take audio.
type audioIn struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

// audioFormats are the audio types sent as input_audio, by format name.
var audioFormats = map[string]string{"audio/wav": "wav", "audio/mpeg": "mp3"}

// filePart is an attached PDF, for servers that read files.
type filePart struct {
	Filename string `json:"filename"`
	FileData string `json:"file_data"`
}

type imageURL struct {
	URL string `json:"url"`
}

type chatToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function chatFunctionCall `json:"function"`
}

type chatFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// MarshalJSON adds the echoed reasoning field, whose name is only known at
// run time.
func (m chatMessage) MarshalJSON() ([]byte, error) {
	type plain chatMessage
	data, err := json.Marshal(plain(m))
	if err != nil || m.reasoningField == "" {
		return data, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	if fields[m.reasoningField], err = json.Marshal(m.reasoning); err != nil {
		return nil, err
	}
	return json.Marshal(fields)
}

func convertMessages(system string, history []llm.Message, echoField string, media mediaInput) []chatMessage {
	var out []chatMessage
	if system != "" {
		out = append(out, chatMessage{Role: "system", Content: system})
	}
	for _, msg := range history {
		if msg.Role == llm.RoleAssistant {
			if converted, ok := assistantMessage(msg, echoField); ok {
				out = append(out, converted)
			}
			continue
		}
		out = append(out, userMessages(msg, media)...)
	}
	return out
}

// userMessages splits tool results into "tool" messages, which must
// directly follow the assistant's tool calls, ahead of any user content.
// mediaInput is what the server takes besides text and images.
type mediaInput struct {
	files, audio bool
}

// Documents the server takes go whole (PDFs as files, WAV and MP3 as
// audio); the rest as their text or description.
func userMessages(msg llm.Message, media mediaInput) []chatMessage {
	var out []chatMessage
	var parts []contentPart
	hasImage := false
	for _, part := range msg.Parts {
		switch {
		case part.Kind == llm.PartToolResult && part.ToolResult != nil:
			out = append(out, chatMessage{Role: "tool", ToolCallID: part.ToolResult.CallID, Content: part.ToolResult.Content})
		case part.Kind == llm.PartText && part.Text != "":
			parts = append(parts, contentPart{Type: "text", Text: part.Text})
		case part.Kind == llm.PartImage && part.Image != nil:
			hasImage = true
			parts = append(parts, contentPart{Type: "image_url", ImageURL: &imageURL{URL: dataURL(part.Image)}})
		case part.Kind == llm.PartDocument && part.Document != nil && media.files && part.Document.MediaType == "application/pdf":
			hasImage = true // needs the content array too
			doc := part.Document
			parts = append(parts, contentPart{Type: "file", File: &filePart{Filename: doc.Name,
				FileData: "data:" + doc.MediaType + ";base64," + base64.StdEncoding.EncodeToString(doc.Data)}})
		case part.Kind == llm.PartDocument && part.Document != nil && media.audio && audioFormats[part.Document.MediaType] != "":
			hasImage = true
			doc := part.Document
			parts = append(parts, contentPart{Type: "input_audio", Audio: &audioIn{
				Data: base64.StdEncoding.EncodeToString(doc.Data), Format: audioFormats[doc.MediaType]}})
		case part.Kind == llm.PartDocument && part.Document != nil:
			parts = append(parts, contentPart{Type: "text", Text: llm.DocumentText(*part.Document)})
		}
	}
	if len(parts) > 0 {
		out = append(out, chatMessage{Role: "user", Content: userContent(parts, hasImage)})
	}
	return out
}

// userContent prefers a plain string, which every server accepts; content
// arrays are only needed for images and files.
func userContent(parts []contentPart, hasImage bool) any {
	if hasImage {
		return parts
	}
	texts := make([]string, len(parts))
	for i, part := range parts {
		texts[i] = part.Text
	}
	return strings.Join(texts, "\n\n")
}

func dataURL(image *llm.Image) string {
	return "data:" + image.MediaType + ";base64," + base64.StdEncoding.EncodeToString(image.Data)
}
