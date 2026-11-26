package secure

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/flynn/noise"
)

// Conn is an open channel: Send and Recv whole messages, encrypted.
// Send is safe from several goroutines; Recv must be called from one.
type Conn struct {
	stream io.ReadWriteCloser
	sendMu sync.Mutex
	send   *noise.CipherState
	recv   *noise.CipherState
	// Peer is the other side's public key.
	Peer []byte
}

const (
	// chunkSize is the most plaintext one frame carries; Noise allows
	// 65535 bytes of ciphertext with its 16-byte tag.
	chunkSize = 65000
	// MaxMessage bounds a whole message, across its frames.
	MaxMessage = 16 << 20
	// moreFrames starts a frame that a later one continues.
	moreFrames, lastFrame = 1, 0
)

// ErrTooLarge is returned for a message over MaxMessage.
var ErrTooLarge = errors.New("message too large")

// Send encrypts msg and writes it.
func (c *Conn) Send(msg []byte) error {
	if len(msg) > MaxMessage {
		return ErrTooLarge
	}
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	for first := true; first || len(msg) > 0; first = false {
		n := min(len(msg), chunkSize)
		flag := byte(lastFrame)
		if n < len(msg) {
			flag = moreFrames
		}
		frame, err := c.send.Encrypt(nil, nil, append([]byte{flag}, msg[:n]...))
		if err != nil {
			return err
		}
		if err := writeFrame(c.stream, frame); err != nil {
			return err
		}
		msg = msg[n:]
	}
	return nil
}

// Recv reads and decrypts the next message.
func (c *Conn) Recv() ([]byte, error) {
	var msg []byte
	for {
		frame, err := readFrame(c.stream)
		if err != nil {
			return nil, err
		}
		plain, err := c.recv.Decrypt(nil, nil, frame)
		if err != nil || len(plain) == 0 {
			return nil, fmt.Errorf("decrypt: corrupt frame")
		}
		msg = append(msg, plain[1:]...)
		if len(msg) > MaxMessage {
			return nil, ErrTooLarge
		}
		if plain[0] == lastFrame {
			return msg, nil
		}
	}
}

// Close closes the stream underneath.
func (c *Conn) Close() error { return c.stream.Close() }

// writeFrame writes a 4-byte big-endian length, then the frame.
func writeFrame(w io.Writer, frame []byte) error {
	buf := make([]byte, 4+len(frame))
	binary.BigEndian.PutUint32(buf, uint32(len(frame)))
	copy(buf[4:], frame)
	_, err := w.Write(buf)
	return err
}

func readFrame(r io.Reader) ([]byte, error) {
	var head [4]byte
	if _, err := io.ReadFull(r, head[:]); err != nil {
		return nil, err
	}
	n := binary.BigEndian.Uint32(head[:])
	if n > 65535 {
		return nil, fmt.Errorf("frame of %d bytes is too large", n)
	}
	frame := make([]byte, n)
	_, err := io.ReadFull(r, frame)
	return frame, err
}
