package window

import (
	"image"
	"image/color"
	"io"
	"strings"

	"optimus/tabs"

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
	"gioui.org/unit"
	"gioui.org/widget/material"
)

const (
	defaultFontSize = 13
	fontFamily      = "Menlo"
	initCols        = 80
	initRows        = 24
)

func cellMetrics(context layout.Context, fontSize int) (cellW, cellH int) {
	fontPx := context.Sp(unit.Sp(fontSize))
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
	activeTabID := -1
	tabHits := []tabHit{}
	uiFontSize := defaultFontSize

	tabManager, err := tabs.New(shell, initCols, initRows, window.Invalidate)
	if err != nil {
		return err
	}
	defer tabManager.CloseAll()

	theme := material.NewTheme()

	for {
		switch event := window.Event().(type) {
		case app.DestroyEvent:
			return event.Err
		case app.FrameEvent:
			session := tabManager.Active()
			if session == nil {
				continue
			}
			if session.ID() != activeTabID {
				activeTabID = session.ID()
				selecting = false
				hasSelection = false
			}
			term := session.Terminal()
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
			cellW, cellH := cellMetrics(context, uiFontSize)
			paint.Fill(&operationList, color.NRGBA{R: 0x28, G: 0x2C, B: 0x34, A: 0xFF})
			tabBarH, hits := drawTabBar(context, theme, tabManager.List())
			tabHits = hits

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
						if isZoomInShortcut(ev) {
							if uiFontSize < 28 {
								uiFontSize++
							}
							continue
						}
						if isZoomOutShortcut(ev) {
							if uiFontSize > 9 {
								uiFontSize--
							}
							continue
						}
						if isNewTabShortcut(ev) {
							_ = tabManager.NewTab()
							selecting, hasSelection = false, false
							continue
						}
						if isCloseTabShortcut(ev) {
							tabManager.CloseActive()
							selecting, hasSelection = false, false
							continue
						}
						if isPrevTabShortcut(ev) {
							tabManager.Prev()
							selecting, hasSelection = false, false
							continue
						}
						if isNextTabShortcut(ev) {
							tabManager.Next()
							selecting, hasSelection = false, false
							continue
						}
						if !isCopyShortcut(ev) && !isPasteShortcut(ev) {
							hasSelection = false
						}
						if isCopyShortcut(ev) {
							if hasSelection {
								session.Lock()
								text := selectionText(term.Buffer(), selStartCol, selStartRow, selEndCol, selEndRow)
								session.Unlock()
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
							session.WriteInput(seq)
						}
					}
				case key.EditEvent:
					if ev.Text != "" {
						session.WriteInput([]byte(ev.Text))
					}
				case transfer.DataEvent:
					if ev.Type == "text/plain" {
						r := ev.Open()
						data, err := io.ReadAll(r)
						r.Close()
						if err == nil && len(data) > 0 {
							session.WriteInput(data)
						}
					}
				case pointer.Event:
					if int(ev.Position.Y) < tabBarH {
						if ev.Kind == pointer.Press && ev.Buttons.Contain(pointer.ButtonPrimary) {
							x := int(ev.Position.X)
							for _, hit := range tabHits {
								if x >= hit.startX && x <= hit.endX {
									if hit.addNew {
										_ = tabManager.NewTab()
									} else if hit.close {
										tabManager.CloseAt(hit.index)
									} else {
										tabManager.ActivateAt(hit.index)
									}
									selecting, hasSelection = false, false
									break
								}
							}
						}
						continue
					}
					ev.Position.Y -= float32(tabBarH)
					col, row := pointerToCell(ev, cellW, cellH)
					session.Lock()
					buf := term.Buffer()
					maxCol := buf.Cols() - 1
					maxRow := buf.Rows() - 1
					session.Unlock()
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

					proto := term.MouseProtocol()
					pressed := ev.Buttons &^ lastButtons
					released := lastButtons &^ ev.Buttons

					forceSelection := ev.Modifiers.Contain(key.ModShift)
					if proto.Enabled && !forceSelection {
						sendPointerToPTY(session, ev, pressed, released, col, row, proto)
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
							session.Lock()
							text := selectionText(term.Buffer(), selStartCol, selStartRow, selEndCol, selEndRow)
							session.Unlock()
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
			rows := (context.Constraints.Max.Y - tabBarH) / cellH
			if cols < 1 {
				cols = 1
			}
			if rows < 1 {
				rows = 1
			}
			tabManager.ResizeAll(cols, rows)

			session.Lock()
			deferY := op.Offset(image.Point{Y: tabBarH}).Push(context.Ops)
			drawCells(
				context,
				term,
				theme,
				uiFontSize,
				cellW,
				cellH,
				hasSelection || selecting,
				selStartCol,
				selStartRow,
				selEndCol,
				selEndRow,
			)
			deferY.Pop()
			session.Unlock()

			event.Frame(context.Ops)
		}
	}
}

var _ ioevent.Filter = key.Filter{}
