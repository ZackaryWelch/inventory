package main

import (
	"syscall/js"
	"time"

	"cogentcore.org/core/core"
	"github.com/nishiki/frontend/app"
)

func main() {
	// Add panic recovery to prevent total app crashes
	defer func() {
		if r := recover(); r != nil {
			js.Global().Get("console").Call("error", "Application panic recovered:", r)
		}
	}()

	// Wait for DOM to be ready before initializing UI
	waitForDOMReady()

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

// waitForDOMReady waits for the DOM to be fully loaded before proceeding
func waitForDOMReady() {
	doc := js.Global().Get("document")
	readyState := doc.Get("readyState").String()

	// If document is already loaded, return immediately
	if readyState == "complete" || readyState == "interactive" {
		return
	}

	// Otherwise, wait for DOMContentLoaded event
	done := make(chan struct{})
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		close(done)
		cb.Release()
		return nil
	})

	doc.Call("addEventListener", "DOMContentLoaded", cb)

	// Timeout after 5 seconds to prevent indefinite blocking
	select {
	case <-done:
		// DOM is ready
	case <-time.After(5 * time.Second):
		// Timeout, proceed anyway
		cb.Release()
	}
}