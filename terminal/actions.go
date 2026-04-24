package terminal

import "image/color"

func (t *Terminal) handleControl(r rune) {
	switch r {
	case '\r': // CR
		t.buf.CarriageReturn()
	case '\n', '\v', '\f': // LF
		t.buf.LineFeed()
	case '\b': // BS
		t.buf.Backspace()
	case '\t': // HT
		t.buf.Tab()
	case 0x07: // BEL — could ring a bell
	case 0x0E, 0x0F: // SO/SI — character set switching (ignore for now)
	}
}

// handleCSI processes a complete CSI sequence (ESC [ params cmd).
func (t *Terminal) handleCSI(cmd byte, params []int) {
	p := func(i, def int) int {
		if i < len(params) && params[i] != 0 {
			return params[i]
		}
		return def
	}

	switch cmd {
	case '@': // ICH - insert blank chars
		t.buf.InsertBlankChars(p(0, 1))
	case 'A': // CUU — cursor up
		t.buf.CursorUp(p(0, 1))
	case 'B': // CUD — cursor down
		t.buf.CursorDown(p(0, 1))
	case 'C': // CUF — cursor forward
		t.buf.CursorForward(p(0, 1))
	case 'a': // HPR — horizontal position relative
		t.buf.CursorForward(p(0, 1))
	case 'D': // CUB — cursor back
		t.buf.CursorBack(p(0, 1))
	case 'e': // VPR — vertical position relative
		t.buf.CursorDown(p(0, 1))
	case 'E': // CNL — cursor next line
		t.buf.CursorDown(p(0, 1))
		t.buf.CarriageReturn()
	case 'F': // CPL — cursor previous line
		t.buf.CursorUp(p(0, 1))
		t.buf.CarriageReturn()
	case 'G': // CHA — cursor horizontal absolute
		t.buf.MoveCursor(p(0, 1)-1, t.buf.CursorRow)
	case 'H', 'f': // CUP / HVP — cursor position (1-based)
		t.buf.MoveCursor(p(1, 1)-1, p(0, 1)-1)
	case 'd': // VPA — line position absolute (1-based)
		t.buf.MoveCursor(t.buf.CursorCol, p(0, 1)-1)
	case 'L': // IL — insert lines
		t.buf.InsertLines(p(0, 1))
	case 'M': // DL — delete lines
		t.buf.DeleteLines(p(0, 1))
	case 'J': // ED — erase in display
		t.buf.EraseInDisplay(p(0, 0))
	case 'K': // EL — erase in line
		t.buf.EraseInLine(p(0, 0))
	case 'P': // DCH — delete chars
		t.buf.DeleteChars(p(0, 1))
	case 'X': // ECH — erase chars
		t.buf.EraseChars(p(0, 1))
	case 'S': // SU — scroll up
		t.buf.ScrollUp(p(0, 1))
	case 'T': // SD — scroll down
		t.buf.ScrollDown(p(0, 1))
	case 'm': // SGR — select graphic rendition (colors + attrs)
		t.handleSGR(params)
	case 'h': // SM — set mode
		t.handleSetMode(params, true)
	case 'l': // RM — reset mode
		t.handleSetMode(params, false)
	case 'r': // DECSTBM — set scrolling region (1-based, inclusive)
		top := p(0, 1) - 1
		bottom := p(1, t.buf.Rows()) - 1
		if top <= bottom {
			t.buf.SetScrollRegion(top, bottom)
			t.buf.MoveCursor(0, 0)
		} else {
			t.buf.ResetScrollRegion()
			t.buf.MoveCursor(0, 0)
		}
	case 's': // SCOSC — save cursor
	case 'u': // SCORC — restore cursor
	}
}

// handleESC processes ESC sequences (ESC + single byte).
func (t *Terminal) handleESC(cmd byte) {
	switch cmd {
	case 'M': // RI — reverse index (scroll down)
		t.buf.ReverseIndex()
	case '7': // DECSC — save cursor
	case '8': // DECRC — restore cursor
	case 'c': // RIS — full reset
		t.buf.EraseInDisplay(2)
		t.buf.MoveCursor(0, 0)
		t.buf.ResetAttr()
	}
}

// handleSGR processes SGR (color/attribute) parameters.
// Reference: https://en.wikipedia.org/wiki/ANSI_escape_code#SGR
func (t *Terminal) handleSGR(params []int) {
	attr := t.buf.CurrentAttr()
	if len(params) == 0 {
		params = []int{0}
	}

	for i := 0; i < len(params); i++ {
		switch params[i] {
		case 0: // reset
			attr = Attr{FG: DefaultFG, BG: DefaultBG}
		case 1:
			attr.Bold = true
		case 3:
			attr.Italic = true
		case 4:
			attr.Underline = true
		case 5:
			attr.Blink = true
		case 7:
			attr.Reverse = true
		case 22:
			attr.Bold = false
		case 23:
			attr.Italic = false
		case 24:
			attr.Underline = false
		case 27:
			attr.Reverse = false

		// Standard foreground colors (30–37, 90–97)
		case 30, 31, 32, 33, 34, 35, 36, 37:
			attr.FG = ansi16[params[i]-30]
		case 39:
			attr.FG = DefaultFG
		case 90, 91, 92, 93, 94, 95, 96, 97:
			attr.FG = ansi16[8+params[i]-90]

		// Standard background colors (40–47, 100–107)
		case 40, 41, 42, 43, 44, 45, 46, 47:
			attr.BG = ansi16[params[i]-40]
		case 49:
			attr.BG = DefaultBG
		case 100, 101, 102, 103, 104, 105, 106, 107:
			attr.BG = ansi16[8+params[i]-100]

		// 256-color / truecolor foreground: ESC[38;5;Nm or ESC[38;2;R;G;Bm
		case 38:
			if i+2 < len(params) && params[i+1] == 5 {
				attr.FG = ansi256[params[i+2]]
				i += 2
			} else if i+4 < len(params) && params[i+1] == 2 {
				attr.FG = color.RGBA{
					R: uint8(params[i+2]),
					G: uint8(params[i+3]),
					B: uint8(params[i+4]),
					A: 0xFF,
				}
				i += 4
			}

		// 256-color / truecolor background: ESC[48;5;Nm or ESC[48;2;R;G;Bm
		case 48:
			if i+2 < len(params) && params[i+1] == 5 {
				attr.BG = ansi256[params[i+2]]
				i += 2
			} else if i+4 < len(params) && params[i+1] == 2 {
				attr.BG = color.RGBA{
					R: uint8(params[i+2]),
					G: uint8(params[i+3]),
					B: uint8(params[i+4]),
					A: 0xFF,
				}
				i += 4
			}
		}
	}

	t.buf.SetAttr(attr)
}

// handleSetMode processes DEC private mode sequences (?h / ?l).
func (t *Terminal) handleSetMode(params []int, set bool) {
	for _, p := range params {
		switch p {
		case 1049: // alternate screen with save/restore cursor
			if set {
				t.buf.SwitchToAltScreen()
			} else {
				t.buf.SwitchToPrimaryScreen()
			}
		case 47, 1047: // alternate screen (simpler variant)
			if set {
				t.buf.SwitchToAltScreen()
			} else {
				t.buf.SwitchToPrimaryScreen()
			}
		case 25: // cursor visibility — stub
		}
	}
}

// ansi16 maps the 16 standard ANSI colors.
var ansi16 = [16]color.RGBA{
	{0x00, 0x00, 0x00, 0xFF}, // 0 black
	{0xCC, 0x00, 0x00, 0xFF}, // 1 red
	{0x00, 0xCC, 0x00, 0xFF}, // 2 green
	{0xCC, 0xCC, 0x00, 0xFF}, // 3 yellow
	{0x00, 0x00, 0xCC, 0xFF}, // 4 blue
	{0xCC, 0x00, 0xCC, 0xFF}, // 5 magenta
	{0x00, 0xCC, 0xCC, 0xFF}, // 6 cyan
	{0xCC, 0xCC, 0xCC, 0xFF}, // 7 white
	{0x55, 0x55, 0x55, 0xFF}, // 8 bright black (gray)
	{0xFF, 0x55, 0x55, 0xFF}, // 9 bright red
	{0x55, 0xFF, 0x55, 0xFF}, // 10 bright green
	{0xFF, 0xFF, 0x55, 0xFF}, // 11 bright yellow
	{0x55, 0x55, 0xFF, 0xFF}, // 12 bright blue
	{0xFF, 0x55, 0xFF, 0xFF}, // 13 bright magenta
	{0x55, 0xFF, 0xFF, 0xFF}, // 14 bright cyan
	{0xFF, 0xFF, 0xFF, 0xFF}, // 15 bright white
}

// ansi256 maps the 256-color palette.
// First 16 are the standard colors above; 16–231 are a 6×6×6 color cube;
// 232–255 are a grayscale ramp.
var ansi256 [256]color.RGBA

func init() {
	// Copy standard 16
	for i := 0; i < 16; i++ {
		ansi256[i] = ansi16[i]
	}
	// 6×6×6 color cube (indices 16–231)
	for r := 0; r < 6; r++ {
		for g := 0; g < 6; g++ {
			for b := 0; b < 6; b++ {
				idx := 16 + 36*r + 6*g + b
				toV := func(v int) uint8 {
					if v == 0 {
						return 0
					}
					return uint8(55 + v*40)
				}
				ansi256[idx] = color.RGBA{toV(r), toV(g), toV(b), 0xFF}
			}
		}
	}
	// Grayscale ramp (indices 232–255)
	for i := 0; i < 24; i++ {
		v := uint8(8 + i*10)
		ansi256[232+i] = color.RGBA{v, v, v, 0xFF}
	}
}
