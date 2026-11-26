package settings

// Phone sets up the phone remote: the GoCode mobile app following and
// steering this desktop's chats.
type Phone struct {
	// Enabled makes the desktop listen for paired phones on the local
	// network.
	Enabled bool `json:"enabled"`
	// Port is where it listens (TCP) and answers discovery (UDP).
	Port int `json:"port"`
	// Relay is the relay Worker's address, for phones away from home;
	// empty for the local network only.
	Relay string `json:"relay"`
}
