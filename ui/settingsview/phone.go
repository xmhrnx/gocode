package settingsview

import (
	"fmt"

	"github.com/xmhrnx/giode"

	"gocode/ui/core"
	"gocode/ui/theme"
)

func phoneEntries(s *core.Services) []entry {
	p := &s.Settings.Phone
	return []entry{
		{title: "Let the GoCode phone app connect", description: "Follow and steer this computer's chats from your phone: see the agent work, send prompts, stop it and answer permission requests. " +
			"Only phones you pair can connect, over a channel encrypted end to end. No account is needed.",
			keywords: "phone mobile remote app android ios pair connect", control: switchControl(s, &p.Enabled)},
		{title: "Port", description: "Where GoCode listens for phones on your local network.",
			keywords: "phone mobile remote port", control: number(s, &p.Port, 1024, 65535)},
		{title: "Relay address", description: "Lets a phone away from home reach this computer while one of its chats is working, through a relay you run on Cloudflare (see relay/README.md in GoCode's source). Pair phones again after changing it. Empty: the local network only.",
			keywords: "phone mobile remote relay worker cloudflare internet away", control: text(s, &p.Relay, "https://gocode-relay.you.workers.dev", 300)},
		{title: "Pair a phone", description: "Shows a code to type in the GoCode phone app. It works once, for ten minutes.",
			keywords: "phone mobile remote pair code wormhole", control: newPairingControl(s), wide: true},
		{title: "Status", description: "Whether phones can connect now.",
			keywords: "phone mobile remote status connected relay", control: phoneStatus{s}, wide: true},
		{title: "Paired phones", description: "Phones that can connect. Remove one to stop it connecting.",
			keywords: "phone mobile remote paired devices remove unpair", control: newPairedPhones(s), wide: true},
	}
}

// phoneStatus shows whether the desktop is listening, the relay room
// and the phones connected.
type phoneStatus struct{ services *core.Services }

func (p phoneStatus) Layout(gtx giode.Context) giode.Dimensions {
	text, color := "Off", theme.Muted
	if host := p.services.Phone; host != nil {
		st := host.Status()
		switch {
		case st.Err != "":
			text, color = st.Err, theme.Danger
		case st.Listening:
			text = fmt.Sprintf("Listening on port %d", st.Port)
			if st.Relay {
				text += ", and in the relay room"
			}
			if st.Phones > 0 {
				text += fmt.Sprintf(" · %d %s connected", st.Phones, plural(st.Phones, "phone", "phones"))
				color = theme.Success
			}
		}
	}
	return giode.Text(text, giode.Styles{Color: color, FontSize: theme.Sizes.Small}).Layout(gtx)
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
