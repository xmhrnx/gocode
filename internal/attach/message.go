package attach

import (
	"strings"

	"gocode/internal/llm"
)

// BuildMessage returns a user message with the attachments followed by
// the typed text: text files inline in <file> tags, images as image parts.
func BuildMessage(text string, attachments []Attachment) llm.Message {
	msg := llm.Message{Role: llm.RoleUser}
	for _, att := range attachments {
		switch att.Kind {
		case KindImage:
			msg.Parts = append(msg.Parts, llm.ImagePart(att.MediaType, att.Data))
		case KindDocument:
			msg.Parts = append(msg.Parts, llm.DocumentPart(llm.Document{
				Name: att.Name, MediaType: att.MediaType, Data: att.Data, Text: att.Text,
			}))
		case KindText:
			msg.Parts = append(msg.Parts, llm.TextPart(fileBlock(att.Name, string(att.Data))))
		}
	}
	if strings.TrimSpace(text) != "" {
		msg.Parts = append(msg.Parts, llm.TextPart(text))
	}
	return msg
}

// FileBlockPrefix starts a text part holding an attached file, so views
// can show it as a chip instead of its content.
const FileBlockPrefix = "<file name=\""

func fileBlock(name, content string) string {
	return FileBlockPrefix + name + "\">\n" + content + "\n</file>"
}

// FileBlockName returns the file name of a part made by BuildMessage, and
// whether the text is such a part.
func FileBlockName(text string) (string, bool) {
	rest, ok := strings.CutPrefix(text, FileBlockPrefix)
	if !ok {
		return "", false
	}
	name, _, ok := strings.Cut(rest, "\">")
	return name, ok
}
