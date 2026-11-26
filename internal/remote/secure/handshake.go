package secure

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"github.com/flynn/noise"
)

// prologue binds the handshake to this protocol and version.
var prologue = []byte("gocode-remote/1")

// ErrRejected is returned to a phone the desktop didn't accept.
var ErrRejected = errors.New("the desktop didn't accept this device")

// Dial opens the channel from the phone's side: self is the phone's key
// pair and desktop the desktop's public key, known from pairing. hello
// goes encrypted with the first message (the phone's name, and a pairing
// token the first time); welcome is what the desktop answered.
func Dial(stream io.ReadWriteCloser, self KeyPair, desktop, hello []byte) (conn *Conn, welcome []byte, err error) {
	hs, err := noise.NewHandshakeState(noise.Config{
		CipherSuite: suite, Random: rand.Reader, Pattern: noise.HandshakeIK, Initiator: true,
		Prologue: prologue, StaticKeypair: self.dh(), PeerStatic: desktop,
	})
	if err != nil {
		return nil, nil, err
	}
	first, _, _, err := hs.WriteMessage(nil, hello)
	if err != nil {
		return nil, nil, err
	}
	if err := writeFrame(stream, first); err != nil {
		return nil, nil, err
	}
	second, err := readFrame(stream)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrRejected, err)
	}
	welcome, send, recv, err := hs.ReadMessage(nil, second)
	if err != nil {
		return nil, nil, fmt.Errorf("handshake: %w", err)
	}
	return &Conn{stream: stream, send: send, recv: recv, Peer: desktop}, welcome, nil
}

// Accept opens the channel from the desktop's side. accept sees the
// phone's key and hello, and returns the welcome to send back, or an
// error to refuse the phone, which closes the stream.
func Accept(stream io.ReadWriteCloser, self KeyPair, accept func(phone, hello []byte) ([]byte, error)) (*Conn, []byte, error) {
	hs, err := noise.NewHandshakeState(noise.Config{
		CipherSuite: suite, Random: rand.Reader, Pattern: noise.HandshakeIK, Initiator: false,
		Prologue: prologue, StaticKeypair: self.dh(),
	})
	if err != nil {
		return nil, nil, err
	}
	first, err := readFrame(stream)
	if err != nil {
		return nil, nil, err
	}
	hello, _, _, err := hs.ReadMessage(nil, first)
	if err != nil {
		return nil, nil, fmt.Errorf("handshake: %w", err)
	}
	phone := hs.PeerStatic()
	welcome, err := accept(phone, hello)
	if err != nil {
		stream.Close()
		return nil, nil, err
	}
	second, recv, send, err := hs.WriteMessage(nil, welcome)
	if err != nil {
		return nil, nil, err
	}
	if err := writeFrame(stream, second); err != nil {
		return nil, nil, err
	}
	return &Conn{stream: stream, send: send, recv: recv, Peer: phone}, hello, nil
}
