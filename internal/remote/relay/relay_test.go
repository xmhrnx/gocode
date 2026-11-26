package relay

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"gocode/internal/remote/relay/relaytest"
)

func TestPhoneReachesDesktopThroughRelay(t *testing.T) {
	server := relaytest.New()
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := Phone(ctx, server.URL, "room1"); !errors.Is(err, ErrOffline) {
		t.Fatalf("phone before the desktop: %v", err)
	}
	control, err := Listen(ctx, server.URL, "room1")
	if err != nil {
		t.Fatal(err)
	}
	defer control.Close()
	time.Sleep(50 * time.Millisecond) // the relay registers the desktop

	accepted := make(chan io.ReadWriteCloser, 1)
	go func() {
		id, err := control.Next()
		if err != nil {
			t.Error(err)
			return
		}
		stream, err := Accept(ctx, server.URL, "room1", id)
		if err != nil {
			t.Error(err)
			return
		}
		accepted <- stream
	}()
	phone, err := Phone(ctx, server.URL, "room1")
	if err != nil {
		t.Fatal(err)
	}
	defer phone.Close()
	if _, err := phone.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	desk := <-accepted
	defer desk.Close()
	buf := make([]byte, 4)
	if _, err := io.ReadFull(desk, buf); err != nil || string(buf) != "ping" {
		t.Fatalf("desktop read %q, %v", buf, err)
	}
	if _, err := desk.Write([]byte("pong")); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(phone, buf); err != nil || string(buf) != "pong" {
		t.Fatalf("phone read %q, %v", buf, err)
	}
}

func TestRelayAddress(t *testing.T) {
	got, err := wsURL("https://relay.example.workers.dev/", "phone/r")
	if err != nil || got != "wss://relay.example.workers.dev/v1/phone/r" {
		t.Fatalf("%q %v", got, err)
	}
	if _, err := wsURL("relay.example", "x"); err == nil {
		t.Error("accepted an address without a scheme")
	}
}
