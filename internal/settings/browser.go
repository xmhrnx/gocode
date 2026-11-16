package settings

// Browser sets up the bridge the GoCode browser extension connects to,
// which gives the agent the browser tool.
type Browser struct {
	// Enabled makes the app listen for the extension on 127.0.0.1.
	Enabled bool `json:"enabled"`
	// Port is where the app listens; the extension must use the same.
	Port int `json:"port"`
}
