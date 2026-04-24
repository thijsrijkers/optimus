package window

import (
	"image"
	"image/color"

	"optimus/tabs"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

const tabBarHeight = 28

type tabHit struct {
	startX int
	endX   int
	index  int
	addNew bool
	close  bool
}

func drawTabBar(context layout.Context, theme *material.Theme, entries []tabs.TabMeta) (int, []tabHit) {
	h := context.Dp(unit.Dp(tabBarHeight))
	if h < 20 {
		h = 20
	}
	padX := context.Dp(unit.Dp(10))
	if padX < 6 {
		padX = 6
	}
	innerGap := context.Dp(unit.Dp(10))
	if innerGap < 6 {
		innerGap = 6
	}
	pillGap := context.Dp(unit.Dp(6))
	if pillGap < 4 {
		pillGap = 4
	}
	verticalPad := context.Dp(unit.Dp(4))
	if verticalPad < 3 {
		verticalPad = 3
	}
	fontPx := context.Sp(12)
	if fontPx < 11 {
		fontPx = 11
	}
	charW := (fontPx * 3) / 5
	if charW < 6 {
		charW = 6
	}
	crossW := context.Dp(unit.Dp(18))
	if crossW < 14 {
		crossW = 14
	}
	addW := context.Dp(unit.Dp(36))
	if addW < 30 {
		addW = 30
	}
	bg := color.NRGBA{R: 0x22, G: 0x26, B: 0x2E, A: 0xFF}
	paint.FillShape(context.Ops, bg, clip.Rect(image.Rect(0, 0, context.Constraints.Max.X, h)).Op())

	addLabel := " + "
	containerX := padX
	containerW := context.Constraints.Max.X - 2*padX
	if containerW < addW {
		containerW = addW
	}
	addStart := containerX + containerW - addW
	tabsW := containerW - addW - innerGap
	if tabsW < 0 {
		tabsW = 0
	}

	hits := make([]tabHit, 0, len(entries))
	tabCount := len(entries)
	if tabCount > 0 && tabsW > 0 {
		slotW := tabsW / tabCount
		pillW := slotW - pillGap
		if pillW < 70 {
			pillW = 70
		}
		x := containerX
		for i, entry := range entries {
			if x >= addStart {
				break
			}
			w := pillW
			pillX := x + (slotW-w)/2
			if pillX+w > addStart {
				w = addStart - pillX
			}
			if w < 40 {
				break
			}

			pillRect := image.Rect(pillX, verticalPad, pillX+w, h-verticalPad)
			fill := color.NRGBA{R: 0x2A, G: 0x2F, B: 0x37, A: 0xFF}
			if entry.Active {
				fill = color.NRGBA{R: 0x3B, G: 0x42, B: 0x4F, A: 0xFF}
			}
			paint.FillShape(context.Ops, fill, clip.UniformRRect(pillRect, 7).Op(context.Ops))
			outline := color.NRGBA{R: 0x3C, G: 0x43, B: 0x4E, A: 0xFF}
			if entry.Active {
				outline = color.NRGBA{R: 0x56, G: 0x5F, B: 0x6D, A: 0xFF}
			}
			outlineRect := image.Rect(pillRect.Min.X+1, pillRect.Min.Y+1, pillRect.Max.X-1, pillRect.Max.Y-1)
			paint.FillShape(context.Ops, outline, clip.Stroke{Path: clip.UniformRRect(outlineRect, 7).Path(context.Ops), Width: 1}.Op())

			title := entry.Title
			if len(title) > 18 {
				title = title[:17] + "…"
			}
			lb := material.Label(theme, unit.Sp(12), title)
			lb.Color = color.NRGBA{R: 0xD6, G: 0xDB, B: 0xE3, A: 0xFF}
			if !entry.Active {
				lb.Color = color.NRGBA{R: 0xB2, G: 0xB9, B: 0xC4, A: 0xFF}
			}
			textY := (h-fontPx)/2 - 1
			if textY < 4 {
				textY = 4
			}
			crossX := pillX + w - crossW - padX + 2
			textW := len(title) * charW
			avail := crossX - (pillX + padX) - 4
			if avail < textW {
				textW = avail
			}
			textX := pillX + (crossX-(pillX+padX)-textW)/2 + padX
			if textX < pillX+padX {
				textX = pillX + padX
			}
			stack := opOffset(context, textX, textY)
			lb.Layout(context)
			stack()

			crossLabel := material.Label(theme, unit.Sp(13), "×")
			crossLabel.Color = color.NRGBA{R: 0x9C, G: 0xA5, B: 0xB3, A: 0xFF}
			if entry.Active {
				crossLabel.Color = color.NRGBA{R: 0xC5, G: 0xCC, B: 0xD8, A: 0xFF}
			}
			crossStack := opOffset(context, crossX+charW/2, textY-3)
			crossLabel.Layout(context)
			crossStack()

			hits = append(hits, tabHit{startX: pillX, endX: crossX - 2, index: i})
			hits = append(hits, tabHit{startX: crossX, endX: pillX + w - 4, index: i, close: true})

			x += slotW
		}
	}

	hits = append(hits, tabHit{startX: addStart, endX: addStart + addW, index: -1, addNew: true})
	addRect := image.Rect(addStart, verticalPad, addStart+addW, h-verticalPad)
	paint.FillShape(context.Ops, color.NRGBA{R: 0x2A, G: 0x2F, B: 0x37, A: 0xFF}, clip.UniformRRect(addRect, 7).Op(context.Ops))
	addOutlineRect := image.Rect(addRect.Min.X+1, addRect.Min.Y+1, addRect.Max.X-1, addRect.Max.Y-1)
	paint.FillShape(context.Ops, color.NRGBA{R: 0x3C, G: 0x43, B: 0x4E, A: 0xFF}, clip.Stroke{Path: clip.UniformRRect(addOutlineRect, 7).Path(context.Ops), Width: 1}.Op())
	add := material.Label(theme, unit.Sp(13), addLabel)
	add.Color = color.NRGBA{R: 0xB8, G: 0xC1, B: 0xCD, A: 0xFF}
	addTextY := (h-context.Sp(13))/2 - 1
	if addTextY < 4 {
		addTextY = 4
	}
	addStack := opOffset(context, addStart+padX, addTextY-2)
	add.Layout(context)
	addStack()

	return h, hits
}

func opOffset(context layout.Context, x, y int) func() {
	st := op.Offset(image.Point{X: x, Y: y}).Push(context.Ops)
	return func() { st.Pop() }
}
