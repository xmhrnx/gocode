package client

import (
	"context"
	"errors"
	"io"
	"net"
	"strconv"
	"time"

	"gocode/internal/remote/host"
)

const (
	// findTime is how long the phone looks on the local network.
	findTime = 1500 * time.Millisecond
	// dialTime bounds one direct connection attempt.
	dialTime = 2 * time.Second
)

var errNotNearby = errors.New("the desktop isn't on this network")

// findLocal connects to the desktop on the local network: at the
// addresses it had, and wherever it answers a broadcast.
func findLocal(ctx context.Context, d *Desktop, keyID string) (io.ReadWriteCloser, string, error) {
	ctx, cancel := context.WithTimeout(ctx, findTime+dialTime)
	defer cancel()
	results := make(chan found, 8)
	tries := 0
	try := func(addr string) {
		tries++
		go func() {
			var dialer net.Dialer
			dctx, done := context.WithTimeout(ctx, dialTime)
			defer done()
			conn, err := dialer.DialContext(dctx, "tcp", addr)
			if err != nil {
				results <- found{}
				return
			}
			results <- found{conn, addr}
		}()
	}
	seen := map[string]bool{}
	for _, addr := range append([]string{d.LastAddr}, d.Addrs...) {
		if addr != "" && !seen[addr] {
			seen[addr] = true
			try(addr)
		}
	}
	answers := broadcast(ctx, d.Port, keyID)
	for {
		select {
		case addr, ok := <-answers:
			if !ok {
				answers = nil
				if tries == 0 {
					return nil, "", errNotNearby
				}
				continue
			}
			if !seen[addr] {
				seen[addr] = true
				try(addr)
			}
		case r := <-results:
			tries--
			if r.conn != nil {
				go drainExtra(results, tries)
				return r.conn, r.addr, nil
			}
			if tries == 0 && answers == nil {
				return nil, "", errNotNearby
			}
		case <-ctx.Done():
			go drainExtra(results, tries)
			return nil, "", errNotNearby
		}
	}
}

// found is one connection attempt's outcome.
type found struct {
	conn net.Conn
	addr string
}

// drainExtra waits for the attempts still going and closes the
// connections they make, which nobody needs any more.
func drainExtra(results <-chan found, left int) {
	for ; left > 0; left-- {
		if r := <-results; r.conn != nil {
			r.conn.Close()
		}
	}
}

// broadcast asks the network for the desktop with keyID and sends the
// address of each answer, until findTime passes.
func broadcast(ctx context.Context, port int, keyID string) <-chan string {
	out := make(chan string, 4)
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		close(out)
		return out
	}
	if port <= 0 {
		port = host.DefaultPort
	}
	target := &net.UDPAddr{IP: net.IPv4bcast, Port: port}
	_, _ = conn.WriteTo(host.FindMessage(keyID), target)
	_ = conn.SetReadDeadline(time.Now().Add(findTime))
	go func() { <-ctx.Done(); conn.Close() }()
	go func() {
		defer close(out)
		defer conn.Close()
		buf := make([]byte, 256)
		for {
			n, from, err := conn.ReadFrom(buf)
			if err != nil {
				return
			}
			udp, ok := from.(*net.UDPAddr)
			if p, match := host.ParseHere(buf[:n], keyID); ok && match {
				out <- net.JoinHostPort(udp.IP.String(), strconv.Itoa(p))
			}
		}
	}()
	return out
}
