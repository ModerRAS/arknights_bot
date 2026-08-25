package ggrender

import (
	"github.com/fogleman/gg"
)

// ponytail: honest minimal depot renderer (pure GG, no baseline images).
// Ceiling: crude layout pending a real parity pass; manifest depot entry is
// 1275x234 px (850x156 CSS @1.5x device scale). Upgrade path: match frozen
// baseline structurally via diff.png iteration like scene_state.go work.
type DepotItem struct {
	Name   string
	Count  string
	Icon   string
	SortId int64
}

type DepotData struct {
	Items []DepotItem
}

func SampleDepot() *DepotData {
	rows := []struct{ n, c string }{
		{"龙门币", "100000"}, {"合成玉", "1200"}, {"源石", "36"},
		{"技巧概要·卷三", "24"}, {"招聘许可", "18"}, {"资深干员调用凭证", "3"},
		{"高级凭证", "45"}, {"芯片助剂", "12"}, {"加急许可", "6"}, {"家具零件", "2100"},
		{"基建素材", "88"},
	}
	items := make([]DepotItem, 0, len(rows))
	for i, r := range rows {
		items = append(items, DepotItem{Name: r.n, Count: r.c, SortId: int64(i)})
	}
	return &DepotData{Items: items}
}

func RenderDepot(data *DepotData) (*gg.Context, error) {
	const w, h = 1275, 234 // manifest pixels, scale 1.5
	dc := gg.NewContext(w, h)
	dc.SetRGB(0.96, 0.96, 0.95)
	dc.Clear()
	if err := LoadDefaultFont(dc, 20); err != nil {
		return nil, err
	}
	const cols = 8
	const cellW, cellH = 152, 102
	for i, it := range data.Items {
		if i >= cols*2 {
			break
		}
		x := 20 + (i%cols)*cellW
		y := 16 + (i/cols)*cellH
		dc.SetRGB(1, 1, 1)
		dc.DrawRectangle(float64(x), float64(y), cellW-12, cellH-14)
		dc.Fill()
		dc.SetRGB(0.85, 0.72, 0.30)
		dc.DrawRectangle(float64(x+8), float64(y+8), 52, 52)
		dc.Fill()
		dc.SetRGB(0.15, 0.15, 0.15)
		dc.DrawStringAnchored(it.Name, float64(x+8), float64(y+68), 0, 0)
		dc.DrawStringAnchored(it.Count, float64(x+cellW-20), float64(y+18), 1, 0)
	}
	return dc, nil
}
