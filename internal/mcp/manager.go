package mcp

import (
	"reflect"
	"sync"

	"gocode/internal/settings"
	"gocode/internal/tools"
)

// Manager connects to the configured servers and exposes their tools. It
// is safe for concurrent use.
type Manager struct {
	mu      sync.Mutex
	servers map[string]*serverConn
	// order is the configured order of server names, for Status.
	order     []string
	listeners []func()
	closed    bool
	// work tracks connect and close goroutines, so Close can wait for them.
	work sync.WaitGroup
}

// NewManager returns a manager with no servers.
func NewManager() *Manager {
	return &Manager{servers: make(map[string]*serverConn)}
}

// Sync connects to newly enabled servers, disconnects removed or disabled
// ones and reconnects servers whose config changed. Connecting happens in
// the background; Sync returns immediately. Servers without a name, and
// repeats of a name, are ignored.
func (m *Manager) Sync(servers []settings.MCPServer) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	wanted := make(map[string]bool, len(servers))
	m.order = make([]string, 0, len(servers))
	for _, config := range servers {
		if config.Name == "" || wanted[config.Name] {
			continue
		}
		wanted[config.Name] = true
		m.order = append(m.order, config.Name)
		m.applyLocked(config)
	}
	for name, conn := range m.servers {
		if !wanted[name] {
			m.retireLocked(conn)
			delete(m.servers, name)
		}
	}
	m.mu.Unlock()
	m.changed()
}

// applyLocked starts a server that is new or whose config changed.
func (m *Manager) applyLocked(config settings.MCPServer) {
	current := m.servers[config.Name]
	if current != nil && reflect.DeepEqual(current.config, config) {
		return
	}
	if current != nil {
		m.retireLocked(current)
	}
	m.servers[config.Name] = m.startLocked(config)
}

// Reconnect restarts the named server; unknown names are ignored.
func (m *Manager) Reconnect(name string) {
	m.mu.Lock()
	conn := m.servers[name]
	if m.closed || conn == nil {
		m.mu.Unlock()
		return
	}
	m.retireLocked(conn)
	m.servers[name] = m.startLocked(conn.config)
	m.mu.Unlock()
	m.changed()
}

// Tools returns the tools of every connected server, named uniquely.
func (m *Manager) Tools() []tools.Tool {
	m.mu.Lock()
	defer m.mu.Unlock()
	used := make(map[string]bool)
	var list []tools.Tool
	for _, name := range m.order {
		conn := m.servers[name]
		if conn.state != StateConnected {
			continue
		}
		for _, info := range conn.tools {
			unique := uniqueName(toolName(name, info.Name), used)
			list = append(list, &serverTool{caller: conn.client, server: name, info: info, name: unique})
		}
	}
	return list
}

// Close disconnects every server and waits for their processes to end.
func (m *Manager) Close() {
	m.mu.Lock()
	m.closed = true
	for name, conn := range m.servers {
		m.retireLocked(conn)
		delete(m.servers, name)
	}
	m.order = nil
	m.mu.Unlock()
	m.work.Wait()
}
