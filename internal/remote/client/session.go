package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"gocode/internal/remote/host"
	"gocode/internal/remote/relay"
	"gocode/internal/remote/secure"
	"gocode/internal/remote/wire"
)

// Route is how a session reached its desktop.
type Route string

const (
	RouteLocal Route = "local network"
	RouteRelay Route = "relay"
)

// Session is an open connection to a desktop.
type Session struct {
	conn *secure.Conn
	// Route is how the desktop was reached, and Name what it calls
	// itself.
	Route Route
	Name  string

	mu      sync.Mutex
	nextID  int64
	pending map[int64]chan wire.Envelope
	pushes  chan wire.Envelope
	done    chan struct{}
	err     error
}

// ErrUnreachable is returned when the desktop is neither on this network
// nor in its relay room.
var ErrUnreachable = errors.New("the desktop can't be reached: it isn't on this network, and it isn't working with a relay set up")

// Connect opens a session with d: directly on the local network when it
// is there, else through its relay. self and name identify the phone;
// token is the pairing token the first time, else "". A direct
// connection's address is remembered in d.LastAddr.
func Connect(ctx context.Context, self secure.KeyPair, name string, d *Desktop, token string) (*Session, error) {
	key, err := secure.DecodeKey(d.Key)
	if err != nil {
		return nil, err
	}
	route := RouteLocal
	stream, addr, err := findLocal(ctx, d, host.KeyID(key))
	if err == nil {
		d.LastAddr = addr
	} else if d.Relay != "" {
		route = RouteRelay
		stream, err = relay.Phone(ctx, d.Relay, d.Room)
	}
	if err != nil {
		if errors.Is(err, relay.ErrOffline) || errors.Is(err, errNotNearby) {
			return nil, ErrUnreachable
		}
		return nil, fmt.Errorf("couldn't reach the desktop: %w", err)
	}
	return open(stream, self, key, wire.Hello{Name: name, Token: token}, route)
}

func open(stream io.ReadWriteCloser, self secure.KeyPair, key []byte, hello wire.Hello, route Route) (*Session, error) {
	helloData, _ := json.Marshal(hello)
	conn, welcomeData, err := secure.Dial(stream, self, key, helloData)
	if err != nil {
		stream.Close()
		if errors.Is(err, secure.ErrRejected) {
			return nil, errors.New("the desktop refused this phone: pair it again")
		}
		return nil, err
	}
	var welcome wire.Welcome
	_ = json.Unmarshal(welcomeData, &welcome)
	s := &Session{conn: conn, Route: route, Name: welcome.Name,
		pending: map[int64]chan wire.Envelope{}, pushes: make(chan wire.Envelope, 16), done: make(chan struct{})}
	go s.readLoop()
	return s, nil
}

// Pushes delivers the views the desktop sends as they change: TypeChats
// with a ChatList, TypeChat with a ChatView. Only the latest matters, so
// older ones are dropped when the receiver falls behind.
func (s *Session) Pushes() <-chan wire.Envelope { return s.pushes }

// Done is closed when the session ends; Err then says why.
func (s *Session) Done() <-chan struct{} { return s.done }

// Err is why the session ended.
func (s *Session) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

// Close ends the session.
func (s *Session) Close() { s.conn.Close() }

// Request sends a request and decodes its reply into out (nil to ignore
// it).
func (s *Session) Request(ctx context.Context, kind string, body, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	reply := make(chan wire.Envelope, 1)
	s.mu.Lock()
	s.nextID++
	id := s.nextID
	s.pending[id] = reply
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.pending, id)
		s.mu.Unlock()
	}()
	msg, _ := json.Marshal(wire.Envelope{Type: kind, ID: id, Body: data})
	if err := s.conn.Send(msg); err != nil {
		return err
	}
	select {
	case env := <-reply:
		var failure wire.Failure
		if json.Unmarshal(env.Body, &failure) == nil && failure.Error != "" {
			return errors.New(failure.Error)
		}
		if out == nil {
			return nil
		}
		return json.Unmarshal(env.Body, out)
	case <-s.done:
		return s.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Session) readLoop() {
	var err error
	defer func() {
		s.mu.Lock()
		s.err = err
		s.mu.Unlock()
		close(s.done)
	}()
	for {
		var data []byte
		data, err = s.conn.Recv()
		if err != nil {
			err = fmt.Errorf("the connection to the desktop ended: %w", err)
			return
		}
		var env wire.Envelope
		if err = json.Unmarshal(data, &env); err != nil {
			return
		}
		if env.Type == wire.TypeReply {
			s.mu.Lock()
			reply := s.pending[env.ID]
			s.mu.Unlock()
			if reply != nil {
				reply <- env
			}
			continue
		}
		s.push(env)
	}
}

// push delivers a pushed view, dropping the oldest if the queue is full.
func (s *Session) push(env wire.Envelope) {
	for {
		select {
		case s.pushes <- env:
			return
		default:
		}
		select {
		case <-s.pushes:
		default:
		}
	}
}
