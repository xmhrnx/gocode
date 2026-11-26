// Package phone connects the desktop's chats to the GoCode mobile app:
// what a paired phone sees and what it may do, done on the UI goroutine.
package phone

import (
	"errors"
	"time"

	"github.com/xmhrnx/giode"

	"gocode/internal/remote/wire"
	"gocode/ui/core"
)

// Chat is a tab a phone can follow: the chat tab.
type Chat interface {
	RemoteID() string
	RemoteSummary() wire.ChatSummary
	RemoteView() wire.ChatView
	RemoteSend(text string) error
	RemoteStop()
	RemoteDecide(d wire.Decide) error
}

// Controller is the host.Controller over the open tabs.
type Controller struct {
	app  *giode.App
	name string
	tabs func() []core.Tab
}

// uiTimeout bounds how long a phone's request waits for the UI.
const uiTimeout = 5 * time.Second

var errBusy = errors.New("the desktop didn't answer in time")
var errNoChat = errors.New("that chat isn't open on the desktop any more")

// NewController returns a controller over the tabs tabs returns, which
// it calls on the UI goroutine.
func NewController(app *giode.App, name string, tabs func() []core.Tab) *Controller {
	return &Controller{app: app, name: name, tabs: tabs}
}

// onUI runs fn on the UI goroutine and waits for it.
func (c *Controller) onUI(fn func()) error {
	done := make(chan struct{})
	c.app.Post(func() { fn(); close(done) })
	select {
	case <-done:
		return nil
	case <-time.After(uiTimeout):
		return errBusy
	}
}

// Chats lists the open chats that have messages.
func (c *Controller) Chats() wire.ChatList {
	list := wire.ChatList{Desktop: c.name}
	_ = c.onUI(func() {
		for _, tab := range c.tabs() {
			if chat, ok := tab.(Chat); ok {
				if summary := chat.RemoteSummary(); summary.Busy || !isEmpty(tab) {
					list.Chats = append(list.Chats, summary)
				}
			}
		}
	})
	return list
}

// isEmpty reports a chat tab nothing was sent in yet.
func isEmpty(tab core.Tab) bool {
	e, ok := tab.(interface{ Empty() bool })
	return ok && e.Empty()
}

// Chat returns a chat's view.
func (c *Controller) Chat(id string) (view wire.ChatView, found bool) {
	_ = c.onUI(func() {
		if chat := c.find(id); chat != nil {
			view, found = chat.RemoteView(), true
		}
	})
	return view, found
}

// Send sends a prompt to a chat.
func (c *Controller) Send(id, text string) error {
	return c.with(id, func(chat Chat) error { return chat.RemoteSend(text) })
}

// Stop stops a chat's run.
func (c *Controller) Stop(id string) error {
	return c.with(id, func(chat Chat) error { chat.RemoteStop(); return nil })
}

// Decide answers a chat's permission request.
func (c *Controller) Decide(d wire.Decide) error {
	return c.with(d.Chat, func(chat Chat) error { return chat.RemoteDecide(d) })
}

// with runs fn on the chat with id, on the UI goroutine.
func (c *Controller) with(id string, fn func(Chat) error) error {
	var err error
	if uiErr := c.onUI(func() {
		chat := c.find(id)
		if chat == nil {
			err = errNoChat
			return
		}
		err = fn(chat)
	}); uiErr != nil {
		return uiErr
	}
	return err
}

func (c *Controller) find(id string) Chat {
	for _, tab := range c.tabs() {
		if chat, ok := tab.(Chat); ok && chat.RemoteID() == id {
			return chat
		}
	}
	return nil
}
