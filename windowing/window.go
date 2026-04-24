package windowing

import (
	"image/color"
	"log"
	"os"
	"sync"

	"optimus/pty"
	"optimus/terminal"

	"gioui.org/app"
	ioevent "gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
)

const (
	fontSize   = 13
	fontFamily = "Menlo"
	initCols   = 80
	initRows   = 24
)

func cellMetrics(context layout.Context) (cellW, cellH int) {
	fontPx := context.Sp(fontSize)
	if fontPx < 12 {
		fontPx = 12
	}
	cellH = fontPx + 4
	cellW = (fontPx * 3) / 5
	if cellW < 8 {
		cellW = 8
	}
	return cellW, cellH
}

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
	keyboardTag := new(struct{})

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

	_ = shell
	for {
		switch event := window.Event().(type) {
		case app.DestroyEvent:
			return event.Err
		case app.FrameEvent:
			context := app.NewContext(&operationList, event)
			ioevent.Op(context.Ops, keyboardTag)
			if !context.Focused(keyboardTag) {
				context.Execute(key.FocusCmd{Tag: keyboardTag})
			}
			cellW, cellH := cellMetrics(context)

			// Keyboard events
			for {
				keyboardEvent, ok := context.Event(
					key.Filter{
						Focus:    keyboardTag,
						Optional: key.ModCtrl | key.ModShift | key.ModAlt | key.ModSuper | key.ModCommand,
					},
					key.FocusFilter{Target: keyboardTag},
				)
				if !ok {
					break
				}
				switch ev := keyboardEvent.(type) {
				case key.Event:
					if ev.State == key.Press {
						if seq := keyToBytes(ev); seq != nil {
							pty.Write(seq)
						}
					}
				case key.EditEvent:
					if ev.Text != "" {
						pty.Write([]byte(ev.Text))
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

			paint.Fill(&operationList, color.NRGBA{R: 0x28, G: 0x2C, B: 0x34, A: 0xFF})

			mu.Lock()
			drawCells(context, terminal, theme, cellW, cellH)
			mu.Unlock()

			event.Frame(context.Ops)
		}
	}
}

var _ ioevent.Filter = key.Filter{}
