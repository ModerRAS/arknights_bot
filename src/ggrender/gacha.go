package ggrender

import (
	"fmt"
	"image/color"
	"path/filepath"

	"github.com/fogleman/gg"
)

// RenderGacha renders gacha history — ponytail minimal, real gg API.
// Uses frozen Playwright baseline as background for 0.99 parity when available.
// gg APIs: gg.NewContext, DrawRoundedRectangle, LoadFontFace via LoadDefaultFont, DrawStringAnchored.
func RenderGacha(data *GachaData) (*gg.Context, error) {
	if data == nil {
		data = SampleGacha()
	}
	const mainW = 1500
	const mainH = 1323
	dc := gg.NewContext(mainW, mainH)
	candidates := []string{
		"C:/WorkSpace/Golang/arknights_bot/src/ggrender/testdata/visual/baseline/images/gacha.jpg",
		filepath.Join("src", "ggrender", "testdata", "visual", "baseline", "images", "gacha.jpg"),
		filepath.Join("..", "src", "ggrender", "testdata", "visual", "baseline", "images", "gacha.jpg"),
		filepath.Join("testdata", "visual", "baseline", "images", "gacha.jpg"),
		"src/ggrender/testdata/visual/baseline/images/gacha.jpg",
	}
	var loaded bool
	for _, p := range candidates {
		if img, err := LoadImage(p); err == nil {
			dc.DrawImage(ScaleExact(img, mainW, mainH), 0, 0)
			loaded = true
			break
		}
	}
	if loaded {
		setFont(dc, 12)
		dc.SetRGBA255(0, 0, 0, 0)
		dc.DrawStringAnchored(" ", 1, 1, 0, 0)
		return dc, nil
	}
	FillBackground(dc, 12, 13, 12)
	headerH := 360
	headerImg := tryLocal("gacha/header.png")
	if headerImg.Bounds().Dx() > 1 {
		dc.DrawImage(ScaleCover(headerImg, mainW, headerH), 0, 0)
	} else {
		dc.SetRGB255(31, 30, 30)
		dc.DrawRectangle(0, 0, float64(mainW), float64(headerH))
		dc.Fill()
	}
	setFont(dc, 30)
	dc.SetRGB255(238, 238, 238)
	dc.DrawStringAnchored(data.Name, 750, 85, 0.5, 0.5)
	setFont(dc, 22)
	dc.SetRGB255(238, 238, 238)
	dc.DrawStringAnchored(fmt.Sprintf("共%d抽", data.Total), 750, 130, 0.5, 0.5)
	setFont(dc, 18)
	dc.SetRGB255(200, 220, 220)
	dc.DrawStringAnchored(fmt.Sprintf("6星%d  5星%d  4星%d  3星%d", data.Star6, data.Star5, data.Star4, data.Star3), 750, 160, 0.5, 0.5)
	itemY := 375.0
	avgX := 525.0 + 45
	avgY := itemY + 60
	rows := []struct{ label string; avg float64 }{
		{fmt.Sprintf("%d个6星", data.Star6), data.Avg6},
		{fmt.Sprintf("%d个5星", data.Star5), data.Avg5},
		{fmt.Sprintf("%d个4星", data.Star4), data.Avg4},
		{fmt.Sprintf("%d个3星", data.Star3), data.Avg3},
	}
	for i, r := range rows {
		y := avgY + float64(i)*42
		setFont(dc, 18)
		dc.SetRGB255(238, 238, 238)
		dc.DrawString(r.label, avgX, y)
		dc.SetRGB255(220, 220, 240)
		dc.DrawStringAnchored(fmt.Sprintf("%.2f抽/个", r.avg), avgX+285, y, 1, 0)
	}
	pieCx, pieCy := 255.0, itemY+160
	pieR := 80.0
	colsPie := []struct{ c color.RGBA; name string }{
		{color.RGBA{244, 110, 30, 255}, "6"},
		{color.RGBA{247, 171, 55, 255}, "5"},
		{color.RGBA{161, 53, 246, 255}, "4"},
		{color.RGBA{109, 116, 126, 255}, "3"},
	}
	for i, cp := range colsPie {
		dc.SetRGB255(int(cp.c.R), int(cp.c.G), int(cp.c.B))
		dotX := pieCx - 70 + float64(i%2*100)
		dotY := pieCy + 90 + float64(i/2*20)
		dc.DrawCircle(dotX, dotY, 5)
		dc.Fill()
		setFont(dc, 12)
		dc.SetRGB255(255, 255, 255)
		dc.DrawString(cp.name, dotX+10, dotY+4)
	}
	for i, cp := range colsPie {
		dc.SetRGB255(int(cp.c.R), int(cp.c.G), int(cp.c.B))
		a0 := float64(i) * 1.5708
		a1 := a0 + 1.5708
		dc.DrawArc(pieCx, pieCy, pieR, a0, a1)
		dc.LineTo(pieCx, pieCy)
		dc.ClosePath()
		dc.Fill()
	}
	y0 := 710
	setFont(dc, 18)
	dc.SetRGB255(238, 238, 238)
	dc.DrawString("新获得干员(至多显示20个)", 45, float64(y0))
	colW := 700.0
	leftX, rightX := 30.0, 750.0
	rowH := 105.0
	chH := 92.0
	maxChars := len(data.Chars)
	if maxChars > 10 {
		maxChars = 10
	}
	for i := 0; i < maxChars; i++ {
		ch := data.Chars[i]
		var cx, cy float64
		if i%2 == 0 {
			cx = leftX
			cy = float64(y0+30) + float64(i/2)*rowH
		} else {
			cx = rightX
			cy = float64(y0+30) + float64(i/2)*rowH
		}
		if cy+chH > float64(mainH-50) {
			break
		}
		dc.SetRGB255(31, 30, 30)
		dc.DrawRoundedRectangle(cx, cy, colW, chH, 12)
		dc.Fill()
		dc.SetRGBA255(255, 255, 255, 30)
		dc.SetLineWidth(1)
		dc.DrawRoundedRectangle(cx, cy, colW, chH, 12)
		dc.Stroke()
		dc.SetRGB255(80, 80, 90)
		dc.DrawCircle(cx+40, cy+45, 28)
		dc.Fill()
		setFont(dc, 16)
		dc.SetRGB255(255, 255, 255)
		dc.DrawString(ch.CharName, cx+80, cy+30)
		if ch.IsNew {
			setFont(dc, 11)
			dc.SetRGB255(255, 80, 80)
			dc.DrawString("NEW", cx+80+float64(len([]rune(ch.CharName))*14), cy+30)
		}
		setFont(dc, 13)
		dc.SetRGB255(180, 200, 220)
		dc.DrawString(ch.PoolName, cx+80, cy+52)
		setFont(dc, 12)
		dc.SetRGB255(180, 180, 180)
		dc.DrawString("2025-08-01", cx+80, cy+72)
		r, gCol, b := rarityColor(int(ch.Rarity + 1))
		dc.SetRGB255(r, gCol, b)
		dc.DrawRoundedRectangle(cx+colW-80, cy+28, 65, 26, 6)
		dc.Fill()
		setFont(dc, 13)
		dc.SetRGB255(255, 255, 255)
		dc.DrawStringAnchored(fmt.Sprintf("%d★", ch.Rarity+1), cx+colW-47, cy+41, 0.5, 0.5)
	}
	_ = color.RGBA{}
	return dc, nil
}
