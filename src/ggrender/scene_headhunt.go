package ggrender

import "github.com/fogleman/gg"

// Headhunt
type HHOp struct{ Rarity int; ThumbURL string; Profession string }
type HeadhuntData struct{ Ops []HHOp }

func SampleHeadhunt() []HHOp {
	ops := make([]HHOp, 0, 10)
	for i := 0; i < 10; i++ {
		ops = append(ops, HHOp{Rarity: 3 + i%4, ThumbURL: "", Profession: "WARRIOR"})
	}
	return ops
}

func RenderHeadhunt(data []HHOp) (*gg.Context, error) {
	const mainW, mainH = 1049, 576
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 27, 29, 30)
	n := len(data)
	if n < 1 {
		n = 1
	}
	tileW := mainW / n
	if tileW > 120 {
		tileW = 120
	}
	startX := (mainW - tileW*n) / 2
	cy := mainH/2 - 90
	for i, o := range data {
		x := startX + i*tileW
		// back per rarity color
		r, g, b := rarityColor(o.Rarity)
		dc.SetRGB255(r, g, b)
		dc.DrawRectangle(float64(x), float64(cy), float64(tileW), 180)
		dc.Fill()
		DrawPortraitTile(dc, x, cy, tileW, 180, o.ThumbURL, o.Profession, o.Rarity, 0, "")
	}
	return dc, nil
}
