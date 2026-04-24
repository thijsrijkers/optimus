package window

import (
	"runtime"
	"strings"

	"gioui.org/io/key"
)

func keyToBytes(e key.Event) []byte {
	// Modifiers
	ctrl := e.Modifiers.Contain(key.ModCtrl)
	shift := e.Modifiers.Contain(key.ModShift)
	_ = shift

	// Named keys
	switch e.Name {
	case key.NameReturn, key.NameEnter:
		return []byte{'\r'}
	case key.NameDeleteBackward:
		return []byte{0x7F}
	case key.NameDeleteForward:
		return []byte{0x1B, '[', '3', '~'}
	case key.NameTab:
		if ctrl {
			return nil
		}
		return []byte{'\t'}
	case key.NameEscape:
		return []byte{0x1B}
	case key.NameUpArrow:
		return []byte{0x1B, '[', 'A'}
	case key.NameDownArrow:
		return []byte{0x1B, '[', 'B'}
	case key.NameRightArrow:
		return []byte{0x1B, '[', 'C'}
	case key.NameLeftArrow:
		return []byte{0x1B, '[', 'D'}
	case key.NameHome:
		return []byte{0x1B, '[', 'H'}
	case key.NameEnd:
		return []byte{0x1B, '[', 'F'}
	case key.NamePageUp:
		return []byte{0x1B, '[', '5', '~'}
	case key.NamePageDown:
		return []byte{0x1B, '[', '6', '~'}
	case key.NameF1:
		return []byte{0x1B, 'O', 'P'}
	case key.NameF2:
		return []byte{0x1B, 'O', 'Q'}
	case key.NameF3:
		return []byte{0x1B, 'O', 'R'}
	case key.NameF4:
		return []byte{0x1B, 'O', 'S'}
	case key.NameF5:
		return []byte{0x1B, '[', '1', '5', '~'}
	}

	// Ctrl+letter, control code
	if ctrl && len(e.Name) == 1 {
		c := e.Name[0]
		if c >= 'A' && c <= 'Z' {
			return []byte{c - 'A' + 1}
		}
		if c >= 'a' && c <= 'z' {
			return []byte{c - 'a' + 1}
		}
		switch c {
		case '[':
			return []byte{0x1B}
		case '\\':
			return []byte{0x1C}
		case ']':
			return []byte{0x1D}
		case '^':
			return []byte{0x1E}
		case '_':
			return []byte{0x1F}
		}
	}

	return nil
}

func isCopyShortcut(e key.Event) bool {
	if !e.Modifiers.Contain(key.ModShortcut) {
		return false
	}
	if runtime.GOOS != "darwin" && !e.Modifiers.Contain(key.ModShift) {
		return false
	}
	return strings.EqualFold(string(e.Name), "c")
}

func isPasteShortcut(e key.Event) bool {
	if !e.Modifiers.Contain(key.ModShortcut) {
		return false
	}
	if runtime.GOOS != "darwin" && !e.Modifiers.Contain(key.ModShift) {
		return false
	}
	return strings.EqualFold(string(e.Name), "v")
}

func isNewTabShortcut(e key.Event) bool {
	if !e.Modifiers.Contain(key.ModShortcut) {
		return false
	}
	return strings.EqualFold(string(e.Name), "t")
}

func isCloseTabShortcut(e key.Event) bool {
	if !e.Modifiers.Contain(key.ModShortcut) {
		return false
	}
	return strings.EqualFold(string(e.Name), "w")
}

func isNextTabShortcut(e key.Event) bool {
	if !e.Modifiers.Contain(key.ModShortcut) {
		return false
	}
	if e.Modifiers.Contain(key.ModShift) {
		return false
	}
	return e.Name == key.NameTab
}

func isPrevTabShortcut(e key.Event) bool {
	if !e.Modifiers.Contain(key.ModShortcut) || !e.Modifiers.Contain(key.ModShift) {
		return false
	}
	return e.Name == key.NameTab
}

func isZoomInShortcut(e key.Event) bool {
	if !e.Modifiers.Contain(key.ModShortcut) {
		return false
	}
	n := string(e.Name)
	return n == "+" || n == "="
}

func isZoomOutShortcut(e key.Event) bool {
	if !e.Modifiers.Contain(key.ModShortcut) {
		return false
	}
	return string(e.Name) == "-"
}
