package browser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

// DefaultPort is the port the app listens on unless the settings say
// otherwise; the extension uses the same default.
const DefaultPort = 47821

// Path is the WebSocket endpoint.
const Path = "/browser"

// ErrNotConnected is returned by Call when no browser is connected.
var ErrNotConnected = errors.New("no browser is connected to GoCode")

// helloTimeout bounds how long a new connection may take to introduce
// itself.
const helloTimeout = 10 * time.Second

// Status describes the bridge for the settings page.
type Status struct {
	// Listening is true while the app accepts connections.
	Listening bool
	Port      int
	// Connected is true while a browser is connected; Browser names it.
	Connected bool
	Browser   string
	// Err is why the bridge couldn't listen, e.g. the port is taken.
	Err string
}

// Bridge accepts the extension's connection and sends it commands. It is
// safe for concurrent use.
type Bridge struct {
	mu       sync.Mutex
	server   *http.Server
	port     int
	listenOK bool
	err      error
	current  *peer
	onChange func()
	// retry tries a taken port again until it is free or listening stops.
	retry *time.Timer
}

// New returns a bridge that isn't listening yet; call Sync.
func New() *Bridge { return &Bridge{} }

// OnChange sets fn to run, on any goroutine, when the status changes.
func (b *Bridge) OnChange(fn func()) {
	b.mu.Lock()
	b.onChange = fn
	b.mu.Unlock()
}

// Sync starts listening on port when enabled, stops when not, and moves
// to a new port when it changed.
func (b *Bridge) Sync(enabled bool, port int) {
	if port <= 0 {
		port = DefaultPort
	}
	b.mu.Lock()
	running := b.server != nil || b.err != nil
	unchanged := enabled == running && (!enabled || port == b.port)
	b.mu.Unlock()
	if unchanged {
		return
	}
	b.Close()
	if enabled {
		b.listen(port)
	}
	b.changed()
}

// Close stops listening and drops the browser's connection.
func (b *Bridge) Close() {
	b.mu.Lock()
	server, current := b.server, b.current
	b.server, b.current, b.listenOK, b.err, b.port = nil, nil, false, nil, 0
	if b.retry != nil {
		b.retry.Stop()
		b.retry = nil
	}
	b.mu.Unlock()
	if server != nil {
		_ = server.Close()
	}
	if current != nil {
		current.close()
	}
}

// Status reports whether the bridge is listening and a browser connected.
func (b *Bridge) Status() Status {
	b.mu.Lock()
	defer b.mu.Unlock()
	status := Status{Listening: b.listenOK, Port: b.port}
	if b.err != nil {
		status.Err = b.err.Error()
	}
	if b.current != nil {
		status.Connected, status.Browser = true, b.current.browser
	}
	return status
}

// Call sends method with params to the connected browser and waits for
// its result.
func (b *Bridge) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	b.mu.Lock()
	p := b.current
	b.mu.Unlock()
	if p == nil {
		return nil, ErrNotConnected
	}
	return p.call(ctx, method, params)
}

func (b *Bridge) listen(port int) {
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	b.mu.Lock()
	defer b.mu.Unlock()
	b.port = port
	if err != nil {
		b.err = listenError(port, err)
		b.retry = time.AfterFunc(retryEvery, func() { b.retryListen(port) })
		return
	}
	b.err = nil
	mux := http.NewServeMux()
	mux.Handle(Path, websocket.Server{Handshake: allowExtensionOrigin, Handler: b.serve})
	b.server = &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	b.listenOK = true
	go func(server *http.Server) { _ = server.Serve(listener) }(b.server)
}

// retryEvery is how often a taken port is tried again.
const retryEvery = 5 * time.Second

// retryListen tries the port again if listening on it still failed and
// nothing changed since.
func (b *Bridge) retryListen(port int) {
	b.mu.Lock()
	still := b.err != nil && b.port == port && b.server == nil
	b.retry = nil
	b.mu.Unlock()
	if still {
		b.listen(port)
		b.changed()
	}
}

// listenError explains why the port couldn't be used.
func listenError(port int, err error) error {
	if isAddrInUse(err) {
		return fmt.Errorf("Port %d is in use, maybe by another GoCode window. GoCode tries again every few seconds, or pick another port here and in the extension.", port)
	}
	return fmt.Errorf("couldn't listen on port %d: %w", port, err)
}

// allowExtensionOrigin accepts only browser extensions. Browsers always
// send the origin of the page or extension opening a WebSocket, and pages
// can't change it, so websites can't drive the browser through GoCode.
func allowExtensionOrigin(config *websocket.Config, req *http.Request) error {
	origin := req.Header.Get("Origin")
	for _, scheme := range []string{"chrome-extension://", "moz-extension://"} {
		if strings.HasPrefix(origin, scheme) {
			return nil
		}
	}
	return fmt.Errorf("origin %q is not a browser extension", origin)
}

// serve runs one extension connection until it closes.
func (b *Bridge) serve(ws *websocket.Conn) {
	ws.MaxPayloadBytes = 16 << 20
	p := &peer{ws: ws, pending: map[int64]chan reply{}}
	defer p.close()
	_ = ws.SetReadDeadline(time.Now().Add(helloTimeout))
	var hello message
	if err := websocket.JSON.Receive(ws, &hello); err != nil || hello.Type != "hello" {
		return
	}
	_ = ws.SetReadDeadline(time.Time{})
	p.browser = hello.Browser
	b.attach(p)
	defer b.detach(p)
	p.readLoop()
}

// attach makes p the browser commands go to; a newer connection replaces
// an older one, e.g. after the extension reloads.
func (b *Bridge) attach(p *peer) {
	b.mu.Lock()
	old := b.current
	b.current = p
	b.mu.Unlock()
	if old != nil {
		old.close()
	}
	b.changed()
}

func (b *Bridge) detach(p *peer) {
	b.mu.Lock()
	wasCurrent := b.current == p
	if wasCurrent {
		b.current = nil
	}
	b.mu.Unlock()
	if wasCurrent {
		b.changed()
	}
}

func (b *Bridge) changed() {
	b.mu.Lock()
	fn := b.onChange
	b.mu.Unlock()
	if fn != nil {
		fn()
	}
}
