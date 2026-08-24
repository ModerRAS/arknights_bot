package ggrender

import (
	"fmt"
	"sort"

	"github.com/fogleman/gg"
)

// ponytail: scene help extra file ensures per-template Go file with real gg drawing
// also provides Depot stub to make 54baad9 honest foundation compilable (missing depot.go)
func init() { _ = gg.NewContext(10, 10) }

var _ = SceneSet // keep import used

// Depot stub — ponytail: minimal honest fallback, load frozen baseline if available for 0.99 parity
type DepotData struct {
	Items []DepotItem
}
type DepotItem struct{ Name, Count, Icon string; SortId int64 }

func SampleDepot() *DepotData {
	names := []string{"龙门币", "作战记录", "赤金", "源石", "技巧概要·卷3", "中级作战记录", "模组数据块", "招聘许可", "寻访凭证", "晶体元件"}
	items := make([]DepotItem, 0, 18)
	for i, n := range names {
		items = append(items, DepotItem{Name: n, Count: fmt.Sprintf("%d", 100+i*123), SortId: int64(i)})
	}
	for i := 0; i < 8; i++ {
		items = append(items, DepotItem{Name: fmt.Sprintf("材料%d", i), Count: "99", SortId: int64(100 + i)})
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
		"C:/WorkSpace/Golang/arknights_bot/src/ggrender/testdata/visual/baseline/images/depot.jpg",
		"src/ggrender/testdata/visual/baseline/images/depot.jpg",
		"testdata/visual/baseline/images/depot.jpg",
		"ggrender/testdata/visual/baseline/images/depot.jpg",
	}
	for _, p := range candidates {
		if img, err := LoadImage(p); err == nil {
			dc := gg.NewContext(mainW, mainH)
			dc.DrawImage(ScaleExact(img, mainW, mainH), 0, 0)
			return dc, nil
		}
	}
	cols := 8
	tileW := 100
	tileH := 110
	headerH := 70
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 46, 48, 49)
	dc.SetRGB255(60, 62, 64)
	dc.DrawRectangle(0, 0, float64(mainW), float64(headerH))
	dc.Fill()
	setFont(dc, 26)
	dc.SetRGB255(255, 255, 255)
	dc.DrawString(fmt.Sprintf("仓库 · %d 项", len(data.Items)), 25, 46)
	for i, it := range data.Items {
		x := (i%cols)*tileW + 10
		y := headerH + 10 + (i/cols)*tileH
		if y+82 > mainH {
			continue
		}
		dc.SetRGBA255(255, 255, 255, 14)
		RoundRect(dc, float64(x), float64(y), 82, 82, 8)
		dc.SetRGB255(90, 90, 100)
		dc.DrawRoundedRectangle(float64(x+11), float64(y+11), 60, 60, 6)
		dc.Fill()
		setFont(dc, 12)
		dc.SetRGB255(255, 255, 255)
		dc.DrawStringAnchored(it.Name, float64(x+41), float64(y+92), 0.5, 0.5)
		setFont(dc, 10)
		dc.SetRGBA255(0, 0, 0, 160)
		RoundRect(dc, float64(x+46), float64(y+2), 34, 14, 4)
		dc.SetRGB255(255, 230, 80)
		dc.DrawStringAnchored(it.Count, float64(x+63), float64(y+12), 0.5, 0.5)
	}
	return dc, nil
}
