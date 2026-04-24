package buffer

// New creates a Buffer with the given dimensions.
func New(cols, rows int) *Buffer {
	b := &Buffer{
		cols:         cols,
		rows:         rows,
		currentAttr:  Attr{FG: DefaultFG, BG: DefaultBG},
		scrollTop:    0,
		scrollBottom: rows - 1,
	}
	b.cells = makeGrid(cols, rows)
	return b
}

// Resize resizes the buffer, preserving content where possible.
func (b *Buffer) Resize(cols, rows int) {
	newCells := resizeGrid(b.cells, b.cols, b.rows, cols, rows)

	if b.usingAltScreen {
		b.altCells = resizeGrid(b.altCells, b.cols, b.rows, cols, rows)
		b.altCursorCol = clamp(b.altCursorCol, 0, cols-1)
		b.altCursorRow = clamp(b.altCursorRow, 0, rows-1)
	}
	b.cols, b.rows = cols, rows
	b.cells = newCells
	b.scrollTop = clamp(b.scrollTop, 0, rows-1)
	b.scrollBottom = clamp(b.scrollBottom, b.scrollTop, rows-1)
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

func (b *Buffer) clampCursor() {
	b.CursorCol = clamp(b.CursorCol, 0, b.cols-1)
	b.CursorRow = clamp(b.CursorRow, 0, b.rows-1)
}

func (b *Buffer) eraseLine(row int) {
	for c := range b.cells[row] {
		b.cells[row][c] = blankCell(b.currentAttr)
	}
}

// SetAttr updates the current drawing style.
func (b *Buffer) SetAttr(a Attr)    { b.currentAttr = a }
func (b *Buffer) CurrentAttr() Attr { return b.currentAttr }
func (b *Buffer) ResetAttr() {
	b.currentAttr = Attr{FG: DefaultFG, BG: DefaultBG}
}
