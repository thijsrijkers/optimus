package buffer

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
	DefaultFG = color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	DefaultBG = color.RGBA{R: 0x28, G: 0x2C, B: 0x34, A: 0xFF}
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

	CursorCol int
	CursorRow int

	currentAttr Attr

	scrollback [][]Cell

	altCells       [][]Cell
	altCursorCol   int
	altCursorRow   int
	usingAltScreen bool

	scrollTop    int
	scrollBottom int
}
