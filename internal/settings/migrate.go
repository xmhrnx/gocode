package settings

// currentVersion is the settings format written by this build.
const currentVersion = 4

// migrate updates settings written by older builds.
func (s *Settings) migrate() {
	if s.Version < 2 {
		s.migrateLimits()
	}
	if s.Version < 3 && s.Chat.SendKey == "shift+enter" {
		// Shift+Enter as the send key isn't offered any more; Ctrl+Enter
		// keeps the point of it: plain Enter never sends.
		s.Chat.SendKey = SendCtrlEnter
	}
	if s.Terminal != nil {
		// The terminal tab is gone; its shell setting was the agent's too.
		if s.Agent.Shell == "" {
			s.Agent.Shell = s.Terminal.Shell
		}
		s.Terminal = nil
	}
	s.Version = currentVersion
}

// migrateLimits drops the first release's step and reply limits (60 steps,
// 16000 tokens) when they are still the defaults: they stopped long tasks
// for no benefit. Values the user chose are kept. Compaction is new, so
// it starts on.
func (s *Settings) migrateLimits() {
	if s.Agent.MaxSteps == 60 {
		s.Agent.MaxSteps = 0
	}
	if s.Agent.MaxOutputTokens == 16000 {
		s.Agent.MaxOutputTokens = 0
	}
	s.Agent.AutoCompact = true
}
