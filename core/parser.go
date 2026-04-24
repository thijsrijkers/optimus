package core

import "fmt"

type State int

const (
	StateGround      State = iota // normal text
	StateEscape                   // received ESC
	StateEscapeInter              // ESC + intermediate byte
	StateCSIEntry                 // ESC [
	StateCSIParam                 // ESC [ <params>
	StateCSIInter                 // ESC [ <params> <intermediate>
	StateCSIIgnore                // malformed CSI — ignore until final
	StateOSCString                // ESC ] ... ST  (operating system command)
	StateDCSEntry                 // ESC P  (device control string)
)

// Action is what the parser tells the terminal to do.
type Action struct {
	Type    ActionType
	Params  []int  // numeric params (e.g. [1;32] → [1, 32])
	Rune    rune   // for Print actions
	Cmd     byte   // final byte of CSI/ESC sequence
	Private bool   // CSI private mode (e.g. ESC[?1049h)
	OSCRaw  string // raw OSC payload (title changes, etc.)
}

type ActionType int

const (
	ActionPrint   ActionType = iota // print a rune to screen
	ActionExecute                   // C0 control (BEL, BS, HT, LF, CR, etc.)
	ActionCSI                       // CSI sequence complete  (ESC [ ... cmd)
	ActionESC                       // ESC sequence (ESC cmd)
	ActionOSC                       // OSC string complete
)

func (a ActionType) String() string {
	return [...]string{"Print", "Execute", "CSI", "ESC", "OSC"}[a]
}

type Parser struct {
	state        State
	params       []int
	currentParam int
	hasParam     bool
	intermediate []byte
	oscBuf       []byte
	privateCSI   bool
}

func NewParser() *Parser {
	return &Parser{state: StateGround}
}

func (p *Parser) Feed(b byte) []Action {
	// C0 controls are processed in most states immediately.
	if b < 0x20 && b != 0x1B {
		return []Action{{Type: ActionExecute, Rune: rune(b)}}
	}

	switch p.state {
	case StateGround:
		if b == 0x1B {
			p.state = StateEscape
			return nil
		}
		// Printable ASCII or UTF-8 lead byte.
		return []Action{{Type: ActionPrint, Rune: rune(b)}}

	case StateEscape:
		switch {
		case b == '[':
			p.enterCSI()
			return nil
		case b == ']':
			p.oscBuf = p.oscBuf[:0]
			p.state = StateOSCString
			return nil
		case b >= 0x20 && b <= 0x2F:
			p.intermediate = append(p.intermediate, b)
			p.state = StateEscapeInter
			return nil
		case b >= 0x30 && b <= 0x7E:
			p.state = StateGround
			return []Action{{Type: ActionESC, Cmd: b}}
		default:
			p.state = StateGround
		}

	case StateEscapeInter:
		if b >= 0x30 && b <= 0x7E {
			p.state = StateGround
			return []Action{{Type: ActionESC, Cmd: b}}
		}

	case StateCSIEntry, StateCSIParam:
		switch {
		case b == '?' && p.state == StateCSIEntry:
			p.privateCSI = true
			p.state = StateCSIParam
		case b >= '0' && b <= '9':
			p.currentParam = p.currentParam*10 + int(b-'0')
			p.hasParam = true
			p.state = StateCSIParam
		case b == ';':
			p.params = append(p.params, p.currentParam)
			p.currentParam = 0
			p.hasParam = false
			p.state = StateCSIParam
		case b >= 0x20 && b <= 0x2F:
			p.intermediate = append(p.intermediate, b)
			p.state = StateCSIInter
		case b >= 0x40 && b <= 0x7E:
			// Final byte, flush current param and dispatch.
			if p.hasParam || len(p.params) > 0 {
				p.params = append(p.params, p.currentParam)
			}
			params := make([]int, len(p.params))
			copy(params, p.params)
			p.state = StateGround
			private := p.privateCSI
			return []Action{{Type: ActionCSI, Cmd: b, Params: params, Private: private}}
		case b == 0x1B:
			p.state = StateEscape
		}

	case StateCSIInter:
		if b >= 0x40 && b <= 0x7E {
			p.state = StateGround
		}

	case StateCSIIgnore:
		if b >= 0x40 && b <= 0x7E {
			p.state = StateGround
		}

	case StateOSCString:
		if b == 0x07 || b == 0x9C {
			raw := string(p.oscBuf)
			p.state = StateGround
			return []Action{{Type: ActionOSC, OSCRaw: raw}}
		}
		if b == 0x1B {
			// just end OSC on ESC too.
			raw := string(p.oscBuf)
			p.state = StateEscape
			return []Action{{Type: ActionOSC, OSCRaw: raw}}
		}
		p.oscBuf = append(p.oscBuf, b)
	}

	return nil
}

func (p *Parser) enterCSI() {
	p.state = StateCSIEntry
	p.params = p.params[:0]
	p.currentParam = 0
	p.hasParam = false
	p.intermediate = p.intermediate[:0]
	p.privateCSI = false
}

func (p *Parser) FeedRune(r rune) []Action {
	if p.state == StateGround && r >= 0x20 {
		return []Action{{Type: ActionPrint, Rune: r}}
	}
	// Fall back to byte-by-byte for non-ground states.
	return []Action{{Type: ActionPrint, Rune: r}}
}

func (a Action) String() string {
	switch a.Type {
	case ActionPrint:
		return fmt.Sprintf("Print(%q)", a.Rune)
	case ActionExecute:
		return fmt.Sprintf("Execute(0x%02X)", a.Rune)
	case ActionCSI:
		if a.Private {
			return fmt.Sprintf("CSI(?%q params=%v)", a.Cmd, a.Params)
		}
		return fmt.Sprintf("CSI(%q params=%v)", a.Cmd, a.Params)
	case ActionESC:
		return fmt.Sprintf("ESC(%q)", a.Cmd)
	case ActionOSC:
		return fmt.Sprintf("OSC(%q)", a.OSCRaw)
	}
	return "Unknown"
}
