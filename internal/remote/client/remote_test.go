package client

import (
	"context"
	"errors"
	"net"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/psanford/wormhole-william/rendezvous/rendezvousservertest"
	"github.com/psanford/wormhole-william/wormhole"

	"gocode/internal/remote/host"
	"gocode/internal/remote/pair"
	"gocode/internal/remote/relay/relaytest"
	"gocode/internal/remote/secure"
	"gocode/internal/remote/wire"
)

// fakeDesktop is a Controller with one chat.
type fakeDesktop struct {
	mu   sync.Mutex
	sent []string
	busy bool
}

func (f *fakeDesktop) Chats() wire.ChatList {
	f.mu.Lock()
	defer f.mu.Unlock()
	return wire.ChatList{Desktop: "Desk", Chats: []wire.ChatSummary{{ID: "c1", Title: "Fix bug", Busy: f.busy}}}
}

func (f *fakeDesktop) Chat(id string) (wire.ChatView, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if id != "c1" {
		return wire.ChatView{}, false
	}
	view := wire.ChatView{ChatSummary: wire.ChatSummary{ID: "c1", Title: "Fix bug", Busy: f.busy}}
	for _, text := range f.sent {
		view.Rows = append(view.Rows, wire.Row{Kind: wire.RowUser, Text: text})
	}
	return view, true
}

func (f *fakeDesktop) Send(id, text string) error {
	f.mu.Lock()
	f.sent = append(f.sent, text)
	f.busy = true
	f.mu.Unlock()
	return nil
}

func (f *fakeDesktop) Stop(string) error        { return nil }
func (f *fakeDesktop) Decide(wire.Decide) error { return errors.New("nothing to decide") }

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func newHost(t *testing.T, ctl host.Controller, relayURL string) (*host.Host, int) {
	t.Helper()
	key, _ := secure.NewKeyPair()
	devices, err := host.OpenDevices(filepath.Join(t.TempDir(), "devices.json"))
	if err != nil {
		t.Fatal(err)
	}
	h := host.New(host.Config{Name: "Desk", Key: key, Devices: devices, Controller: ctl, Room: pair.NewSecret()})
	port := freePort(t)
	h.Sync(true, port, relayURL)
	t.Cleanup(h.Close)
	if st := h.Status(); !st.Listening {
		t.Fatalf("host not listening: %+v", st)
	}
	return h, port
}

// pairPhone pairs a new phone with h through a test wormhole server.
func pairPhone(t *testing.T, ctx context.Context, h *host.Host, port int) (secure.KeyPair, Desktop) {
	t.Helper()
	server := rendezvousservertest.NewServer()
	t.Cleanup(server.Close)
	wormhole.DefaultTransitRelayAddress = ""
	code, result, err := h.Pair(ctx, server.WebSocketURL())
	if err != nil {
		t.Fatal(err)
	}
	offer, err := pair.Receive(ctx, server.WebSocketURL(), code)
	if err != nil {
		t.Fatal(err)
	}
	desk := FromOffer(offer)
	desk.Addrs = []string{"127.0.0.1:" + strconv.Itoa(port)}
	phone, _ := secure.NewKeyPair()
	s, err := Connect(ctx, phone, "Pixel", &desk, offer.Token)
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	if r := <-result; r.Err != nil || r.Device != "Pixel" {
		t.Fatalf("pairing result %+v", r)
	}
	return phone, desk
}

func TestPairConnectAndFollowAChat(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ctl := &fakeDesktop{}
	h, port := newHost(t, ctl, "")
	phone, desk := pairPhone(t, ctx, h, port)

	s, err := Connect(ctx, phone, "Pixel", &desk, "") // no token: paired now
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if s.Route != RouteLocal || s.Name != "Desk" {
		t.Errorf("route %q name %q", s.Route, s.Name)
	}
	var list wire.ChatList
	if err := s.Request(ctx, wire.TypeList, nil, &list); err != nil || len(list.Chats) != 1 {
		t.Fatalf("list %+v %v", list, err)
	}
	var view wire.ChatView
	if err := s.Request(ctx, wire.TypeWatch, wire.Watch{Chat: "c1"}, &view); err != nil || view.Title != "Fix bug" {
		t.Fatalf("watch %+v %v", view, err)
	}
	if err := s.Request(ctx, wire.TypeSend, wire.Send{Chat: "c1", Text: "add tests"}, nil); err != nil {
		t.Fatal(err)
	}
	select {
	case env := <-s.Pushes():
		if env.Type != wire.TypeChat {
			t.Fatalf("push %s", env.Type)
		}
	case <-ctx.Done():
		t.Fatal("no update after sending")
	}
	if err := s.Request(ctx, wire.TypeDecide, wire.Decide{Chat: "c1"}, nil); err == nil || err.Error() != "nothing to decide" {
		t.Errorf("decide error %v", err)
	}
}

func TestUnpairedPhoneIsRefused(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	h, port := newHost(t, &fakeDesktop{}, "")
	_, desk := pairPhone(t, ctx, h, port)
	stranger, _ := secure.NewKeyPair()
	if _, err := Connect(ctx, stranger, "Other", &desk, "made-up"); err == nil {
		t.Fatal("an unpaired phone connected")
	}
}

func TestRelayOnlyWhileWorking(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	server := relaytest.New()
	defer server.Close()
	h, port := newHost(t, &fakeDesktop{}, server.URL)
	phone, desk := pairPhone(t, ctx, h, port)
	desk.Addrs, desk.LastAddr, desk.Port = nil, "", freePort(t) // away from home

	if _, err := Connect(ctx, phone, "Pixel", &desk, ""); !errors.Is(err, ErrUnreachable) {
		t.Fatalf("reached an idle desktop: %v", err)
	}
	h.SetWorking(true)
	deadline := time.Now().Add(5 * time.Second)
	for !h.Status().Relay && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	s, err := Connect(ctx, phone, "Pixel", &desk, "")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if s.Route != RouteRelay {
		t.Errorf("route %q", s.Route)
	}
	var list wire.ChatList
	if err := s.Request(ctx, wire.TypeList, nil, &list); err != nil || list.Desktop != "Desk" {
		t.Fatalf("list over the relay: %+v %v", list, err)
	}
}
