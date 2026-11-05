package settingsview

import (
	"gioui.org/layout"
	"github.com/xmhrnx/giode"

	"gocode/internal/mcp"
	"gocode/ui/core"
	"gocode/ui/theme"
)

// mcpPage lists the configured MCP servers with their status, and the
// form that adds one.
type mcpPage struct {
	services *core.Services
	rows     map[string]*mcpServerRow
	form     *mcpServerForm
}

func mcpEntries(s *core.Services) []entry {
	return []entry{{
		title:       "MCP servers",
		description: "Model Context Protocol servers give the agent more tools: databases, browsers, issue trackers, docs. Their tools appear as mcp__server__tool and ask for permission like commands do, unless the server marks them read-only.",
		keywords:    "mcp model context protocol server tools stdio http plugin",
		wide:        true,
		control:     newMCPPage(s),
	}}
}

func newMCPPage(s *core.Services) *mcpPage {
	p := &mcpPage{services: s, rows: map[string]*mcpServerRow{}}
	p.form = newMCPServerForm(p)
	return p
}

func (p *mcpPage) Layout(gtx layout.Context) layout.Dimensions {
	statuses := map[string]mcp.ServerStatus{}
	for _, status := range p.services.MCP.Status() {
		statuses[status.Name] = status
	}
	children := []giode.Element{p.form}
	for _, server := range p.services.Settings.MCPServers {
		row, ok := p.rows[server.Name]
		if !ok {
			row = newMCPServerRow(p, server.Name)
			p.rows[server.Name] = row
		}
		row.status = statuses[server.Name]
		children = append(children, row)
	}
	if len(p.services.Settings.MCPServers) == 0 {
		children = append(children, giode.Text("No MCP servers yet.", theme.Caption()))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 8}, children...).Layout(gtx)
}
