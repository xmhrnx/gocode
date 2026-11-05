package chat

import (
	"slices"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"gocode/internal/attach"
	"gocode/internal/llm"
	"gocode/ui/components"
	"gocode/ui/modelpicker"
	"gocode/ui/theme"
)

// composer is the message box at the bottom of the chat: attachments, the
// editor, and the toolbar with the mode, model and send controls.
type composer struct {
	tab         *ChatTab
	editor      *input.Input
	attachments []pendingAttachment
	nextID      int
	chips       *giode.KeyedCache[int, *attachmentChip]
	attachFile  *components.IconButton
	send        *sendButton
	modes       *modeMenu
	model       *components.PillButton
	picker      *modelpicker.Picker
}

// pendingAttachment is an attachment waiting in the composer.
type pendingAttachment struct {
	id int
	attach.Attachment
}

func newComposer(tab *ChatTab) *composer {
	c := &composer{tab: tab, chips: giode.NewKeyedCache[int, *attachmentChip]()}
	c.editor = components.NewMessageField("Ask anything, or describe a task…", false).
		Submit(func(string) { tab.send() })
	c.attachFile = components.NewIconButton("paperclip", "Attach files or images", c.pickFiles)
	c.send = newSendButton(tab)
	c.modes = newModeMenu(tab)
	c.picker = modelpicker.New(tab.services, func(ref string) { tab.selectModel(ref) })
	c.picker.Current = tab.model.String
	c.model = components.NewPillButton("cpu", "Choose a model  ("+giode.FormatShortcut("Mod+K")+")", c.picker.Open)
	return c
}

// takeMessage empties the composer into a user message; ok is false when
// there is nothing to send.
func (c *composer) takeMessage() (llm.Message, bool) {
	text := c.editor.Text()
	if len(c.attachments) == 0 && isBlank(text) {
		return llm.Message{}, false
	}
	files := make([]attach.Attachment, len(c.attachments))
	for i, p := range c.attachments {
		files[i] = p.Attachment
	}
	msg := attach.BuildMessage(text, files)
	c.editor.SetText("")
	c.attachments = nil
	return msg, true
}

// restoreQueued puts messages that were never delivered back into the
// editor, after anything typed since.
func (c *composer) restoreQueued(messages []llm.Message) {
	texts := []string{}
	if current := c.editor.Text(); !isBlank(current) {
		texts = append(texts, current)
	}
	for _, msg := range messages {
		texts = append(texts, msg.Text())
	}
	c.editor.SetText(strings.Join(texts, "\n\n"))
}

// restore puts an unsent message's text back.
func (c *composer) restore(msg llm.Message) {
	c.editor.SetText(msg.Text())
}

// applySettings follows the send key and text size settings.
func (c *composer) applySettings() {
	c.editor.SubmitOn(components.SubmitMode(c.tab.services.Settings.Chat.SendKey))
	c.editor.Styles(components.MessageFieldStyle(false))
}

func (c *composer) view() giode.Element {
	c.model.SetLabel(modelLabel(c.tab.model.Model))
	return giode.Row(giode.Styles{Width: -1, Justify: giode.JustifyCenter, Padding: giode.Inset{Left: 20, Right: 20, Bottom: 16}},
		giode.Stack(giode.Styles{Width: -1, MaxWidth: chatWidth(c.tab), Gap: 8},
			todoPanel(c.tab),
			queuedPanel(c.tab),
			c.card(),
			c.picker,
		),
	)
}

func (c *composer) card() giode.Element {
	return giode.Stack(giode.Styles{
		Width: -1, Background: theme.Panel, BorderRadius: 14, BorderWidth: 1, BorderColor: theme.Border,
		Padding: giode.Inset{Top: 12, Left: 14, Right: 10, Bottom: 8}, Gap: 8,
		BoxShadow: giode.BoxShadow{Y: 6, Blur: 24, Color: "#0000002a"},
	},
		c.attachmentRow(),
		giode.Box(giode.Styles{Width: -1, Padding: giode.Inset{Right: 4}}, c.editor),
		c.toolbar(),
	)
}

func (c *composer) toolbar() giode.Element {
	return giode.Row(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 4, Wrap: giode.WrapNoWrap},
		c.attachFile,
		c.modes,
		giode.Box(giode.Styles{FlexGrow: 1}),
		c.model,
		c.send,
	)
}

func (c *composer) attachmentRow() giode.Element {
	if len(c.attachments) == 0 {
		return giode.Box(giode.Styles{Display: giode.DisplayNone})
	}
	chips := make([]giode.Element, 0, len(c.attachments))
	for _, att := range c.attachments {
		chips = append(chips, c.chips.Get(att.id, func() *attachmentChip { return newAttachmentChip(c, att) }))
	}
	return giode.Row(giode.Styles{Gap: 6}, chips...)
}

func (c *composer) add(att attach.Attachment) {
	c.nextID++
	c.attachments = append(c.attachments, pendingAttachment{id: c.nextID, Attachment: att})
}

func (c *composer) remove(id int) {
	giode.AfterLayout(func() {
		c.attachments = slices.DeleteFunc(c.attachments, func(p pendingAttachment) bool { return p.id == id })
	})
}
