package ggrender

import (
	"image"
	"math"

	xdraw "golang.org/x/image/draw"

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

func scaleSmoothCR(img image.Image, w, h int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), xdraw.Over, nil)
	return dst
}

func RenderOperator(data *OperatorInfo) (*gg.Context, error) {
	const cssW, cssH = 1200, 800
	dc := gg.NewContext(1800, 1200) // manifest pixels
	// browser body default white behind everything (bg.png has transparency)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	// bg + painting drawn 1:1 in REAL pixel space BEFORE the CSS scale transform
	// (gg DrawImage resamples under transform; identity matrix keeps pixels exact)
	bg := scaleSmooth(tryLocal("operator/bg.png"), 2130, 1200)
	dc.DrawImage(bg, -2, 0)
	paint := tryLocal(data.Painting)
	dc.DrawImage(scaleSmooth(paint, 978, 974), 89, 225)
	dc.Scale(1.5, 1.5)

	// ---- attr table (top-left): 3 rows x (b101.3,w71.3)x3, y18.7 rowH30 ----
	attrY := 20.0
	rowH := 30.0
	for r, row := range data.AttrRows {
		y := attrY + float64(r)*rowH
		x := 0.0
		for c, cell := range row {
			w := 101.3
			if c%2 == 1 {
				w = 71.3
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
			setFont(dc, 16)
			drawString(dc, cell, x+5.3, y+21)
			x += w
		}
	}

	// ---- potential table ----
	dc.SetRGBA255(0, 0, 0, 204)
	dc.DrawRectangle(0, 120, 136.7, 25.3)
	dc.Fill()
	dc.SetRGB255(255, 255, 255)
	setFont(dc, 16)
	drawString(dc, "潜能提升", 5.3, 139)
	potY := []float64{146, 190}
	potH := []float64{44, 25.3}
	potBase := []float64{168, 211}
	for i, p := range data.Potentials {
		dc.SetRGBA255(0xef, 0xee, 0xef, 102)
		dc.DrawRectangle(0, potY[i], 136.7, potH[i])
		dc.Fill()
		pot := tryLocal("box/Potential_2.png")
		drawImageReal(dc, scaleSmooth(pot, 30, 30), 8, (potY[i]+5)*1.5)
		dc.SetRGB255(0, 0, 0)
		setFont(dc, 16)
		drawString(dc, p, 30, potBase[i])
	}

	// ---- right tables (x=600 w=600) ----
	rtX := 600.0
	rtW := 600.0
	// talent header
	dc.SetRGB255(0, 0, 0)
	dc.DrawRectangle(rtX, 20, rtW, 20)
	dc.Fill()
	dc.SetRGB255(255, 255, 255)
	setFont(dc, 15)
	drawString(dc, "天赋", rtX+3, 35.5)
	// talent row
	dc.SetRGBA255(0xb0, 0xb1, 0xb1, 178)
	dc.DrawRectangle(rtX, 40, rtW, 20)
	dc.Fill()
	dc.SetRGB255(0, 0, 0)
	setFont(dc, 12)
	drawString(dc, "精英化1", rtX+3, 55)
	drawStringAnchored(dc, data.TalentName, rtX+190, 55, 0.5, 0.5)
	drawString(dc, data.TalentDesc, rtX+347, 55)
	// building header
	dc.SetRGB255(0, 0, 0)
	dc.DrawRectangle(rtX, 60, rtW, 20)
	dc.Fill()
	dc.SetRGB255(255, 255, 255)
	setFont(dc, 15)
	drawString(dc, "基建技能", rtX+3, 75.5)
	// building row
	dc.SetRGBA255(0xb0, 0xb1, 0xb1, 178)
	dc.DrawRectangle(rtX, 80, rtW, 40)
	dc.Fill()
	dc.SetRGB255(0, 0, 0)
	setFont(dc, 12)
	drawString(dc, "精英化0", rtX+3, 104)
	drawImageReal(dc, scaleSmooth(tryLocal("operator/building.png"), 54, 54), 693*1.5, 86*1.5)
	drawString(dc, data.BuildingName, rtX+173, 104)
	drawString(dc, data.BuildingDesc, rtX+273, 104)
	// skill block
	dc.SetRGBA255(63, 63, 62, 252)
	dc.DrawRectangle(rtX, 120, rtW, 73.3)
	dc.Fill()
	drawImageReal(dc, scaleSmooth(tryLocal("operator/skill.png"), 80, 80), 630*1.5, 121*1.5)
	dc.SetRGB255(255, 255, 255)
	setFont(dc, 12)
	drawStringAnchored(dc, data.SkillName, rtX+56, 181, 0.5, 0.5)
	drawString(dc, "自动回复", rtX+240, 137)
	dc.DrawRectangle(rtX+288, 126, 59, 15)
	dc.Stroke()
	drawString(dc, "技力0/45", rtX+291, 137.5)
	dc.DrawRectangle(rtX+353, 126, 76, 15)
	dc.Stroke()
	drawString(dc, "持续时间30s", rtX+356, 137.5)
	drawString(dc, data.SkillDesc, rtX+123, 167)
	dc.DrawRectangle(rtX+580, 150, 10, 10)
	dc.Stroke()

	// ---- bottom-left identity ----
	// class icon black box + white staff
	dc.SetRGB255(0, 0, 0)
	dc.DrawRectangle(13.3, 483.3, 75.4, 75.4)
	dc.Fill()
	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(4)
	dc.DrawLine(30, 545, 68, 505)
	dc.Stroke()
	dc.DrawCircle(72, 500, 6)
	dc.Fill()
	dc.SetLineWidth(2.5)
	dc.DrawLine(24, 552, 33, 543)
	dc.Stroke()
	// stars
	dc.SetRGB255(0xfd, 0xdb, 0x1e)
	for i := 0; i < data.Rarity; i++ {
		drawStar(dc, 99+float64(i)*11.7, 496.5, 7.2)
	}
	// class name bold
	dc.SetRGB255(0, 0, 0)
	setFont(dc, 20)
	drawStringBold(dc, data.ClassName, 93, 530)
	// position/tag
	setFont(dc, 16)
	drawString(dc, data.Position+" "+data.Tag, 13.3, 578.5)
	drawStringBold(dc, data.Desc, 13.3, 608)
	// big name
	setFont(dc, 32)
	drawStringBold(dc, data.Name, 13.3, 693)
	// checkbox
	dc.SetLineWidth(1.5)
	dc.DrawRectangle(113, 680, 13, 13)
	dc.Stroke()
	// code badge
	dc.SetRGB255(255, 255, 255)
	dc.DrawRectangle(13.3, 733, 40, 19.4)
	dc.Fill()
	dc.SetRGB255(0, 0, 0)
	setFont(dc, 15)
	drawStringBold(dc, data.Code, 16, 748.5)
	// en name
	setFont(dc, 16)
	drawString(dc, data.EnName, 113, 749)
	return dc, nil
}
