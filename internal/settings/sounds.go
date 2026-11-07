package settings

// Sounds sets which events play a sound, and how loud.
type Sounds struct {
	Enabled bool `json:"enabled"`
	// Done plays when the agent finishes, Permission when it asks to run
	// something, Errors when a run fails.
	Done       bool `json:"done"`
	Permission bool `json:"permission"`
	Errors     bool `json:"errors"`
	// Volume is a percentage.
	Volume int `json:"volume"`
}
