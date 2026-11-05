package chat

import (
	"strings"

	"github.com/xmhrnx/giode/styles/properties"

	"gocode/internal/catalog"
)

func isBlank(s string) bool { return strings.TrimSpace(s) == "" }

func (c *composer) isEmpty() bool {
	return len(c.attachments) == 0 && isBlank(c.editor.Text())
}

// modelLabel is what the model button shows.
func modelLabel(model string) string {
	if model == "" {
		return "Choose a model"
	}
	return model
}

// chatWidth is the readable column width from the settings.
func chatWidth(c *ChatTab) properties.MaxWidth {
	return properties.MaxWidth(c.services.Settings.Appearance.ChatWidth)
}

// selectModel switches the chat to the model ("provider/model").
func (c *ChatTab) selectModel(ref string) {
	c.model = catalog.ParseRef(ref)
	c.session.Model = ref
	c.services.Settings.Models.MarkUsed(ref)
	c.services.SettingsChanged()
	c.composer.editor.Focus()
}
