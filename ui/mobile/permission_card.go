package mobile

import (
	"unicode"

	"github.com/xmhrnx/giode"

	"gocode/internal/remote/wire"
	"gocode/ui/theme"
)

// detailShown bounds the command or diff shown with a request.
const detailShown = 1500

// permissionCard shows the tool call the agent waits on, with answers.
type permissionCard struct {
	app                 *App
	allow, always, deny *giode.ButtonElement
}

func newPermissionCard(a *App) *permissionCard {
	p := &permissionCard{app: a}
	p.allow = primaryButton("Allow", func() { p.answer(wire.DecisionAllow) })
	p.always = outlineButton("Always", func() { p.answer(wire.DecisionAlways) })
	p.deny = outlineButton("Deny", func() { p.answer(wire.DecisionDeny) })
	return p
}

func (p *permissionCard) answer(decision string) {
	if l := p.app.link; l != nil && l.watching != "" {
		l.act(wire.TypeDecide, wire.Decide{Chat: l.watching, Decision: decision})
	}
}

func (p *permissionCard) Layout(gtx giode.Context) giode.Dimensions {
	l := p.app.link
	if l == nil || l.view.Permission == nil {
		return giode.Dimensions{}
	}
	req := l.view.Permission
	detail := req.Detail
	if runes := []rune(detail); len(runes) > detailShown {
		detail = string(runes[:detailShown]) + "\n…"
	}
	children := []giode.Element{
		giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
			theme.Icon("shield-alert", 16, theme.Warning),
			giode.Text("The agent asks to "+lowerFirst(req.Summary), giode.Styles{Color: theme.Text, FontWeight: giode.FontWeightMedium, FlexGrow: 1})),
	}
	if detail != "" {
		children = append(children, giode.Box(giode.Styles{Width: -1, MaxHeight: 180, Padding: giode.SymmetricInset(10, 8),
			BorderRadius: 8, Background: theme.CodeBack},
			giode.Text(detail, giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Small, FontFamily: "JetBrains Mono"})))
	}
	children = append(children, giode.Row(giode.Styles{Gap: 8}, p.allow, p.always, p.deny))
	return giode.Box(giode.Styles{Width: -1, Padding: giode.SymmetricInset(10, 6)},
		giode.Stack(giode.Styles{Width: -1, Gap: 10, Padding: giode.SymmetricInset(14, 12), BorderRadius: 12,
			Background: theme.WarningSoft, BorderWidth: 1, BorderColor: theme.Border}, children...)).Layout(gtx)
}

// lowerFirst lowercases a summary's first letter ("Run make" → "run
// make"), unless it starts an acronym ("NPM install").
func lowerFirst(s string) string {
	runes := []rune(s)
	if len(runes) == 0 || (len(runes) > 1 && unicode.IsUpper(runes[1])) {
		return s
	}
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
