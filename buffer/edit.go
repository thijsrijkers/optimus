package buffer

func (b *Buffer) InsertBlankChars(n int) {
	if n <= 0 {
		n = 1
	}
	if b.CursorCol >= b.cols {
		return
	}
	if n > b.cols-b.CursorCol {
		n = b.cols - b.CursorCol
	}
	row := b.cells[b.CursorRow]
	copy(row[b.CursorCol+n:], row[b.CursorCol:b.cols-n])
	for c := b.CursorCol; c < b.CursorCol+n; c++ {
		row[c] = blankCell(b.currentAttr)
	}
}

func (b *Buffer) DeleteChars(n int) {
	if n <= 0 {
		n = 1
	}
	if b.CursorCol >= b.cols {
		return
	}
	if n > b.cols-b.CursorCol {
		n = b.cols - b.CursorCol
	}
	row := b.cells[b.CursorRow]
	copy(row[b.CursorCol:], row[b.CursorCol+n:])
	for c := b.cols - n; c < b.cols; c++ {
		row[c] = blankCell(b.currentAttr)
	}
}

func (b *Buffer) EraseChars(n int) {
	if n <= 0 {
		n = 1
	}
	end := min(b.cols, b.CursorCol+n)
	for c := b.CursorCol; c < end; c++ {
		b.cells[b.CursorRow][c] = blankCell(b.currentAttr)
	}
}

func (b *Buffer) InsertLines(n int) {
	if b.CursorRow < b.scrollTop || b.CursorRow > b.scrollBottom {
		return
	}
	if n <= 0 {
		n = 1
	}
	if n > b.scrollBottom-b.CursorRow+1 {
		n = b.scrollBottom - b.CursorRow + 1
	}
	for r := b.scrollBottom; r >= b.CursorRow+n; r-- {
		b.cells[r] = b.cells[r-n]
	}
	for r := b.CursorRow; r < b.CursorRow+n; r++ {
		b.cells[r] = makeRow(b.cols)
	}
}

func (b *Buffer) DeleteLines(n int) {
	if b.CursorRow < b.scrollTop || b.CursorRow > b.scrollBottom {
		return
	}
	if n <= 0 {
		n = 1
	}
	if n > b.scrollBottom-b.CursorRow+1 {
		n = b.scrollBottom - b.CursorRow + 1
	}
	for r := b.CursorRow; r <= b.scrollBottom-n; r++ {
		b.cells[r] = b.cells[r+n]
	}
	for r := b.scrollBottom - n + 1; r <= b.scrollBottom; r++ {
		b.cells[r] = makeRow(b.cols)
	}
}

// EraseInLine erases part of the current line.
// mode 0 = cursor to end, 1 = start to cursor, 2 = whole line
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
// mode 0 = cursor to end, 1 = start to cursor, 2 = whole screen, 3 = whole screen + scrollback
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
