package settings

// MCPServer is one Model Context Protocol server whose tools the agent can
// use. A server is reached either by starting Command (stdio) or at URL
// (Streamable HTTP).
type MCPServer struct {
	Name      string            `json:"name"`
	Transport string            `json:"transport"` // "stdio" or "http"
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	URL       string            `json:"url,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Enabled   bool              `json:"enabled"`
}

// Skills holds skill preferences. Installed skills live as folders in the
// skills directory; only the ones switched off are recorded here.
type Skills struct {
	Disabled []string `json:"disabled,omitempty"`
}
