// Package pair is how a phone is paired with a desktop once: the desktop
// sends an Offer through Magic Wormhole, under a short code it shows; the
// phone, given the code, receives it and then connects with the token in
// it, which the desktop accepts once.
package pair

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Version is the offer format; a phone refuses offers newer than it
// knows.
const Version = 1

// Offer is what a phone needs to reach a desktop: who it is, where it
// listens and, for the first connection, the pairing token.
type Offer struct {
	Version int    `json:"v"`
	Name    string `json:"name"`
	// Key is the desktop's public key (secure.EncodeKey).
	Key string `json:"key"`
	// Token lets one phone in, once, until it expires.
	Token string `json:"token"`
	// Addrs are the desktop's addresses on its local networks
	// (host:port), and Port the port it listens on.
	Addrs []string `json:"addrs,omitempty"`
	Port  int      `json:"port"`
	// Relay is the relay Worker's address, or "" without one, and Room
	// the desktop's room there.
	Relay string `json:"relay,omitempty"`
	Room  string `json:"room"`
}

// Encode writes the offer for sending.
func (o Offer) Encode() (string, error) {
	data, err := json.Marshal(o)
	return string(data), err
}

// Decode reads an offer a desktop sent.
func Decode(text string) (Offer, error) {
	var o Offer
	if err := json.Unmarshal([]byte(text), &o); err != nil {
		return Offer{}, fmt.Errorf("not a GoCode pairing offer")
	}
	if o.Version > Version {
		return Offer{}, fmt.Errorf("this desktop needs a newer GoCode mobile app")
	}
	if o.Key == "" || o.Token == "" {
		return Offer{}, fmt.Errorf("the pairing offer is incomplete")
	}
	return o, nil
}

// NewSecret returns 16 random bytes as hex: a pairing token or a relay
// room.
func NewSecret() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err) // crypto/rand doesn't fail on supported systems
	}
	return hex.EncodeToString(b)
}
