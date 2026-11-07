package settings

// Policy is what happens when the agent wants to use a kind of tool.
type Policy string

// Agent holds the agent loop's behavior.
type Agent struct {
	// MaxSteps stops a run after this many model calls; 0 means no limit.
	MaxSteps int `json:"max_steps"`
	// MaxOutputTokens caps one model reply; 0 leaves it to the provider.
	MaxOutputTokens int `json:"max_output_tokens"`
	// AutoCompact summarizes the conversation when it nears the model's
	// context window, and when the provider rejects it as too long.
	AutoCompact bool `json:"auto_compact"`
	// CompactAtPercent is how full the context may get before compacting.
	CompactAtPercent int `json:"compact_at_percent"`
	// ContextWindow is used for models whose provider doesn't report one.
	ContextWindow   int     `json:"context_window"`
	Temperature     float64 `json:"temperature"`
	UseTemperature  bool    `json:"use_temperature"`
	ReasoningEffort string  `json:"reasoning_effort"`
	// Shell runs the agent's commands; empty picks the platform default.
	Shell                 string   `json:"shell"`
	CommandTimeoutSeconds int      `json:"command_timeout_seconds"`
	MaxToolOutputKB       int      `json:"max_tool_output_kb"`
	LoadProjectRules      bool     `json:"load_project_rules"`
	ReadPolicy            Policy   `json:"read_policy"`
	WritePolicy           Policy   `json:"write_policy"`
	ExecutePolicy         Policy   `json:"execute_policy"`
	NetworkPolicy         Policy   `json:"network_policy"`
	DisabledTools         []string `json:"disabled_tools"`
}

// Prompts holds the system prompts. Empty fields use the built-in ones.
type Prompts struct {
	Agent       string `json:"agent"`
	Chat        string `json:"chat"`
	Title       string `json:"title"`
	Compact     string `json:"compact"`
	CustomRules string `json:"custom_rules"`
}

const (
	PolicyAsk   Policy = "ask"
	PolicyAllow Policy = "allow"
	PolicyDeny  Policy = "deny"
)
