package buffer

func makeGrid(cols, rows int) [][]Cell {
	g := make([][]Cell, rows)
	for i := range g {
		g[i] = makeRow(cols)
	}
	return g
}

func resizeGrid(old [][]Cell, oldCols, oldRows, cols, rows int) [][]Cell {
	newCells := makeGrid(cols, rows)
	for r := 0; r < min(rows, oldRows); r++ {
		for c := 0; c < min(cols, oldCols); c++ {
			newCells[r][c] = old[r][c]
		}
	}
	return newCells
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
