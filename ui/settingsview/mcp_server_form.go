package settingsview

import (
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/segmented"

	"gocode/internal/settings"
	"gocode/ui/theme"

	"gocode/ui/components"
)

// mcpServerForm adds a server started as a program (stdio) or reached at
// a URL (Streamable HTTP).
type mcpServerForm struct {
	page              *mcpPage
	open              bool
	transport         *segmented.Control
	name, command     *input.Input
	env, url, headers *input.Input
	show, add, cancel *giode.ButtonElement
	presets           []giode.Element
}

func newMCPServerForm(p *mcpPage) *mcpServerForm {
	f := &mcpServerForm{page: p}
	f.transport = giode.Input.Segmented("Program (stdio)", "URL (HTTP)")
	f.name = giode.Input.Text("Name, e.g. github", fieldStyle(-1))
	f.command = giode.Input.Text("Command, e.g. npx -y @modelcontextprotocol/server-filesystem ~/code", fieldStyle(-1))
	f.env = giode.Input.Text("Environment, one KEY=value per line (optional)", fieldStyle(-1)).Lines(1, 6)
	f.url = giode.Input.Text("URL, e.g. https://mcp.example.com/mcp", fieldStyle(-1))
	f.headers = giode.Input.Text("Headers, one Name: value per line (optional)", fieldStyle(-1)).Lines(1, 6)
	for _, field := range []*input.Input{f.name, f.command, f.env, f.url, f.headers} {
		components.AddEditMenu(p.services.App, field, components.EditOptions{})
	}
	f.show = smallButton("Add MCP server", false).Icon(theme.Icon("plus", 14, "")).OnClick(func() { f.open = true })
	f.add = smallButton("Add server", true).OnClick(f.submit)
	f.cancel = smallButton("Cancel", false).OnClick(func() { f.open = false })
	f.presets = presetButtons(f)
	return f
}

// fill puts a preset into the fields.
func (f *mcpServerForm) fill(preset mcpPreset) {
	f.name.SetText(preset.name)
	if preset.url != "" {
		f.transport.SetSelected(1)
		f.url.SetText(preset.url)
		f.headers.SetText("")
		return
	}
	f.transport.SetSelected(0)
	f.command.SetText(preset.command)
	f.env.SetText(preset.env)
	f.env.Focus() // the key goes after the "="
}

func (f *mcpServerForm) submit() {
	s := f.page.services
	server := settings.MCPServer{Name: strings.TrimSpace(f.name.Text()), Enabled: true}
	if f.transport.Selected() == 0 {
		words := splitCommandLine(f.command.Text())
		if len(words) > 0 {
			server.Transport, server.Command, server.Args = "stdio", words[0], words[1:]
		}
		server.Env = parsePairs(f.env.Text(), "=")
	} else {
		server.Transport, server.URL = "http", strings.TrimSpace(f.url.Text())
		server.Headers = parsePairs(f.headers.Text(), ":")
	}
	if err := validateServer(s.Settings.MCPServers, server); err != "" {
		s.Error(err)
		return
	}
	s.Settings.MCPServers = append(s.Settings.MCPServers, server)
	s.SettingsChanged()
	for _, field := range []*input.Input{f.name, f.command, f.env, f.url, f.headers} {
		field.SetText("")
	}
	f.open = false
}

func validateServer(existing []settings.MCPServer, server settings.MCPServer) string {
	switch {
	case server.Name == "":
		return "Give the server a name."
	case server.Command == "" && server.URL == "":
		return "Enter the command that starts the server, or its URL."
	}
	for _, other := range existing {
		if other.Name == server.Name {
			return "A server named " + server.Name + " already exists."
		}
	}
	return ""
}

func (f *mcpServerForm) Layout(gtx giode.Context) giode.Dimensions {
	if !f.open {
		return f.show.Layout(gtx)
	}
	fields := []giode.Element{field("Name", f.name), field("Command", f.command), field("Environment", f.env)}
	if f.transport.Selected() == 1 {
		fields = []giode.Element{field("Name", f.name), field("URL", f.url), field("Headers", f.headers)}
	}
	presets := append([]giode.Element{giode.Text("Web search:", giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small})}, f.presets...)
	children := append([]giode.Element{
		giode.Text("New MCP server", giode.Styles{Color: theme.Text, FontWeight: giode.FontWeightSemiBold}),
		giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapWrap}, presets...),
		f.transport,
	}, fields...)
	children = append(children, giode.Row(giode.Styles{Gap: 8}, f.add, f.cancel))
	return giode.Stack(giode.Styles{Width: -1, Gap: 12, Padding: giode.UniformInset(16), BorderRadius: 10,
		Background: theme.Raised, BorderWidth: 1, BorderColor: theme.SoftBorder}, children...).Layout(gtx)
}
