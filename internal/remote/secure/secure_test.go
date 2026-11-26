package secure

import (
	"bytes"
	"errors"
	"net"
	"testing"
)

func pair(t *testing.T) (KeyPair, KeyPair) {
	t.Helper()
	phone, err := NewKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	desktop, err := NewKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	return phone, desktop
}

func TestHandshakeAndLargeMessages(t *testing.T) {
	phone, desktop := pair(t)
	a, b := net.Pipe()
	type result struct {
		conn  *Conn
		hello []byte
		err   error
	}
	done := make(chan result, 1)
	go func() {
		conn, hello, err := Accept(b, desktop, func(key, hello []byte) ([]byte, error) {
			if !bytes.Equal(key, phone.Public) {
				t.Error("desktop saw the wrong phone key")
			}
			return []byte("welcome"), nil
		})
		done <- result{conn, hello, err}
	}()
	client, welcome, err := Dial(a, phone, desktop.Public, []byte("hi"))
	if err != nil {
		t.Fatal(err)
	}
	server := <-done
	if server.err != nil || string(server.hello) != "hi" || string(welcome) != "welcome" {
		t.Fatalf("hello %q welcome %q err %v", server.hello, welcome, server.err)
	}

	big := bytes.Repeat([]byte("0123456789"), 30000) // several frames
	go func() { _ = client.Send(big) }()
	got, err := server.conn.Recv()
	if err != nil || !bytes.Equal(got, big) {
		t.Fatalf("got %d bytes, err %v", len(got), err)
	}
	go func() { _ = server.conn.Send(nil) }()
	if got, err := client.Recv(); err != nil || len(got) != 0 {
		t.Fatalf("empty message: %q %v", got, err)
	}
}

func TestWrongDesktopKeyFails(t *testing.T) {
	phone, desktop := pair(t)
	other, _ := NewKeyPair()
	a, b := net.Pipe()
	go func() {
		_, _, _ = Accept(b, desktop, func(_, _ []byte) ([]byte, error) { return nil, nil })
		b.Close()
	}()
	if _, _, err := Dial(a, phone, other.Public, nil); err == nil {
		t.Fatal("handshake with the wrong desktop key succeeded")
	}
}

func TestRefusedPhone(t *testing.T) {
	phone, desktop := pair(t)
	a, b := net.Pipe()
	go func() {
		_, _, _ = Accept(b, desktop, func(_, _ []byte) ([]byte, error) { return nil, errors.New("not paired") })
	}()
	if _, _, err := Dial(a, phone, desktop.Public, nil); !errors.Is(err, ErrRejected) {
		t.Fatalf("err = %v", err)
	}
}
