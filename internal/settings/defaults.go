package settings

// Default returns the settings a fresh install starts with.
func Default() Settings {
	return Settings{
		Version: currentVersion,
		General: General{RestoreTabs: true, ConfirmCloseBusy: true, AutoTitle: true},
		Appearance: Appearance{
			Theme: "system", UIFontSize: 14, CodeFontSize: 13, ChatWidth: 820,
			TitleBarStyle: "windows", MaxFPS: 120,
		},
		Chat: Chat{
			SendKey: SendEnter, ShowReasoning: true, ShowTokenUsage: true,
			MaxAttachmentMB: 20, InlineTextFilesKB: 256,
		},
		Agent: Agent{
			AutoCompact: true, CompactAtPercent: 80, ContextWindow: 128000, Temperature: 0.2,
			CommandTimeoutSeconds: 120, MaxToolOutputKB: 30, LoadProjectRules: true,
			ReadPolicy: PolicyAllow, WritePolicy: PolicyAsk,
			ExecutePolicy: PolicyAsk, NetworkPolicy: PolicyAsk,
		},
		Network:   Network{RequestTimeoutSeconds: 300, MaxRetries: 3},
		WebSearch: WebSearch{Results: 5},
		Sounds:    Sounds{Enabled: true, Done: true, Permission: true, Errors: true, Volume: 60},
	}
}

// fillDefaults replaces zero values that have no meaning as zero with the
// defaults, so a file written by an older version still loads sensibly.
func (s *Settings) fillDefaults() {
	def := Default()
	s.fillWebSearchDefaults(def)
	setString(&s.Appearance.Theme, def.Appearance.Theme)
	setString(&s.Appearance.TitleBarStyle, def.Appearance.TitleBarStyle)
	setInt(&s.Appearance.UIFontSize, def.Appearance.UIFontSize)
	setInt(&s.Appearance.CodeFontSize, def.Appearance.CodeFontSize)
	setInt(&s.Appearance.ChatWidth, def.Appearance.ChatWidth)
	if s.Appearance.MaxFPS == 0 {
		s.Appearance.MaxFPS = def.Appearance.MaxFPS
	}
	if s.Chat.SendKey == "" {
		s.Chat.SendKey = def.Chat.SendKey
	}
	setInt(&s.Chat.MaxAttachmentMB, def.Chat.MaxAttachmentMB)
	setInt(&s.Chat.InlineTextFilesKB, def.Chat.InlineTextFilesKB)
	setInt(&s.Agent.CompactAtPercent, def.Agent.CompactAtPercent)
	setInt(&s.Agent.ContextWindow, def.Agent.ContextWindow)
	setInt(&s.Agent.CommandTimeoutSeconds, def.Agent.CommandTimeoutSeconds)
	setInt(&s.Agent.MaxToolOutputKB, def.Agent.MaxToolOutputKB)
	setPolicy(&s.Agent.ReadPolicy, def.Agent.ReadPolicy)
	setPolicy(&s.Agent.WritePolicy, def.Agent.WritePolicy)
	setPolicy(&s.Agent.ExecutePolicy, def.Agent.ExecutePolicy)
	setPolicy(&s.Agent.NetworkPolicy, def.Agent.NetworkPolicy)
	setInt(&s.Network.RequestTimeoutSeconds, def.Network.RequestTimeoutSeconds)
}

func setString(field *string, fallback string) {
	if *field == "" {
		*field = fallback
	}
}

func (s *Settings) fillWebSearchDefaults(def Settings) {
	setInt(&s.WebSearch.Results, def.WebSearch.Results)
}

func setInt(field *int, fallback int) {
	if *field <= 0 {
		*field = fallback
	}
}

func setPolicy(field *Policy, fallback Policy) {
	if *field == "" {
		*field = fallback
	}
}
