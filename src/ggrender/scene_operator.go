package ggrender

import (
	"math"

	"github.com/fogleman/gg"
)

// Operator scene: manifest 1800x1200 px = 1200x800 CSS @1.5x (template/Operator.tmpl).
// bg.png cover, half-body painting bottom-left, attr/potential tables top-left,
// talent/building/skill tables right, class icon + stars + names bottom-left.

type OperatorInfo struct {
	Name, EnName, Code     string
	ClassName              string // 中坚术师
	Position, Tag, Desc    string
	Rarity                 int
	Painting, ClassIcon    string // asset rel paths
	AttrRows               [][6]string
	Potentials             []string
	TalentName, TalentDesc string
	BuildingName           string
	BuildingDesc           string
	SkillName              string
	SkillMeta              string // 自动回复 技力0/45 持续时间30s
	SkillDesc              string
}

func SampleOperator() *OperatorInfo {
	return &OperatorInfo{
		Name: "阿米娅", EnName: "Amiya", Code: "R001",
		ClassName: "中坚术师", Position: "远程位", Tag: "输出",
		Desc:     "攻击造成法术伤害",
		Rarity:   6,
		Painting: "operator/painting.png", ClassIcon: "operator/class.png",
		AttrRows: [][6]string{
			{"最大生命值", "1742", "攻击力", "699", "防御力", "121"},
			{"法抗", "10", "攻击间隔", "1.6s", "再部署时间", "70s"},
			{"阻挡数", "1", "部署费用", "18", "所属", "罗德岛"},
		},
		Potentials:   []string{"部署费用-1", "再部署时间-4秒"},
		TalentName:   "精神融合",
		TalentDesc:   "攻击力+10%",
		BuildingName: "合作协议",
		BuildingDesc: "控制中枢内线索搜集速度提升",
		SkillName:    "战术咏唱",
		SkillMeta:    "自动回复 技力0/45 持续时间30s",
		SkillDesc:    "攻击力提升",
	}
}

func drawStar(dc *gg.Context, cx, cy, r float64) {
	var pts []float64
	for i := 0; i < 10; i++ {
		ang := -math.Pi/2 + float64(i)*math.Pi/5
		rad := r
		if i%2 == 1 {
			rad = r * 0.45
		}
		pts = append(pts, cx+rad*math.Cos(ang), cy+rad*math.Sin(ang))
	}
	dc.MoveTo(pts[0], pts[1])
	for i := 2; i < len(pts); i += 2 {
		dc.LineTo(pts[i], pts[i+1])
	}
	dc.ClosePath()
	dc.Fill()
}

func RenderOperator(data *OperatorInfo) (*gg.Context, error) {
	const cssW, cssH = 1200, 800
	dc := gg.NewContext(1800, 1200) // manifest pixels
	dc.Scale(1.5, 1.5)
	// bg cover
	dc.DrawImage(ScaleCover(tryLocal("operator/bg.png"), cssW, cssH), 0, 0)
	// painting: height 650, bottom 0, left 5%
	paint := tryLocal(data.Painting)
	pw := 650.0
	dc.DrawImage(ScaleExact(paint, int(pw), int(pw)), 60, cssH-650)

	// ---- attr table (top-left, opacity .8) ----
	attrY := 20.0
	rowH := 25.7
	setFont(dc, 15)
	for r, row := range data.AttrRows {
		y := attrY + float64(r)*rowH
		x := 0.0
		for c, cell := range row {
			w := 100.0
			if c%2 == 1 {
				w = 70.0
			}
			if c%2 == 0 {
				dc.SetRGBA255(0, 0, 0, 204)
			} else {
				dc.SetRGBA255(0xef, 0xee, 0xef, 204)
			}
			dc.DrawRectangle(x, y, w, rowH)
			dc.Fill()
			if c%2 == 0 {
				dc.SetRGB255(255, 255, 255)
			} else {
				dc.SetRGB255(0, 0, 0)
			}
			drawString(dc, cell, x+3, y+18)
			x += w
		}
	}

	// ---- potential table ----
	potY := attrY + 3*rowH + 20
	dc.SetRGBA255(0, 0, 0, 204)
	dc.DrawRectangle(0, potY, 205, rowH)
	dc.Fill()
	dc.SetRGB255(255, 255, 255)
	drawString(dc, "潜能提升", 3, potY+18)
	for i, p := range data.Potentials {
		y := potY + rowH + float64(i)*24
		dc.SetRGBA255(0xef, 0xee, 0xef, 102)
		dc.DrawRectangle(0, y, 205, 24)
		dc.Fill()
		pot := tryLocal("box/Potential_2.png")
		dc.DrawImage(ScaleExact(pot, 20, 20), 2, int(y)+2)
		dc.SetRGB255(0, 0, 0)
		setFont(dc, 14)
		drawString(dc, p, 26, y+17)
	}

	// ---- right tables (x=600 w=600) ----
	rtX := 600.0
	rtW := 600.0
	setFont(dc, 15)
	// talent
	dc.SetRGBA255(0, 0, 0, 204)
	dc.DrawRectangle(rtX, 20, rtW, rowH)
	dc.Fill()
	dc.SetRGB255(255, 255, 255)
	drawString(dc, "天赋", rtX+3, 38)
	dc.SetRGBA255(0xb0, 0xb1, 0xb1, 178)
	dc.DrawRectangle(rtX, 20+rowH, rtW, 24)
	dc.Fill()
	dc.SetRGB255(0, 0, 0)
	setFont(dc, 12)
	drawString(dc, "精英化1", rtX+3, 20+rowH+16)
	drawStringAnchored(dc, data.TalentName, rtX+250, 20+rowH+16, 0.5, 0.5)
	drawString(dc, data.TalentDesc, rtX+380, 20+rowH+16)
	// building
	bY := 20 + rowH + 24 + 8
	dc.SetRGBA255(0, 0, 0, 204)
	dc.DrawRectangle(rtX, bY, rtW, rowH)
	dc.Fill()
	dc.SetRGB255(255, 255, 255)
	setFont(dc, 15)
	drawString(dc, "基建技能", rtX+3, bY+18)
	dc.SetRGBA255(0xb0, 0xb1, 0xb1, 178)
	dc.DrawRectangle(rtX, bY+rowH, rtW, 48)
	dc.Fill()
	dc.SetRGB255(0, 0, 0)
	setFont(dc, 12)
	drawString(dc, "精英化0", rtX+3, bY+rowH+28)
	bicon := tryLocal("operator/building.png")
	dc.DrawImage(ScaleExact(bicon, 36, 36), int(rtX)+110, int(bY+rowH)+6)
	drawString(dc, data.BuildingName, rtX+160, bY+rowH+28)
	drawString(dc, data.BuildingDesc, rtX+260, bY+rowH+28)
	// skill block
	sY := bY + rowH + 48 + 8
	dc.SetRGBA255(63, 63, 62, 252)
	dc.DrawRectangle(rtX, sY, rtW, 96)
	dc.Fill()
	sicon := ScaleExact(tryLocal("operator/skill.png"), 56, 56)
	dc.DrawImage(sicon, int(rtX)+40, int(sY)+8)
	dc.SetRGB255(255, 255, 255)
	setFont(dc, 12)
	drawStringAnchored(dc, data.SkillName, rtX+68, sY+80, 0.5, 0.5)
	drawStringAnchored(dc, data.SkillMeta, rtX+330, sY+20, 0.5, 0.5)
	drawStringAnchored(dc, data.SkillDesc, rtX+330, sY+56, 0.5, 0.5)

	// ---- bottom-left identity ----
	// class icon black box
	dc.SetRGBA255(0, 0, 0, 230)
	dc.DrawRectangle(13, 487, 76, 76)
	dc.Fill()
	// white staff glyph
	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(5)
	dc.DrawLine(30, 548, 72, 505)
	dc.Stroke()
	dc.DrawCircle(76, 501, 7)
	dc.Fill()
	dc.SetLineWidth(3)
	dc.DrawLine(24, 554, 34, 544)
	dc.Stroke()
	// stars
	dc.SetRGB255(0xfd, 0xdb, 0x1e)
	for i := 0; i < data.Rarity; i++ {
		drawStar(dc, 110+float64(i)*23, 500, 11)
	}
	// class name
	dc.SetRGB255(0, 0, 0)
	setFont(dc, 26)
	drawString(dc, data.ClassName, 96, 540)
	// position/tag
	setFont(dc, 16)
	drawString(dc, data.Position+" "+data.Tag, 13, 600)
	setFont(dc, 16)
	drawString(dc, data.Desc, 13, 626)
	// big name
	setFont(dc, 40)
	dc.SetRGB255(0, 0, 0)
	drawString(dc, data.Name, 13, 715)
	// checkbox
	dc.SetLineWidth(1.5)
	dc.DrawRectangle(120, 692, 16, 16)
	dc.Stroke()
	// code badge
	dc.SetRGB255(255, 255, 255)
	dc.DrawRectangle(13, 745, 46, 26)
	dc.Fill()
	dc.SetRGB255(0, 0, 0)
	setFont(dc, 20)
	drawString(dc, data.Code, 16, 766)
	// en name
	drawString(dc, data.EnName, 120, 765)
	return dc, nil
}
