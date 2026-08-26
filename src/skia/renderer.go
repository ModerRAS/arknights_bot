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

// Recruit — 71 capsule rrect + boxShadow 0 3px 5px gray, via yoga flex, MeasureText, LRU frozen26, sigma 5
type RecruitOperator struct {
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
	Profession string `json:"profession"`
	Rarity     int    `json:"rarity"`
}
type RecruitGroup struct {
	Tags      []string          `json:"tags"`
	Operators []RecruitOperator `json:"operators"`
}

func BuildRecruit(groups []RecruitGroup) *YogaNode {
	root := NewYogaNode()
	root.Style.Width = 900
	root.Style.Height = 356
	root.Style.FlexDirection = "column"
	root.Style.Display = "flex"
	header := NewYogaNode()
	header.Style.Width = 900
	header.Style.Height = 42
	header.Style.Display = "flex"
	header.Style.FlexDirection = "row"
	for i, lab := range []string{"标签", "干员"} {
		n := NewYogaNode()
		txt := lab
		n.SetMeasureFunc(func(w float32, wm MeasureMode, h float32, hm MeasureMode) Size {
			return Size{Width: float32(MeasureText(txt, 18)), Height: 18}
		})
		n.Text = lab
		if i == 0 {
			n.Style.Width = 275
		} else {
			n.Style.Width = 625
		}
		n.Style.Height = 42
		header.AddChild(n)
	}
	root.AddChild(header)
	for _, g := range groups {
		row := NewYogaNode()
		row.Style.Width = 900
		row.Style.Height = 130
		row.Style.Display = "flex"
		row.Style.FlexDirection = "row"
		left := NewYogaNode()
		left.Style.Width = 275
		left.Style.Height = 130
		left.Style.Display = "flex"
		left.Style.AlignItems = "center"
		left.Style.JustifyContent = "center"
		left.Style.Gap = 16
		for _, tag := range g.Tags {
			t := NewYogaNode()
			txt := tag
			w := float32(MeasureText(txt, 18)) + 16
			if w < 30 {
				w = 30
			}
			t.Style.Width = w
			t.Style.Height = 28
			t.Text = txt
			t.SetMeasureFunc(func(w float32, wm MeasureMode, h float32, hm MeasureMode) Size {
				return Size{Width: w, Height: 28}
			})
			left.AddChild(t)
		}
		row.AddChild(left)
		right := NewYogaNode()
		right.Style.Width = 625
		right.Style.Height = 130
		right.Style.Display = "flex"
		right.Style.FlexWrap = "wrap"
		right.Style.Gap = 2
		for _, op := range g.Operators {
			o := NewYogaNode()
			o.Style.Width = 100
			o.Style.Height = 100
			o.Text = op.Avatar
			o.IconPath = op.Avatar
			right.AddChild(o)
		}
		row.AddChild(right)
	}
	root.CalculateLayout(900, 356)
	return root
}

func RenderRecruit(groups []RecruitGroup, scale float64) *Canvas {
	if scale <= 0 {
		scale = 1.5
	}
	if len(groups) == 0 {
		groups = []RecruitGroup{
			{Tags: []string{"先锋干员", "输出"}, Operators: []RecruitOperator{{Name: "阿米娅", Avatar: "https://media.prts.wiki/3/36/%E5%A4%B4%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85.png?image_process=format,webp/quality,Q_90", Profession: "CASTER", Rarity: 4}}},
			{Tags: []string{"近卫", "支援机械"}, Operators: []RecruitOperator{{Name: "能天使", Avatar: "https://media.prts.wiki/a/ad/%E5%A4%B4%E5%83%8F_%E8%83%BD%E5%A4%A9%E4%BD%BF.png?image_process=format,webp/quality,Q_90", Profession: "SNIPER", Rarity: 5}}},
		}
	}
	W, H := int(900*scale+0.5), int(356*scale+0.5)
	c := NewCanvas(W, H)
	c.Clear(color.RGBA{0xff, 0xff, 0xff, 255})
	root := BuildRecruit(groups)
	ld, _ := NewLoader(findRepoRoot())
	header := root.Children[0]
	for i, cell := range header.Children {
		cx := header.Layout.X + cell.Layout.X
		cy := header.Layout.Y + cell.Layout.Y
		lab := []string{"标签", "干员"}[i]
		tw := MeasureText(lab, 18)
		tx := cx + (cell.Layout.Width-float32(tw))/2
		ty := cy + (42-18)/2
		c.DrawText(lab, tx, ty, 18, color.RGBA{0, 0, 0, 255}, scale)
		if i == 0 {
			c.DrawRect(Rect{X: 0, Y: (cy + 41) * float32(scale), W: 900 * float32(scale), H: 1 * float32(scale)}, Paint{Color: color.RGBA{0xaa, 0xaa, 0xaa, 255}})
		}
	}
	for _, row := range root.Children[1:] {
		ry := row.Layout.Y
		left := row.Children[0]
		right := row.Children[1]
		// row border
		c.DrawRect(Rect{X: 0, Y: (ry + 129) * float32(scale), W: 900 * float32(scale), H: 1 * float32(scale)}, Paint{Color: color.RGBA{0xaa, 0xaa, 0xaa, 255}})
		// tags: capsule rrect + shadow sigma 5, gray
		for _, tagNode := range left.Children {
			tx := row.Layout.X + left.Layout.X + tagNode.Layout.X
			ty := row.Layout.Y + left.Layout.Y + tagNode.Layout.Y + (130-tagNode.Layout.Height)/2 - 10
			rr := RRect{Rect: Rect{X: tx * float32(scale), Y: ty * float32(scale), W: tagNode.Layout.Width * float32(scale), H: tagNode.Layout.Height * float32(scale)}, Radius: 14 * float32(scale)}
			c.DrawDropShadow(rr, color.RGBA{0x80, 0x80, 0x80, 255}, 0, 3*float32(scale), 5)
			c.DrawRRect(rr, Paint{Color: color.RGBA{0x31, 0x31, 0x31, 255}})
			tw := MeasureText(tagNode.Text, 18)
			tx2 := tx + (tagNode.Layout.Width-float32(tw))/2
			ty2 := ty + (28-18)/2
			c.DrawText(tagNode.Text, tx2, ty2, 18, color.RGBA{255, 255, 255, 255}, scale)
		}
		// operators: 100x100 avatar + profession 30 + rarity 20
		gIdx := -1
		for i, ch := range root.Children[1:] {
			if ch == row {
				gIdx = i
				break
			}
		}
		var ops []RecruitOperator
		if gIdx >= 0 && gIdx < len(groups) {
			ops = groups[gIdx].Operators
		}
		for oi, opNode := range right.Children {
			if oi >= len(ops) {
				break
			}
			op := ops[oi]
			ox := row.Layout.X + right.Layout.X + opNode.Layout.X
			oy := row.Layout.Y + right.Layout.Y + opNode.Layout.Y
			if img, err := ld.Load(op.Avatar); err == nil {
				_ = c.DrawImageRect(img, Rect{X: ox * float32(scale), Y: oy * float32(scale), W: 100 * float32(scale), H: 100 * float32(scale)})
			} else {
				c.DrawRect(Rect{X: ox * float32(scale), Y: oy * float32(scale), W: 100 * float32(scale), H: 100 * float32(scale)}, Paint{Color: color.RGBA{0x55, 0x55, 0x55, 255}})
			}
			if pimg, err := ld.Load(fmt.Sprintf("assets/box/%s.png", op.Profession)); err == nil {
				_ = c.DrawImageRect(pimg, Rect{X: ox * float32(scale), Y: oy * float32(scale), W: 30 * float32(scale), H: 30 * float32(scale)})
			}
			if rimg, err := ld.Load(fmt.Sprintf("assets/box/Rarity_%d.png", op.Rarity)); err == nil {
				_ = c.DrawImageRect(rimg, Rect{X: (ox + 30) * float32(scale), Y: oy * float32(scale), W: 40 * float32(scale), H: 20 * float32(scale)})
			}
		}
	}
	return c
}

// Missing — 复用 box 72行结构, via yoga flex, MeasureText, LRU frozen26, sigma 5
type MissingChar struct {
	Name       string `json:"name"`
	SkinId     string `json:"skinId"`
	Rarity     int    `json:"rarity"`
	Profession string `json:"profession"`
}
type MissingInfo struct {
	Name  string        `json:"name"`
	Chars []MissingChar `json:"chars"`
}

func BuildMissing(info MissingInfo) *YogaNode {
	root := NewYogaNode()
	root.Style.Width = 700
	root.Style.Height = 357
	root.Style.FlexDirection = "column"
	root.Style.Display = "flex"
	header := NewYogaNode()
	header.Style.Width = 700
	header.Style.Height = 76
	header.Style.Display = "flex"
	txt := "Dr " + info.Name + "(未获取)"
	header.SetMeasureFunc(func(w float32, wm MeasureMode, h float32, hm MeasureMode) Size {
		return Size{Width: float32(MeasureText(txt, 30)), Height: 30}
	})
	header.Text = txt
	root.AddChild(header)
	wrap := NewYogaNode()
	wrap.Style.Width = 700
	wrap.Style.Display = "flex"
	wrap.Style.FlexWrap = "wrap"
	for _, ch := range info.Chars {
		n := NewYogaNode()
		n.Style.Width = 70
		n.Style.Height = 140
		n.Text = ch.Name
		n.IconPath = ch.SkinId
		wrap.AddChild(n)
	}
	root.AddChild(wrap)
	root.CalculateLayout(700, 357)
	return root
}

func RenderMissing(info MissingInfo, scale float64) *Canvas {
	if scale <= 0 {
		scale = 1.5
	}
	if len(info.Chars) == 0 {
		info = MissingInfo{Name: "Test", Chars: []MissingChar{
			{Name: "阿米娅", SkinId: "https://media.prts.wiki/a/a0/%E5%8D%8A%E8%BA%AB%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85_1.png?image_process=format,webp/quality,Q_90", Profession: "CASTER", Rarity: 5},
			{Name: "能天使", SkinId: "https://media.prts.wiki/a/ad/%E5%A4%B4%E5%83%8F_%E8%83%BD%E5%A4%A9%E4%BD%BF.png?image_process=format,webp/quality,Q_90", Profession: "SNIPER", Rarity: 5},
		}}
	}
	W, H := int(700*scale+0.5), int(357*scale+0.5)
	c := NewCanvas(W, H)
	c.Clear(color.RGBA{0x2e, 0x30, 0x31, 255})
	root := BuildMissing(info)
	ld, _ := NewLoader(findRepoRoot())
	// header with label bg
	if label, err := ld.Load("assets/help/label.png"); err == nil {
		_ = c.DrawImageRect(label, Rect{X: 0, Y: 0, W: 700 * float32(scale), H: 76 * float32(scale)})
	}
	hdr := root.Children[0]
	title := "Dr " + info.Name + "(未获取)"
	c.DrawText(title, hdr.Layout.X+25, hdr.Layout.Y+(76-30)/2, 30, color.RGBA{255, 255, 255, 255}, scale)
	_ = MeasureText(title, 30)
	wrap := root.Children[1]
	for idx, node := range wrap.Children {
		if idx >= len(info.Chars) {
			break
		}
		ch := info.Chars[idx]
		nx := wrap.Layout.X + node.Layout.X
		ny := wrap.Layout.Y + node.Layout.Y
		// card shadow like box: sigma 5
		rr := RRect{Rect: Rect{X: nx * float32(scale), Y: ny * float32(scale), W: 70 * float32(scale), H: 140 * float32(scale)}, Radius: 0}
		c.DrawDropShadow(rr, color.RGBA{0, 0, 0, 100}, 0, 3*float32(scale), 5)
		c.DrawRect(Rect{X: nx * float32(scale), Y: ny * float32(scale), W: 70 * float32(scale), H: 140 * float32(scale)}, Paint{Color: color.RGBA{0x2e, 0x30, 0x31, 255}})
		if img, err := ld.Load(ch.SkinId); err == nil {
			_ = c.DrawImageRect(img, Rect{X: nx * float32(scale), Y: ny * float32(scale), W: 70 * float32(scale), H: 140 * float32(scale)})
		} else {
			c.DrawRect(Rect{X: nx * float32(scale), Y: ny * float32(scale), W: 70 * float32(scale), H: 140 * float32(scale)}, Paint{Color: color.RGBA{0x55, 0x55, 0x55, 255}})
		}
		if pimg, err := ld.Load(fmt.Sprintf("assets/box/%s.png", ch.Profession)); err == nil {
			_ = c.DrawImageRect(pimg, Rect{X: (nx + 3) * float32(scale), Y: (ny + 5) * float32(scale), W: 15 * float32(scale), H: 15 * float32(scale)})
		}
		if rimg, err := ld.Load(fmt.Sprintf("assets/box/Rarity_%d.png", ch.Rarity)); err == nil {
			_ = c.DrawImageRect(rimg, Rect{X: (nx + 20) * float32(scale), Y: (ny + 5) * float32(scale), W: 40 * float32(scale), H: 15 * float32(scale)})
		}
		// name bar at bottom -3 height 14 bg rgba0,0,0,.7
		barY := ny + 140 - 14
		c.DrawRect(Rect{X: nx * float32(scale), Y: barY * float32(scale), W: 70 * float32(scale), H: 14 * float32(scale)}, Paint{Color: color.RGBA{0, 0, 0, 179}})
		tw := MeasureText(ch.Name, 10)
		tx := nx + (70-float32(tw))/2
		ty := barY + (14-10)/2
		c.DrawText(ch.Name, tx, ty, 10, color.RGBA{255, 255, 255, 255}, scale)
	}
	return c
}

// Box — 70x140 单图 + 进度条 3px + 稀有度色条, via yoga flexWrap + MeasureText + DrawRect/DrawImageRect
type BoxChar struct {
	CharId        string `json:"charId"`
	SkinId        string `json:"skinId"`
	Name          string `json:"name"`
	Level         int    `json:"level"`
	EvolvePhase   int    `json:"evolvePhase"`
	PotentialRank int    `json:"potentialRank"`
	FavorPercent  int    `json:"favorPercent"`
	Rarity        int    `json:"rarity"`
	Profession    string `json:"profession"`
}
type BoxInfo struct {
	Name  string    `json:"name"`
	Chars []BoxChar `json:"chars"`
}

func boxRarityColor(r int) color.RGBA {
	switch r {
	case 6:
		return color.RGBA{240, 180, 40, 255}
	case 5:
		return color.RGBA{170, 110, 220, 255}
	case 4:
		return color.RGBA{90, 160, 230, 255}
	case 3:
		return color.RGBA{150, 150, 150, 255}
	default:
		return color.RGBA{150, 150, 150, 255}
	}
}

func BuildBox(info BoxInfo) *YogaNode {
	root := NewYogaNode()
	root.Style.Width = 700
	root.Style.Height = 357
	root.Style.FlexDirection = "column"
	root.Style.Display = "flex"
	header := NewYogaNode()
	header.Style.Width = 700
	header.Style.Height = 76
	header.Style.Display = "flex"
	txt := "Dr " + info.Name
	header.SetMeasureFunc(func(w float32, wm MeasureMode, h float32, hm MeasureMode) Size {
		return Size{Width: float32(MeasureText(txt, 30)), Height: 30}
	})
	header.Text = txt
	root.AddChild(header)
	wrap := NewYogaNode()
	wrap.Style.Width = 700
	wrap.Style.Display = "flex"
	wrap.Style.FlexWrap = "wrap"
	wrap.Style.Gap = 0
	for _, ch := range info.Chars {
		n := NewYogaNode()
		n.Style.Width = 70
		n.Style.Height = 140
		n.Text = ch.Name
		n.IconPath = ch.SkinId
		wrap.AddChild(n)
	}
	root.AddChild(wrap)
	root.CalculateLayout(700, 357)
	return root
}

func RenderBox(info BoxInfo, scale float64) *Canvas {
	if scale <= 0 {
		scale = 1.5
	}
	if len(info.Chars) == 0 {
		info = BoxInfo{Name: "Test", Chars: []BoxChar{
			{Name: "阿米娅", SkinId: "char_002_amiya", Profession: "CASTER", Rarity: 5, Level: 90, EvolvePhase: 2, PotentialRank: 5, FavorPercent: 100},
			{Name: "能天使", SkinId: "char_103_angel", Profession: "SNIPER", Rarity: 5, Level: 80, EvolvePhase: 1, PotentialRank: 3, FavorPercent: 50},
		}}
	}
	W, H := int(700*scale+0.5), int(357*scale+0.5)
	c := NewCanvas(W, H)
	c.Clear(color.RGBA{0x2e, 0x30, 0x31, 255})
	root := BuildBox(info)
	ld, _ := NewLoader(findRepoRoot())
	if label, err := ld.Load("assets/help/label.png"); err == nil {
		_ = c.DrawImageRect(label, Rect{X: 0, Y: 0, W: 700 * float32(scale), H: 76 * float32(scale)})
	}
	hdr := root.Children[0]
	title := "Dr " + info.Name
	c.DrawText(title, hdr.Layout.X+25, hdr.Layout.Y+(76-30)/2, 30, color.RGBA{255, 255, 255, 255}, scale)
	_ = MeasureText(title, 30)
	wrap := root.Children[1]
	for idx, node := range wrap.Children {
		if idx >= len(info.Chars) {
			break
		}
		ch := info.Chars[idx]
		nx := wrap.Layout.X + node.Layout.X
		ny := wrap.Layout.Y + node.Layout.Y
		rr := RRect{Rect: Rect{X: nx * float32(scale), Y: ny * float32(scale), W: 70 * float32(scale), H: 140 * float32(scale)}, Radius: 0}
		c.DrawDropShadow(rr, color.RGBA{0, 0, 0, 100}, 0, 3*float32(scale), 5)
		c.DrawRect(Rect{X: nx * float32(scale), Y: ny * float32(scale), W: 70 * float32(scale), H: 140 * float32(scale)}, Paint{Color: color.RGBA{0x2e, 0x30, 0x31, 255}})
		portraitURL := "https://web.hycdn.cn/arknights/game/assets/char_skin/portrait/" + url.PathEscape(ch.SkinId) + ".png"
		if img, err := ld.Load(portraitURL); err == nil {
			_ = c.DrawImageRect(img, Rect{X: nx * float32(scale), Y: ny * float32(scale), W: 70 * float32(scale), H: 140 * float32(scale)})
		} else if fb, err := ld.Load("assets/common/amiya.png"); err == nil {
			_ = c.DrawImageRect(fb, Rect{X: nx * float32(scale), Y: ny * float32(scale), W: 70 * float32(scale), H: 140 * float32(scale)})
		} else {
			c.DrawRect(Rect{X: nx * float32(scale), Y: ny * float32(scale), W: 70 * float32(scale), H: 140 * float32(scale)}, Paint{Color: color.RGBA{0x55, 0x55, 0x55, 255}})
		}
		if pimg, err := ld.Load(fmt.Sprintf("assets/box/%s.png", ch.Profession)); err == nil {
			_ = c.DrawImageRect(pimg, Rect{X: (nx + 3) * float32(scale), Y: (ny + 5) * float32(scale), W: 15 * float32(scale), H: 15 * float32(scale)})
		}
		if rimg, err := ld.Load(fmt.Sprintf("assets/box/Rarity_%d.png", ch.Rarity)); err == nil {
			_ = c.DrawImageRect(rimg, Rect{X: (nx + 20) * float32(scale), Y: (ny + 5) * float32(scale), W: 40 * float32(scale), H: 15 * float32(scale)})
		}
		if ev, err := ld.Load(fmt.Sprintf("assets/box/Evolve_%d.png", ch.EvolvePhase)); err == nil {
			_ = c.DrawImageRect(ev, Rect{X: (nx + 15) * float32(scale), Y: (ny + 80) * float32(scale), W: 40 * float32(scale), H: 15 * float32(scale)})
		}
		if pt, err := ld.Load(fmt.Sprintf("assets/box/Potential_%d.png", ch.PotentialRank)); err == nil {
			_ = c.DrawImageRect(pt, Rect{X: (nx + 40) * float32(scale), Y: (ny + 100) * float32(scale), W: 30 * float32(scale), H: 30 * float32(scale)})
		}
		rc := boxRarityColor(ch.Rarity)
		c.DrawRect(Rect{X: nx * float32(scale), Y: (ny + 140 - 14 - 3) * float32(scale), W: 70 * float32(scale), H: 3 * float32(scale)}, Paint{Color: rc})
		c.DrawRect(Rect{X: nx * float32(scale), Y: (ny + 140 - 14 - 6) * float32(scale), W: 70 * float32(scale), H: 3 * float32(scale)}, Paint{Color: color.RGBA{0x33, 0x33, 0x33, 255}})
		frac := float32(ch.FavorPercent) / 100
		if frac < 0 {
			frac = 0
		}
		if frac > 1 {
			frac = 1
		}
		c.DrawRect(Rect{X: nx * float32(scale), Y: (ny + 140 - 14 - 6) * float32(scale), W: 70 * frac * float32(scale), H: 3 * float32(scale)}, Paint{Color: color.RGBA{0x54, 0x70, 0xc6, 255}})
		levelCX := (nx + 12) * float32(scale)
		levelCY := (ny + 140 - 12) * float32(scale)
		c.DrawCircle(Circle{CX: levelCX, CY: levelCY, R: 10 * float32(scale)}, Paint{Color: color.RGBA{0, 0, 0, 179}})
		lvStr := fmt.Sprintf("%d", ch.Level)
		tw := MeasureText(lvStr, 11)
		tx := nx + 12 - float32(tw)/2
		ty := ny + 140 - 12 - 5 - 5
		c.DrawText(lvStr, tx, ty, 11, color.RGBA{255, 255, 255, 255}, scale)
		barY := ny + 140 - 14
		c.DrawRect(Rect{X: nx * float32(scale), Y: barY * float32(scale), W: 70 * float32(scale), H: 14 * float32(scale)}, Paint{Color: color.RGBA{0, 0, 0, 179}})
		tw2 := MeasureText(ch.Name, 10)
		tx2 := nx + (70-float32(tw2))/2
		ty2 := barY + (14-10)/2
		c.DrawText(ch.Name, tx2, ty2, 10, color.RGBA{255, 255, 255, 255}, scale)
	}
	return c
}

// BoxSummary — 8x4 矩阵 900x482(1350x723@1.5), border-spacing 20x5 38px 行距, via yoga column/row + MeasureText + DrawRect
type BoxSummary struct {
	Name                   string        `json:"name"`
	AllCharCnt             string        `json:"allCharCnt"`
	AllEvolvePhase2Cnt     int           `json:"allEvolvePhase2Cnt"`
	AllSkill10Cnt          int           `json:"allSkill10Cnt"`
	AllSkill9Cnt           int           `json:"allSkill9Cnt"`
	AllSkill8Cnt           int           `json:"allSkill8Cnt"`
	AllEquipStage3Cnt      int           `json:"allEquipStage3Cnt"`
	AllEquipStage2Cnt      int           `json:"allEquipStage2Cnt"`
	AllEquipStage1Cnt      int           `json:"allEquipStage1Cnt"`
	Star6CharCnt           string        `json:"star6CharCnt"`
	Star6EvolvePhase2Cnt   int           `json:"star6EvolvePhase2Cnt"`
	Star6Skill10Cnt        int           `json:"star6Skill10Cnt"`
	Star6Skill9Cnt         int           `json:"star6Skill9Cnt"`
	Star6Skill8Cnt         int           `json:"star6Skill8Cnt"`
	Star6EquipStage3Cnt    int           `json:"star6EquipStage3Cnt"`
	Star6EquipStage2Cnt    int           `json:"star6EquipStage2Cnt"`
	Star6EquipStage1Cnt    int           `json:"star6EquipStage1Cnt"`
	Star5CharCnt           string        `json:"star5CharCnt"`
	Star5EvolvePhase2Cnt   int           `json:"star5EvolvePhase2Cnt"`
	Star5Skill10Cnt        int           `json:"star5Skill10Cnt"`
	Star5Skill9Cnt         int           `json:"star5Skill9Cnt"`
	Star5Skill8Cnt         int           `json:"star5Skill8Cnt"`
	Star5EquipStage3Cnt    int           `json:"star5EquipStage3Cnt"`
	Star5EquipStage2Cnt    int           `json:"star5EquipStage2Cnt"`
	Star5EquipStage1Cnt    int           `json:"star5EquipStage1Cnt"`
	Star4CharCnt           string        `json:"star4CharCnt"`
	Star4EvolvePhase2Cnt   int           `json:"star4EvolvePhase2Cnt"`
	Star4Skill10Cnt        int           `json:"star4Skill10Cnt"`
	Star4Skill9Cnt         int           `json:"star4Skill9Cnt"`
	Star4Skill8Cnt         int           `json:"star4Skill8Cnt"`
	Star4EquipStage3Cnt    int           `json:"star4EquipStage3Cnt"`
	Star4EquipStage2Cnt    int           `json:"star4EquipStage2Cnt"`
	Star4EquipStage1Cnt    int           `json:"star4EquipStage1Cnt"`
	MissingChars           []MissingChar `json:"missingChars"`
}

func BuildBoxSummary(s BoxSummary) *YogaNode {
	root := NewYogaNode()
	root.Style.Width = 900
	root.Style.Height = 482
	root.Style.FlexDirection = "column"
	root.Style.Display = "flex"
	header := NewYogaNode()
	header.Style.Width = 900
	header.Style.Height = 60
	header.Style.Display = "flex"
	txt := "Dr " + s.Name
	header.SetMeasureFunc(func(w float32, wm MeasureMode, h float32, hm MeasureMode) Size {
		return Size{Width: float32(MeasureText(txt, 30)), Height: 30}
	})
	header.Text = txt
	root.AddChild(header)
	table := NewYogaNode()
	table.Style.Width = 900
	table.Style.Display = "flex"
	table.Style.FlexDirection = "column"
	table.Style.Gap = 0
	hdrRow := NewYogaNode()
	hdrRow.Style.Width = 900
	hdrRow.Style.Height = 30
	hdrRow.Style.Display = "flex"
	hdrRow.Style.FlexDirection = "row"
	hdrRow.Style.Gap = 20
	for _, lab := range []string{"全部干员", "六星干员", "五星干员", "四星干员"} {
		n := NewYogaNode()
		n.Style.Width = 210
		n.Style.Height = 30
		txt2 := lab
		n.SetMeasureFunc(func(w float32, wm MeasureMode, h float32, hm MeasureMode) Size {
			return Size{Width: float32(MeasureText(txt2, 12)), Height: 12}
		})
		n.Text = lab
		hdrRow.AddChild(n)
	}
	table.AddChild(hdrRow)
	metrics := []struct {
		label string
		vals  [4]string
	}{
		{"招募干员数量", [4]string{s.AllCharCnt, s.Star6CharCnt, s.Star5CharCnt, s.Star4CharCnt}},
		{"精英阶段2干员", [4]string{fmt.Sprintf("%d", s.AllEvolvePhase2Cnt), fmt.Sprintf("%d", s.Star6EvolvePhase2Cnt), fmt.Sprintf("%d", s.Star5EvolvePhase2Cnt), fmt.Sprintf("%d", s.Star4EvolvePhase2Cnt)}},
		{"专精三技能数量", [4]string{fmt.Sprintf("%d", s.AllSkill10Cnt), fmt.Sprintf("%d", s.Star6Skill10Cnt), fmt.Sprintf("%d", s.Star5Skill10Cnt), fmt.Sprintf("%d", s.Star4Skill10Cnt)}},
		{"专精二技能数量", [4]string{fmt.Sprintf("%d", s.AllSkill9Cnt), fmt.Sprintf("%d", s.Star6Skill9Cnt), fmt.Sprintf("%d", s.Star5Skill9Cnt), fmt.Sprintf("%d", s.Star4Skill9Cnt)}},
		{"专精一技能数量", [4]string{fmt.Sprintf("%d", s.AllSkill8Cnt), fmt.Sprintf("%d", s.Star6Skill8Cnt), fmt.Sprintf("%d", s.Star5Skill8Cnt), fmt.Sprintf("%d", s.Star4Skill8Cnt)}},
		{"三级模组数量", [4]string{fmt.Sprintf("%d", s.AllEquipStage3Cnt), fmt.Sprintf("%d", s.Star6EquipStage3Cnt), fmt.Sprintf("%d", s.Star5EquipStage3Cnt), fmt.Sprintf("%d", s.Star4EquipStage3Cnt)}},
		{"二级模组数量", [4]string{fmt.Sprintf("%d", s.AllEquipStage2Cnt), fmt.Sprintf("%d", s.Star6EquipStage2Cnt), fmt.Sprintf("%d", s.Star5EquipStage2Cnt), fmt.Sprintf("%d", s.Star4EquipStage2Cnt)}},
		{"一级模组数量", [4]string{fmt.Sprintf("%d", s.AllEquipStage1Cnt), fmt.Sprintf("%d", s.Star6EquipStage1Cnt), fmt.Sprintf("%d", s.Star5EquipStage1Cnt), fmt.Sprintf("%d", s.Star4EquipStage1Cnt)}},
	}
	for _, m := range metrics {
		row := NewYogaNode()
		row.Style.Width = 900
		row.Style.Height = 38
		row.Style.Display = "flex"
		row.Style.FlexDirection = "row"
		row.Style.Gap = 20
		for j := 0; j < 4; j++ {
			cell := NewYogaNode()
			cell.Style.Width = 210
			cell.Style.Height = 38
			txt3 := m.vals[j]
			cell.SetMeasureFunc(func(w float32, wm MeasureMode, h float32, hm MeasureMode) Size {
				return Size{Width: float32(MeasureText(txt3, 12)), Height: 12}
			})
			cell.Text = txt3
			row.AddChild(cell)
		}
		// attach label for rendering via row index (stored in first cell text is count, but we need label for left side)
		_ = m.label
			table.AddChild(row)
	}
	root.AddChild(table)
	if len(s.MissingChars) > 0 {
		titleN := NewYogaNode()
		titleN.Style.Width = 900
		titleN.Style.Height = 22
		txt4 := "未招募干员"
		titleN.SetMeasureFunc(func(w float32, wm MeasureMode, h float32, hm MeasureMode) Size {
			return Size{Width: float32(MeasureText(txt4, 14)), Height: 14}
		})
		titleN.Text = txt4
		root.AddChild(titleN)
		wrap := NewYogaNode()
		wrap.Style.Width = 900
		wrap.Style.Display = "flex"
		wrap.Style.FlexWrap = "wrap"
		wrap.Style.Gap = 2
		for _, ch := range s.MissingChars {
			n := NewYogaNode()
			n.Style.Width = 40
			n.Style.Height = 40
			n.Text = ch.Name
			n.IconPath = ch.SkinId
			wrap.AddChild(n)
		}
		root.AddChild(wrap)
	}
	root.CalculateLayout(900, 482)
	return root
}

func RenderBoxSummary(s BoxSummary, scale float64) *Canvas {
	if scale <= 0 {
		scale = 1.5
	}
	if s.Name == "" {
		s.Name = "Test"
	}
	if s.AllCharCnt == "" {
		s.AllCharCnt = "10/20"
		s.Star6CharCnt = "5/10"
		s.Star5CharCnt = "3/5"
		s.Star4CharCnt = "2/5"
	}
	W, H := int(900*scale+0.5), int(482*scale+0.5)
	c := NewCanvas(W, H)
	c.Clear(color.RGBA{0x2e, 0x30, 0x31, 255})
	root := BuildBoxSummary(s)
	ld, _ := NewLoader(findRepoRoot())
	if label, err := ld.Load("assets/help/label.png"); err == nil {
		_ = c.DrawImageRect(label, Rect{X: 0, Y: 0, W: 900 * float32(scale), H: 60 * float32(scale)})
	}
	hdr := root.Children[0]
	title := "Dr " + s.Name
	c.DrawText(title, hdr.Layout.X+25, hdr.Layout.Y+(60-30)/2, 30, color.RGBA{255, 255, 255, 255}, scale)
	_ = MeasureText(title, 30)
	table := root.Children[1]
	hdrRow := table.Children[0]
	labels := []string{"全部干员", "六星干员", "五星干员", "四星干员"}
	for i, cell := range hdrRow.Children {
		cx := hdrRow.Layout.X + cell.Layout.X
		cy := hdrRow.Layout.Y + cell.Layout.Y
		tw := MeasureText(labels[i], 12)
		tx := cx + (210-float32(tw))/2
		ty := cy + (30-12)/2
		c.DrawText(labels[i], tx, ty, 12, color.RGBA{255, 255, 255, 255}, scale)
	}
	metricLabels := []string{"招募干员数量", "精英阶段2干员", "专精三技能数量", "专精二技能数量", "专精一技能数量", "三级模组数量", "二级模组数量", "一级模组数量"}
	for ri, row := range table.Children[1:] {
		ry := row.Layout.Y
		if ri%2 == 0 {
			c.DrawRect(Rect{X: 15 * float32(scale), Y: ry * float32(scale), W: float32(900-30) * float32(scale), H: 38 * float32(scale)}, Paint{Color: color.RGBA{255, 255, 255, 10}})
		}
		c.DrawRect(Rect{X: 15 * float32(scale), Y: (ry + 37) * float32(scale), W: float32(900-30) * float32(scale), H: 1 * float32(scale)}, Paint{Color: color.RGBA{0x55, 0x55, 0x55, 255}})
		vals := []string{}
		switch ri {
		case 0:
			vals = []string{s.AllCharCnt, s.Star6CharCnt, s.Star5CharCnt, s.Star4CharCnt}
		case 1:
			vals = []string{fmt.Sprintf("%d", s.AllEvolvePhase2Cnt), fmt.Sprintf("%d", s.Star6EvolvePhase2Cnt), fmt.Sprintf("%d", s.Star5EvolvePhase2Cnt), fmt.Sprintf("%d", s.Star4EvolvePhase2Cnt)}
		case 2:
			vals = []string{fmt.Sprintf("%d", s.AllSkill10Cnt), fmt.Sprintf("%d", s.Star6Skill10Cnt), fmt.Sprintf("%d", s.Star5Skill10Cnt), fmt.Sprintf("%d", s.Star4Skill10Cnt)}
		case 3:
			vals = []string{fmt.Sprintf("%d", s.AllSkill9Cnt), fmt.Sprintf("%d", s.Star6Skill9Cnt), fmt.Sprintf("%d", s.Star5Skill9Cnt), fmt.Sprintf("%d", s.Star4Skill9Cnt)}
		case 4:
			vals = []string{fmt.Sprintf("%d", s.AllSkill8Cnt), fmt.Sprintf("%d", s.Star6Skill8Cnt), fmt.Sprintf("%d", s.Star5Skill8Cnt), fmt.Sprintf("%d", s.Star4Skill8Cnt)}
		case 5:
			vals = []string{fmt.Sprintf("%d", s.AllEquipStage3Cnt), fmt.Sprintf("%d", s.Star6EquipStage3Cnt), fmt.Sprintf("%d", s.Star5EquipStage3Cnt), fmt.Sprintf("%d", s.Star4EquipStage3Cnt)}
		case 6:
			vals = []string{fmt.Sprintf("%d", s.AllEquipStage2Cnt), fmt.Sprintf("%d", s.Star6EquipStage2Cnt), fmt.Sprintf("%d", s.Star5EquipStage2Cnt), fmt.Sprintf("%d", s.Star4EquipStage2Cnt)}
		case 7:
			vals = []string{fmt.Sprintf("%d", s.AllEquipStage1Cnt), fmt.Sprintf("%d", s.Star6EquipStage1Cnt), fmt.Sprintf("%d", s.Star5EquipStage1Cnt), fmt.Sprintf("%d", s.Star4EquipStage1Cnt)}
		}
		for ci, cell := range row.Children {
			cx := row.Layout.X + cell.Layout.X
			cy := row.Layout.Y + cell.Layout.Y
			lab := metricLabels[ri]
			twLab := MeasureText(lab, 12)
			c.DrawText(lab, cx+10, cy+(38-12)/2, 12, color.RGBA{200, 200, 200, 255}, scale)
			_ = twLab
			twVal := MeasureText(vals[ci], 12)
			txVal := cx + 210 - float32(twVal) - 10
			tyVal := cy + (38-12)/2
			col := color.RGBA{230, 230, 230, 255}
			if ci == 1 {
				col = color.RGBA{240, 180, 40, 255}
			} else if ci == 2 {
				col = color.RGBA{170, 110, 220, 255}
			} else if ci == 3 {
				col = color.RGBA{90, 160, 230, 255}
			}
			c.DrawText(vals[ci], txVal, tyVal, 12, col, scale)
		}
	}
	if len(s.MissingChars) > 0 && len(root.Children) > 3 {
		titleN := root.Children[2]
		c.DrawText("未招募干员", titleN.Layout.X+15, titleN.Layout.Y+5, 14, color.RGBA{255, 255, 255, 255}, scale)
		wrap := root.Children[3]
		for idx, node := range wrap.Children {
			if idx >= len(s.MissingChars) {
				break
			}
			ch := s.MissingChars[idx]
			nx := wrap.Layout.X + node.Layout.X
			ny := wrap.Layout.Y + node.Layout.Y
			if img, err := ld.Load(ch.SkinId); err == nil {
				_ = c.DrawImageRect(img, Rect{X: nx * float32(scale), Y: ny * float32(scale), W: 40 * float32(scale), H: 40 * float32(scale)})
			} else {
				c.DrawRect(Rect{X: nx * float32(scale), Y: ny * float32(scale), W: 40 * float32(scale), H: 40 * float32(scale)}, Paint{Color: color.RGBA{0x55, 0x55, 0x55, 255}})
			}
		}
	}
	return c
}

