package chat

// noticeKind styles a notice.
type noticeKind int

// notice is a line in the transcript that isn't part of the conversation
// sent to the model: an error, or "Stopped.".
type notice struct {
	kind noticeKind
	text string
	// after is how many messages preceded the notice, which places it.
	after int
	id    int
}

const (
	noticeInfo noticeKind = iota
	noticeError
)

func (c *ChatTab) addNotice(kind noticeKind, text string) {
	c.notices = append(c.notices, notice{kind: kind, text: text, after: len(c.session.Messages), id: len(c.notices)})
	c.changed()
}
