// Package history is the dialog that lists past chats and reopens them.
package history

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/scroll"

	"gocode/internal/session"
	"gocode/ui/core"
	"gocode/ui/theme"

	"gocode/ui/components"
)

// Dialog lists stored chats, newest first, with a search field.
type Dialog struct {
	services  *core.Services
	onOpen    func(*session.Session)
	open      bool
	loading   bool
	summaries []session.Summary
	dialog    *dialog.Dialog
	search    *input.Input
	list      *scroll.Scroll
	rows      *giode.KeyedCache[string, *row]
}

// New returns a closed dialog that calls onOpen with the chosen session.
func New(services *core.Services, onOpen func(*session.Session)) *Dialog {
	d := &Dialog{services: services, onOpen: onOpen, rows: giode.NewKeyedCache[string, *row]()}
	d.search = giode.Input.Text("Search chats…", giode.Styles{
		Background: theme.Canvas, BorderColor: theme.Border, BorderRadius: 9, Color: theme.Text,
	})
	components.AddEditMenu(services.App, d.search, components.EditOptions{})
	d.list = giode.Scroll(d.listView, giode.Styles{ScrollBar: giode.ScrollBar{Width: 6, AutoHide: true, TrackColor: "transparent"}})
	d.dialog = giode.Dialog(&d.open, d.content, giode.Styles{
		Background: theme.Panel, BorderRadius: 14, BorderWidth: 1, BorderColor: theme.Border,
		BoxShadow: giode.BoxShadow{Y: 18, Blur: 48, Color: "#00000066"},
	}).Padding(giode.UniformInset(14))
	return d
}

// Open shows the dialog and reloads the list.
func (d *Dialog) Open() {
	d.open, d.loading = true, true
	d.search.SetText("")
	d.search.Focus()
	store, app := d.services.Sessions, d.services.App
	go func() {
		summaries, err := store.List()
		app.Post(func() {
			d.loading, d.summaries = false, summaries
			if err != nil {
				d.services.Error("Couldn't list chats: " + err.Error())
			}
		})
	}()
}

// Layout lays out the dialog; it takes no space.
func (d *Dialog) Layout(gtx giode.Context) giode.Dimensions { return d.dialog.Layout(gtx) }

func (d *Dialog) content() giode.Element {
	return giode.Stack(giode.Styles{Width: 620, Gap: 10},
		giode.Text("Chat history", giode.Styles{Color: theme.Text, FontSize: theme.Sizes.Title, FontWeight: giode.FontWeightSemiBold}),
		d.search,
		giode.Box(giode.Styles{Width: -1, Height: 440}, d.list),
	)
}

// choose loads the session and opens it.
func (d *Dialog) choose(id string) {
	d.open = false
	store, app := d.services.Sessions, d.services.App
	go func() {
		sess, err := store.Load(id)
		if err != nil {
			d.services.Error("Couldn't open the chat: " + err.Error())
			return
		}
		app.Post(func() { d.onOpen(sess) })
	}()
}

// remove deletes the session after the frame.
func (d *Dialog) remove(id string) {
	giode.AfterLayout(func() {
		if err := d.services.Sessions.Delete(id); err != nil {
			d.services.Error("Couldn't delete the chat: " + err.Error())
			return
		}
		for i, s := range d.summaries {
			if s.ID == id {
				d.summaries = append(d.summaries[:i], d.summaries[i+1:]...)
				return
			}
		}
	})
}
