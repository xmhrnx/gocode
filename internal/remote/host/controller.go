package host

import "gocode/internal/remote/wire"

// Controller is the desktop app as a phone sees it. Its methods are
// called from the host's goroutines; an implementation does the work on
// its UI goroutine and waits for it.
type Controller interface {
	Chats() wire.ChatList
	// Chat returns a chat's view, or false when no such chat is open.
	Chat(id string) (wire.ChatView, bool)
	Send(id, text string) error
	Stop(id string) error
	Decide(d wire.Decide) error
}
