package mcp

import "context"

// refreshTools re-lists conn's tools after the server announced a change.
func (m *Manager) refreshTools(conn *serverConn) {
	m.mu.Lock()
	c := conn.client
	m.mu.Unlock()
	if c == nil {
		return // still connecting; the initial listing is under way
	}
	ctx, cancel := context.WithTimeout(context.Background(), refreshTimeout)
	defer cancel()
	list, err := c.listTools(ctx)
	m.mu.Lock()
	if m.servers[conn.config.Name] != conn || conn.client != c {
		m.mu.Unlock()
		return
	}
	if err != nil {
		conn.err = "refresh tools: " + err.Error()
	} else {
		conn.tools, conn.err = list, ""
	}
	m.mu.Unlock()
	m.changed()
}

// markLost records that a connected server went away, e.g. its process
// crashed, so the UI shows why and its tools disappear.
func (m *Manager) markLost(conn *serverConn, err error) {
	m.mu.Lock()
	if m.servers[conn.config.Name] != conn || conn.state != StateConnected {
		m.mu.Unlock()
		return
	}
	conn.state, conn.err, conn.tools = StateFailed, err.Error(), nil
	m.retireLocked(conn)
	conn.client = nil
	m.mu.Unlock()
	m.changed()
}
