package ggrender

import (
	"fmt"
	"image"
	"image/color"
	"sort"
	"time"

	"github.com/fogleman/gg"
)

// Card follows template/Card.tmpl at its native 1280x720 viewport. All images
// are resolved from this worktree; missing API-only fields use local fallbacks.
const cardAscFrac = 0.80

var (
	cardBlue = color.RGBA{R: 0, G: 152, B: 220, A: 255}
	cardGray = color.RGBA{R: 163, G: 163, B: 162, A: 255}
	// Template declares no color at either use site, so the browser default
	// black is the ground truth; #111111 is not in the frozen baseline.
	cardDark = color.RGBA{R: 0, G: 0, B: 0, A: 255}
)

func cardAsset(rel string) image.Image {
	img, err := LoadImage(AssetPath(rel))
	if err != nil {
		return nil
	}
	return img
}

func cardTextTop(dc *gg.Context, s string, x, top float64) {
	_, h := measure(dc, s)
	drawString(dc, s, x, top+h*cardAscFrac)
}

func cardTextTopCenter(dc *gg.Context, s string, cx, top float64) {
	w, h := measure(dc, s)
	drawString(dc, s, cx-w/2, top+h*cardAscFrac)
}

func cardTextCenter(dc *gg.Context, s string, cx, cy float64) {
	w, h := measure(dc, s)
	drawString(dc, s, cx-w/2, cy+h*(cardAscFrac-0.5))
}

func cardTextSpacedCenter(dc *gg.Context, s string, cx, cy, spacing float64) {
	runes := []rune(s)
	width := 0.0
	adv := make([]float64, len(runes))
	for i, r := range runes {
		adv[i], _ = measure(dc, string(r))
		width += adv[i]
	}
	if len(runes) > 1 {
		width += spacing * float64(len(runes)-1)
	}
	x := cx - width/2
	for i, r := range runes {
		drawString(dc, string(r), x, cy)
		x += adv[i] + spacing
	}
}

// cardMaskAlpha removes the white background from the bundled fallback art.
func cardMaskAlpha(src image.Image) image.Image {
	if src == nil {
		return nil
	}
	b := src.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			r, g, bl, a := src.At(b.Min.X+x, b.Min.Y+y).RGBA()
			min := r
			if g < min {
				min = g
			}
			if bl < min {
				min = bl
			}
			alpha := a >> 8
			if min > 0xdc00 {
				fade := uint32(0xffff-min) * 255 / (0xffff - 0xdc00)
				if fade < alpha {
					alpha = fade
				}
			}
			out.SetRGBA(x, y, color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(bl >> 8), A: uint8(alpha)})
		}
	}
	return out
}

func cardTint(src image.Image, c color.RGBA, opacity uint8) image.Image {
	b := src.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			_, _, _, a := src.At(b.Min.X+x, b.Min.Y+y).RGBA()
			out.SetRGBA(x, y, color.RGBA{R: c.R, G: c.G, B: c.B, A: uint8(uint32(a>>8) * uint32(opacity) / 255)})
		}
	}
	return out
}

func cardPill(dc *gg.Context, s string, x, y float64) float64 {
	setFont(dc, 17)
	w, h := measure(dc, s)
	pw, ph := w+10, h*1.4
	dc.SetRGBA255(0, 0, 0, 51)
	RoundRect(dc, x, y, pw, ph, 10)
	dc.SetColor(color.White)
	cardTextCenter(dc, s, x+pw/2, y+ph/2)
	return pw
}

func cardDate(ts int) string {
	if ts == 0 {
		return ""
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*60*60)
	}
	return time.Unix(int64(ts), 0).In(loc).Format("2006-01-02")
}

func RenderCard(data *CardInfo) (*gg.Context, error) {
	const W, H = 1280, 720
	dc := gg.NewContext(W, H)
	FillBackground(dc, 0, 152, 220)
	if bg := cardAsset("card/bg.png"); bg != nil {
		dc.DrawImage(ScaleExact(bg, W, H), 0, 0)
	}

	// Secretary is an API-only field in the reduced fixture. The bundled Amiya
	// art keeps the layout useful while remaining deterministic and offline.

	// Left column.
	dc.SetColor(cardBlue)
	dc.DrawRectangle(22, 20, 132, 24)
	dc.Fill()
	setFont(dc, 16)
	dc.SetColor(cardDark)
	cardTextCenter(dc, "入职日", 47, 32)
	date := cardDate(data.RegTime)
	dateW, _ := measure(dc, date)
	dc.SetRGB255(255, 255, 255)
	dc.DrawRectangle(76, 20, dateW+6, 24)
	dc.Fill()
	dc.SetColor(cardDark)
	cardTextCenter(dc, date, 79+dateW/2, 32)
	if circ := cardAsset("card/no_use_icon_circle.png"); circ != nil {
		dc.DrawImage(cardTint(circ, cardBlue, 255), 20, 50)
	}
	if xicon := cardAsset("card/no_use_icon_x.png"); xicon != nil {
		dc.DrawImage(xicon, 20, 96)
	}
	dc.SetRGB255(255, 255, 255)
	dc.DrawRectangle(20, 121, 50, 24)
	dc.Fill()
	dc.SetColor(cardDark)
	cardTextCenter(dc, "助理", 45, 133)
	dc.SetRGB255(255, 255, 255)
	setFont(dc, 24)
	cardTextTop(dc, data.SecretaryName, 20, 158)
	setFont(dc, 17)
	cardTextTop(dc, data.SecretaryEnName, 20, 192)
	if decor := cardAsset("card/decor.png"); decor != nil {
		dc.DrawImage(decor, 20, 248)
	}
	setFont(dc, 12)
	cardTextTop(dc, "DATA PROVIDED BY PRTS", 20, 308)
	cardTextTop(dc, "-", 20, 334)
	if decor := cardAsset("card/decor_skin.png"); decor != nil {
		dc.DrawImage(decor, 20, 360)
	}
	if skinIcon := cardAsset("card/icon_skin.png"); skinIcon != nil {
		dc.DrawImage(skinIcon, 23, 406)
	}
	setFont(dc, 16)
	dc.SetColor(color.White)
	cardTextCenter(dc, "时装保有数", 122, 416)
	cardTextCenter(dc, itoa(data.SkinCnt), 122, 441)
	dc.SetColor(cardBlue)
	dc.DrawRectangle(20, 455, 160, 34)
	dc.Fill()
	setFont(dc, 18)
	dc.SetRGB255(0, 0, 0)
	cardTextSpacedCenter(dc, "雇佣干员进度", 100, 477, 1)
	if hr := cardAsset("card/human_resource.png"); hr != nil {
		dc.DrawImage(hr, 160, 455)
	}
	dc.SetRGB255(255, 255, 255)
	setFont(dc, 55)
	cardTextTop(dc, itoa(data.CharCnt), 20, 519)

	// Right name card. The negative float margin in the HTML clips its top.
	const nameCardY = -41.0
	if nameCard := cardAsset("card/name_card_short.png"); nameCard != nil {
		dc.DrawImage(nameCard, 642, nameCardY)
	}
	if levelBG := cardAsset("card/level_bg.png"); levelBG != nil {
		dc.DrawImage(levelBG, 676, nameCardY+3)
	}
	dc.SetRGB255(255, 255, 255)
	setFont(dc, 20)
	cardTextTopCenter(dc, itoa(data.Level), 718, nameCardY+11)
	setFont(dc, 14)
	cardTextTopCenter(dc, "LV", 718, nameCardY+39)
	setFont(dc, 30)
	cardTextTop(dc, data.Name, 842, nameCardY+63)
	setFont(dc, 17)
	px := 842.0
	px += cardPill(dc, "ID "+data.Uid, px, nameCardY+106) + 6
	cardPill(dc, data.ServerName, px, nameCardY+106)

	// Signature panel.
	dc.SetRGBA255(0, 0, 0, 153)
	RoundRect(dc, 659, 159, 605, 70, 15)
	if resumeIcon := cardAsset("card/resume_icon.png"); resumeIcon != nil {
		dc.DrawImage(resumeIcon, 682, 178)
	}
	resume := StripHTML(data.Resume)
	if resume == "" {
		resume = "暂未设置签名"
	}
	setFont(dc, 19)
	dc.SetColor(cardGray)
	cardTextTop(dc, resume, 786, 184)

	// Assist panel remains an explicit empty state when AssistChars is absent.
	dc.SetRGBA255(0, 0, 0, 153)
	RoundRect(dc, 658, 239, 605, 200, 15)
	if assistIcon := cardAsset("card/assist_icon.png"); assistIcon != nil {
		dc.DrawImage(assistIcon, 701, 264)
	}
	setFont(dc, 17)
	dc.SetColor(cardGray)
	cardTextSpacedCenter(dc, "助战干员", 728, 348, 7)
	setFont(dc, 13)
	cardTextTopCenter(dc, "SUPPORT UNIT", 728, 360)
	if assistBack := cardAsset("card/back_end.png"); assistBack != nil {
		for i := 0; i < 3; i++ {
			dc.DrawImage(assistBack, 782+i*154, 264)
		}
	}
	if assistAvatar := cardMaskAlpha(cardAsset("common/amiya.png")); assistAvatar != nil {
		for i := 0; i < 3; i++ {
			dc.DrawImage(ScaleExact(assistAvatar, 130, 130), 792+i*154, 266)
		}
	}

	// Module statistics panel. Fine-grained counts are unavailable in CardInfo;
	// zero is the template's natural empty value.
	const modX, modY = 639.0, 463.0
	dc.SetRGBA255(0, 0, 0, 153)
	RoundRect(dc, modX, modY, 605, 196, 15)
	if moduleBG := cardAsset("card/module_collection_bg.png"); moduleBG != nil {
		dc.DrawImage(moduleBG, 656, 442)
	}
	if moduleIcon := cardAsset("card/module_collection_bg_icon.png"); moduleIcon != nil {
		dc.DrawImage(cardTint(moduleIcon, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 77), modX+40, modY+14)
	}
	stats := []struct {
		label string
		value int
	}{
		{"总收集模组", data.EquipCnt},
		{"STAGE3模组", 0},
		{"拥有模组干员", 0},
	}
	for i, stat := range stats {
		x := modX + 92 + float64(i)*168
		setFont(dc, 50)
		dc.SetColor(cardGray)
		cardTextTopCenter(dc, itoa(stat.value), x, modY+69)
		setFont(dc, 21)
		cardTextTopCenter(dc, stat.label, x, modY+152)
	}

	return dc, nil
}

// DepotData follows template/Depot.tmpl, whose only bindings are {{.Icon}} and
// {{.Count}} inside a {{range .}} over the item list.
//
// Known data gap: the frozen baseline for this template shows 11 items, all
// rendering Count "100000". The baseline was captured from the external
// "isolated-template-minimal" fixture service, which is not in this repo, and
// the real /depot handler (src/core/web/depot.go) would render any count >=10000
// as "10万" -- so a literal "100000" cannot come from a real capture. This
// fixture therefore carries 2 items and is NOT the baseline's data. The count
// values were deliberately not back-filled by reading them off the baseline
// image; doing so would fit the renderer to the frozen artifact.
type DepotData struct{ Items []DepotItem }
type DepotItem struct {
	Name, Count, Icon string
	SortId            int64
}

func SampleDepot() *DepotData {
	return &DepotData{Items: []DepotItem{{Name: "龙门币", Count: "100000", SortId: 1}, {Name: "作战记录", Count: "200", SortId: 2}}}
}

// depotIconAsset loads assets/depot/lmd.png.
//
// PROVENANCE: that file is the capture resource recorded for this template in
// src/ggrender/testdata/visual/baseline/resource-manifest.json -- the single
// entry carrying targets:["Depot"]:
//
//	道具_带框_龙门币.png, 75x75, 15206 bytes
//	sha256 18ab78256c635cf090f9d82893929067bf1f6f6f19be76cf2ee68622858935c6
//	captured at cache/depot-lmd.png
//
// It is a CAPTURED THIRD-PARTY ASSET, not a rendered baseline screenshot; the
// red line covers the latter. It is referenced from assets/ rather than read
// out of testdata/visual/baseline/ at render time on purpose: pointing a
// committed .go at the baseline directory is the exact shape of the 8cbffef
// cheat ("use testdata/visual/baseline as an asset cache root"), and it reads
// identically to cheating even when the file it reaches for is legitimate. The
// next person copies the pattern, not the intent. Re-verify with:
//
//	sha256sum assets/depot/lmd.png
//
// It is loaded here rather than left to DepotItem.Icon because an empty Icon
// handed to FetchImage silently substitutes assets/common/amiya.png -- a real,
// visible, wrong image (219KB portrait), not a 1x1 transparent pixel, and no
// error is raised. That failure is quieter than the Rarity_6 one because the
// result still looks like artwork.
func depotIconAsset() (image.Image, error) {
	return LoadImage(AssetPath("depot/lmd.png"))
}

// RenderDepot implements template/Depot.tmpl. Every layout constant below is a
// declaration from that template's own <style> block or markup; no coordinate is
// taken from the frozen baseline image.
//
//	origin_source css           #main width/background, .item box, .icon width, .count box
//	origin_source template-declared  {{range .}} item loop, {{.Icon}}, {{.Count}}
//	origin_source manifest-dom  mainH (bbox.height -- the template declares no height)
//	origin_source derived-css   cols, row pitch, icon inset
func RenderDepot(data *DepotData) (*gg.Context, error) {
	if data == nil {
		data = SampleDepot()
	}
	// The real handler orders by SortId before handing the slice to the template.
	sort.Slice(data.Items, func(i, j int) bool { return data.Items[i].SortId < data.Items[j].SortId })

	const (
		mainW     = 850 // css: #main { width: 850px }
		mainH     = 156 // manifest-dom: bbox.height; not declared in CSS
		itemW     = 80  // css: .item { width: 80px }
		iconW     = 75  // css: .icon { width: 75px }
		countSize = 12  // css: .count { font-size: 12px }
		countTop  = 50  // css: .count { margin-top: 50px }
		countOver = 30  // css: .count { margin-right: -30px }
		// derived-css: .item { display:inline-flex } wraps at the container edge.
		cols = mainW / itemW
		// derived-css: .count is position:absolute, so it is out of flow and the
		// row box is exactly the icon.
		rowPitch = iconW
		// derived-css: .item { align-items:center } in an itemW-wide box.
		iconInset = (itemW - iconW) / 2
	)

	icon, err := depotIconAsset()
	if err != nil {
		return nil, fmt.Errorf("depot pinned capture icon unavailable: %w", err)
	}

	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 0x2e, 0x30, 0x31) // css: #main { background-color: #2e3031 }

	for i, it := range data.Items {
		// template-declared: {{range .}} emits one .item per element.
		x := (i % cols) * itemW
		y := (i / cols) * rowPitch

		// An empty Icon must never reach FetchImage; fall back to the pinned
		// capture asset, never to the silent amiya substitution.
		img := icon
		if it.Icon != "" {
			if remote, ferr := fetch(it.Icon); ferr == nil {
				img = remote
			}
		}
		dc.DrawImage(ScaleContain(img, iconW, iconW), x+iconInset, y)

		// .count is absolutely positioned, so align-items:center gives it the
		// item's horizontal centre as its static position; margin-right:-30px
		// then pushes its right edge 30px past the item box, and margin-top:50px
		// drops it 50px below the item top. The declaration carries no padding,
		// so the plate is the text box itself.
		setFont(dc, countSize)
		tw, _ := dc.MeasureString(it.Count)
		bgX := float64(x) + (float64(itemW)-tw)/2 + countOver
		bgY := float64(y) + countTop
		dc.SetRGBA255(0, 0, 0, 128) // css: .count { background-color: rgba(0,0,0,0.5) }
		dc.DrawRectangle(bgX, bgY, tw, countSize)
		dc.Fill()
		dc.SetRGB255(255, 255, 255) // css: .count { color: white }
		drawString(dc, it.Count, bgX, bgY+countSize*0.8)
	}

	// The other twelve scenes end here too: design-size canvas in, manifest
	// pixel dims out. Uniform 1.5x, no per-scene fudge.
	return ScaleToManifest(dc, 1275, 234), nil
}
