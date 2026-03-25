package main

import (
	"log"

	"gioui.org/app"

	gioapp "github.com/nishiki/frontend/app"
)

func main() {
	ga := gioapp.NewGioApp()
	go func() {
		if err := ga.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	app.Main()
}
