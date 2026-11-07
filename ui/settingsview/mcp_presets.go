package settingsview

import "github.com/xmhrnx/giode"

// mcpPreset fills the new-server form with a known server; the user adds
// the key before adding it.
type mcpPreset struct {
	label, name string
	// command or url, and env, are what the form's fields get.
	command, env, url string
}

// searchPresets are web search MCP servers. Brave and Tavily need their
// key in the environment; Exa's hosted server works without one.
var searchPresets = []mcpPreset{
	{label: "Brave Search", name: "brave-search",
		command: "npx -y @brave/brave-search-mcp-server --transport stdio", env: "BRAVE_API_KEY="},
	{label: "Tavily", name: "tavily", command: "npx -y tavily-mcp@latest", env: "TAVILY_API_KEY="},
	{label: "Exa", name: "exa", url: "https://mcp.exa.ai/mcp"},
}

// presetButtons returns a button per preset that fills form.
func presetButtons(form *mcpServerForm) []giode.Element {
	buttons := make([]giode.Element, 0, len(searchPresets))
	for _, preset := range searchPresets {
		buttons = append(buttons, smallButton(preset.label, false).OnClick(func() { form.fill(preset) }))
	}
	return buttons
}
