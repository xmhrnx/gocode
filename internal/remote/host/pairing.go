package host

import (
	"context"
	"crypto/subtle"
	"errors"
	"time"

	"gocode/internal/remote/pair"
	"gocode/internal/remote/secure"
)

// pairingWindow is how long a pairing code works.
const pairingWindow = 10 * time.Minute

// pending is a pairing in progress: the token a phone must bring, and
// where to report the phone that did.
type pending struct {
	token   string
	expires time.Time
	result  chan PairResult
}

// PairResult is how a pairing ended: the phone that paired, or why none
// did.
type PairResult struct {
	Device string
	Err    error
}

// ErrPairingExpired is the result of a code nobody used in time.
var ErrPairingExpired = errors.New("the pairing code expired")

// Pair starts pairing a phone: it returns the code to show and a channel
// that reports the result. server is the Magic Wormhole server ("" for
// the public one). A new pairing replaces one in progress. The desktop
// must be listening, so the phone can connect after receiving the offer.
func (h *Host) Pair(ctx context.Context, server string) (string, <-chan PairResult, error) {
	h.mu.Lock()
	port, listening, relay := h.port, h.listener != nil, h.relayURL
	h.mu.Unlock()
	if !listening {
		return "", nil, errors.New("turn on the phone remote first")
	}
	p := &pending{token: pair.NewSecret(), expires: time.Now().Add(pairingWindow), result: make(chan PairResult, 1)}
	offer := pair.Offer{
		Version: pair.Version, Name: h.cfg.Name, Key: secure.EncodeKey(h.cfg.Key.Public),
		Token: p.token, Addrs: localAddrs(port), Port: port, Relay: relay, Room: h.cfg.Room,
	}
	ctx, cancel := context.WithTimeout(ctx, pairingWindow)
	code, sent, err := pair.Send(ctx, server, offer)
	if err != nil {
		cancel()
		return "", nil, err
	}
	h.mu.Lock()
	h.pairing = p
	h.mu.Unlock()
	go func() {
		defer cancel()
		if err := <-sent; err != nil {
			h.endPairing(p, PairResult{Err: err})
			return
		}
		select {
		case <-time.After(time.Until(p.expires)):
			h.endPairing(p, PairResult{Err: ErrPairingExpired})
		case <-ctx.Done():
		}
	}()
	return code, p.result, nil
}

// CancelPairing stops the pairing in progress.
func (h *Host) CancelPairing() {
	h.mu.Lock()
	p := h.pairing
	h.mu.Unlock()
	if p != nil {
		h.endPairing(p, PairResult{Err: context.Canceled})
	}
}

// redeem reports whether token is the pairing in progress, and if so
// pairs the phone and ends the pairing.
func (h *Host) redeem(token, name string, key []byte) bool {
	h.mu.Lock()
	p := h.pairing
	h.mu.Unlock()
	if p == nil || token == "" || time.Now().After(p.expires) ||
		subtle.ConstantTimeCompare([]byte(token), []byte(p.token)) != 1 {
		return false
	}
	if err := h.cfg.Devices.add(name, key); err != nil {
		h.endPairing(p, PairResult{Err: err})
		return false
	}
	h.endPairing(p, PairResult{Device: name})
	return true
}

func (h *Host) endPairing(p *pending, result PairResult) {
	h.mu.Lock()
	current := h.pairing == p
	if current {
		h.pairing = nil
	}
	h.mu.Unlock()
	if current {
		p.result <- result
		h.changed()
	}
}
