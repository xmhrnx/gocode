// Command gocode-remote is the GoCode phone app: it pairs with GoCode on
// a desktop and follows and steers the agent there.
package main

import (
	"log"

	"github.com/xmhrnx/giode"
	_ "github.com/xmhrnx/giode/formula" // typesets the math in replies

	"gocode/assets"
	"gocode/mobile/remoteui"
	"gocode/ui/theme"
)

func main() {
	giode.SecureStore.Service = "dev.gocode.remote"
	if err := assets.Load(); err != nil {
		log.Fatal(err)
	}
	theme.Apply(theme.SystemID)
	theme.SetBaseSize(15, 13)
	app := giode.New(giode.Config{Title: "GoCode", Background: "theme.background", Icon: assets.AppIcon})
	ui := remoteui.New(app)
	if err := app.Run(ui.View); err != nil {
		log.Fatal(err)
	}
}
