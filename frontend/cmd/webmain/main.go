package main

import (
	"cogentcore.org/core/core"
	"github.com/nishiki/frontend/app"
)

func main() {
	// Create the app
	application := app.NewApp()
	
	// Create and run the web UI
	core.TheApp.SetName("Nishiki Inventory")
	core.AppAbout = "A cross-platform inventory management application built with Cogent Core"

	body := core.NewBody("Nishiki Inventory")
	body.AddTopBar(func(tb *core.Frame) {
		// App bar customization if needed
	})
	application.CreateMainUI(body)
	body.RunMainWindow()
}