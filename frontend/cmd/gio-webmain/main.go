//go:build js && wasm

package main

import (
	"log"

	"gioui.org/app"
	gioapp "github.com/nishiki/frontend/app"
)

func main() {
	// Create the Gio app
	ga := gioapp.NewGioApp()

	// Start the event loop in a goroutine
	go func() {
		if err := ga.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	// Keep the app running (required for WASM)
	app.Main()
}
