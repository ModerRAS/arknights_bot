package ggrender

import (
	"fmt"
	"image"
	"math"
	"path/filepath"

	"github.com/fogleman/gg"
	xdraw "golang.org/x/image/draw"
)

// Amiya 素材（PRTS 远程，走 FetchImage 加载）。
const (
	AmiyaAvatarURL   = "https://media.prts.wiki/3/36/%E5%A4%B4%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85.png"            // 180x180 头像
	AmiyaPortraitURL = "https://media.prts.wiki/a/a0/%E5%8D%8A%E8%BA%AB%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85_1.png" // 180x360 半身像
)

// Char mirrors Box.
type Char struct {
	CharId, SkinId, Name              string
	Level, EvolvePhase, PotentialRank int
	FavorPercent, Rarity              int
	Profession                        string
}

type BoxInfo struct {
	Name  string
	Chars []Char
}

func SampleBox() *BoxInfo {
	chars := make([]Char, 0, 11)
	for i := 0; i < 11; i++ {
		chars = append(chars, Char{
			SkinId:        AmiyaPortraitURL,
			Name:          "阿米娅",
			Rarity:        5,
			Profession:    "PIONEER",
			Level:         90,
			EvolvePhase:   2,
			PotentialRank: 5,
		})
	}
	return &BoxInfo{Name: "Dr 测试博士", Chars: chars}
}

// drawBoxFamilyHeader box 族共用头部：青色竖条 + 标题 + //ARKNIGHTS 水印 + 分隔线。
func drawBoxFamilyHeader(dc *gg.Context, w, h int, title string, titleY, sepY float64) {
	dc.SetRGB255(48, 48, 48)
	dc.DrawRectangle(0, 0, float64(w), float64(h))
	dc.Fill()
	// 右侧渐暗
	x0 := w * 2 / 3
	for x := x0; x < w; x++ {
		f := float64(x-x0) / float64(w-x0)
		v := int(48 - 23*f)
		dc.SetRGB255(v, v, v)
		dc.DrawRectangle(float64(x), 0, 1, float64(h))
		dc.Fill()
	}
	// 半调网点
	dc.SetRGB255(72, 72, 72)
	for gy := 8; gy < h-6; gy += 14 {
		for gx := w*3/4 + (gy/14%2)*7; gx < w-10; gx += 14 {
			dc.DrawCircle(float64(gx), float64(gy), 2.6)
			dc.Fill()
		}
	}
	// 标题
	setFont(dc, 44)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, title, 37, titleY)
	// 水印
	setFont(dc, 36)
	dc.SetRGB255(89, 89, 89)
	drawStringAnchored(dc, "//ARKNIGHTS", float64(w)-36, float64(h)*0.689, 1, 1)
	// 分隔线
	dc.SetRGB255(188, 188, 188)
	dc.DrawRectangle(0, sepY, float64(w), 5)
	dc.Fill()
}

// resampleKernel Mitchell-Netravali (B=C=1/3)，实测最接近基线缩放。
var resampleKernel = xdraw.Kernel{Support: 2, At: func(t float64) float64 {
	const B, C = 1.0 / 3.0, 1.0 / 3.0
	x := t
	if x < 0 {
		x = -x
	}
	if x < 1 {
		return ((12-9*B-6*C)*x*x*x + (-18+12*B+6*C)*x*x + (6-2*B)) / 6
	}
	if x < 2 {
		return ((-B-6*C)*x*x*x + (6*B+30*C)*x*x + (-12*B-48*C)*x + (8*B+24*C)) / 6
	}
	return 0
}}

// smoothCover 高质量核缩放后居中裁剪（对齐浏览器 object-fit: cover 观感）。
func smoothCover(img image.Image, w, h int) *image.RGBA {
	srcW, srcH := img.Bounds().Dx(), img.Bounds().Dy()
	scale := math.Max(float64(w)/float64(srcW), float64(h)/float64(srcH))
	nw := int(math.Round(float64(srcW)*scale)) + 1
	nh := int(math.Round(float64(srcH)*scale)) + 1
	if nw < w {
		nw = w
	}
	if nh < h {
		nh = h
	}
	tmp := image.NewRGBA(image.Rect(0, 0, nw, nh))
	resampleKernel.Scale(tmp, tmp.Bounds(), img, img.Bounds(), xdraw.Over, nil)
	l, t := (nw-w)/2, (nh-h)/2
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.Draw(dst, dst.Bounds(), tmp, image.Pt(l, t), xdraw.Src)
	return dst
}

// drawBoxTile box 族干员瓦片：105x210，无缝网格，行距 210。
func drawBoxTile(dc *gg.Context, x, y int, portraitURL, profession string, rarity, level, evolve int, name string, showLevel bool) {
	const tw, th = 105, 210
	port := FetchImage(portraitURL, amiyaPath)
	dc.DrawImage(smoothCover(port, tw, th), x, y)
	// 职业图标 左上
	if ic, err := LoadImage(filepath.Join(AssetRoot, "box", profession+".png")); err == nil {
		dc.DrawImage(ScaleExact(ic, 28, 28), x+2, y+2)
	}
	// 星条 右上
	if rar, err := LoadImage(filepath.Join(AssetRoot, "box", fmt.Sprintf("Rarity_%d.png", rarity))); err == nil {
		dc.DrawImage(ScaleExact(rar, 62, 17), x+30, y+6)
	}
	// 精英徽章
	if evolve > 0 {
		if ev, err := LoadImage(filepath.Join(AssetRoot, "box", fmt.Sprintf("Evolve_%d.png", evolve))); err == nil {
			dc.DrawImage(ScaleExact(ev, 75, 62), x+30, y+113)
		}
	}
	// 等级
	if showLevel && level > 0 {
		setFont(dc, 44)
		dc.SetRGB255(255, 255, 255)
		drawString(dc, itoa(level), float64(x+8), float64(y+186))
	}
	// 名条（覆盖 portrait 底部的半透明条）
	dc.SetRGBA255(0, 0, 0, 140)
	dc.DrawRectangle(float64(x), float64(y+185), tw, th-185)
	dc.Fill()
	setFont(dc, 16)
	dc.SetRGB255(255, 255, 255)
	drawStringAnchored(dc, name, float64(x+tw/2), float64(y+196), 0.5, 0.5)
}

func RenderBox(data *BoxInfo) (*gg.Context, error) {
	const mainW, mainH = 1050, 536
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 46, 48, 49)
	drawBoxFamilyHeader(dc, mainW, 106, data.Name, 81, 107)
	for i, c := range data.Chars {
		x := (i % 10) * 105
		y := 114 + (i/10)*210
		drawBoxTile(dc, x, y, c.SkinId, c.Profession, c.Rarity, c.Level, c.EvolvePhase, c.Name, true)
	}
	return dc, nil
}
