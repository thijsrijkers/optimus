package buffer

func (b *Buffer) SwitchToAltScreen() {
	if b.usingAltScreen {
		return
	}
	b.altCells = b.cells
	b.altCursorCol = b.CursorCol
	b.altCursorRow = b.CursorRow
	b.cells = makeGrid(b.cols, b.rows)
	b.CursorCol, b.CursorRow = 0, 0
	b.ResetScrollRegion()
	b.usingAltScreen = true
}

func (b *Buffer) SwitchToPrimaryScreen() {
	if !b.usingAltScreen {
		return
	}
	b.cells = b.altCells
	b.CursorCol = b.altCursorCol
	b.CursorRow = b.altCursorRow
	b.ResetScrollRegion()
	b.usingAltScreen = false
}

func (b *Buffer) UsingAltScreen() bool {
	return b.usingAltScreen
}
