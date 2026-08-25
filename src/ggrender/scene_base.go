package ggrender

import (
	"fmt"
	"image"

	xdraw "golang.org/x/image/draw"

	"github.com/fogleman/gg"
)

// drawImageReal draws a real-resolution image at real coords under the CSS scale transform.
func drawImageReal(dc *gg.Context, img image.Image, realX, realY float64) {
	dc.Push()
	dc.Scale(1/1.5, 1/1.5)
	dc.DrawImage(img, int(realX), int(realY))
	dc.Pop()
}

// scaleSmooth bilinear-downscale (Chrome-like photo scaling, unlike nearest ScaleExact).
func scaleSmooth(img image.Image, w, h int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.BiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), xdraw.Over, nil)
	return dst
}

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

func drawStringBold(dc *gg.Context, s string, x, y float64) {
	// synthetic bold (Chrome fake-bold for CJK): overlap 4 offset fills
	drawString(dc, s, x, y)
	drawString(dc, s, x+1.0, y)
	drawString(dc, s, x, y+0.6)
	drawString(dc, s, x+1.0, y+0.6)
}

// drawNoteIcon approximates the svg glyph left of a room note.
func drawNoteIcon(dc *gg.Context, x, y, s float64, c [3]int) {
	dc.SetRGB255(c[0], c[1], c[2])
	dc.SetLineWidth(2)
	dc.DrawRoundedRectangle(x, y, s, s, 4)
	dc.Stroke()
	dc.DrawCircle(x+s/2, y+s/2, s*0.22)
	dc.Fill()
}

func RenderBase(data *BaseInfo) (*gg.Context, error) {
	const cssW, cssH = 1110, 612
	dc := gg.NewContext(1665, 918) // manifest pixels
	dc.Scale(1.5, 1.5)
	FillBackground(dc, 0x2b, 0x33, 0x3c)
	// header (h3 18.7px bold; labor right: glyph + text + 100px bar)
	setFont(dc, 19)
	dc.SetRGB255(255, 255, 255)
	drawStringBold(dc, "基建信息", 8.7, 22.5)
	lx := 977.0
	dc.SetRGB255(0x85, 0x2c, 0xd3)
	dc.DrawRectangle(lx, 4, 4, 4)
	dc.Fill()
	dc.DrawRectangle(lx+13, 4, 4, 4)
	dc.Fill()
	dc.DrawRectangle(lx+6.5, 11, 4, 4)
	dc.Fill()
	setFont(dc, 15)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, fmt.Sprintf("%d/%d", data.LaborCur, data.LaborTotal), 1002, 17.5)
	dc.SetRGBA255(255, 255, 255, 25)
	dc.DrawRectangle(980, 22.5, 100, 4)
	dc.Fill()
	dc.SetRGB255(255, 255, 255)
	dc.DrawRectangle(980, 22.5, 100*float64(data.LaborCur)/float64(data.LaborTotal), 4)
	dc.Fill()

	// cards: full 0..1108; halves left 0..552 / right 556..1108; h=112 pitch 117.33
	cardY := func(i int) float64 { return 32 + float64(i)*117.33 }
	halfIdx := 0
	row := 0
	for _, r := range data.Rooms {
		var x, w float64
		var idx int
		if !r.Half {
			x, w = 0, 1108
			idx = row
			row++
		} else {
			x, w = 0, 552
			if halfIdx%2 == 1 {
				x = 556
				w = 552
			}
			idx = row
			if halfIdx%2 == 1 {
				row++
			}
			halfIdx++
		}
		y := cardY(idx)
		dc.SetRGB255(0x21, 0x26, 0x2f)
		dc.DrawRoundedRectangle(x, y, w, 112, 15)
		dc.Fill()
		// title h3 bold
		setFont(dc, 19)
		dc.SetRGB255(255, 255, 255)
		drawStringBold(dc, r.Title, x+10, y+40.2)
		// note right: text ends w-22.5, icon 20px at textStart-40
		if r.Note != "" && r.SkillIcon == "" && len(r.Board) == 0 {
			iconW := 20.0
			iconGap := 40.0
			if r.Title == "发电站 Lv.3" {
				iconW, iconGap = 24, 31
			}
			setFont(dc, 15)
			dc.SetRGB255(r.NoteColor[0], r.NoteColor[1], r.NoteColor[2])
			nw, _ := dc.MeasureString(r.Note)
			nx := x + w - 22.5 - nw
			drawString(dc, r.Note, nx, y+37.5)
			drawNoteIcon(dc, nx-iconGap, y+22.2, iconW, r.NoteColor)
		}
		// meeting: orange note, 线索 label, board boxes
		if len(r.Board) > 0 {
			bx := x + w - 22
			for i := len(r.Board) - 1; i >= 0; i-- {
				setFont(dc, 15)
				dc.SetRGB255(255, 255, 255)
				dc.SetLineWidth(2)
				dc.DrawRoundedRectangle(bx-28, y+19, 28, 28, 8)
				dc.Stroke()
				drawStringAnchored(dc, fmt.Sprintf("%d", r.Board[i]), bx-14, y+38, 0.5, 0.5)
				bx -= 32
			}
			drawString(dc, "线索", bx-36, y+37.5)
			if r.Note != "" {
				setFont(dc, 18)
				dc.SetRGB255(r.NoteColor[0], r.NoteColor[1], r.NoteColor[2])
				nw, _ := dc.MeasureString(r.Note)
				nx := x + w - 162 - nw
				drawString(dc, r.Note, nx, y+39)
				drawNoteIcon(dc, nx-22, y+22, 20, r.NoteColor)
			}
		}
		// training: Lv text + 30px skill icon
		if r.SkillIcon != "" {
			setFont(dc, 15)
			dc.SetRGB255(255, 255, 255)
			nw, _ := dc.MeasureString(r.Note)
			drawString(dc, r.Note, x+w-58-nw, y+35.2)
			drawImageReal(dc, scaleSmooth(tryLocal(r.SkillIcon), 45, 45), (x+w-52)*1.5, (y+18)*1.5)
		}
		// chars
		for _, c := range r.Chars {
			drawImageReal(dc, scaleSmooth(tryLocal(c.Avatar), 60, 60), (x+6)*1.5, (y+65.3)*1.5)
			drawMoodIcon(dc, x+51.7, y+75.3, c.AP)
			setFont(dc, 15)
			dc.SetRGB255(255, 255, 255)
			drawString(dc, c.Name, x+72, y+90.2)
			dc.SetRGB255(0x80, 0x81, 0x85)
			dc.DrawRectangle(x+11.3, y+105.3, 149, 2.7)
			dc.Fill()
			dc.SetRGB255(255, 255, 255)
			dc.DrawRectangle(x+11.3, y+105.3, 149*float64(c.AP)/100, 2.7)
			dc.Fill()
		}
	}
	return dc, nil
}

func measureString(dc *gg.Context, s string) float64 {
	w, _ := dc.MeasureString(s)
	return w
}
