// Package settings holds the user's preferences and persists them as JSON.
package settings

// Settings is every user preference. Zero values are replaced by defaults
// on load, so new fields work with old files.
type Settings struct {
	// Version is the settings format, for migrating older files.
	Version    int        `json:"version"`
	Appearance Appearance `json:"appearance"`
	Chat       Chat       `json:"chat"`
	Agent      Agent      `json:"agent"`
	Prompts    Prompts    `json:"prompts"`
	// Terminal is read only to move its shell to Agent.Shell; see migrate.
	Terminal   *legacyTerminal    `json:"terminal,omitempty"`
	Providers  []ProviderSettings `json:"providers"`
	Models     Models             `json:"models"`
	Network    Network            `json:"network"`
	General    General            `json:"general"`
	MCPServers []MCPServer        `json:"mcp_servers"`
	Skills     Skills             `json:"skills"`
}

// General holds app-wide behavior.
type General struct {
	DefaultWorkDir   string `json:"default_work_dir"`
	RestoreTabs      bool   `json:"restore_tabs"`
	ConfirmCloseBusy bool   `json:"confirm_close_busy"`
	AutoTitle        bool   `json:"auto_title"`
}

// Appearance holds the look of the app.
type Appearance struct {
	Theme         string `json:"theme"`
	UIFontSize    int    `json:"ui_font_size"`
	CodeFontSize  int    `json:"code_font_size"`
	ChatWidth     int    `json:"chat_width"`
	Compact       bool   `json:"compact"`
	TitleBarStyle string `json:"title_bar_style"` // "windows" or "mac"
	// MaxFPS caps the frame rate; NoFrameCap draws at the display's rate.
	MaxFPS int `json:"max_fps"`
}

// SendKey is the key combination that sends a message.
type SendKey string

// Chat holds composer and transcript behavior.
type Chat struct {
	SendKey           SendKey `json:"send_key"`
	ShowReasoning     bool    `json:"show_reasoning"`
	CollapseReasoning bool    `json:"collapse_reasoning"`
	ExpandToolOutput  bool    `json:"expand_tool_output"`
	ShowTokenUsage    bool    `json:"show_token_usage"`
	MaxAttachmentMB   int     `json:"max_attachment_mb"`
	InlineTextFilesKB int     `json:"inline_text_files_kb"`
}

// Network holds HTTP behavior for provider calls.
type Network struct {
	RequestTimeoutSeconds int `json:"request_timeout_seconds"`
	MaxRetries            int `json:"max_retries"`
}

// NoFrameCap is the MaxFPS value for drawing at the display's refresh
// rate; 0 means unset and becomes the default.
const NoFrameCap = -1

const (
	SendEnter     SendKey = "enter"
	SendCtrlEnter SendKey = "ctrl+enter"
)

// legacyTerminal is the settings of the removed terminal tab.
type legacyTerminal struct {
	Shell string `json:"shell"`
}
