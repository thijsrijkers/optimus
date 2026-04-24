package window

import (
	"image"
	"image/color"
	"io"
	"strings"
	"sync"

	core "optimus/core"
	"optimus/pty"

	"gioui.org/app"
	"gioui.org/io/clipboard"
	ioevent "gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
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

func Run(window *app.Window, shell string) error {
	var operationList op.Ops
	keyboardTag := new(struct{})
	mouseTag := new(struct{})
	clipboardTag := new(struct{})

	selecting := false
	hasSelection := false
	selStartCol, selStartRow := 0, 0
	selEndCol, selEndRow := 0, 0
	lastButtons := pointer.Buttons(0)

	pty, err := pty.New(shell, initCols, initRows)
	if err != nil {
		return err
	}
	defer pty.Close()

	terminal := core.New(initCols, initRows)
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

	for {
		switch event := window.Event().(type) {
		case app.DestroyEvent:
			return event.Err
		case app.FrameEvent:
			context := app.NewContext(&operationList, event)
			ioevent.Op(context.Ops, keyboardTag)
			ioevent.Op(context.Ops, clipboardTag)
			mouseArea := clip.Rect(image.Rect(0, 0, context.Constraints.Max.X, context.Constraints.Max.Y)).Push(context.Ops)
			ioevent.Op(context.Ops, mouseTag)
			pointer.CursorText.Add(context.Ops)
			mouseArea.Pop()
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
					pointer.Filter{
						Target:  mouseTag,
						Kinds:   pointer.Press | pointer.Release | pointer.Move | pointer.Drag | pointer.Scroll,
						ScrollX: pointer.ScrollRange{Min: -1 << 20, Max: 1 << 20},
						ScrollY: pointer.ScrollRange{Min: -1 << 20, Max: 1 << 20},
					},
					transfer.TargetFilter{Target: clipboardTag, Type: "text/plain"},
				)
				if !ok {
					break
				}
				switch ev := keyboardEvent.(type) {
				case key.Event:
					if ev.State == key.Press {
						if !isCopyShortcut(ev) && !isPasteShortcut(ev) {
							hasSelection = false
						}
						if isCopyShortcut(ev) {
							if hasSelection {
								mu.Lock()
								text := selectionText(terminal.Buffer(), selStartCol, selStartRow, selEndCol, selEndRow)
								mu.Unlock()
								if text != "" {
									context.Execute(clipboard.WriteCmd{Type: "text/plain", Data: io.NopCloser(strings.NewReader(text))})
								}
							}
							continue
						}
						if isPasteShortcut(ev) {
							context.Execute(clipboard.ReadCmd{Tag: clipboardTag})
							continue
						}
						if seq := keyToBytes(ev); seq != nil {
							pty.Write(seq)
						}
					}
				case key.EditEvent:
					if ev.Text != "" {
						pty.Write([]byte(ev.Text))
					}
				case transfer.DataEvent:
					if ev.Type == "text/plain" {
						r := ev.Open()
						data, err := io.ReadAll(r)
						r.Close()
						if err == nil && len(data) > 0 {
							pty.Write(data)
						}
					}
				case pointer.Event:
					col, row := pointerToCell(ev, cellW, cellH)
					mu.Lock()
					buf := terminal.Buffer()
					maxCol := buf.Cols() - 1
					maxRow := buf.Rows() - 1
					mu.Unlock()
					if col < 0 {
						col = 0
					}
					if row < 0 {
						row = 0
					}
					if col > maxCol {
						col = maxCol
					}
					if row > maxRow {
						row = maxRow
					}

					proto := terminal.MouseProtocol()
					pressed := ev.Buttons &^ lastButtons
					released := lastButtons &^ ev.Buttons

					forceSelection := ev.Modifiers.Contain(key.ModShift)
					if proto.Enabled && !forceSelection {
						sendPointerToPTY(pty, ev, pressed, released, col, row, proto)
					} else {
						if ev.Kind == pointer.Press && pressed.Contain(pointer.ButtonPrimary) {
							selecting = true
							hasSelection = false
							selStartCol, selStartRow = col, row
							selEndCol, selEndRow = col, row
						}
						if selecting && (ev.Kind == pointer.Drag || ev.Kind == pointer.Move) {
							selEndCol, selEndRow = col, row
						}
						if selecting && ev.Kind == pointer.Release && released.Contain(pointer.ButtonPrimary) {
							selecting = false
							selEndCol, selEndRow = col, row
							mu.Lock()
							text := selectionText(terminal.Buffer(), selStartCol, selStartRow, selEndCol, selEndRow)
							mu.Unlock()
							if text != "" {
								hasSelection = true
								context.Execute(clipboard.WriteCmd{
									Type: "text/plain",
									Data: io.NopCloser(strings.NewReader(text)),
								})
							} else {
								hasSelection = false
							}
						}
						if ev.Kind == pointer.Press && (pressed.Contain(pointer.ButtonSecondary) || pressed.Contain(pointer.ButtonTertiary)) {
							context.Execute(clipboard.ReadCmd{Tag: clipboardTag})
						}
					}
					lastButtons = ev.Buttons
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
			drawCells(
				context,
				terminal,
				theme,
				cellW,
				cellH,
				hasSelection || selecting,
				selStartCol,
				selStartRow,
				selEndCol,
				selEndRow,
			)
			mu.Unlock()

			event.Frame(context.Ops)
		}
	}
}

var _ ioevent.Filter = key.Filter{}
