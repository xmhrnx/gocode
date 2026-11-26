package host

import (
	"context"
	"sync"
	"time"

	"gocode/internal/remote/relay"
)

// relayGrace keeps the desktop in its relay room this long after its
// agent stops working, so a phone away from home sees how it ended.
const relayGrace = 2 * time.Minute

// relayRetry is how long the desktop waits before reconnecting.
const relayRetry = 10 * time.Second

// relayLoop keeps the desktop in its relay room until stopped.
type relayLoop struct {
	cancel context.CancelFunc
	url    string
	mu     sync.Mutex
	in     bool
}

// SetWorking tells the host whether an agent is working: the desktop is
// in its relay room only then, and for a while after.
func (h *Host) SetWorking(working bool) {
	h.mu.Lock()
	changed := h.working != working
	h.working = working
	h.mu.Unlock()
	if !changed {
		return
	}
	if working {
		h.syncRelay()
		return
	}
	time.AfterFunc(relayGrace, h.syncRelay)
}

// syncRelay joins or leaves the relay room to match the settings and
// whether an agent works.
func (h *Host) syncRelay() {
	h.mu.Lock()
	want := h.relayURL != "" && h.listener != nil && h.working
	loop := h.relay
	if loop != nil && (!want || loop.url != h.relayURL) {
		h.relay = nil
		loop.cancel()
		loop = nil
	}
	if want && loop == nil {
		ctx, cancel := context.WithCancel(context.Background())
		loop = &relayLoop{cancel: cancel, url: h.relayURL}
		h.relay = loop
		go loop.run(ctx, h)
	}
	h.mu.Unlock()
	h.changed()
}

func (l *relayLoop) connected() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.in
}

func (l *relayLoop) set(in bool, h *Host) {
	l.mu.Lock()
	l.in = in
	l.mu.Unlock()
	h.changed()
}

// run holds the room, taking each phone that arrives, and reconnects
// after a failure.
func (l *relayLoop) run(ctx context.Context, h *Host) {
	for ctx.Err() == nil {
		control, err := relay.Listen(ctx, l.url, h.cfg.Room)
		if err == nil {
			l.set(true, h)
			stop := context.AfterFunc(ctx, func() { control.Close() })
			for {
				id, err := control.Next()
				if err != nil {
					break
				}
				go func() {
					stream, err := relay.Accept(ctx, l.url, h.cfg.Room, id)
					if err == nil {
						h.serve(stream)
					}
				}()
			}
			stop()
			control.Close()
			l.set(false, h)
		}
		select {
		case <-ctx.Done():
		case <-time.After(relayRetry):
		}
	}
}
