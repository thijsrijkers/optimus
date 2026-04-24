package gui

import (
	"log"
	"os"

	"gioui.org/app"

	windowengine "optimus/window"
)

func CreateWindow(shell string) {
	go func() {
		var w app.Window
		w.Option(app.Title("Optimus"))
		if err := windowengine.Run(&w, shell); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
