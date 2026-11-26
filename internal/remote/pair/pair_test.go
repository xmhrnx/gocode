package pair

import (
	"context"
	"testing"
	"time"

	"github.com/psanford/wormhole-william/rendezvous/rendezvousservertest"
	"github.com/psanford/wormhole-william/wormhole"
)

func TestOfferTravelsThroughWormhole(t *testing.T) {
	server := rendezvousservertest.NewServer()
	defer server.Close()
	wormhole.DefaultTransitRelayAddress = "" // text needs no transit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sent := Offer{Version: Version, Name: "Desk", Key: "k", Token: NewSecret(), Port: 47822, Room: NewSecret(),
		Addrs: []string{"192.168.1.5:47822"}}
	code, done, err := Send(ctx, server.WebSocketURL(), sent)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Receive(ctx, server.WebSocketURL(), code)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Desk" || got.Token != sent.Token || got.Addrs[0] != sent.Addrs[0] {
		t.Errorf("got %+v", got)
	}
	if err := <-done; err != nil {
		t.Errorf("sender: %v", err)
	}
}

func TestDecodeRefusesNewerOffers(t *testing.T) {
	if _, err := Decode(`{"v":2,"key":"k","token":"t"}`); err == nil {
		t.Error("accepted a newer offer")
	}
}
