package settingsview

import (
	"fmt"
	"slices"
	"strings"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/toggle"

	"gocode/internal/mcp"
	"gocode/internal/settings"
	"gocode/ui/components"
	"gocode/ui/theme"
)

// mcpServerRow shows one server: a switch, its address, its state and
// tools, and buttons to reconnect or remove it.
type mcpServerRow struct {
	page      *mcpPage
	name      string
	status    mcp.ServerStatus
	enabled   *toggle.Switch
	reconnect *components.IconButton
	remove    *components.IconButton
}

func newMCPServerRow(p *mcpPage, name string) *mcpServerRow {
	s := p.services
	r := &mcpServerRow{page: p, name: name}
	r.enabled = giode.Input.Switch("")
	r.enabled.SetChecked(r.config().Enabled)
	r.enabled.OnChange(func(on bool) {
		if server := r.config(); server != nil {
			server.Enabled = on
			s.SettingsChanged()
		}
	})
	r.reconnect = components.NewIconButton("rotate-cw", "Reconnect", func() { s.MCP.Reconnect(name) }).SetSize(14)
	r.remove = components.NewIconButton("trash-2", "Remove server", func() {
		s.Settings.MCPServers = slices.DeleteFunc(s.Settings.MCPServers, func(m settings.MCPServer) bool { return m.Name == name })
		s.SettingsChanged()
	}).SetSize(14)
	return r
}

// config returns the server's settings, or nil once it was removed.
func (r *mcpServerRow) config() *settings.MCPServer {
	servers := r.page.services.Settings.MCPServers
	for i := range servers {
		if servers[i].Name == r.name {
			return &servers[i]
		}
	}
	return nil
}

func (r *mcpServerRow) Layout(gtx layout.Context) layout.Dimensions {
	server := r.config()
	if server == nil {
		return layout.Dimensions{}
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(4, 6)},
		r.enabled,
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2},
			giode.Text(server.Name, giode.Styles{Color: theme.Text, FontWeight: giode.FontWeightMedium}),
			giode.Text(address(*server), giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Tiny, FontFamily: "JetBrains Mono",
				TextOverflow: giode.TextOverflowEllipsis}),
			r.statusLine(),
		),
		r.reconnect, r.remove,
	).Layout(gtx)
}

func (r *mcpServerRow) statusLine() giode.Element {
	text, color := r.status.State, theme.Muted
	switch r.status.State {
	case mcp.StateConnected:
		text, color = fmt.Sprintf("connected · %d tools: %s", len(r.status.ToolNames), strings.Join(r.status.ToolNames, ", ")), theme.Success
	case mcp.StateFailed:
		text, color = "failed: "+r.status.Err, theme.Danger
	}
	return giode.Text(text, giode.Styles{Color: color, FontSize: theme.Sizes.Small, MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis})
}

func address(server settings.MCPServer) string {
	if server.URL != "" {
		return server.URL
	}
	return strings.Join(append([]string{server.Command}, server.Args...), " ")
}
