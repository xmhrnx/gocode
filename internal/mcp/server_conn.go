package mcp

import (
	"context"
	"encoding/json"
	"time"

	"gocode/internal/settings"
)

// serverConn is the manager's record of one configured server. Its fields
// are guarded by Manager.mu. A record is never reused: reconnecting
// replaces it, so background work checks that its record is still current
// before writing.
type serverConn struct {
	config settings.MCPServer
	state  string
	err    string
	client *client
	tools  []toolInfo
	// cancel stops a connection attempt still in progress.
	cancel context.CancelFunc
}

const (
	// connectTimeout bounds starting a server and listing its tools;
	// npx and uvx may download the server first.
	connectTimeout = 60 * time.Second
	// refreshTimeout bounds re-listing tools after the server changed them.
	refreshTimeout = 30 * time.Second
)

// startLocked creates the record for config and, if the server is
// enabled, starts connecting in the background.
func (m *Manager) startLocked(config settings.MCPServer) *serverConn {
	conn := &serverConn{config: config, state: StateDisabled}
	if !config.Enabled {
		return conn
	}
	ctx, cancel := context.WithCancel(context.Background())
	conn.state, conn.cancel = StateConnecting, cancel
	m.work.Add(1)
	go m.connect(ctx, conn)
	return conn
}

// retireLocked stops conn's connection attempt and closes its client in
// the background, since a server may take a moment to exit.
func (m *Manager) retireLocked(conn *serverConn) {
	if conn.cancel != nil {
		conn.cancel()
	}
	if conn.client == nil {
		return
	}
	m.work.Add(1)
	go func(c *client) {
		defer m.work.Done()
		// The server is being dropped; if killing it fails there is
		// nobody left to report to.
		_ = c.close()
	}(conn.client)
}

// connect starts conn's server, lists its tools and records the outcome.
func (m *Manager) connect(ctx context.Context, conn *serverConn) {
	defer m.work.Done()
	attempt, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	c, list, err := dial(attempt, conn.config, m.eventsFor(conn))
	m.mu.Lock()
	if m.servers[conn.config.Name] != conn {
		m.mu.Unlock()
		if c != nil {
			_ = c.close() // superseded while connecting; nobody needs it
		}
		return
	}
	if err != nil {
		conn.state, conn.err = StateFailed, err.Error()
	} else {
		conn.state, conn.client, conn.tools = StateConnected, c, list
	}
	m.mu.Unlock()
	m.changed()
}

// dial connects to a server and lists its tools.
func dial(ctx context.Context, config settings.MCPServer, events serverEvents) (*client, []toolInfo, error) {
	c, err := connect(ctx, config, events)
	if err != nil {
		return nil, nil, err
	}
	list, err := c.listTools(ctx)
	if err != nil {
		return nil, nil, joinClose(err, c.transport)
	}
	return c, list, nil
}

// eventsFor routes conn's server events back to the manager.
func (m *Manager) eventsFor(conn *serverConn) serverEvents {
	return serverEvents{
		notify: func(method string, _ json.RawMessage) {
			if method == "notifications/tools/list_changed" {
				go m.refreshTools(conn)
			}
		},
		lost: func(err error) { m.markLost(conn, err) },
	}
}
