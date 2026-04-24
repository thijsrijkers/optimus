package ansi

import "optimus/buffer"

type ModeState struct {
	MouseReport bool
	MouseDrag   bool
	MouseMotion bool
	MouseSGR    bool
}

func ApplyPrivateMode(state *ModeState, buf *buffer.Buffer, params []int, set bool) {
	for _, p := range params {
		switch p {
		case 1049, 47, 1047:
			if set {
				buf.SwitchToAltScreen()
			} else {
				buf.SwitchToPrimaryScreen()
			}
		case 1000:
			state.MouseReport = set
		case 1002:
			state.MouseDrag = set
		case 1003:
			state.MouseMotion = set
		case 1006:
			state.MouseSGR = set
		}
	}
}
