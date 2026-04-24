package ansi

import (
	"image/color"

	"optimus/buffer"
)

func ApplySGR(attr buffer.Attr, params []int) buffer.Attr {
	if len(params) == 0 {
		params = []int{0}
	}
	for i := 0; i < len(params); i++ {
		switch params[i] {
		case 0:
			attr = buffer.Attr{FG: buffer.DefaultFG, BG: buffer.DefaultBG}
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
		case 30, 31, 32, 33, 34, 35, 36, 37:
			attr.FG = ansi16[params[i]-30]
		case 39:
			attr.FG = buffer.DefaultFG
		case 90, 91, 92, 93, 94, 95, 96, 97:
			attr.FG = ansi16[8+params[i]-90]
		case 40, 41, 42, 43, 44, 45, 46, 47:
			attr.BG = ansi16[params[i]-40]
		case 49:
			attr.BG = buffer.DefaultBG
		case 100, 101, 102, 103, 104, 105, 106, 107:
			attr.BG = ansi16[8+params[i]-100]
		case 38:
			if i+2 < len(params) && params[i+1] == 5 {
				attr.FG = ansi256[params[i+2]]
				i += 2
			} else if i+5 < len(params) && params[i+1] == 2 && params[i+2] == 0 {
				attr.FG = color.RGBA{R: uint8(params[i+3]), G: uint8(params[i+4]), B: uint8(params[i+5]), A: 0xFF}
				i += 5
			} else if i+4 < len(params) && params[i+1] == 2 {
				attr.FG = color.RGBA{R: uint8(params[i+2]), G: uint8(params[i+3]), B: uint8(params[i+4]), A: 0xFF}
				i += 4
			}
		case 48:
			if i+2 < len(params) && params[i+1] == 5 {
				attr.BG = ansi256[params[i+2]]
				i += 2
			} else if i+5 < len(params) && params[i+1] == 2 && params[i+2] == 0 {
				attr.BG = color.RGBA{R: uint8(params[i+3]), G: uint8(params[i+4]), B: uint8(params[i+5]), A: 0xFF}
				i += 5
			} else if i+4 < len(params) && params[i+1] == 2 {
				attr.BG = color.RGBA{R: uint8(params[i+2]), G: uint8(params[i+3]), B: uint8(params[i+4]), A: 0xFF}
				i += 4
			}
		}
	}
	return attr
}
