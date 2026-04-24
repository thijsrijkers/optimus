package core

import (
	"optimus/ansi"
)

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
func (t *Terminal) handleCSI(cmd byte, params []int, private bool) {
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
		t.buf.SetAttr(ansi.ApplySGR(t.buf.CurrentAttr(), params))
	case 'h': // SM — set mode
		if private {
			ansi.ApplyPrivateMode(&t.modeState, t.buf, params, true)
		}
	case 'l': // RM — reset mode
		if private {
			ansi.ApplyPrivateMode(&t.modeState, t.buf, params, false)
		}
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
