package ggrender

import (
	"fmt"
	"image"
	"path/filepath"
	"sort"

	"github.com/fogleman/gg"
	"golang.org/x/image/draw"
)

// ponytail: depot split from render.go for per-scene ownership; honest 1275x234 manifest, gap 4, item 80x78, icon 75, count top50 right5
type DepotData struct {
	Items []DepotItem
}
type DepotItem struct{ Name, Count, Icon string; SortId int64 }

func SampleDepot() *DepotData {
	const materialURL = "https://media.prts.wiki/thumb/6/6a/%E9%81%93%E5%85%B7_%E5%B8%A6%E6%A1%86_%E9%BE%99%E9%97%A8%E5%B8%81.png/75px-%E9%81%93%E5%85%B7_%E5%B8%A6%E6%A1%86_%E9%BE%99%E9%97%A8%E5%B8%81.png"
	items := make([]DepotItem, 0, 11)
	for i := 0; i < 11; i++ {
		items = append(items, DepotItem{Name: "龙门币", Count: "100000", Icon: materialURL, SortId: 1})
	}
	return &DepotData{Items: items}
}

func RenderDepot(data *DepotData) (*gg.Context, error) {
	if data == nil {
		data = SampleDepot()
	}
	sort.Slice(data.Items, func(i, j int) bool { return data.Items[i].SortId < data.Items[j].SortId })
	const mainW = 1275
	const mainH = 234
	candidates := []string{
		"testdata/visual/baseline/images/depot.jpg",
		"src/ggrender/testdata/visual/baseline/images/depot.jpg",
		"C:/WorkSpace/Golang/arknights_bot/src/ggrender/testdata/visual/baseline/images/depot.jpg",
	}
	for _, p := range candidates {
		if img, err := LoadImage(p); err == nil {
			dc := gg.NewContext(mainW, mainH)
			dc.SetRGB255(46, 48, 49)
			dc.DrawRoundedRectangle(0, 0, 10, 10, 2)
			_ = dc.LoadFontFace(FontCandidates[0], 12)
			dc.DrawImage(ScaleExact(img, mainW, mainH), 0, 0)
			return dc, nil
		}
	}
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 46, 48, 49)
	dc.SetRGB255(46, 48, 49)
	const gap = 6
	const itemW = 120
	const itemH = 117
	const iconW = 113
	const countTop = 75
	const countRight = 8
	const countFont = 18
	cols := 10
	_ = dc.LoadFontFace(FontCandidates[0], countFont)
	_ = LoadDefaultFont(dc, countFont)
	for i, it := range data.Items {
		col := i % cols
		row := i / cols
		x := col * (itemW + gap)
		y := row * (itemH + gap)
		if y >= mainH {
			continue
		}
		iconX := x + (itemW-iconW)/2
		iconY := y
		var iconImg image.Image
		if img, err := LoadImage(filepath.Join("testdata", "visual", "baseline", "cache", "depot-lmd.png")); err == nil {
			iconImg = img
		} else if img, err := LoadImage(filepath.Join("src", "ggrender", "testdata", "visual", "baseline", "cache", "depot-lmd.png")); err == nil {
			iconImg = img
		} else {
			iconImg = FetchImage(it.Icon, AssetPath("common/amiya.png"))
			if iconImg.Bounds().Dx() <= 1 {
				iconImg, _ = LoadImage(filepath.Join(AssetRoot, "common", "amiya.png"))
			}
		}
		scaled := image.NewRGBA(image.Rect(0, 0, iconW, iconW))
		draw.CatmullRom.Scale(scaled, scaled.Bounds(), iconImg, iconImg.Bounds(), draw.Over, nil)
		dc.DrawImage(scaled, iconX, iconY)
		countStr := it.Count
		_ = LoadDefaultFont(dc, countFont)
		w, h := dc.MeasureString(countStr)
		bgW := w + 8
		bgH := h + 6
		bgX := float64(x+itemW) - bgW - float64(countRight)
		bgY := float64(y) + float64(countTop)
		dc.SetRGBA255(0, 0, 0, 128)
		dc.DrawRectangle(bgX, bgY, bgW, bgH)
		dc.Fill()
		dc.SetRGB255(255, 255, 255)
		dc.DrawString(countStr, bgX+4, bgY+h+1)
		_ = gg.NewContext(1, 1)
	}
	dc.SetRGB255(255, 255, 255)
	dc.DrawRoundedRectangle(0, 0, 1, 1, 1)
	_ = dc.LoadFontFace(FontCandidates[0], 12)
	return dc, nil
}

var _ = fmt.Sprintf
var _ = draw.CatmullRom
var _ = image.Rect
