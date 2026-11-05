package mcp

// ServerStatus is the state of one configured server, for the settings UI.
type ServerStatus struct {
	Name string
	// State is one of StateConnecting, StateConnected, StateFailed or
	// StateDisabled.
	State string
	// Err explains a failure; it may also be set on a connected server
	// whose tool list could not be refreshed.
	Err string
	// ToolNames are the server's own names for its tools.
	ToolNames []string
}

// Server states reported in ServerStatus.State.
const (
	StateConnecting = "connecting"
	StateConnected  = "connected"
	StateFailed     = "failed"
	StateDisabled   = "disabled"
)

// Status returns the state of every configured server, in config order.
func (m *Manager) Status() []ServerStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	statuses := make([]ServerStatus, 0, len(m.order))
	for _, name := range m.order {
		conn := m.servers[name]
		status := ServerStatus{Name: name, State: conn.state, Err: conn.err}
		for _, info := range conn.tools {
			status.ToolNames = append(status.ToolNames, info.Name)
		}
		statuses = append(statuses, status)
	}
	return statuses
}

// OnChange registers fn to be called, from any goroutine, whenever server
// states or tool lists change.
func (m *Manager) OnChange(fn func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, fn)
}

// changed calls the listeners. It runs without the lock held, so a
// listener may call back into the manager.
func (m *Manager) changed() {
	m.mu.Lock()
	listeners := append([]func(){}, m.listeners...)
	m.mu.Unlock()
	for _, fn := range listeners {
		fn()
	}
}
