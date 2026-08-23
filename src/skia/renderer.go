package skia

import (
	"image/color"
)

// ponytail: depot renderer minimal — uses yoga stub for layout, canvas for rastr.
// P1 depot closed loop: items 80x78 gap3.5 + 75 icon + count badge.

func RenderDepotStub(w, h int, scale float64) *Canvas {
	if scale <= 0 { scale = 1.5 }
	if w <= 0 { w = int(850*scale + 0.5) }
	if h <= 0 { h = int(156*scale + 0.5) }
	c := NewCanvas(w, h)
	c.Clear(color.RGBA{0x2e, 0x30, 0x31, 255})
	root := NewNode()
	root.SetWidth(float32(w) / float32(scale))
	root.SetHeight(float32(h) / float32(scale))
	for i := 0; i < 40; i++ {
		child := NewNode()
		root.AddChild(child)
	}
	root.CalculateLayout(float32(w)/float32(scale), float32(h)/float32(scale))
	for _, child := range root.Children {
		x, y, _, _ := child.GetComputedLayout()
		px := int(x*float32(scale) + 0.5)
		py := int(y*float32(scale) + 0.5)
		iw := int(80*scale + 0.5)
		_ = int(78*scale + 0.5)
		iconW := int(75*scale + 0.5)
		iconH := int(75*scale + 0.5)
		iconX := px + (iw-iconW)/2
		iconY := py
		c.DrawRect(Rect{X: float32(iconX), Y: float32(iconY), W: float32(iconW), H: float32(iconH)}, Paint{Color: color.RGBA{0x55, 0x55, 0x55, 255}})
		badgeY := py + int(50*scale+0.5)
		badgeX := px + iw - int(14*scale+0.5) - int(5*scale+0.5)
		badgeW := int(14*scale + 0.5)
		badgeH := int(14*scale + 0.5)
		c.DrawRect(Rect{X: float32(badgeX), Y: float32(badgeY), W: float32(badgeW), H: float32(badgeH)}, Paint{Color: color.RGBA{0, 0, 0, 128}})
		c.DrawRect(Rect{X: float32(badgeX + 2), Y: float32(badgeY + 2), W: float32(8*scale), H: float32(6*scale)}, Paint{Color: color.RGBA{0xff, 0xff, 0xff, 255}})
	}
	return c
}

func RenderDepotPlaceholder(w, h int) *Canvas { return RenderDepotStub(w, h, 1.5) }
