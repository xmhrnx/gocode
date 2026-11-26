// Package relaytest is a stand-in for the relay Worker, for tests: the
// same paths and behaviour, in one process.
package relaytest

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

// Server is a running stand-in relay.
type Server struct {
	*httptest.Server
	mu       sync.Mutex
	desktops map[string]*websocket.Conn
	waiting  map[string]chan acceptance
}

// New starts a relay; Close stops it.
func New() *Server {
	s := &Server{desktops: map[string]*websocket.Conn{}, waiting: map[string]chan acceptance{}}
	mux := http.NewServeMux()
	mux.Handle("/v1/desktop/", websocket.Handler(s.desktop))
	mux.HandleFunc("/v1/phone/", s.phone)
	mux.Handle("/v1/accept/", websocket.Handler(s.accept))
	s.Server = httptest.NewServer(mux)
	return s
}

func (s *Server) desktop(ws *websocket.Conn) {
	room := strings.TrimPrefix(ws.Request().URL.Path, "/v1/desktop/")
	s.mu.Lock()
	if old := s.desktops[room]; old != nil {
		old.Close()
	}
	s.desktops[room] = ws
	s.mu.Unlock()
	_, _ = io.Copy(io.Discard, ws) // until it closes
	s.mu.Lock()
	if s.desktops[room] == ws {
		delete(s.desktops, room)
	}
	s.mu.Unlock()
}

func (s *Server) phone(w http.ResponseWriter, r *http.Request) {
	room := strings.TrimPrefix(r.URL.Path, "/v1/phone/")
	s.mu.Lock()
	desk := s.desktops[room]
	s.mu.Unlock()
	if desk == nil {
		http.Error(w, "desktop offline", http.StatusNotFound)
		return
	}
	websocket.Handler(func(ws *websocket.Conn) {
		ws.PayloadType = websocket.BinaryFrame
		id := newID()
		taken := make(chan acceptance, 1)
		s.mu.Lock()
		s.waiting[room+"/"+id] = taken
		s.mu.Unlock()
		_ = websocket.Message.Send(desk, `{"type":"incoming","id":"`+id+`"}`)
		select {
		case other := <-taken:
			pipe(ws, other.ws)
			close(other.done)
		case <-time.After(15 * time.Second):
		}
		ws.Close()
	}).ServeHTTP(w, r)
}

func (s *Server) accept(ws *websocket.Conn) {
	key := strings.TrimPrefix(ws.Request().URL.Path, "/v1/accept/")
	ws.PayloadType = websocket.BinaryFrame
	s.mu.Lock()
	taken := s.waiting[key]
	delete(s.waiting, key)
	s.mu.Unlock()
	if taken == nil {
		ws.Close()
		return
	}
	done := make(chan struct{})
	taken <- acceptance{ws: ws, done: done}
	<-done
}

// acceptance is a desktop's side of a phone's stream, and a channel
// closed when the pipe is done with it.
type acceptance struct {
	ws   *websocket.Conn
	done chan struct{}
}

// pipe copies both ways until either side closes.
func pipe(a, b *websocket.Conn) {
	done := make(chan struct{}, 2)
	go func() { _, _ = io.Copy(a, b); done <- struct{}{} }()
	go func() { _, _ = io.Copy(b, a); done <- struct{}{} }()
	<-done
	a.Close()
	b.Close()
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
