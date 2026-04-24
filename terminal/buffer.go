// Package buffer implements the terminal screen buffer:
// a 2D grid of cells, each holding a rune + style.
// Handles cursor movement, scrolling, and primary/alternate screens.
package terminal

import "image/color"

// Attr holds text rendering attributes.
type Attr struct {
	FG        color.RGBA
	BG        color.RGBA
	Bold      bool
	Italic    bool
	Underline bool
	Blink     bool
	Reverse   bool
}

// DefaultFG / DefaultBG are the terminal defaults.
var (
	DefaultFG = color.RGBA{R: 0xCC, G: 0xCC, B: 0xCC, A: 0xFF}
	DefaultBG = color.RGBA{R: 0x1E, G: 0x1E, B: 0x1E, A: 0xFF}
)

// Cell is a single character position on screen.
type Cell struct {
	Char rune
	Attr Attr
}

// Buffer is the terminal screen: a grid of cells plus cursor state.
type Buffer struct {
	cols, rows int
	cells      [][]Cell // [row][col]

	// Cursor position (0-based)
	CursorCol int
	CursorRow int

	// Current drawing style
	currentAttr Attr

	// Scrollback (primary screen only)
	scrollback [][]Cell

	// Alternate screen (used by vim, htop, etc.)
	altCells      [][]Cell
	altCursorCol  int
	altCursorRow  int
	usingAltScreen bool
}

// New creates a Buffer with the given dimensions.
func NewBuffer(cols, rows int) *Buffer {
	b := &Buffer{
		cols: cols,
		rows: rows,
		currentAttr: Attr{FG: DefaultFG, BG: DefaultBG},
	}
	b.cells = makeGrid(cols, rows)
	return b
}

// Resize resizes the buffer, preserving content where possible.
func (b *Buffer) Resize(cols, rows int) {
	newCells := makeGrid(cols, rows)
	for r := 0; r < min(rows, b.rows); r++ {
		for c := 0; c < min(cols, b.cols); c++ {
			newCells[r][c] = b.cells[r][c]
		}
	}
	b.cols, b.rows = cols, rows
	b.cells = newCells
	b.clampCursor()
}

// Cell returns the cell at (col, row). Returns empty cell if out of bounds.
func (b *Buffer) Cell(col, row int) Cell {
	if row < 0 || row >= b.rows || col < 0 || col >= b.cols {
		return Cell{Char: ' ', Attr: Attr{FG: DefaultFG, BG: DefaultBG}}
	}
	return b.cells[row][col]
}

// Cols / Rows return current dimensions.
func (b *Buffer) Cols() int { return b.cols }
func (b *Buffer) Rows() int { return b.rows }

// --- Writing ---

// PutRune writes a rune at the current cursor position and advances the cursor.
func (b *Buffer) PutRune(r rune) {
	if b.CursorCol >= b.cols {
		b.CursorCol = 0
		b.lineFeed()
	}
	b.cells[b.CursorRow][b.CursorCol] = Cell{Char: r, Attr: b.currentAttr}
	b.CursorCol++
}

// --- Cursor movement ---

func (b *Buffer) MoveCursor(col, row int) {
	b.CursorCol = col
	b.CursorRow = row
	b.clampCursor()
}

func (b *Buffer) CursorUp(n int)    { b.CursorRow = max(0, b.CursorRow-n) }
func (b *Buffer) CursorDown(n int)  { b.CursorRow = min(b.rows-1, b.CursorRow+n) }
func (b *Buffer) CursorForward(n int) { b.CursorCol = min(b.cols-1, b.CursorCol+n) }
func (b *Buffer) CursorBack(n int)  { b.CursorCol = max(0, b.CursorCol-n) }

func (b *Buffer) CarriageReturn() { b.CursorCol = 0 }
func (b *Buffer) LineFeed()       { b.lineFeed() }

func (b *Buffer) Backspace() {
	if b.CursorCol > 0 {
		b.CursorCol--
	}
}

func (b *Buffer) Tab() {
	next := ((b.CursorCol / 8) + 1) * 8
	b.CursorCol = min(next, b.cols-1)
}

// --- Erasing ---

// EraseInLine erases part of the current line.
//
//	mode 0 = cursor to end, 1 = start to cursor, 2 = whole line
func (b *Buffer) EraseInLine(mode int) {
	start, end := 0, b.cols
	switch mode {
	case 0:
		start = b.CursorCol
	case 1:
		end = b.CursorCol + 1
	}
	for c := start; c < end; c++ {
		b.cells[b.CursorRow][c] = blankCell(b.currentAttr)
	}
}

// EraseInDisplay erases part of the display.
//
//	mode 0 = cursor to end, 1 = start to cursor, 2 = whole screen, 3 = whole screen + scrollback
func (b *Buffer) EraseInDisplay(mode int) {
	switch mode {
	case 0:
		b.EraseInLine(0)
		for r := b.CursorRow + 1; r < b.rows; r++ {
			b.eraseLine(r)
		}
	case 1:
		for r := 0; r < b.CursorRow; r++ {
			b.eraseLine(r)
		}
		b.EraseInLine(1)
	case 2, 3:
		for r := 0; r < b.rows; r++ {
			b.eraseLine(r)
		}
		if mode == 3 {
			b.scrollback = nil
		}
	}
}

// --- Scrolling ---

func (b *Buffer) lineFeed() {
	if b.CursorRow < b.rows-1 {
		b.CursorRow++
		return
	}
	// Scroll up: push top line into scrollback, shift rows up.
	if !b.usingAltScreen {
		b.scrollback = append(b.scrollback, b.cells[0])
	}
	copy(b.cells, b.cells[1:])
	b.cells[b.rows-1] = makeRow(b.cols)
}

// --- Alternate screen ---

func (b *Buffer) SwitchToAltScreen() {
	if b.usingAltScreen {
		return
	}
	b.altCells = b.cells
	b.altCursorCol = b.CursorCol
	b.altCursorRow = b.CursorRow
	b.cells = makeGrid(b.cols, b.rows)
	b.CursorCol, b.CursorRow = 0, 0
	b.usingAltScreen = true
}

func (b *Buffer) SwitchToPrimaryScreen() {
	if !b.usingAltScreen {
		return
	}
	b.cells = b.altCells
	b.CursorCol = b.altCursorCol
	b.CursorRow = b.altCursorRow
	b.usingAltScreen = false
}

// --- Styling ---

// SetAttr updates the current drawing style.
func (b *Buffer) SetAttr(a Attr) { b.currentAttr = a }
func (b *Buffer) CurrentAttr() Attr { return b.currentAttr }
func (b *Buffer) ResetAttr() {
	b.currentAttr = Attr{FG: DefaultFG, BG: DefaultBG}
}

// --- Helpers ---

func (b *Buffer) clampCursor() {
	b.CursorCol = clamp(b.CursorCol, 0, b.cols-1)
	b.CursorRow = clamp(b.CursorRow, 0, b.rows-1)
}

func (b *Buffer) eraseLine(row int) {
	for c := range b.cells[row] {
		b.cells[row][c] = blankCell(b.currentAttr)
	}
}

func makeGrid(cols, rows int) [][]Cell {
	g := make([][]Cell, rows)
	for i := range g {
		g[i] = makeRow(cols)
	}
	return g
}

func makeRow(cols int) []Cell {
	row := make([]Cell, cols)
	for i := range row {
		row[i] = Cell{Char: ' ', Attr: Attr{FG: DefaultFG, BG: DefaultBG}}
	}
	return row
}

func blankCell(a Attr) Cell {
	return Cell{Char: ' ', Attr: Attr{FG: a.FG, BG: a.BG}}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
