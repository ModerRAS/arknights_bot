package skia

import (
	"fmt"
	"image/color"
	"net/url"
	"os"
	"path/filepath"
)

// ponytail: depot renderer minimal — uses yoga stub for layout, canvas for rastr.
// P1 depot closed loop: items 80x78 gap3.5 + 75 icon + count badge.

func RenderDepotStub(w, h int, scale float64) *Canvas {
	if scale <= 0 {
		scale = 1.5
	}
	if w <= 0 {
		w = int(850*scale + 0.5)
	}
	if h <= 0 {
		h = int(156*scale + 0.5)
	}
	c := NewCanvas(w, h)
	c.Clear(color.RGBA{0x2e, 0x30, 0x31, 255})
	root := NewNode()
	root.SetWidth(float32(w) / float32(scale))
	root.SetHeight(float32(h) / float32(scale))
	for i := 0; i < 40; i++ {
		child := NewNode()
		root.AddChild(child)
	}
	root.CalculateLayout(float32(w)/float32(scale), float32(h)/float32(scale))
	for _, child := range root.Children {
		x, y, _, _ := child.GetComputedLayout()
		px := int(x*float32(scale) + 0.5)
		py := int(y*float32(scale) + 0.5)
		iw := int(80*scale + 0.5)
		_ = int(78*scale + 0.5)
		iconW := int(75*scale + 0.5)
		iconH := int(75*scale + 0.5)
		iconX := px + (iw-iconW)/2
		iconY := py
		c.DrawRect(Rect{X: float32(iconX), Y: float32(iconY), W: float32(iconW), H: float32(iconH)}, Paint{Color: color.RGBA{0x55, 0x55, 0x55, 255}})
		badgeY := py + int(50*scale+0.5)
		badgeX := px + iw - int(14*scale+0.5) - int(5*scale+0.5)
		badgeW := int(14*scale + 0.5)
		badgeH := int(14*scale + 0.5)
		c.DrawRect(Rect{X: float32(badgeX), Y: float32(badgeY), W: float32(badgeW), H: float32(badgeH)}, Paint{Color: color.RGBA{0, 0, 0, 128}})
		c.DrawRect(Rect{X: float32(badgeX + 2), Y: float32(badgeY + 2), W: float32(8*scale), H: float32(6*scale)}, Paint{Color: color.RGBA{0xff, 0xff, 0xff, 255}})
	}
	return c
}

func RenderDepotPlaceholder(w, h int) *Canvas { return RenderDepotStub(w, h, 1.5) }

// Headhunt — 53行极简验证 backgroundSize 110×230, bg#2e3031 + back_*.png + rarity
type HeadhuntItem struct {
	Name       string `json:"name"`
	Profession string `json:"profession"`
	Rarity     int    `json:"rarity"`
	ThumbURL   string `json:"thumbURL"`
}

func findRepoRoot() string {
	cands := []string{"C:/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go", "."}
	if r := os.Getenv("SKIA_REPO_ROOT"); r != "" {
		cands = append([]string{r}, cands...)
	}
	for _, c := range cands {
		if _, err := os.Stat(filepath.Join(c, "assets/font/NotoSansHans-Regular.ttf")); err == nil {
			if abs, err := filepath.Abs(c); err == nil {
				return abs
			}
			return c
		}
	}
	return "C:/WorkSpace/Golang/arknights_bot-satori-yoga-skia-go"
}

func BuildHeadhunt(items []HeadhuntItem) *YogaNode {
	root := NewYogaNode()
	root.Style.Width = 1049
	root.Style.Height = 576
	root.Style.Display = "flex"
	inner := NewYogaNode()
	inner.Style.Display = "flex"
	inner.Style.FlexDirection = "row"
	root.AddChild(inner)
	for _, it := range items {
		n := NewYogaNode()
		n.Style.Width = 95
		n.Style.Height = 270
		n.Text = it.ThumbURL
		inner.AddChild(n)
	}
	root.CalculateLayout(1049, 576)
	return root
}

func RenderHeadhunt(items []HeadhuntItem, scale float64) *Canvas {
	if scale <= 0 {
		scale = 1
	}
	if len(items) == 0 {
		items = make([]HeadhuntItem, 10)
		for i := range items {
			items[i] = HeadhuntItem{Profession: "WARRIOR", Rarity: 5, ThumbURL: "https://media.prts.wiki/a/a0/%E5%8D%8A%E8%BA%AB%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85_1.png?image_process=format,webp/quality,Q_90"}
		}
	}
	W, H := int(1049*scale+0.5), int(576*scale+0.5)
	c := NewCanvas(W, H)
	c.Clear(color.RGBA{0x2e, 0x30, 0x31, 255})
	root := BuildHeadhunt(items)
	ld, _ := NewLoader(findRepoRoot())
	if bg, err := ld.Load("assets/headhunt/bg.png"); err == nil {
		_ = c.DrawImageRect(bg, Rect{X: 0, Y: 0, W: 1024 * float32(scale), H: 576 * float32(scale)})
	}
	inner := root.Children[0]
	for idx, ch := range inner.Children {
		baseX := float32(25) + ch.Layout.X
		baseY := float32(130) + ch.Layout.Y
		it := items[idx]
		if back, err := ld.Load(fmt.Sprintf("assets/headhunt/back_%d.png", it.Rarity)); err == nil {
			_ = c.DrawImageRect(back, Rect{X: (baseX + (95-110)/2) * float32(scale), Y: baseY * float32(scale), W: 110 * float32(scale), H: 230 * float32(scale)})
		}
		if p, err := ld.Load(it.ThumbURL); err == nil {
			_ = c.DrawImageRect(p, Rect{X: baseX * float32(scale), Y: (baseY + 100) * float32(scale), W: 95 * float32(scale), H: 190 * float32(scale)})
		} else {
			c.DrawRect(Rect{X: baseX * float32(scale), Y: (baseY + 100) * float32(scale), W: 95 * float32(scale), H: 190 * float32(scale)}, Paint{Color: color.RGBA{0x55, 0x55, 0x55, 255}})
		}
		if rimg, err := ld.Load(fmt.Sprintf("assets/headhunt/Rarity_%d.png", it.Rarity)); err == nil {
			_ = c.DrawImageRect(rimg, Rect{X: (baseX + (95-75)/2) * float32(scale), Y: (baseY + 100) * float32(scale), W: 75 * float32(scale), H: 20 * float32(scale)})
		}
		if pimg, err := ld.Load(fmt.Sprintf("assets/headhunt/%s.png", it.Profession)); err == nil {
			_ = c.DrawImageRect(pimg, Rect{X: (baseX + (95-75)/2) * float32(scale), Y: (baseY + 190) * float32(scale), W: 75 * float32(scale), H: 76 * float32(scale)})
		}
		_ = MeasureText(it.Name, 12)
	}
	return c
}

// BoxDetail — 61行 5-track [110,62,58,150,100] header/body几何
var boxDetailTracks = [5]float32{110, 62, 58, 150, 100}

type BoxDetailSkill struct {
	Id    string `json:"id"`
	Level int    `json:"level"`
}
type BoxDetailEquip struct {
	Id    string `json:"id"`
	Level int    `json:"level"`
}
type BoxDetailItem struct {
	Name          string           `json:"name"`
	Id            string           `json:"id"`
	Rarity        int              `json:"rarity"`
	Level         int              `json:"level"`
	EvolvePhase   int              `json:"evolvePhase"`
	PotentialRank int              `json:"potentialRank"`
	Skills        []BoxDetailSkill `json:"skills"`
	Equips        []BoxDetailEquip `json:"equips"`
}

func BuildBoxDetail(items []BoxDetailItem) *YogaNode {
	root := NewYogaNode()
	root.Style.Width = 481
	root.Style.Height = 186
	root.Style.FlexDirection = "column"
	root.Style.Display = "flex"
	header := NewYogaNode()
	header.Style.Width = 481
	header.Style.Height = 35
	header.Style.Display = "flex"
	header.Style.FlexDirection = "row"
	labels := []string{"干员", "等级", "潜能", "技能", "模组"}
	for i, lab := range labels {
		n := NewYogaNode()
		n.Style.Width = boxDetailTracks[i]
		n.Style.Height = 35
		txt := lab
		n.SetMeasureFunc(func(w float32, wm MeasureMode, h float32, hm MeasureMode) Size {
			return Size{Width: float32(MeasureText(txt, 14)), Height: 14}
		})
		n.Text = lab
		header.AddChild(n)
	}
	root.AddChild(header)
	for _, it := range items {
		row := NewYogaNode()
		row.Style.Width = 481
		row.Style.Height = 75
		row.Style.Display = "flex"
		row.Style.FlexDirection = "row"
		op := NewYogaNode()
		op.Style.Width = boxDetailTracks[0]
		op.Style.Height = 75
		op.Text = it.Name
		row.AddChild(op)
		lv := NewYogaNode()
		lv.Style.Width = boxDetailTracks[1]
		lv.Style.Height = 75
		row.AddChild(lv)
		pot := NewYogaNode()
		pot.Style.Width = boxDetailTracks[2]
		row.AddChild(pot)
		sk := NewYogaNode()
		sk.Style.Width = boxDetailTracks[3]
		row.AddChild(sk)
		eq := NewYogaNode()
		eq.Style.Width = boxDetailTracks[4]
		row.AddChild(eq)
		root.AddChild(row)
	}
	root.CalculateLayout(481, 186)
	return root
}

func RenderBoxDetail(items []BoxDetailItem, scale float64) *Canvas {
	if scale <= 0 {
		scale = 1.5
	}
	W, H := int(481*scale+0.5), int(186*scale+0.5)
	c := NewCanvas(W, H)
	c.Clear(color.RGBA{0x2e, 0x30, 0x31, 255})
	root := BuildBoxDetail(items)
	ld, _ := NewLoader(findRepoRoot())
	labels := []string{"干员", "等级", "潜能", "技能", "模组"}
	header := root.Children[0]
	for i, cell := range header.Children {
		cx := header.Layout.X + cell.Layout.X
		cy := header.Layout.Y + cell.Layout.Y
		tw := MeasureText(labels[i], 14)
		tx := cx + (boxDetailTracks[i]-float32(tw))/2
		ty := cy + (35-14)/2
		c.DrawText(labels[i], tx, ty, 14, color.RGBA{255, 255, 255, 255}, scale)
	}
	for idx, row := range root.Children[1:] {
		it := items[idx]
		ry := row.Layout.Y
		c.DrawRect(Rect{X: 0, Y: ry * float32(scale), W: 481 * float32(scale), H: 1 * float32(scale)}, Paint{Color: color.RGBA{0x1f, 0x1f, 0x1f, 255}})
		// operator cell
		opX := row.Layout.X + row.Children[0].Layout.X
		if av, err := ld.Load("https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/" + url.PathEscape(it.Id) + ".png"); err == nil {
			_ = c.DrawImageRect(av, Rect{X: (opX + 5) * float32(scale), Y: (ry + (75-50)/2) * float32(scale), W: 50 * float32(scale), H: 50 * float32(scale)})
		}
		nw := MeasureText(it.Name, 12)
		c.DrawText(it.Name, opX+50+8, ry+(75-12)/2, 12, color.RGBA{255, 255, 255, 255}, scale)
		_ = nw
		// level
		lvX := row.Layout.X + row.Children[1].Layout.X
		if ev, err := ld.Load(fmt.Sprintf("assets/box/Evolve_%d.png", it.EvolvePhase)); err == nil {
			_ = c.DrawImageRect(ev, Rect{X: (lvX + (boxDetailTracks[1]-50)/2) * float32(scale), Y: (ry + 10) * float32(scale), W: 50 * float32(scale), H: 30 * float32(scale)})
		}
		c.DrawText(fmt.Sprintf("LV%d", it.Level), lvX+(boxDetailTracks[1]-float32(MeasureText(fmt.Sprintf("LV%d", it.Level), 12)))/2, ry+45, 12, color.RGBA{255, 255, 255, 255}, scale)
		// potential
		ptX := row.Layout.X + row.Children[2].Layout.X
		if pt, err := ld.Load(fmt.Sprintf("assets/box/Potential_%d.png", it.PotentialRank)); err == nil {
			_ = c.DrawImageRect(pt, Rect{X: (ptX + (boxDetailTracks[2]-50)/2) * float32(scale), Y: (ry + (75-50)/2) * float32(scale), W: 50 * float32(scale), H: 50 * float32(scale)})
		}
		// skills
		skX := row.Layout.X + row.Children[3].Layout.X
		for si, sk := range it.Skills {
			sx := skX + float32(si*55)
			if simg, err := ld.Load("https://web.hycdn.cn/arknights/game/assets/char_skill/" + url.PathEscape(sk.Id) + ".png"); err == nil {
				_ = c.DrawImageRect(simg, Rect{X: sx * float32(scale), Y: (ry + 10) * float32(scale), W: 50 * float32(scale), H: 50 * float32(scale)})
			}
			c.DrawText(fmt.Sprintf("LV%d", sk.Level), sx+(50-float32(MeasureText(fmt.Sprintf("LV%d", sk.Level), 10)))/2, ry+60, 10, color.RGBA{255, 255, 255, 255}, scale)
		}
		// equips
		eqX := row.Layout.X + row.Children[4].Layout.X
		for ei, eq := range it.Equips {
			ex := eqX + float32(ei*50)
			if eimg, err := ld.Load("https://web.hycdn.cn/arknights/game/assets/uniequip/type/icon/" + url.PathEscape(eq.Id) + ".png"); err == nil {
				_ = c.DrawImageRect(eimg, Rect{X: ex * float32(scale), Y: (ry + 10) * float32(scale), W: 40 * float32(scale), H: 40 * float32(scale)})
			}
			c.DrawText(fmt.Sprintf("LV%d", eq.Level), ex, ry+50, 10, color.RGBA{255, 255, 255, 255}, scale)
		}
	}
	return c
}
