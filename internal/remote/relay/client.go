// Package relay connects a phone and a desktop through the relay Worker
// (relay/ in this repository) when they can't reach each other directly.
// The Worker only pipes bytes between the two: the channel over it is
// end-to-end encrypted.
//
// A desktop holds a control connection to its room while its agent
// works; the Worker tells it of each phone that arrives, and the desktop
// opens a connection to take that phone's stream.
package relay

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/websocket"
)

// ErrOffline is returned to a phone whose desktop isn't connected to the
// relay.
var ErrOffline = errors.New("the desktop isn't reachable through the relay: it's off, or its agent isn't working")

// origin is sent with every connection; the Worker doesn't check it.
const origin = "https://gocode.app"

// Phone opens a stream to the desktop in room.
func Phone(ctx context.Context, base, room string) (io.ReadWriteCloser, error) {
	ws, err := dial(ctx, base, "phone/"+room)
	var dialErr *websocket.DialError
	if errors.As(err, &dialErr) && dialErr.Err == websocket.ErrBadStatus {
		return nil, ErrOffline // the Worker answers 404 when no desktop is in the room
	}
	return ws, err
}

// Control is a desktop's connection to its room.
type Control struct{ ws *websocket.Conn }

// Listen connects the desktop to its room.
func Listen(ctx context.Context, base, room string) (*Control, error) {
	ws, err := dial(ctx, base, "desktop/"+room)
	if err != nil {
		return nil, err
	}
	return &Control{ws: ws}, nil
}

// Next waits for a phone to arrive and returns its connection ID.
func (c *Control) Next() (string, error) {
	for {
		var msg struct{ Type, ID string }
		if err := websocket.JSON.Receive(c.ws, &msg); err != nil {
			return "", err
		}
		if msg.Type == "incoming" && msg.ID != "" {
			return msg.ID, nil
		}
	}
}

// Close disconnects the desktop from its room.
func (c *Control) Close() error { return c.ws.Close() }

// Accept takes the stream of the phone that arrived as id.
func Accept(ctx context.Context, base, room, id string) (io.ReadWriteCloser, error) {
	return dial(ctx, base, "accept/"+room+"/"+url.PathEscape(id))
}

func dial(ctx context.Context, base, path string) (*websocket.Conn, error) {
	target, err := wsURL(base, path)
	if err != nil {
		return nil, err
	}
	config, err := websocket.NewConfig(target, origin)
	if err != nil {
		return nil, err
	}
	ws, err := config.DialContext(ctx)
	if err != nil {
		return nil, err
	}
	ws.PayloadType = websocket.BinaryFrame
	return ws, nil
}

// wsURL turns the relay's address (https://… or http://…) into the
// WebSocket address of path.
func wsURL(base, path string) (string, error) {
	u, err := url.Parse(strings.TrimRight(strings.TrimSpace(base), "/"))
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("the relay address %q isn't a web address", base)
	}
	switch u.Scheme {
	case "https", "wss":
		u.Scheme = "wss"
	case "http", "ws":
		u.Scheme = "ws"
	default:
		return "", fmt.Errorf("the relay address %q isn't a web address", base)
	}
	return u.String() + "/v1/" + path, nil
}
