// Package host is the desktop side of the phone remote: it listens on the
// local network, holds the relay room while the agent works, pairs
// phones and serves each connected phone.
package host

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	"gocode/internal/remote/secure"
)

// DefaultPort is where the desktop listens for phones (TCP) and answers
// discovery (UDP).
const DefaultPort = 47822

// Config is what a host needs.
type Config struct {
	// Name is the desktop's name as phones show it.
	Name       string
	Key        secure.KeyPair
	Devices    *Devices
	Controller Controller
	// Room is the desktop's relay room.
	Room string
}

// Status is the host's state, for the settings page.
type Status struct {
	Listening bool
	Port      int
	Err       string
	// Relay is true while the desktop is in its relay room.
	Relay bool
	// Phones counts the phones connected now.
	Phones int
}

// Host serves phones. Its methods are safe from any goroutine.
type Host struct {
	cfg Config
	id  string

	mu        sync.Mutex
	listener  net.Listener
	discovery net.PacketConn
	port      int
	err       error
	relayURL  string
	relay     *relayLoop
	working   bool
	pairing   *pending
	conns     map[*session]bool
	onChange  func()

	// revision changes whenever a chat changes; sessions push views
	// when it does.
	revision atomic.Int64
}

// New returns a host that isn't listening yet; call Sync.
func New(cfg Config) *Host {
	return &Host{cfg: cfg, id: KeyID(cfg.Key.Public), conns: map[*session]bool{}}
}

// OnChange sets fn to run, on any goroutine, when the status changes.
func (h *Host) OnChange(fn func()) {
	h.mu.Lock()
	h.onChange = fn
	h.mu.Unlock()
}

// Changed tells the host that a chat changed, so phones get updates.
func (h *Host) Changed() { h.revision.Add(1) }

// Sync listens on port when enabled and stops when not; relay is the
// relay Worker's address, or "" for the local network only.
func (h *Host) Sync(enabled bool, port int, relay string) {
	if port <= 0 {
		port = DefaultPort
	}
	h.mu.Lock()
	running := h.listener != nil || h.err != nil
	same := enabled == running && (!enabled || port == h.port)
	h.relayURL = relay
	h.mu.Unlock()
	if !same {
		h.stopListening()
		if enabled {
			h.listen(port)
		}
	}
	h.syncRelay()
	h.changed()
}

// Close stops serving and drops every phone.
func (h *Host) Close() {
	h.mu.Lock()
	h.relayURL = ""
	h.mu.Unlock()
	h.stopListening()
	h.syncRelay()
	h.mu.Lock()
	conns := h.conns
	h.conns = map[*session]bool{}
	h.mu.Unlock()
	for s := range conns {
		s.close()
	}
}

// Status reports what the host is doing.
func (h *Host) Status() Status {
	h.mu.Lock()
	defer h.mu.Unlock()
	st := Status{Listening: h.listener != nil, Port: h.port, Phones: len(h.conns), Relay: h.relay != nil && h.relay.connected()}
	if h.err != nil {
		st.Err = h.err.Error()
	}
	return st
}

// ID is the desktop's discovery ID.
func (h *Host) ID() string { return h.id }

func (h *Host) listen(port int) {
	addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(port))
	listener, err := net.Listen("tcp", addr)
	var discovery net.PacketConn
	if err == nil {
		discovery, err = net.ListenPacket("udp4", addr)
		if err != nil {
			listener.Close()
		}
	}
	h.mu.Lock()
	h.port = port
	if err != nil {
		h.err = fmt.Errorf("couldn't listen on port %d: %w", port, err)
		h.mu.Unlock()
		return
	}
	h.listener, h.discovery, h.err = listener, discovery, nil
	h.mu.Unlock()
	go answerDiscovery(discovery, h.id, port)
	go h.acceptLoop(listener)
}

func (h *Host) stopListening() {
	h.mu.Lock()
	listener, discovery := h.listener, h.discovery
	h.listener, h.discovery, h.err, h.port = nil, nil, nil, 0
	h.mu.Unlock()
	if listener != nil {
		listener.Close()
	}
	if discovery != nil {
		discovery.Close()
	}
}

func (h *Host) acceptLoop(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go h.serve(conn)
	}
}

func (h *Host) changed() {
	h.mu.Lock()
	fn := h.onChange
	h.mu.Unlock()
	if fn != nil {
		fn()
	}
}

// Devices is the list of paired phones.
func (h *Host) Devices() *Devices { return h.cfg.Devices }

// Unpair removes a paired phone and drops its connections.
func (h *Host) Unpair(key string) error {
	if err := h.cfg.Devices.Remove(key); err != nil {
		return err
	}
	h.mu.Lock()
	var drop []*session
	for s := range h.conns {
		if secure.EncodeKey(s.conn.Peer) == key {
			drop = append(drop, s)
		}
	}
	h.mu.Unlock()
	for _, s := range drop {
		s.close()
	}
	h.changed()
	return nil
}
