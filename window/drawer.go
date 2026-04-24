package window

import (
	"image"
	"image/color"
	core "optimus/core"
	"unicode/utf8"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

func drawCells(context layout.Context, terminal *core.Terminal, theme *material.Theme, fontSize int, cellW, cellH int, selectionActive bool, selStartCol, selStartRow, selEndCol, selEndRow int) {
	buf := terminal.Buffer()
	cols := buf.Cols()
	rows := buf.Rows()
	if selectionActive && (selStartRow > selEndRow || (selStartRow == selEndRow && selStartCol > selEndCol)) {
		selStartCol, selEndCol = selEndCol, selStartCol
		selStartRow, selEndRow = selEndRow, selStartRow
	}

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := buf.Cell(col, row)

			x := col * cellW
			y := row * cellH
			cellRect := image.Rect(x, y, x+cellW, y+cellH)

			// Draw background
			background := cell.Attr.BG
			if cell.Attr.Reverse {
				background = cell.Attr.FG
			}
			bgNRGBA := color.NRGBA{R: background.R, G: background.G, B: background.B, A: background.A}
			paint.FillShape(context.Ops,
				bgNRGBA,
				clip.Rect(cellRect).Op(),
			)

			if selectionActive && isCellSelected(col, row, selStartCol, selStartRow, selEndCol, selEndRow) {
				paint.FillShape(context.Ops,
					color.NRGBA{R: 0x5A, G: 0x7A, B: 0xA8, A: 0x88},
					clip.Rect(cellRect).Op(),
				)
			}

			// Draw cursor block
			if col == buf.CursorCol && row == buf.CursorRow {
				paint.FillShape(context.Ops,
					color.NRGBA{R: 0xCC, G: 0xCC, B: 0xCC, A: 0x88},
					clip.Rect(cellRect).Op(),
				)
			}

			// Draw glyph (skip space)
			if cell.Char != ' ' && cell.Char != 0 {
				fg := cell.Attr.FG
				if cell.Attr.Reverse {
					fg = cell.Attr.BG
				}
				glyph(context, theme, cell.Char, x, y, fg, cell.Attr.Bold, fontSize, cellW, cellH)
			}
		}
	}
}

func isCellSelected(col, row, startCol, startRow, endCol, endRow int) bool {
	if row < startRow || row > endRow {
		return false
	}
	if startRow == endRow {
		return row == startRow && col >= startCol && col <= endCol
	}
	if row == startRow {
		return col >= startCol
	}
	if row == endRow {
		return col <= endCol
	}
	return true
}

func glyph(context layout.Context, theme *material.Theme, r rune, x, y int, foreground color.RGBA, bold bool, fontSize int, cellW, cellH int) {
	// Encode the rune as a  for the label widget.
	var buf [4]byte
	n := utf8.EncodeRune(buf[:], r)
	str := string(buf[:n])

	fgNRGBA := color.NRGBA{R: foreground.R, G: foreground.G, B: foreground.B, A: foreground.A}

	// Push a translation offset for this cell.
	defer op.Offset(image.Point{X: x, Y: y}).Push(context.Ops).Pop()

	// Constrain to cell size.
	context.Constraints = layout.Exact(image.Point{X: cellW, Y: cellH})

	label := material.Label(theme, unit.Sp(fontSize), str)
	label.Font.Typeface = fontFamily
	label.Color = fgNRGBA
	if bold {
		label.Font.Weight = font.Bold
	}
	label.Layout(context)
}
