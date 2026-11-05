package terminal

import (
	"fmt"

	"gioui.org/layout"
	"github.com/xmhrnx/giode"

	"gocode/ui/components"
	"gocode/ui/theme"
)

// blockView draws one block: the command, its status and its output, with
// copy, rerun and ask-the-agent actions.
type blockView struct {
	tab        *Tab
	block      *block
	copyOutput giode.Element
	rerun      *components.IconButton
	askAgent   *components.IconButton
	stop       *components.IconButton
}

func newBlockView(t *Tab, b *block) *blockView {
	v := &blockView{tab: t, block: b}
	v.copyOutput = giode.CopyButton(b.text, giode.Styles{Color: theme.Muted})
	v.rerun = components.NewIconButton("rotate-cw", "Run again", func() {
		if t.running() == nil {
			t.run(b.command)
		}
	}).SetSize(14)
	v.askAgent = components.NewIconButton("sparkles", "Ask the agent about this", func() {
		t.services.Open().OpenChat(b.workDir, agentPrompt(b))
	}).SetSize(14)
	v.stop = components.NewIconButton("square", "Stop  (Esc)", t.stop).SetSize(14)
	return v
}

func (v *blockView) Layout(gtx layout.Context) layout.Dimensions {
	b := v.block
	children := []giode.Element{v.header()}
	if text := b.text(); text != "" {
		output := theme.Mono()
		if b.dropped > 0 {
			children = append(children, giode.Text(fmt.Sprintf("… %d earlier lines not kept", b.dropped), theme.Caption()))
		}
		children = append(children, giode.Text(text, output).Selectable(b.id))
	}
	border := theme.SoftBorder
	if !b.running && b.exitCode != 0 {
		border = giode.Mix("theme.danger", "theme.background", 0.45)
	}
	return giode.Box(giode.Styles{Width: -1, Padding: giode.Inset{Left: 16, Right: 16, Top: 10}},
		giode.Stack(giode.Styles{Width: -1, Gap: 8, Padding: giode.SymmetricInset(14, 10), BorderRadius: 10,
			Background: theme.Panel, BorderWidth: 1, BorderColor: border}, children...),
	).Layout(gtx)
}

func (v *blockView) header() giode.Element {
	b := v.block
	actions := []giode.Element{v.copyOutput, v.rerun, v.askAgent}
	if b.running {
		actions = []giode.Element{v.stop}
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2},
			giode.Text(shortDir(b.workDir), giode.Styles{Color: theme.Muted, FontSize: theme.Sizes.Tiny, FontFamily: "JetBrains Mono"}),
			giode.Text(b.command, giode.Styles{Color: theme.Text, FontFamily: "JetBrains Mono", FontSize: theme.Sizes.Code,
				FontWeight: giode.FontWeightBold}),
		),
		v.status(),
		giode.Row(giode.Styles{Gap: 0, Align: giode.AlignCenter}, actions...),
	)
}

func (v *blockView) status() giode.Element {
	b := v.block
	switch {
	case b.running:
		return giode.Spinner(giode.Styles{Width: 14, Height: 14, Color: theme.Primary})
	case b.exitCode != 0:
		return badge(fmt.Sprintf("exit %d · %s", b.exitCode, formatDuration(b.duration)), theme.Danger, theme.DangerSoft)
	}
	return badge(formatDuration(b.duration), theme.Muted, theme.Hover)
}

func badge(text, color, background string) giode.Element {
	return giode.Box(giode.Styles{Padding: giode.SymmetricInset(7, 2), BorderRadius: 6, Background: background},
		giode.Text(text, giode.Styles{Color: color, FontSize: theme.Sizes.Tiny, FontFamily: "JetBrains Mono"}))
}
