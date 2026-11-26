// Package client is the phone side of the remote: it finds a paired
// desktop on the local network or through the relay, opens the encrypted
// channel and makes requests.
package client

import (
	"gocode/internal/remote/pair"
)

// Desktop is a paired desktop as the phone remembers it.
type Desktop struct {
	Name  string   `json:"name"`
	Key   string   `json:"key"`
	Port  int      `json:"port"`
	Addrs []string `json:"addrs,omitempty"`
	Relay string   `json:"relay,omitempty"`
	Room  string   `json:"room"`
	// LastAddr is where the phone last reached it directly.
	LastAddr string `json:"last_addr,omitempty"`
}

// FromOffer is the desktop a pairing offer describes.
func FromOffer(o pair.Offer) Desktop {
	return Desktop{Name: o.Name, Key: o.Key, Port: o.Port, Addrs: o.Addrs, Relay: o.Relay, Room: o.Room}
}
