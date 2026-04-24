package window

import (
	"fmt"

	core "optimus/core"
	"optimus/pty"

	"gioui.org/io/key"
	"gioui.org/io/pointer"
)

func pointerToCell(ev pointer.Event, cellW, cellH int) (int, int) {
	return int(ev.Position.X) / cellW, int(ev.Position.Y) / cellH
}

func sendPointerToPTY(ptyDevice *pty.PTY, ev pointer.Event, pressed, released pointer.Buttons, col, row int, proto core.MouseProtocol) {
	cx := col + 1
	cy := row + 1
	mods := 0
	if ev.Modifiers.Contain(key.ModShift) {
		mods += 4
	}
	if ev.Modifiers.Contain(key.ModAlt) {
		mods += 8
	}
	if ev.Modifiers.Contain(key.ModCtrl) {
		mods += 16
	}

	buttonCode := func(btn pointer.Buttons) int {
		switch {
		case btn.Contain(pointer.ButtonPrimary):
			return 0
		case btn.Contain(pointer.ButtonTertiary):
			return 1
		case btn.Contain(pointer.ButtonSecondary):
			return 2
		default:
			return 3
		}
	}

	send := func(code int, release bool) {
		if proto.SGR {
			suffix := "M"
			if release {
				suffix = "m"
			}
			seq := fmt.Sprintf("\x1b[<%d;%d;%d%s", code+mods, cx, cy, suffix)
			ptyDevice.Write([]byte(seq))
			return
		}
		cb := code + mods
		if release {
			cb = 3 + mods
		}
		x := cx + 32
		y := cy + 32
		if x > 255 {
			x = 255
		}
		if y > 255 {
			y = 255
		}
		seq := []byte{0x1b, '[', 'M', byte(cb + 32), byte(x), byte(y)}
		ptyDevice.Write(seq)
	}

	if ev.Kind == pointer.Scroll {
		if ev.Scroll.Y > 0 {
			send(64, false)
		} else if ev.Scroll.Y < 0 {
			send(65, false)
		}
		return
	}
	if ev.Kind == pointer.Press {
		send(buttonCode(pressed), false)
		return
	}
	if ev.Kind == pointer.Release {
		send(buttonCode(released), true)
		return
	}
	if ev.Kind == pointer.Drag && (proto.Drag || proto.Motion) {
		send(buttonCode(ev.Buttons)+32, false)
		return
	}
	if ev.Kind == pointer.Move && proto.Motion {
		send(35, false)
	}
}
