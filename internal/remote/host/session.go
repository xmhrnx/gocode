package host

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"gocode/internal/remote/secure"
	"gocode/internal/remote/wire"
)

// session is one connected phone.
type session struct {
	h    *Host
	conn *secure.Conn

	mu sync.Mutex
	// watching is the chat the phone shows, listWatch for the list, or
	// "" for nothing.
	watching string
	// last is the view last pushed, to skip pushing it unchanged.
	last     []byte
	lastPush time.Time
	done     chan struct{}
	once     sync.Once
}

const (
	listWatch = "\x00list"
	// pushEvery is how often a session looks for changes to push, and
	// tickEvery how often a working chat is pushed anyway, for its time.
	pushEvery = 250 * time.Millisecond
	tickEvery = time.Second
	// handshakeTimeout bounds a phone's handshake.
	handshakeTimeout = 15 * time.Second
)

var errNotPaired = errors.New("not a paired phone")

// deadliner is a stream with a deadline, as TCP connections have.
type deadliner interface{ SetDeadline(time.Time) error }

// serve runs one phone's connection until it closes.
func (h *Host) serve(stream io.ReadWriteCloser) {
	if d, ok := stream.(deadliner); ok {
		_ = d.SetDeadline(time.Now().Add(handshakeTimeout))
	}
	welcome, _ := json.Marshal(wire.Welcome{Name: h.cfg.Name})
	conn, _, err := secure.Accept(stream, h.cfg.Key, func(phone, raw []byte) ([]byte, error) {
		var hello wire.Hello
		_ = json.Unmarshal(raw, &hello)
		if h.cfg.Devices.paired(phone) || h.redeem(hello.Token, deviceName(hello.Name), phone) {
			return welcome, nil
		}
		return nil, errNotPaired
	})
	if err != nil {
		stream.Close()
		return
	}
	if d, ok := stream.(deadliner); ok {
		_ = d.SetDeadline(time.Time{})
	}
	s := &session{h: h, conn: conn, done: make(chan struct{})}
	h.mu.Lock()
	h.conns[s] = true
	h.mu.Unlock()
	h.changed()
	go s.pushLoop()
	s.readLoop()
	s.close()
	h.mu.Lock()
	delete(h.conns, s)
	h.mu.Unlock()
	h.changed()
}

// deviceName keeps a phone's name short and printable.
func deviceName(name string) string {
	name = strings.TrimSpace(strings.Map(func(r rune) rune {
		if r < ' ' {
			return -1
		}
		return r
	}, name))
	if name == "" {
		return "Phone"
	}
	if runes := []rune(name); len(runes) > 60 {
		name = string(runes[:60])
	}
	return name
}

func (s *session) close() {
	s.once.Do(func() {
		close(s.done)
		s.conn.Close()
	})
}

func (s *session) readLoop() {
	for {
		data, err := s.conn.Recv()
		if err != nil {
			return
		}
		var req wire.Envelope
		if json.Unmarshal(data, &req) != nil {
			return
		}
		body, err := s.handle(req)
		if err != nil {
			body, _ = json.Marshal(wire.Failure{Error: err.Error()})
		}
		if s.send(wire.Envelope{Type: wire.TypeReply, ID: req.ID, Body: body}) != nil {
			return
		}
	}
}

// handle carries out one request and returns its result.
func (s *session) handle(req wire.Envelope) (json.RawMessage, error) {
	ctl := s.h.cfg.Controller
	var err error
	switch req.Type {
	case wire.TypeList:
		s.watch(listWatch, nil)
		return s.marshalChats()
	case wire.TypeWatch:
		var ref wire.ChatRef
		if err := json.Unmarshal(req.Body, &ref); err != nil {
			return nil, err
		}
		view, ok := ctl.Chat(ref.Chat)
		if !ok {
			return nil, errors.New("that chat isn't open on the desktop any more")
		}
		data, err := json.Marshal(view)
		s.watch(ref.Chat, data)
		return data, err
	case wire.TypeUnwatch:
		s.watch("", nil)
		return json.RawMessage("{}"), nil
	case wire.TypeSend:
		var msg wire.Send
		if err = json.Unmarshal(req.Body, &msg); err == nil {
			err = ctl.Send(msg.Chat, msg.Text)
		}
	case wire.TypeStop:
		var ref wire.ChatRef
		if err = json.Unmarshal(req.Body, &ref); err == nil {
			err = ctl.Stop(ref.Chat)
		}
	case wire.TypeDecide:
		var d wire.Decide
		if err = json.Unmarshal(req.Body, &d); err == nil {
			err = ctl.Decide(d)
		}
	default:
		return nil, errors.New("unknown request " + req.Type)
	}
	s.h.Changed()
	return json.RawMessage("{}"), err
}

func (s *session) marshalChats() (json.RawMessage, error) {
	data, err := json.Marshal(s.h.cfg.Controller.Chats())
	s.mu.Lock()
	s.last, s.lastPush = data, time.Now()
	s.mu.Unlock()
	return data, err
}

// watch sets what the phone follows; sent is the view it just got.
func (s *session) watch(what string, sent []byte) {
	s.mu.Lock()
	s.watching, s.last, s.lastPush = what, sent, time.Now()
	s.mu.Unlock()
}

// pushLoop sends the watched view whenever it changes.
func (s *session) pushLoop() {
	ticker := time.NewTicker(pushEvery)
	defer ticker.Stop()
	seen := s.h.revision.Load()
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
		}
		rev := s.h.revision.Load()
		s.mu.Lock()
		watching, stale := s.watching, time.Since(s.lastPush) >= tickEvery
		s.mu.Unlock()
		if watching == "" || (rev == seen && !(stale && watching != listWatch)) {
			continue
		}
		seen = rev
		if s.push(watching) != nil {
			s.close()
			return
		}
	}
}

// push sends the current view of what the phone watches, if it changed.
func (s *session) push(watching string) error {
	ctl := s.h.cfg.Controller
	kind, view := wire.TypeChats, any(nil)
	if watching == listWatch {
		view = ctl.Chats()
	} else {
		v, ok := ctl.Chat(watching)
		if !ok {
			return nil
		}
		kind, view = wire.TypeChat, v
	}
	data, err := json.Marshal(view)
	if err != nil {
		return err
	}
	s.mu.Lock()
	unchanged := s.watching != watching || bytes.Equal(data, s.last)
	if !unchanged {
		s.last = data
	}
	s.lastPush = time.Now()
	s.mu.Unlock()
	if unchanged {
		return nil
	}
	return s.send(wire.Envelope{Type: kind, Body: data})
}

func (s *session) send(env wire.Envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return s.conn.Send(data)
}
