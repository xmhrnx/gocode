package settingsview

import (
	"github.com/xmhrnx/giode"

	"gocode/internal/settings"
	"gocode/internal/tools"
	"gocode/ui/core"
)

var policies = []option[settings.Policy]{
	{settings.PolicyAsk, "Ask every time"},
	{settings.PolicyAllow, "Allow"},
	{settings.PolicyDeny, "Never allow"},
}

func permissionEntries(s *core.Services) []entry {
	a := &s.Settings.Agent
	entries := []entry{
		{title: "Reading files", description: "Reading, listing and searching files in the project.",
			keywords: "read permission access", control: choice(s, &a.ReadPolicy, policies)},
		{title: "Changing files", description: "Creating and editing files. You see a diff before approving.",
			keywords: "write edit permission access", control: choice(s, &a.WritePolicy, policies)},
		{title: "Running commands", description: "Running shell commands, tests and builds.",
			keywords: "bash shell execute permission access", control: choice(s, &a.ExecutePolicy, policies)},
		{title: "Fetching web pages", description: "Downloading pages and documentation from the internet.",
			keywords: "web fetch network internet permission", control: choice(s, &a.NetworkPolicy, policies)},
	}
	for _, spec := range tools.Default().Specs(nil) {
		entries = append(entries, entry{
			title: "Tool: " + spec.Name, description: firstSentence(spec.Description),
			keywords: "tool enable disable " + spec.Name, control: toolSwitch(s, spec.Name),
		})
	}
	return entries
}

// toolSwitch turns one tool on or off for the agent.
func toolSwitch(s *core.Services, name string) giode.Element {
	sw := giode.Input.Switch("")
	sw.SetChecked(!contains(s.Settings.Agent.DisabledTools, name))
	sw.OnChange(func(on bool) {
		disabled := remove(s.Settings.Agent.DisabledTools, name)
		if !on {
			disabled = append(disabled, name)
		}
		s.Settings.Agent.DisabledTools = disabled
		s.SettingsChanged()
	})
	return sw
}
