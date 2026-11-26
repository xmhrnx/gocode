// Package secure is the encrypted, authenticated channel between a phone
// and a desktop: a Noise IK handshake, in which the phone already knows
// the desktop's key and proves its own, then length-prefixed encrypted
// messages over any stream (a TCP connection or a relayed WebSocket).
package secure

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/flynn/noise"
)

// KeyPair is a device's long-term identity.
type KeyPair struct {
	Private []byte `json:"private"`
	Public  []byte `json:"public"`
}

// suite is the Noise cipher suite both sides use.
var suite = noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashBLAKE2s)

// NewKeyPair makes a new identity.
func NewKeyPair() (KeyPair, error) {
	key, err := suite.GenerateKeypair(rand.Reader)
	if err != nil {
		return KeyPair{}, fmt.Errorf("generate key: %w", err)
	}
	return KeyPair{Private: key.Private, Public: key.Public}, nil
}

// Valid reports whether the pair holds keys of the right size.
func (k KeyPair) Valid() bool { return len(k.Private) == 32 && len(k.Public) == 32 }

func (k KeyPair) dh() noise.DHKey { return noise.DHKey{Private: k.Private, Public: k.Public} }

// EncodeKey writes a public key as text, for pairing offers and lists.
func EncodeKey(key []byte) string { return base64.RawURLEncoding.EncodeToString(key) }

// DecodeKey reads a public key written by EncodeKey.
func DecodeKey(text string) ([]byte, error) {
	key, err := base64.RawURLEncoding.DecodeString(text)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("not a device key")
	}
	return key, nil
}
