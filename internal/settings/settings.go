// Package settings holds the user's preferences and persists them as JSON.
package settings

// Settings is every user preference. Zero values are replaced by defaults
// on load, so new fields work with old files.
type Settings struct {
	// Version is the settings format, for migrating older files.
	Version    int                `json:"version"`
	Appearance Appearance         `json:"appearance"`
	Chat       Chat               `json:"chat"`
	Agent      Agent              `json:"agent"`
	Prompts    Prompts            `json:"prompts"`
	Terminal   Terminal           `json:"terminal"`
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
}

// SendKey is the key combination that sends a message.
type SendKey string

// Chat holds composer and transcript behavior.
type Chat struct {
	SendKey           SendKey `json:"send_key"`
	ShowReasoning     bool    `json:"show_reasoning"`
	ExpandToolOutput  bool    `json:"expand_tool_output"`
	ShowTokenUsage    bool    `json:"show_token_usage"`
	MaxAttachmentMB   int     `json:"max_attachment_mb"`
	InlineTextFilesKB int     `json:"inline_text_files_kb"`
}

// Terminal holds the command tab's behavior.
type Terminal struct {
	Shell         string `json:"shell"`
	MaxBlockLines int    `json:"max_block_lines"`
}

// Network holds HTTP behavior for provider calls.
type Network struct {
	RequestTimeoutSeconds int `json:"request_timeout_seconds"`
	MaxRetries            int `json:"max_retries"`
}

const (
	SendEnter     SendKey = "enter"
	SendCtrlEnter SendKey = "ctrl+enter"
)
