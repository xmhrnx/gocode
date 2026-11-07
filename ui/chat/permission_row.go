package chat

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"gocode/internal/agent"
	"gocode/internal/tools"
	"gocode/ui/theme"

	"gocode/ui/components"
)

// permissionRow asks the user to approve a tool call, showing what it
// would do and, for edits, the diff.
type permissionRow struct {
	rowBase
	feedback *input.Input
	allow    *giode.ButtonElement
	always   *giode.ButtonElement
	deny     *giode.ButtonElement
}

func newPermissionRow(tab *ChatTab) *permissionRow {
	r := &permissionRow{rowBase: rowBase{tab: tab}}
	r.feedback = giode.Input.Text("Optional: tell the agent what to do instead", giode.Styles{
		Background: theme.Canvas, BorderColor: theme.Border, BorderRadius: 8, Color: theme.Text, FontSize: theme.Sizes.Small,
	}).Submit(func(string) { r.answer(agent.DecisionDeny) })
	components.AddEditMenu(tab.services.App, r.feedback, components.EditOptions{})
	r.allow = giode.Button("Allow", giode.Styles{Background: theme.Primary, Color: theme.OnPrimary, BorderRadius: 8,
		FontSize: theme.Sizes.Small, Padding: giode.SymmetricInset(14, 7)}).OnClick(func() { r.answer(agent.DecisionAllow) })
	r.always = giode.Button("Always allow", giode.Styles{BorderColor: theme.Border, Color: theme.Text, BorderRadius: 8,
		FontSize: theme.Sizes.Small, Padding: giode.SymmetricInset(14, 7)}).Variant(giode.VariantOutlined).
		OnClick(func() { r.answer(agent.DecisionAllowAlways) })
	r.deny = giode.Button("Deny", giode.Styles{Background: theme.Danger, Color: theme.Danger, BorderRadius: 8,
		FontSize: theme.Sizes.Small, Padding: giode.SymmetricInset(14, 7)}).Variant(giode.VariantGhost).
		OnClick(func() { r.answer(agent.DecisionDeny) })
	return r
}

func (r *permissionRow) version() int { return 0 }

func (r *permissionRow) answer(decision agent.Decision) {
	r.tab.answerPermission(decision, r.feedback.Text())
	r.feedback.SetText("")
}

func (r *permissionRow) Layout(gtx giode.Context) giode.Dimensions {
	pending := r.tab.pendingRequest()
	if pending == nil {
		return giode.Dimensions{}
	}
	req := pending.request
	r.always.Label("Always allow " + req.Call.Name)
	children := []giode.Element{
		giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter},
			theme.Icon("shield-alert", 17, theme.Warning),
			giode.Text(accessQuestion(req.Access), giode.Styles{Color: theme.Text, FontWeight: giode.FontWeightSemiBold})),
		giode.Box(giode.Styles{Width: -1, Background: theme.CodeBack, BorderRadius: 8, Padding: giode.SymmetricInset(12, 8)},
			giode.Text(req.Summary, theme.Mono())),
	}
	if req.Preview != "" {
		children = append(children, giode.Box(giode.Styles{Width: -1, BorderRadius: 8, Background: theme.CodeBack,
			Overflow: giode.OverflowHidden}, diffView(req.Preview)))
	}
	children = append(children, r.feedback,
		giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter}, r.allow, r.always, r.deny))
	return giode.Box(giode.Styles{Width: -1, Padding: giode.SymmetricInset(0, 8)},
		giode.Stack(giode.Styles{Width: -1, Gap: 12, Padding: giode.UniformInset(16), BorderRadius: 12,
			Background: theme.WarningSoft, BorderWidth: 1, BorderColor: giode.Mix("theme.warning", "theme.surface", 0.45)},
			children...),
	).Layout(gtx)
}

func accessQuestion(access tools.Access) string {
	switch access {
	case tools.AccessWrite:
		return "Allow this file change?"
	case tools.AccessExecute:
		return "Allow this command?"
	case tools.AccessNetwork:
		return "Allow this network request?"
	}
	return "Allow this action?"
}
