//go:build !skia

package skia

import (
	"image/color"
	"path/filepath"
)

// Backend is the minimal Skia backend façade.
// Default build (no cgo) — pure Go; -tags skia delegates to real Skia (skia_backend_skia.go).
type Backend struct {
	font   *Typeface
	loader *Loader
}

func NewBackend(repoRoot string) (*Backend, error) {
	if repoRoot == "" {
		repoRoot = "."
	}
	repoRoot, _ = filepath.Abs(repoRoot)
	tf, err := LoadTypeface(filepath.Join(repoRoot, FontPath))
	if err != nil {
		return nil, err
	}
	ld, _ := NewLoader(repoRoot)
	return &Backend{font: tf, loader: ld}, nil
}

func (b *Backend) Font() *Typeface { return b.font }
func (b *Backend) Loader() *Loader { return b.loader }

// RenderCardStub renders a minimal card-like scene to prove Canvas pipeline.
// Mirrors card.mjs 1280x720 with bg + rrect cards + progress — used by tests as pixel oracle.
func (b *Backend) RenderCardStub(w, h int) *Canvas {
	c := NewCanvas(w, h)
	c.Clear(color.RGBA{12, 13, 12, 255}) // #0c0d0c
	// 3 columns like card: secret bg card
	c.DrawRRect(RRect{Rect: Rect{X: 20, Y: 150, W: 300, H: 250}, Radius: 20}, Paint{Color: color.RGBA{0x1f, 0x1e, 0x1e, 255}})
	c.DrawRRect(RRect{Rect: Rect{X: 340, Y: 150, W: 300, H: 250}, Radius: 20}, Paint{Color: color.RGBA{0x1f, 0x1e, 0x1e, 255}})
	c.DrawRRect(RRect{Rect: Rect{X: 660, Y: 150, W: 300, H: 250}, Radius: 20}, Paint{Color: color.RGBA{0x1f, 0x1e, 0x1e, 255}})
	// progress bar 150x3 with 70% fill (base scenario)
	c.DrawRect(Rect{X: 20, Y: 240, W: 150, H: 3}, Paint{Color: color.RGBA{0x33, 0x33, 0x33, 255}})
	c.DrawRect(Rect{X: 20, Y: 240, W: 105, H: 3}, Paint{Color: color.RGBA{0x54, 0x70, 0xc6, 255}})
	// circle avatar placeholder (r=40)
	c.DrawCircle(Circle{CX: 60, CY: 400, R: 40}, Paint{Color: color.RGBA{0x54, 0x70, 0xc6, 255}})
	// path: gacha pie sector approx
	c.DrawPath([]PathCmd{
		{Op: "M", Args: []float32{150, 150}},
		{Op: "L", Args: []float32{278, 150}},
		{Op: "A", Args: []float32{128, 128, 0, 0, 1, 150, 278}},
		{Op: "Z", Args: nil},
	}, Paint{Color: color.RGBA{0xff, 0x66, 0x66, 255}})
	// shadow (lottery/recruit) — proves DropShadow path
	c.DrawDropShadow(RRect{Rect: Rect{X: 100, Y: 600, W: 200, H: 80}, Radius: 16}, color.RGBA{0, 0, 0, 204}, 0, 10, 25)
	c.DrawRRect(RRect{Rect: Rect{X: 100, Y: 600, W: 200, H: 80}, Radius: 16}, Paint{Color: color.RGBA{0x22, 0x22, 0x22, 255}})
	return c
}
