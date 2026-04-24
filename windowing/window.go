package windowing

import (
	"image/color"
	"log"
	"os"
	"sync"

	"optimus/pty"
	"optimus/terminal"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget/material"
)

const (
	fontSize  = 14 
	cellW     = 8   
	cellH     = 18 
	initCols  = 80
	initRows  = 24
)

func CreateWindow(shell string) {
	go func() {
		var window app.Window
		window.Option(app.Title("Optimus"))
		if err := eventLoop(&window, shell); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func eventLoop(window *app.Window, shell string) error {
	var operationList op.Ops

	pty, err := pty.New(shell, initCols, initRows)
	if err != nil {
		return err
	}
	defer pty.Close()
 
	terminal := terminal.New(initCols, initRows)
	var mu sync.Mutex 
 
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := pty.Read(buf)
			if err != nil {
				return
			}
			mu.Lock()
			terminal.Write(buf[:n])
			mu.Unlock()
			window.Invalidate() 
		}
	}()
 
	theme := material.NewTheme()
	theme.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	_ = shell
	for {
		switch event := window.Event().(type) {
		case app.DestroyEvent:
			return event.Err
		case app.FrameEvent:
			context := app.NewContext(&operationList, event)
      
			// Keyboard events
			for {
				keyboardEvent, ok := context.Event(key.Filter{})
				if !ok {
					break
				}
				if keyEvent, ok := keyboardEvent.(key.Event); ok && keyEvent.State == key.Press {
					if seq := keyToBytes(keyEvent); seq != nil {
						pty.Write(seq)
					}
				}
			}
 
			// Drawing to screen
			cols := context.Constraints.Max.X / cellW
			rows := context.Constraints.Max.Y / cellH
			if cols < 1 {
				cols = 1
			}
			if rows < 1 {
				rows = 1
			}
			mu.Lock()
			terminal.Resize(cols, rows)
			mu.Unlock()
			pty.Resize(cols, rows)
 
			paint.Fill(&operationList, color.NRGBA{R: 0x1E, G: 0x1E, B: 0x1E, A: 0xFF})
 
			mu.Lock()
			drawCells(context, terminal, theme)
			mu.Unlock()
 
			event.Frame(context.Ops)
		}
	}
}

var _ event.Filter = key.Filter{}
