package remoteui

import (
	"context"
	"time"

	"gocode/internal/remote/client"
	"gocode/internal/remote/wire"
)

// linkState is how a link to a desktop stands.
type linkState int

const (
	linkConnecting linkState = iota
	linkOpen
	linkFailed
)

// retryEvery is how often a failed link tries again while shown.
const retryEvery = 5 * time.Second

// link is the connection to the desktop being shown, and what it last
// sent. It is changed on the UI goroutine only.
type link struct {
	app     *App
	desktop client.Desktop
	session *client.Session
	state   linkState
	err     string
	route   client.Route
	chats   wire.ChatList
	// watching is the chat shown, "" for the list; view is its view.
	watching string
	view     wire.ChatView
	cancel   context.CancelFunc
	revision int
}

// openLink starts connecting to d.
func openLink(app *App, d client.Desktop) *link {
	l := &link{app: app, desktop: d}
	l.connect()
	return l
}

// close ends the link.
func (l *link) close() {
	if l.cancel != nil {
		l.cancel()
	}
	if l.session != nil {
		l.session.Close()
	}
}

func (l *link) connect() {
	ctx, cancel := context.WithCancel(context.Background())
	l.cancel, l.state, l.err = cancel, linkConnecting, ""
	desk, key, name := l.desktop, l.app.store.key, phoneName()
	go func() {
		s, err := client.Connect(ctx, key, name, &desk, "")
		l.app.post(func() {
			if ctx.Err() != nil {
				if s != nil {
					s.Close()
				}
				return
			}
			if err != nil {
				l.fail(ctx, err)
				return
			}
			l.opened(ctx, s, desk)
		})
	}()
}

// opened starts using a new session: remembers the address that worked,
// asks for what is shown, and follows the pushes.
func (l *link) opened(ctx context.Context, s *client.Session, desk client.Desktop) {
	l.session, l.state, l.route = s, linkOpen, s.Route
	if desk.LastAddr != l.desktop.LastAddr {
		l.desktop.LastAddr = desk.LastAddr
		l.app.store.put(l.desktop)
	}
	l.refresh()
	go func() {
		for {
			select {
			case env := <-s.Pushes():
				l.app.post(func() { l.apply(env) })
			case <-s.Done():
				l.app.post(func() {
					if l.session == s {
						l.session = nil
						l.fail(ctx, s.Err())
					}
				})
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// fail shows err and tries again later while the link is open.
func (l *link) fail(ctx context.Context, err error) {
	l.state, l.err = linkFailed, err.Error()
	l.changed()
	time.AfterFunc(retryEvery, func() {
		l.app.post(func() {
			if ctx.Err() == nil && l.state == linkFailed {
				l.connect()
				l.changed()
			}
		})
	})
}

// refresh asks again for what is shown.
func (l *link) refresh() {
	if l.watching == "" {
		request(l, wire.TypeList, nil, func(list wire.ChatList) { l.chats = list })
		return
	}
	id := l.watching
	request(l, wire.TypeWatch, wire.Watch{Chat: id}, func(v wire.ChatView) {
		if l.watching == id {
			l.view = v
		}
	})
}

// watch shows a chat, or the list for "".
func (l *link) watch(id string) {
	l.watching, l.view = id, wire.ChatView{}
	if l.state == linkOpen {
		l.refresh()
	}
	l.changed()
}

// apply takes a pushed view.
func (l *link) apply(env wire.Envelope) {
	switch env.Type {
	case wire.TypeChats:
		_ = decode(env.Body, &l.chats)
	case wire.TypeChat:
		var v wire.ChatView
		if decode(env.Body, &v) == nil && v.ID == l.watching {
			l.view = v
		}
	}
	l.changed()
}

// act sends a request that changes something, reporting a failure.
func (l *link) act(kind string, body any) {
	if l.session == nil {
		l.app.toast("Not connected to the desktop.")
		return
	}
	s := l.session
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := s.Request(ctx, kind, body, nil); err != nil {
			l.app.post(func() { l.app.toast(err.Error()) })
		}
	}()
}

func (l *link) changed() { l.revision++ }
