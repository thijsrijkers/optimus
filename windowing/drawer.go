package windowing

import (
	"image"
	"image/color"
	"optimus/terminal"
	"unicode/utf8"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

func drawCells(context layout.Context, terminal *terminal.Terminal, theme *material.Theme) {
	buf := terminal.Buffer()
	cols := buf.Cols()
	rows := buf.Rows()

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
				glyph(context, theme, cell.Char, x, y, fg, cell.Attr.Bold)
			}
		}
	}
}

func glyph(context layout.Context, theme *material.Theme, r rune, x, y int, foreground color.RGBA, bold bool) {
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
	label.Color = fgNRGBA
	if bold {
		label.Font.Weight = font.Bold
	}
	label.Layout(context)
}
