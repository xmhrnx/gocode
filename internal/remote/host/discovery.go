package host

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

// Discovery: a phone on the same network broadcasts "GOCODE1 FIND <id>"
// to the desktop's port over UDP; the desktop with that id answers
// "GOCODE1 HERE <id> <port>", and the phone connects to it over TCP.
const discoveryPrefix = "GOCODE1"

// KeyID is a short public name for a desktop key, for discovery: it says
// which desktop is meant without sending the key itself.
func KeyID(key []byte) string {
	sum := sha256.Sum256(key)
	return hex.EncodeToString(sum[:8])
}

// FindMessage is what a phone broadcasts to find the desktop with id.
func FindMessage(id string) []byte { return []byte(discoveryPrefix + " FIND " + id) }

// ParseHere reads a desktop's answer, reporting its port when it is the
// desktop with id.
func ParseHere(msg []byte, id string) (port int, ok bool) {
	var gotID string
	n, err := fmt.Sscanf(string(msg), discoveryPrefix+" HERE %s %d", &gotID, &port)
	return port, err == nil && n == 2 && gotID == id
}

// answerDiscovery answers finds for this desktop until the socket closes.
func answerDiscovery(conn net.PacketConn, id string, port int) {
	find := discoveryPrefix + " FIND " + id
	here := []byte(fmt.Sprintf("%s HERE %s %d", discoveryPrefix, id, port))
	buf := make([]byte, 256)
	for {
		n, from, err := conn.ReadFrom(buf)
		if err != nil {
			return
		}
		if strings.TrimSpace(string(buf[:n])) == find {
			_, _ = conn.WriteTo(here, from)
		}
	}
}
