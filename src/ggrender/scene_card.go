package ggrender

import (
	"image"
	"image/color"
	"strings"

	"github.com/fogleman/gg"
)

// Card — mirrors template/Card.tmpl rendered at 1280x720 CSS, scale 1 -> 1280x720.
// Layout coordinates follow the frozen Playwright baseline structure: absolute-positioned
// left info column (regTime/circle icons/助理/secretary/decor/skin/hire/nations),
// right-floating name card (headicon/level/avatar/name/uid pills), resume strip,
// assist-characters panel and module-collection panel.
type CardAssist struct {
	Name, CharId, SkinId string
	Level                int
	EvolvePhase          int
	PotentialRank        int
	MainSkillLvl         int
	SpecializeLevel      int
	IsSpecMax            bool
}

type CardNation struct {
	Name string
	Flag int64
}

type CardInfo struct {
	Name, Uid, ServerName, Resume            string
	Level                                    int
	RegTime                                  int
	RegisteredOn                             string
	MainStageProgress                        string
	Avatar, Secretary                        string
	SecretaryName, SecretaryEnName           string
	CharCnt, FurnitureCnt, SkinCnt, EquipCnt int
	EquipOperatorCnt, EquipStage3Cnt         int
	NationList                               []CardNation
	AssistChars                              []CardAssist
}

// SampleCard mirrors the frozen card-minimal fixture (same inputs the Playwright
// baseline captured): amiya secretary painting, amiya#1 player portrait,
// amiya/angel/texas assist avatars, rhodes/lungmen/yan nation flags.
func SampleCard() *CardInfo {
	return &CardInfo{
		Name: "冻结博士", Uid: "10000001", ServerName: "官服", Resume: "稳定基线签名",
		Level: 120, RegTime: 1704067200, RegisteredOn: "2024-01-01", MainStageProgress: "14-21",
		Avatar:        "https://web.hycdn.cn/arknights/game/assets/char_skin/portrait/char_002_amiya%231.png",
		Secretary:     "https://media.prts.wiki/d/dd/%E7%AB%8B%E7%BB%98_%E9%98%BF%E7%B1%B3%E5%A8%85_1.png",
		SecretaryName: "阿米娅", SecretaryEnName: "Amiya",
		CharCnt: 289, FurnitureCnt: 100, SkinCnt: 88, EquipCnt: 46,
		EquipOperatorCnt: 23, EquipStage3Cnt: 12,
		NationList: []CardNation{{"rhodes", 1}, {"lungmen", 0}, {"yan", -1}},
		AssistChars: []CardAssist{
			{Name: "阿米娅", CharId: "char_002_amiya", SkinId: "char_002_amiya#1", Level: 90, EvolvePhase: 2, PotentialRank: 5, MainSkillLvl: 10, SpecializeLevel: 3, IsSpecMax: true},
			{Name: "能天使", CharId: "char_103_angel", SkinId: "char_103_angel#1", Level: 90, EvolvePhase: 2, PotentialRank: 1, MainSkillLvl: 7, IsSpecMax: false},
			{Name: "德克萨斯", CharId: "char_102_texas", SkinId: "char_102_texas#1", Level: 80, EvolvePhase: 2, PotentialRank: 3, MainSkillLvl: 10, SpecializeLevel: 3, IsSpecMax: true},
		},
	}
}

// ---- card-local drawing helpers (self-contained; shared helpers.go untouched) ----

// ascFrac: baseline offset as fraction of measured (Ascent+Descent) height when
// anchoring text by its visual top. Tuned against the frozen baseline.
const cardAscFrac = 0.80

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

// cardTextSpacedCenter draws s with letter-spacing, centered horizontally at cx.
func cardTextSpacedCenter(dc *gg.Context, s string, cx, cy, spacing float64) {
	runes := []rune(s)
	total := 0.0
	adv := make([]float64, len(runes))
	for i, r := range runes {
		w, _ := measure(dc, string(r))
		adv[i] = w
		total += w
	}
	if len(runes) > 1 {
		total += spacing * float64(len(runes)-1)
	}
	x := cx - total/2
	for i, r := range runes {
		drawString(dc, string(r), x, cy)
		x += adv[i] + spacing
	}
}

// cardBold spreads synthetic bold like enemy's drawStringBoldW.
func cardBold(dc *gg.Context, s string, x, y, spread float64) {
	drawString(dc, s, x-spread/2, y)
	drawString(dc, s, x+spread/2, y)
}

// cardTintedByMask recolors a mask image (scaled to w×h) with solid color c,
// preserving the mask's alpha. Used for the drop-shadow-recolored circle icon
// and flag==1 nation silhouettes.
func cardTintedByMask(maskSrc image.Image, w, h int, c color.RGBA) image.Image {
	m := ScaleExact(maskSrc, w, h)
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	b := m.Bounds()
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			_, _, _, a := m.At(x, y).RGBA()
			if a > 0 {
				out.SetRGBA(x, y, color.RGBA{R: c.R, G: c.G, B: c.B, A: uint8(a >> 8)})
			}
		}
	}
	return out
}

// cardWithOpacity returns a copy of src at uniform opacity a (0..255).
func cardWithOpacity(src image.Image, a int) image.Image {
	b := src.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			r, g, bl, al := src.At(x, y).RGBA()
			na := al * uint32(a) / 0xFFFF
			if na > 0 {
				out.SetRGBA(x, y, color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(bl >> 8), A: uint8(na >> 8)})
			}
		}
	}
	return out
}

// cardFadeBottom applies the template's mask linear-gradient(to top, transparent, #fff 50%):
// fully transparent at the bottom edge, fully opaque from mid-height upward.
func cardFadeBottom(src image.Image) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	half := float64(h) / 2
	for y := 0; y < h; y++ {
		frac := (float64(h-y) - 0.5) / half
		if frac > 1 {
			frac = 1
		}
		if frac < 0 {
			frac = 0
		}
		for x := 0; x < w; x++ {
			r, g, bl, al := src.At(b.Min.X+x, b.Min.Y+y).RGBA()
			na := uint32(float64(al) * frac)
			if na > 0 {
				out.SetRGBA(x, y, color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(bl >> 8), A: uint8(na >> 8)})
			}
		}
	}
	return out
}

// cardPanelLayer paints into a w×h layer, clips to a rounded rect of radius r
// (overflow hidden semantics), returns the composited layer.
func cardPanelLayer(w, h int, r float64, paint func(*image.RGBA)) image.Image {
	layer := image.NewRGBA(image.Rect(0, 0, w, h))
	paint(layer)
	mask := gg.NewContext(w, h)
	mask.DrawRoundedRectangle(0, 0, float64(w), float64(h), r)
	mask.SetColor(color.White)
	mask.Fill()
	mi := mask.Image()
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			_, _, _, ma := mi.At(x, y).RGBA()
			r, g, b, a := layer.At(x, y).RGBA()
			na := a * ma / 0xFFFF
			if na > 0 {
				out.SetRGBA(x, y, color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(na >> 8)})
			}
		}
	}
	return out
}

func cardPill(dc *gg.Context, s string, x, yTop float64) float64 {
	setFont(dc, 17)
	w, h := measure(dc, s)
	pillH := h * 1.4
	dc.SetRGBA255(0, 0, 0, 51)
	RoundRect(dc, x, yTop, w+10, pillH, 10)
	dc.Fill()
	cardTextCenter(dc, s, x+(w+10)/2, yTop+pillH/2)
	return w + 10
}

func cardAvatarURL(skinId string) string {
	const base = "https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/"
	return base + strings.ReplaceAll(skinId, "#", "%23") + ".png"
}

var (
	cardBlue   = color.RGBA{R: 0, G: 152, B: 220, A: 255}   // #0098dc
	cardDark   = color.RGBA{R: 17, G: 17, B: 17, A: 255}    // #111
	cardGray   = color.RGBA{R: 163, G: 163, B: 162, A: 255} // #a3a3a2
	cardWhiteC = color.RGBA{R: 255, G: 255, B: 255, A: 255}
)

func RenderCard(data *CardInfo) (*gg.Context, error) {
	const W, H = 1280, 720
	dc := gg.NewContext(W, H)
	// background: assets/card/bg.png covers whole canvas; #0098dc body fallback
	FillBackground(dc, 0, 152, 220)
	if bg, err := LoadImage(AssetPath("card/bg.png")); err == nil {
		dc.DrawImage(ScaleCover(bg, W, H), 0, 0)
	}

	// ---- secretary painting: height 720 at left 5%, bottom half fades out ----
	sec := FetchImage(data.Secretary, AssetPath("common/amiya.png"))
	dc.DrawImage(cardFadeBottom(ScaleExactCR(sec, sec.Bounds().Dx()*720/sec.Bounds().Dy(), 720)), 64, 0)

	// ================= left info column =================
	// 入职日 bar: #0098dc 132x24 at (22,20); dark text, date on white chip
	dc.SetRGB255(0, 152, 220)
	dc.DrawRectangle(22, 20, 132, 24)
	dc.Fill()
	setFont(dc, 16)
	cardTextCenter(dc, "入职日", 47, 32)
	dateW, _ := measure(dc, data.RegisteredOn)
	chipX := 76.0
	chipW := dateW + 6
	dc.SetRGB255(255, 255, 255)
	dc.DrawRectangle(chipX, 20, chipW, 24)
	dc.Fill()
	dc.SetColor(cardDark)
	cardTextCenter(dc, data.RegisteredOn, chipX+chipW/2, 32)

	// circle icon recolored #0098dc via its own alpha (drop-shadow trick equivalent)
	if circ, err := LoadImage(AssetPath("card/no_use_icon_circle.png")); err == nil {
		dc.DrawImage(cardTintedByMask(circ, 16, 42, cardBlue), 20, 50)
	}
	if xi, err := LoadImage(AssetPath("card/no_use_icon_x.png")); err == nil {
		dc.DrawImage(ScaleExact(xi, 16, 15), 20, 96)
	}
	// 助理 white chip
	dc.SetRGB255(255, 255, 255)
	dc.DrawRectangle(20, 121, 50, 24)
	dc.Fill()
	dc.SetColor(cardDark)
	cardTextCenter(dc, "助理", 45, 133)

	dc.SetRGB255(255, 255, 255)
	setFont(dc, 24)
	cardTextTop(dc, data.SecretaryName, 20, 151)
	setFont(dc, 17)
	cardTextTop(dc, data.SecretaryEnName, 20, 185)

	if d, err := LoadImage(AssetPath("card/decor.png")); err == nil {
		dc.DrawImage(ScaleExact(d, 177, 47), 20, 248)
	}
	setFont(dc, 12)
	cardTextTop(dc, "DATA PROVIDED BY PRTS", 20, 297)
	cardTextTop(dc, "-", 20, 320)
	if ds, err := LoadImage(AssetPath("card/decor_skin.png")); err == nil {
		dc.DrawImage(ScaleExact(ds, 96, 27), 20, 360)
	}
	// 时装保有数 table: icon 57x41 + two centered rows to its right
	if ic, err := LoadImage(AssetPath("card/icon_skin.png")); err == nil {
		dc.DrawImage(ScaleExact(ic, 57, 41), 23, 406)
	}
	setFont(dc, 16)
	dc.SetRGB255(255, 255, 255)
	cardTextCenter(dc, "时装保有数", 122, 416)
	cardTextCenter(dc, itoa(data.SkinCnt), 122, 437)

	// 雇佣干员进度 bar + human_resource art
	dc.SetRGB255(0, 152, 220)
	dc.DrawRectangle(20, 455, 160, 34)
	dc.Fill()
	setFont(dc, 17)
	dc.SetRGB255(255, 255, 255)
	cardTextSpacedCenter(dc, "雇佣干员进度", 100, 477, 3)
	if hr, err := LoadImage(AssetPath("card/human_resource.png")); err == nil {
		dc.DrawImage(ScaleExact(hr, 114, 35), 160, 455)
	}

	// char count 55px
	setFont(dc, 55)
	cardTextTop(dc, itoa(data.CharCnt), 20, 504)

	// nation flags 30x30 pitch 37: flag==1 -> blue silhouette, flag==-1 -> 20% opacity
	nx, ny := 17.0, 573.0
	for _, n := range data.NationList {
		scaled := ScaleExact(tryLocal("card/"+n.Name+".png"), 30, 30)
		var img image.Image
		switch n.Flag {
		case 1:
			img = cardTintedByMask(scaled, 30, 30, cardBlue)
		case -1:
			img = cardWithOpacity(scaled, 51)
		default:
			img = scaled
		}
		dc.DrawImage(img, int(nx), int(ny))
		nx += 37
	}

	// ================= right name card =================
	// bg 638x223 floats right: x=642, effective top 0
	if nc, err := LoadImage(AssetPath("card/name_card_short.png")); err == nil {
		dc.DrawImage(ScaleExact(nc, 638, 223), 642, 0)
	}
	if hd, err := LoadImage(AssetPath("card/headicon_back.png")); err == nil {
		dc.DrawImage(ScaleExact(hd, 147, 149), 672, 30)
	}
	// level widget: bg 84x84 at (676,3), number fs20 pad-top 8, LV fs14 below
	if lb, err := LoadImage(AssetPath("card/level_bg.png")); err == nil {
		dc.DrawImage(ScaleExact(lb, 84, 84), 676, 3)
	}
	dc.SetRGB255(255, 255, 255)
	setFont(dc, 20)
	cardTextTopCenter(dc, itoa(data.Level), 718, 11)
	setFont(dc, 14)
	cardTextTopCenter(dc, "LV", 718, 39)
	// player portrait 180x360 -> 130x260 at (681,11)
	portrait := FetchImage(data.Avatar, AssetPath("common/amiya.png"))
	dc.DrawImage(ScaleExact(portrait, 130, 260), 681, 11)
	// name fs30 at (842,55)
	setFont(dc, 30)
	cardTextTop(dc, data.Name, 842, 55)
	// uid/server pills fs17 at (842,102)
	setFont(dc, 17)
	px := 842.0
	px += cardPill(dc, "ID "+data.Uid, px, 102) + 6
	cardPill(dc, data.ServerName, px, 102)

	// ================= resume strip =================
	dc.SetRGBA255(0, 0, 0, 153)
	RoundRect(dc, 659, 159, 605, 70, 15)
	if ri, err := LoadImage(AssetPath("card/resume_icon.png")); err == nil {
		dc.DrawImage(ScaleExact(ri, 45, 32), 682, 178)
	}
	resume := StripHTML(data.Resume)
	if resume == "" {
		resume = "暂未设置签名"
	}
	setFont(dc, 19)
	dc.SetColor(cardGray)
	cardTextTop(dc, resume, 786, 184)

	// ================= assist characters panel =================
	dc.SetRGBA255(0, 0, 0, 153)
	RoundRect(dc, 659, 239, 605, 200, 15)
	// label column centered at cx 729.5
	if ai, err := LoadImage(AssetPath("card/assist_icon.png")); err == nil {
		dc.DrawImage(ScaleExact(ai, 54, 64), 702, 264)
	}
	setFont(dc, 17)
	dc.SetColor(cardGray)
	cardTextSpacedCenter(dc, "助战干员", 729, 348, 7)
	setFont(dc, 13)
	cardTextTopCenter(dc, "SUPPORT UNIT", 729, 360)
	for i, a := range data.AssistChars {
		if i >= 3 {
			break
		}
		cx := 783 + i*154
		cy := 264
		if be, err := LoadImage(AssetPath("card/back_end.png")); err == nil {
			dc.DrawImage(ScaleExact(be, 150, 150), cx, cy)
		}
		av := FetchImage(cardAvatarURL(a.SkinId), AssetPath("common/amiya.png"))
		dc.DrawImage(ScaleExact(av, 130, 130), cx+10, cy+2)
		if a.IsSpecMax {
			if sm, err := LoadImage(AssetPath("card/spec_max_icon.png")); err == nil {
				dc.DrawImage(ScaleExact(sm, 50, 46), cx+95, cy+2)
			}
		}
		if ev, err := LoadImage(AssetPath("box/Evolve_" + itoa(a.EvolvePhase) + ".png")); err == nil {
			dc.DrawImage(ScaleExact(ev, 40, 33), cx+100, cy+97)
		}
		// LV badge top-left
		dc.SetRGB255(255, 255, 255)
		setFont(dc, 10)
		cardTextTopCenter(dc, "LV", float64(cx)+30, float64(cy)+3)
		setFont(dc, 17)
		cardTextTopCenter(dc, itoa(a.Level), float64(cx)+30, float64(cy)+13)
	}

	// ================= module collection panel (overflow hidden) =================
	modules := cardPanelLayer(605, 196, 15, func(layer *image.RGBA) {
		lg := gg.NewContextForImage(layer)
		if mb, err := LoadImage(AssetPath("card/module_collection_bg.png")); err == nil {
			lg.DrawImage(ScaleExact(mb, 612, 178), -7, 0)
		}
		if mi, err := LoadImage(AssetPath("card/module_collection_bg_icon.png")); err == nil {
			lg.DrawImage(cardWithOpacity(ScaleExact(mi, 175, 163), 77), 40, 18)
		}
	})
	dc.DrawImage(modules, 659, 450)
	// numbers row right-aligned at 1264, number tops 519, titles 606
	cols := []struct {
		title string
		val   int
	}{{"总收集模组", data.EquipCnt}, {"STAGE3模组", data.EquipStage3Cnt}, {"拥有模组干员", data.EquipOperatorCnt}}
	widths := make([]float64, len(cols))
	for i, c := range cols {
		setFont(dc, 21)
		tw, _ := measure(dc, c.title)
		setFont(dc, 50)
		nw, _ := measure(dc, itoa(c.val))
		widths[i] = tw
		if nw > widths[i] {
			widths[i] = nw
		}
	}
	right := 1264.0
	for i := len(cols) - 1; i >= 0; i-- {
		c := cols[i]
		cx := right - widths[i]/2
		setFont(dc, 50)
		dc.SetColor(cardGray)
		nw, _ := measure(dc, itoa(c.val))
		cardBold(dc, itoa(c.val), cx-nw/2, 519+50*cardAscFrac, 1.5)
		setFont(dc, 21)
		cardTextTopCenter(dc, c.title, cx, 606)
		right -= widths[i] + 20
	}
	return dc, nil
}
