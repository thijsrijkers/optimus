package window

import (
	"strings"

	"optimus/buffer"
)

func selectionText(buf *buffer.Buffer, col1, row1, col2, row2 int) string {
	if row1 > row2 || (row1 == row2 && col1 > col2) {
		col1, col2 = col2, col1
		row1, row2 = row2, row1
	}
	var out strings.Builder
	for row := row1; row <= row2; row++ {
		startCol := 0
		endCol := buf.Cols() - 1
		if row == row1 {
			startCol = col1
		}
		if row == row2 {
			endCol = col2
		}
		if startCol < 0 {
			startCol = 0
		}
		if endCol >= buf.Cols() {
			endCol = buf.Cols() - 1
		}
		if endCol < startCol {
			continue
		}
		var line strings.Builder
		for col := startCol; col <= endCol; col++ {
			ch := buf.Cell(col, row).Char
			if ch == 0 {
				ch = ' '
			}
			line.WriteRune(ch)
		}
		out.WriteString(strings.TrimRight(line.String(), " "))
		if row < row2 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}
