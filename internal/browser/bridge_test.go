package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/websocket"

	"gocode/internal/tools"
)

// freePort returns a port nothing listens on.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func startBridge(t *testing.T) (*Bridge, int) {
	t.Helper()
	port := freePort(t)
	b := New()
	b.Sync(true, port)
	t.Cleanup(b.Close)
	if st := b.Status(); !st.Listening {
		t.Fatalf("not listening: %+v", st)
	}
	return b, port
}

func dial(port int, origin string) (*websocket.Conn, error) {
	return websocket.Dial(fmt.Sprintf("ws://127.0.0.1:%d%s", port, Path), "", origin)
}

// fakeExtension answers every command with the method's name and args.
func fakeExtension(t *testing.T, port int) *websocket.Conn {
	t.Helper()
	ws, err := dial(port, "chrome-extension://abc")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ws.Close() })
	if err := websocket.JSON.Send(ws, message{Type: "hello", Browser: "Test"}); err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			var req struct {
				ID     int64           `json:"id"`
				Method string          `json:"method"`
				Params json.RawMessage `json:"params"`
			}
			if err := websocket.JSON.Receive(ws, &req); err != nil {
				return
			}
			if req.Method == "fail" {
				_ = websocket.JSON.Send(ws, map[string]any{"id": req.ID, "error": "boom"})
				continue
			}
			_ = websocket.JSON.Send(ws, map[string]any{"id": req.ID, "result": req.Method + " " + string(req.Params)})
		}
	}()
	return ws
}

func waitConnected(t *testing.T, b *Bridge, want bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for b.Status().Connected != want {
		if time.Now().After(deadline) {
			t.Fatalf("connected never became %v", want)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestCallRoundTrip(t *testing.T) {
	b, port := startBridge(t)
	fakeExtension(t, port)
	waitConnected(t, b, true)
	if got := b.Status().Browser; got != "Test" {
		t.Fatalf("browser = %q", got)
	}
	result, err := b.Call(context.Background(), "snapshot", map[string]int{"tab": 3})
	if err != nil {
		t.Fatal(err)
	}
	var text string
	_ = json.Unmarshal(result, &text)
	if text != `snapshot {"tab":3}` {
		t.Fatalf("result = %q", text)
	}
	if _, err := b.Call(context.Background(), "fail", nil); err == nil || err.Error() != "boom" {
		t.Fatalf("err = %v", err)
	}
}

func TestRejectsWebPages(t *testing.T) {
	_, port := startBridge(t)
	if _, err := dial(port, "https://evil.example"); err == nil {
		t.Fatal("a web page connected")
	}
}

func TestNotConnected(t *testing.T) {
	b, _ := startBridge(t)
	if _, err := b.Call(context.Background(), "tabs", nil); err != ErrNotConnected {
		t.Fatalf("err = %v", err)
	}
	res := Tool{Bridge: b}.Run(context.Background(), tools.Env{}, json.RawMessage(`{"action":"tabs"}`))
	if !res.IsError || !strings.Contains(res.Output, "No browser is connected") {
		t.Fatalf("result = %+v", res)
	}
}

func TestDisconnectFailsPendingCall(t *testing.T) {
	b, port := startBridge(t)
	ws, err := dial(port, "chrome-extension://abc")
	if err != nil {
		t.Fatal(err)
	}
	_ = websocket.JSON.Send(ws, message{Type: "hello"})
	waitConnected(t, b, true)
	go func() {
		var req request
		_ = websocket.JSON.Receive(ws, &req)
		ws.Close()
	}()
	if _, err := b.Call(context.Background(), "tabs", nil); err == nil {
		t.Fatal("call succeeded after the browser left")
	}
	waitConnected(t, b, false)
}

func TestSyncStopsListening(t *testing.T) {
	b, port := startBridge(t)
	b.Sync(false, port)
	if b.Status().Listening {
		t.Fatal("still listening")
	}
	if _, err := dial(port, "chrome-extension://abc"); err == nil {
		t.Fatal("connected after stopping")
	}
}

func TestToolRejectsUnknownAction(t *testing.T) {
	res := Tool{Bridge: New()}.Run(context.Background(), tools.Env{}, json.RawMessage(`{"action":"fly"}`))
	if !res.IsError {
		t.Fatal("unknown action accepted")
	}
}

func TestToolSummary(t *testing.T) {
	got := Tool{}.Summary(json.RawMessage(`{"action":"type","text":"hello","ref":4,"key":"Enter"}`))
	if got != `Browser: type "hello" [4] Enter` {
		t.Fatalf("summary = %q", got)
	}
}

func TestTakenPortIsExplainedAndRetried(t *testing.T) {
	port := freePort(t)
	taken, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatal(err)
	}
	b := New()
	b.Sync(true, port)
	t.Cleanup(b.Close)
	if st := b.Status(); st.Listening || !strings.Contains(st.Err, "in use") {
		t.Fatalf("status = %+v", st)
	}
	taken.Close()
	b.retryListen(port) // what the timer does
	if st := b.Status(); !st.Listening || st.Err != "" {
		t.Fatalf("after the port was freed: %+v", st)
	}
}
