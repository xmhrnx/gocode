// Package mobile is GoCode on a phone: pair with desktops, see
// their chats, follow the agent and steer it.
package mobile

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/toaster"

	"gocode/internal/remote/client"
	"gocode/ui/theme"
)

// App is the phone app's state and screens.
type App struct {
	gapp   *giode.App
	store  store
	router *giode.Router
	toasts *toaster.Toaster
	// link is the desktop being shown, or nil on the home screen.
	link *link
	// ready is false until the phone's identity is loaded; loadErr says
	// why it couldn't be.
	ready   bool
	loadErr string

	home    *homeScreen
	pairing *pairScreen
	chats   *chatsScreen
	chat    *chatScreen
}

// New builds the app on gapp.
func New(gapp *giode.App) *App {
	a := &App{gapp: gapp, store: store{desktops: loadDesktops()}}
	a.toasts = giode.Toaster().Placement(giode.ToastBottomCenter).Max(2).Inset(giode.Inset{Bottom: 90})
	a.home = newHomeScreen(a)
	a.pairing = newPairScreen(a)
	a.chats = newChatsScreen(a)
	a.chat = newChatScreen(a)
	a.router = giode.NewRouter(
		giode.Route{Path: "/", View: func() giode.Element { return a.home }},
		giode.Route{Path: "/pair", View: func() giode.Element { return a.pairing }},
		giode.Route{Path: "/desktop", View: func() giode.Element { return a.chats }},
		giode.Route{Path: "/chat", View: func() giode.Element { return a.chat }},
	)
	go func() {
		key, err := loadIdentity()
		a.post(func() {
			a.ready = err == nil
			if err != nil {
				a.loadErr = "Couldn't open this phone's key store: " + err.Error()
			}
			a.store.key = key
		})
	}()
	return a
}

// View is the whole window.
func (a *App) View() giode.Element {
	return giode.Stack(giode.Styles{Width: -1, Height: -1, Background: theme.Canvas},
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, a.router),
		a.toasts,
	)
}

// post runs fn on the UI goroutine.
func (a *App) post(fn func()) { a.gapp.Post(fn) }

// toast shows a short message.
func (a *App) toast(text string) { a.toasts.Show(giode.ToastInfo, text) }

// openDesktop connects to a paired desktop and shows its chats.
func (a *App) openDesktop(d client.Desktop) {
	if a.link != nil {
		a.link.close()
	}
	a.link = openLink(a, d)
	a.router.Push("/desktop")
}

// closeDesktop ends the connection when the home screen shows again.
func (a *App) closeDesktop() {
	if a.link != nil {
		a.link.close()
		a.link = nil
	}
}
