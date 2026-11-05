package chat

import (
	"os"
	"runtime"
	"slices"

	"gocode/internal/agent"
	"gocode/internal/prompts"
	"gocode/internal/skills"
	"gocode/internal/tools"
)

type promptInputs struct {
	template, rules string
	loadRules       bool
	vars            prompts.Vars
	skills          func() []skills.Skill
	useSkills       bool
}

// agentConfig snapshots the settings for a run; it is called on the UI
// goroutine so the run never reads settings the user is changing.
func (c *ChatTab) agentConfig() agent.Config {
	cfg := c.services.Settings.Agent
	run := agent.Config{
		Model: c.model.Model, Permissions: c.perms, Ask: c.ask,
		MaxSteps: cfg.MaxSteps, MaxTokens: cfg.MaxOutputTokens,
		ReasoningEffort: cfg.ReasoningEffort, MaxRetries: c.services.Settings.Network.MaxRetries,
		ContextWindow: c.contextWindow(),
		CompactPrompt: prompts.Or(c.services.Settings.Prompts.Compact, prompts.DefaultCompact()),
		Env: tools.Env{
			WorkDir: c.session.WorkDir, Shell: c.shell(),
			CommandTimeoutSeconds: cfg.CommandTimeoutSeconds,
			MaxOutputBytes:        cfg.MaxToolOutputKB * 1024,
			Todos: func(todos []tools.Todo) {
				c.services.App.Post(func() { c.setTodos(todos) })
			},
		},
	}
	if cfg.AutoCompact {
		run.CompactAtPercent = cfg.CompactAtPercent
	}
	if cfg.UseTemperature {
		temperature := cfg.Temperature
		run.Temperature = &temperature
	}
	if c.useTools {
		disabled := slices.Clone(cfg.DisabledTools)
		c.registry = c.toolRegistry()
		run.Tools = c.registry
		run.ToolEnabled = func(t tools.Tool) bool { return !slices.Contains(disabled, t.Spec().Name) }
	}
	return run
}

// promptVars snapshots what the system prompt needs.
func (c *ChatTab) promptVars() promptInputs {
	p := c.services.Settings.Prompts
	template := prompts.Or(p.Agent, prompts.DefaultAgent())
	if !c.useTools {
		template = prompts.Or(p.Chat, prompts.DefaultChat())
	}
	return promptInputs{
		template: template, rules: p.CustomRules,
		loadRules: c.useTools && c.services.Settings.Agent.LoadProjectRules,
		vars:      prompts.Vars{WorkDir: c.session.WorkDir, Shell: c.shell()},
		skills:    c.skillLister(),
		useSkills: c.useTools,
	}
}

// systemPrompt builds the prompt; it reads project files, so it runs on
// the run's goroutine.
func (c *ChatTab) systemPrompt(in promptInputs) string {
	prompt := prompts.Build(in.template, in.rules, in.vars, in.loadRules)
	if section := skills.PromptSection(in.skills()); in.useSkills && section != "" {
		prompt += "\n\n" + section
	}
	return prompt
}

func (c *ChatTab) shell() string {
	if shell := c.services.Settings.Terminal.Shell; shell != "" {
		return shell
	}
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	if runtime.GOOS == "windows" {
		return "powershell.exe"
	}
	return "/bin/sh"
}
