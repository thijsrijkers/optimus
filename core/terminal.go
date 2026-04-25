package core

import (
	"strings"
	"unicode/utf8"

	"optimus/ansi"
	"optimus/buffer"
)

type Terminal struct {
	buf     *buffer.Buffer
	parser  *Parser
	utf8Buf [4]byte
	utf8Len int
	utf8Rem int

	modeState ansi.ModeState
	title     string
}

type MouseProtocol struct {
	Enabled bool
	Drag    bool
	Motion  bool
	SGR     bool
}

func New(cols, rows int) *Terminal {
	return &Terminal{
		buf:    buffer.New(cols, rows),
		parser: NewParser(),
	}
}

func (terminal *Terminal) Buffer() *buffer.Buffer { return terminal.buf }

func (terminal *Terminal) Resize(cols, rows int) { terminal.buf.Resize(cols, rows) }

func (terminal *Terminal) MouseProtocol() MouseProtocol {
	enabled := terminal.modeState.MouseReport || terminal.modeState.MouseDrag || terminal.modeState.MouseMotion || terminal.modeState.MouseSGR
	return MouseProtocol{
		Enabled: enabled,
		Drag:    terminal.modeState.MouseDrag,
		Motion:  terminal.modeState.MouseMotion,
		SGR:     terminal.modeState.MouseSGR,
	}
}

func (terminal *Terminal) Title() string { return terminal.title }

func (terminal *Terminal) UsingAltScreen() bool {
	return terminal.buf.UsingAltScreen()
}

func (terminal *Terminal) Write(data []byte) {
	for _, b := range data {
		terminal.feedByte(b)
	}
}

func (terminal *Terminal) feedByte(b byte) {
	if terminal.utf8Rem > 0 {
		terminal.utf8Buf[terminal.utf8Len] = b
		terminal.utf8Len++
		terminal.utf8Rem--
		if terminal.utf8Rem == 0 {
			r, _ := utf8.DecodeRune(terminal.utf8Buf[:terminal.utf8Len])
			terminal.applyActions(terminal.parser.FeedRune(r))
			terminal.utf8Len = 0
		}
		return
	}

	if b >= 0xF0 {
		terminal.utf8Buf[0], terminal.utf8Len, terminal.utf8Rem = b, 1, 3
		return
	} else if b >= 0xE0 {
		terminal.utf8Buf[0], terminal.utf8Len, terminal.utf8Rem = b, 1, 2
		return
	} else if b >= 0xC0 {
		terminal.utf8Buf[0], terminal.utf8Len, terminal.utf8Rem = b, 1, 1
		return
	}

	terminal.applyActions(terminal.parser.Feed(b))
}

func (terminal *Terminal) applyActions(actions []Action) {
	for _, a := range actions {
		switch a.Type {

		case ActionPrint:
			terminal.buf.PutRune(a.Rune)

		case ActionExecute:
			terminal.handleControl(a.Rune)

		case ActionCSI:
			terminal.handleCSI(a.Cmd, a.Params, a.Private)

		case ActionESC:
			terminal.handleESC(a.Cmd)

		case ActionOSC:
			terminal.handleOSC(a.OSCRaw)

		}
	}
}

func (terminal *Terminal) handleOSC(raw string) {
	if raw == "" {
		return
	}
	parts := strings.SplitN(raw, ";", 2)
	if len(parts) != 2 {
		return
	}
	if parts[0] == "0" || parts[0] == "2" {
		terminal.title = parts[1]
	}
}
