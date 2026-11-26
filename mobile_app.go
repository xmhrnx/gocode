//go:build android || ios

package main

import (
	"github.com/xmhrnx/giode"
	_ "github.com/xmhrnx/giode/formula" // typesets the math in replies

	"gocode/assets"
	"gocode/ui/mobile"
	"gocode/ui/theme"
)

// run starts the phone app: it pairs with desktops and follows and
// steers their agents. The agent itself runs only on desktops.
func run() error {
	tuneMemory()
	giode.SecureStore.Service = "dev.gocode.app"
	if err := assets.Load(); err != nil {
		return err
	}
	theme.Apply(theme.SystemID)
	theme.SetBaseSize(15, 13)
	app := giode.New(giode.Config{Title: "GoCode", Background: "theme.background", Icon: assets.AppIcon})
	return app.Run(mobile.New(app).View)
}
