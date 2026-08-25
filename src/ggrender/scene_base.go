package ggrender

import (
	"fmt"

	"github.com/fogleman/gg"
)

// Base scene: manifest 1665x918 px = 1110x612 CSS @1.5x (template/Base.tmpl).
// Layout: header + stacked #21262f rounded cards on #2b333d; two half-width
// cards per row from row 3. Avatars from assets/base (production CDN cache).

type BaseChar struct {
	Name   string
	Avatar string // asset rel path under assets/
	AP     int    // 0..100 mood
}

type BaseRoom struct {
	Title     string // "控制中枢 Lv.5"
	Note      string // right-side colored note text
	NoteColor [3]int
	Chars     []BaseChar
	Board     []int  // 会客室 clue boxes
	SkillLv   int    // 训练室 Lv.10
	SkillIcon string // asset rel path
	Half      bool   // half-width card (two per row)
}

type BaseInfo struct {
	LaborCur, LaborTotal int
	Rooms                []BaseRoom
}

func moodColor(ap int) (int, int, int) {
	switch {
	case ap <= 0:
		return 0xe4, 0x2e, 0x20
	case ap < 100:
		return 0xf0, 0xab, 0x22
	default:
		return 0x3c, 0xd6, 0x27
	}
}

func drawMoodIcon(dc *gg.Context, x, y float64, ap int) {
	r, g, b := moodColor(ap)
	dc.SetRGB255(r, g, b)
	dc.SetLineWidth(2)
	dc.DrawCircle(x+9, y+9, 8)
	dc.Stroke()
	// eyes
	dc.DrawCircle(x+5.5, y+7, 1.2)
	dc.Fill()
	dc.DrawCircle(x+12.5, y+7, 1.2)
	dc.Fill()
	// mouth: frown when tired, smile when ok
	if ap <= 0 {
		dc.DrawArc(x+9, y+14, 3, 1.15*3.14159, 1.85*3.14159)
		dc.Stroke()
	} else if ap < 100 {
		dc.MoveTo(x+6, y+12.5)
		dc.LineTo(x+12, y+12.5)
		dc.Stroke()
	} else {
		dc.DrawArc(x+9, y+10, 3.5, 0.15*3.14159, 0.85*3.14159)
		dc.Stroke()
	}
}

func drawAPBar(dc *gg.Context, x, y, w float64, ap int) {
	dc.SetRGBA255(255, 255, 255, 90)
	dc.DrawRectangle(x, y, w, 4)
	dc.Fill()
	dc.SetRGB255(255, 255, 255)
	dc.DrawRectangle(x, y, w*float64(ap)/100, 4)
	dc.Fill()
}

func drawBaseChar(dc *gg.Context, c BaseChar, x, y float64) {
	dc.DrawImage(ScaleExact(tryLocal(c.Avatar), 40, 40), int(x), int(y))
	drawMoodIcon(dc, x+44, y+10, c.AP)
	setFont(dc, 15)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, c.Name, x+68, y+25)
	drawAPBar(dc, x, y+44, 147, c.AP)
}

func SampleBase() *BaseInfo {
	return &BaseInfo{
		LaborCur: 42, LaborTotal: 99,
		Rooms: []BaseRoom{
			{Title: "控制中枢 Lv.5", Chars: []BaseChar{{"阿米娅", "base/avatar-amiya.png", 0}}},
			{Title: "宿舍 Lv.5", Note: "舒适度20000", NoteColor: [3]int{0x66, 0xc0, 0x2f}, Chars: []BaseChar{{"能天使", "base/avatar-angel.png", 0}}},
			{Title: "贸易站 Lv.3", Note: "贵金属订单 3/3", NoteColor: [3]int{0x8c, 0xd1, 0xff}, Chars: []BaseChar{{"德克萨斯", "base/avatar-texas.png", 100}}, Half: true},
			{Title: "制造站 Lv.3", Note: "赤金 1/3", NoteColor: [3]int{0xd7, 0x9d, 0x13}, Chars: []BaseChar{{"推进之王", "base/avatar-siege.png", 0}}, Half: true},
			{Title: "发电站 Lv.3", Note: "270", NoteColor: [3]int{0xad, 0xfe, 0x2e}, Chars: []BaseChar{{"塞雷娅", "base/avatar-saria.png", 50}}, Half: true},
			{Title: "会客室 Lv.3", Note: "线索交流开启中", NoteColor: [3]int{0xee, 0x81, 0x0a}, Board: []int{1, 7}, Chars: []BaseChar{{"凯尔希", "base/avatar-kalts.png", 100}}, Half: true},
			{Title: "办公室 Lv.3", Note: "刷新次数3", NoteColor: [3]int{255, 255, 255}, Chars: []BaseChar{{"银灰", "base/avatar-silverash.png", 0}}, Half: true},
			{Title: "训练室 Lv.3", Note: "Lv.10", NoteColor: [3]int{255, 255, 255}, SkillIcon: "base/skill-amiya.png", Chars: []BaseChar{{"艾雅法拉", "base/avatar-eyjafjalla.png", 0}}, Half: true},
		},
	}
}

func RenderBase(data *BaseInfo) (*gg.Context, error) {
	const cssW, cssH = 1110, 612
	dc := gg.NewContext(1665, 918) // manifest pixels
	dc.Scale(1.5, 1.5)
	FillBackground(dc, 0x2b, 0x33, 0x3d)
	// header
	setFont(dc, 17)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, "基建信息", 10, 23)
	// labor right: purple glyph + 42/99 + bar
	lx := 962.0
	dc.SetRGB255(0x85, 0x2c, 0xd3)
	dc.DrawRectangle(lx, 8, 4, 4)
	dc.Fill()
	dc.DrawRectangle(lx+14, 8, 4, 4)
	dc.Fill()
	dc.DrawRectangle(lx+7, 15, 4, 4)
	dc.Fill()
	setFont(dc, 15)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, fmt.Sprintf("%d/%d", data.LaborCur, data.LaborTotal), lx+24, 21)
	drawAPBar(dc, 980, 26, 100, data.LaborCur*100/data.LaborTotal)

	// cards
	cardY := func(i int) float64 { return 45 + float64(i)*114 }
	halfIdx := 0
	row := 0
	for _, r := range data.Rooms {
		var x, w float64
		var idx int
		if !r.Half {
			x, w = 2.5, 1102
			idx = row
			row++
		} else {
			x, w = 2.5, 548
			if halfIdx%2 == 1 {
				x = 560
				w = 545
			}
			idx = row
			if halfIdx%2 == 1 {
				row++
			}
			halfIdx++
		}
		y := cardY(idx)
		dc.SetRGB255(0x21, 0x26, 0x2f)
		dc.DrawRoundedRectangle(x, y, w, 110, 15)
		dc.Fill()
		// title
		setFont(dc, 17)
		dc.SetRGB255(255, 255, 255)
		drawString(dc, r.Title, x+10, y+24)
		// note right
		if r.Note != "" {
			setFont(dc, 15)
			dc.SetRGB255(r.NoteColor[0], r.NoteColor[1], r.NoteColor[2])
			nw, _ := dc.MeasureString(r.Note)
			drawString(dc, r.Note, x+w-20-nw, y+23)
		}
		// board boxes
		if len(r.Board) > 0 {
			bx := x + w - 20
			for i := len(r.Board) - 1; i >= 0; i-- {
				setFont(dc, 15)
				dc.SetRGB255(255, 255, 255)
				dc.DrawRectangle(bx-22, y+8, 22, 24)
				dc.Stroke()
				drawStringAnchored(dc, fmt.Sprintf("%d", r.Board[i]), bx-11, y+25, 0.5, 0.5)
				bx -= 26
			}
			setFont(dc, 15)
			dc.SetRGB255(255, 255, 255)
			drawString(dc, "线索", bx-40, y+23)
		}
		// skill lv + icon
		if r.SkillIcon != "" {
			setFont(dc, 15)
			dc.SetRGB255(255, 255, 255)
			drawString(dc, r.Note, x+w-70, y+23)
			dc.DrawImage(ScaleExact(tryLocal(r.SkillIcon), 30, 30), int(x+w-50), int(y+6))
		}
		// chars
		for j, c := range r.Chars {
			drawBaseChar(dc, c, x+10, y+48)
			_ = j
		}
	}
	return dc, nil
}
