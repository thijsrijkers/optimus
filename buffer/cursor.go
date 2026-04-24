package buffer

// PutRune writes a rune at the current cursor position and advances the cursor.
func (b *Buffer) PutRune(r rune) {
	if b.CursorCol >= b.cols {
		b.CursorCol = 0
		b.lineFeed()
	}
	b.cells[b.CursorRow][b.CursorCol] = Cell{Char: r, Attr: b.currentAttr}
	b.CursorCol++
}

func (b *Buffer) MoveCursor(col, row int) {
	b.CursorCol = col
	b.CursorRow = row
	b.clampCursor()
}

func (b *Buffer) CursorUp(n int)      { b.CursorRow = max(0, b.CursorRow-n) }
func (b *Buffer) CursorDown(n int)    { b.CursorRow = min(b.rows-1, b.CursorRow+n) }
func (b *Buffer) CursorForward(n int) { b.CursorCol = min(b.cols-1, b.CursorCol+n) }
func (b *Buffer) CursorBack(n int)    { b.CursorCol = max(0, b.CursorCol-n) }

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
