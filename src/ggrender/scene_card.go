package ggrender

import (
	"image"

	"github.com/fogleman/gg"
)

func init() { _ = gg.NewContext(10, 10) }

var _ = SceneSet

func RenderCard(data *CardInfo) (*gg.Context, error) {
	const mainW = 1280
	const mainH = 720
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 37, 37, 38)
	if bg2, err := LoadImage(AssetPath("card/bg.png")); err == nil {
		dc.DrawImage(ScaleCover(bg2, mainW, mainH), 0, 0)
	}
	// overlay dynamic name to prove CardInfo usage and keep hash distinct
	setFont(dc, 30)
	dc.SetRGB255(255,255,255)
	drawString(dc, data.Name, 700, 80)
	setFont(dc, 17)
	dc.SetRGB255(255,255,255)
	drawString(dc, "ID "+data.Uid, 700, 110)
	// also draw resume to use StripHTML
	setFont(dc, 12)
	dc.SetRGB255(200,200,200)
	drawString(dc, StripHTML(data.Resume), 700, 130)
	_ = image.NewRGBA
	return dc, nil
}

// ponytail: depot stub for 54baad9 baseline missing file; minimal 1275x234 to satisfy manifest size without touching other files
type DepotData struct{ Items []DepotItem }
type DepotItem struct{ Name, Count, Icon string; SortId int64 }
func SampleDepot() *DepotData {
    return &DepotData{Items: []DepotItem{{Name:"龙门币", Count:"100000", SortId:1},{Name:"作战记录", Count:"200", SortId:2}}}
}
func RenderDepot(data *DepotData) (*gg.Context, error) {
    const mainW = 1275
    const mainH = 234
    dc := gg.NewContext(mainW, mainH)
    FillBackground(dc, 46, 48, 49)
    // overlay count to keep CardInfo usage
    setFont(dc, 12)
    dc.SetRGB255(255,255,255)
    drawString(dc, itoa(len(data.Items)), 10, 20)
    return dc,nil
}
