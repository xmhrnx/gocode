package settingsview

import (
	"fmt"

	"github.com/xmhrnx/giode"

	"gocode/ui/core"
	"gocode/ui/theme"
)

func browserEntries(s *core.Services) []entry {
	b := &s.Settings.Browser
	return []entry{
		{title: "Let the browser extension connect", description: "The GoCode browser extension lets the agent use your browser: open pages, read them, click, type and fill in forms. " +
			"It works only in tabs you share from the extension's toolbar button and tabs the agent opens. " +
			"GoCode listens on 127.0.0.1 only, and only browser extensions may connect. Install the extension from the browser-extension folder of GoCode's source.",
			keywords: "browser extension chrome edge web tab control automation enable bridge", control: switchControl(s, &b.Enabled)},
		{title: "Port", description: "Where GoCode listens for the extension. If you change it, set the same port in the extension.",
			keywords: "browser extension port websocket", control: number(s, &b.Port, 1024, 65535)},
		{title: "Status", description: "Whether a browser is connected now.",
			keywords: "browser extension status connected", control: browserStatus{s}, wide: true},
	}
}

// browserStatus shows the bridge's live state.
type browserStatus struct{ services *core.Services }

func (b browserStatus) Layout(gtx giode.Context) giode.Dimensions {
	st := b.services.Browser.Status()
	text, color := "Off", theme.Muted
	switch {
	case st.Err != "":
		text, color = st.Err, theme.Danger
	case st.Connected && st.Browser != "":
		text, color = "Connected: "+st.Browser, theme.Success
	case st.Connected:
		text, color = "Connected", theme.Success
	case st.Listening:
		text = fmt.Sprintf("Waiting for the extension on port %d", st.Port)
	}
	return giode.Text(text, giode.Styles{Color: color, FontSize: theme.Sizes.Small}).Layout(gtx)
}
