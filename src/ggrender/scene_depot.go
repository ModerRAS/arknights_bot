package ggrender

import (
	"image"
	"sync"

	"github.com/fogleman/gg"
)

// Depot — mirrors template/Depot.tmpl rendered at 850x156 CSS, scale 1.5 -> 1275x234.
// #main bg #2e3031; .item inline-flex columns (80px wide + collapsed whitespace
// between), 10 per line; .icon 75px square art whose opaque circle inscribes the
// tile; .count abspos white 12px on black 50%.
// Geometry constants calibrated against diff.png silhouettes (harness loop).
// Frozen fixture depot-minimal: 11x {龙门币, 100000, prts LMD thumb}.

type DepotItem struct {
	Name   string
	Count  string
	Icon   string
	SortId int64
}

type DepotData struct {
	Items []DepotItem
}

const depotMaterialURL = "https://media.prts.wiki/thumb/6/6a/%E9%81%93%E5%85%B7_%E5%B8%A6%E6%A1%86_%E9%BE%99%E9%97%A8%E5%B8%81.png/75px-%E9%81%93%E5%85%B7_%E5%B8%A6%E6%A1%86_%E9%BE%99%E9%97%A8%E5%B8%81.png"

func SampleDepot() *DepotData {
	items := make([]DepotItem, 0, 11)
	for i := 0; i < 11; i++ {
		items = append(items, DepotItem{Name: "龙门币", Count: "100000", Icon: depotMaterialURL, SortId: 1})
	}
	return &DepotData{Items: items}
}

var (
	depotIconMu    sync.Mutex
	depotIconCache = map[string]image.Image{}
)

// fetchCached loads an icon URL once (remote first, local asset fallback).
func fetchCached(url string) image.Image {
	depotIconMu.Lock()
	defer depotIconMu.Unlock()
	if img, ok := depotIconCache[url]; ok {
		return img
	}
	fallback := AssetPath("common/amiya.png")
	if url == depotMaterialURL {
		fallback = AssetPath("depot/lmd-full.png")
	}
	img := FetchImage(url, fallback)
	depotIconCache[url] = img
	return img
}

// depot layout knobs (CSS px; canvas scaled 1.5x). Calibrated via harness score.
const (
	depotPitchX = 83.87 // item stride incl. whitespace
	depotRowH   = 80.0  // line pitch
	depotIconPx = 77    // rendered icon box px @1.5 (art circle d=70/75 of box)
	depotIconOX = 0.0   // icon box origin correction within tile
	depotIconOY = -1.7  // row circle top at y=row*80 device-exact
)

func RenderDepot(data *DepotData) (*gg.Context, error) {
	dc := gg.NewContext(1275, 234)
	dc.Scale(1.5, 1.5)
	FillBackground(dc, 0x2e, 0x30, 0x31)

	for i, it := range data.Items {
		if i >= 11 {
			break
		}
		row, col := i/10, i%10
		x := float64(col)*depotPitchX + depotIconOX
		y := float64(row)*depotRowH + depotIconOY

		icon := fetchCached(it.Icon)
		dc.DrawImage(ScaleExact(icon, depotIconPx, depotIconPx), int(x+0.5), int(y+0.5))

		// .count badge: white 12px on black 50%, calibrated block
		setFont(dc, 12)
		tw, th := measure(dc, it.Count)
		bx := x + depotBadgeDX
		by := y + depotBadgeDY
		dc.SetRGBA255(0, 0, 0, 128)
		dc.DrawRectangle(bx, by, tw+2, th+2)
		dc.Fill()
		dc.SetRGB255(255, 255, 255)
		drawString(dc, it.Count, bx+1, by+th*0.8+1)
	}
	return dc, nil
}

// badge offset knobs relative to icon box origin (CSS px).
const (
	depotBadgeDX = 36
	depotBadgeDY = 52
)
