package buffer

func (b *Buffer) lineFeed() {
	if b.CursorRow < b.scrollBottom {
		b.CursorRow++
		return
	}
	if b.CursorRow < b.scrollTop || b.CursorRow > b.scrollBottom {
		b.CursorRow = min(b.rows-1, b.CursorRow+1)
		return
	}
	b.scrollUpOne()
	if b.CursorRow > b.scrollBottom {
		b.CursorRow = b.scrollBottom
	}
}

func (b *Buffer) scrollUpOne() {
	if b.scrollTop == 0 && b.scrollBottom == b.rows-1 && !b.usingAltScreen {
		b.scrollback = append(b.scrollback, b.cells[0])
	}
	for r := b.scrollTop; r < b.scrollBottom; r++ {
		b.cells[r] = b.cells[r+1]
	}
	b.cells[b.scrollBottom] = makeRow(b.cols)
}

func (b *Buffer) scrollDownOne() {
	for r := b.scrollBottom; r > b.scrollTop; r-- {
		b.cells[r] = b.cells[r-1]
	}
	b.cells[b.scrollTop] = makeRow(b.cols)
}

func (b *Buffer) SetScrollRegion(top, bottom int) {
	b.scrollTop = clamp(top, 0, b.rows-1)
	b.scrollBottom = clamp(bottom, b.scrollTop, b.rows-1)
	b.clampCursor()
}

func (b *Buffer) ResetScrollRegion() {
	b.scrollTop = 0
	b.scrollBottom = b.rows - 1
}

func (b *Buffer) ScrollUp(n int) {
	if n <= 0 {
		n = 1
	}
	for i := 0; i < n; i++ {
		b.scrollUpOne()
	}
}

func (b *Buffer) ScrollDown(n int) {
	if n <= 0 {
		n = 1
	}
	for i := 0; i < n; i++ {
		b.scrollDownOne()
	}
}

func (b *Buffer) ReverseIndex() {
	if b.CursorRow > b.scrollTop {
		b.CursorRow--
		return
	}
	if b.CursorRow == b.scrollTop {
		b.scrollDownOne()
	}
	if b.CursorRow < b.scrollTop {
		b.CursorRow = b.scrollTop
	}
	if b.CursorRow >= b.rows {
		b.CursorRow = b.rows - 1
	}
	if b.CursorRow < 0 {
		b.CursorRow = 0
	}
	if b.CursorCol >= b.cols {
		b.CursorCol = b.cols - 1
	}
	if b.CursorCol < 0 {
		b.CursorCol = 0
	}
	if b.scrollBottom >= b.rows {
		b.scrollBottom = b.rows - 1
	}
	if b.scrollTop < 0 {
		b.scrollTop = 0
	}
	if b.scrollTop > b.scrollBottom {
		b.scrollTop = b.scrollBottom
	}
}
