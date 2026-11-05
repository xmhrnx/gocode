package settingsview

import "gocode/ui/core"

func agentEntries(s *core.Services) []entry {
	a := &s.Settings.Agent
	efforts := []option[string]{{"", "Model default"}, {"low", "Low"}, {"medium", "Medium"}, {"high", "High"}}
	return []entry{
		{title: "Reasoning effort", description: "How hard reasoning models think before answering. Higher is slower and costs more.",
			keywords: "thinking effort reasoning budget", control: choice(s, &a.ReasoningEffort, efforts)},
		{title: "Step limit", description: "Stop a run after this many model calls. 0 means no limit: the agent works until the task is done. A run that repeats the same tool call 3 times in a row is paused for your approval either way.",
			keywords: "loop steps iterations limit turns continue", control: number(s, &a.MaxSteps, 0, 100000)},
		{title: "Reply token limit", description: "Caps a single model reply, not the task; a task is many replies. 0 lets the provider decide (Anthropic, which requires a value, gets 32,000).",
			keywords: "max tokens output length reply", control: number(s, &a.MaxOutputTokens, 0, 1000000)},
		{title: "Compact automatically", description: "When the conversation nears the model's context window, or the provider rejects it as too long, summarize it and carry on. The transcript stays visible.",
			keywords: "compact compaction context summarize long conversation window", control: switchControl(s, &a.AutoCompact)},
		{title: "Compact at (% of context)", description: "How full the context window may get before compacting.",
			keywords: "compact compaction context threshold percent", control: number(s, &a.CompactAtPercent, 30, 95)},
		{title: "Context window when unknown (tokens)", description: "Used for models whose provider doesn't report their context size.",
			keywords: "context window size tokens compact", control: number(s, &a.ContextWindow, 4000, 10000000)},
		{title: "Set temperature", description: "Off leaves sampling to the provider. Some models reject custom temperatures.",
			keywords: "temperature sampling randomness", control: switchControl(s, &a.UseTemperature)},
		{title: "Temperature", description: "Used only when the switch above is on. 0 is focused, 1 is creative.",
			keywords: "temperature sampling randomness", control: decimal(s, &a.Temperature, 0, 2)},
		{title: "Command timeout (seconds)", description: "Commands the agent runs are stopped after this long unless it asks for more.",
			keywords: "bash shell command timeout", control: number(s, &a.CommandTimeoutSeconds, 5, 3600)},
		{title: "Tool output limit (KB)", description: "Longer tool output is cut before it reaches the model, keeping its start and end.",
			keywords: "tool output truncate context limit", control: number(s, &a.MaxToolOutputKB, 4, 1024)},
		{title: "Follow project instructions", description: "Add AGENTS.md, CLAUDE.md or .gocode/rules.md from the project folder to the prompt.",
			keywords: "agents.md claude.md rules project instructions", control: switchControl(s, &a.LoadProjectRules)},
	}
}
