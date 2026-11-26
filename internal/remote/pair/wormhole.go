package pair

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/psanford/wormhole-william/wormhole"
)

// AppID keeps GoCode's codes apart from other Magic Wormhole apps on the
// same server.
const AppID = "gocode.dev/remote-pairing"

// DefaultServer is the public Magic Wormhole server, used only to swap
// the offer; the code and the server never see the devices' secrets.
const DefaultServer = "ws://relay.magic-wormhole.io:4000/v1"

// maxOffer bounds what a phone reads from a wormhole.
const maxOffer = 64 << 10

// Send offers o through server (empty: the public one) and returns the
// code to show, and a channel that reports when the phone received it.
func Send(ctx context.Context, server string, o Offer) (code string, done <-chan error, err error) {
	text, err := o.Encode()
	if err != nil {
		return "", nil, err
	}
	client := wormhole.Client{AppID: AppID, RendezvousURL: server}
	code, status, err := client.SendText(ctx, text)
	if err != nil {
		return "", nil, fmt.Errorf("couldn't reach the pairing server: %w", err)
	}
	result := make(chan error, 1)
	go func() {
		r := <-status
		if r.Error == nil && !r.OK {
			r.Error = fmt.Errorf("pairing didn't finish")
		}
		result <- r.Error
	}()
	return code, result, nil
}

// Receive takes the offer sent under code through server (empty: the
// public one).
func Receive(ctx context.Context, server, code string) (Offer, error) {
	client := wormhole.Client{AppID: AppID, RendezvousURL: server}
	msg, err := client.Receive(ctx, strings.TrimSpace(code))
	if err != nil {
		return Offer{}, fmt.Errorf("couldn't get the offer: %w", err)
	}
	if msg.Type != wormhole.TransferText {
		return Offer{}, fmt.Errorf("the code isn't for a GoCode desktop")
	}
	data, err := io.ReadAll(io.LimitReader(msg, maxOffer))
	if err != nil {
		return Offer{}, err
	}
	return Decode(string(data))
}
