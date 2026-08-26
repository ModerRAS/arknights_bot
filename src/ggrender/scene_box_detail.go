package ggrender

import (
	"fmt"

	"github.com/fogleman/gg"
)

type Skill struct {
	Id    string
	Level int
}
type Equip struct {
	Id    string
	Level int
}
type Detail struct {
	Name, Id                                  string
	Rarity, Level, EvolvePhase, PotentialRank int
	Skills                                    []Skill
	Equips                                    []Equip
}
type BoxDetailList struct{ Items []Detail }

func SampleBoxDetail() []Detail {
	return []Detail{
		{Name: "阿米娅", Id: AmiyaAvatarURL, Rarity: 5, Level: 90, EvolvePhase: 2, PotentialRank: 5, Skills: []Skill{{Id: "ska", Level: 10}}, Equips: []Equip{{Id: "eqA", Level: 1}}},
		{Name: "阿米娅", Id: AmiyaAvatarURL, Rarity: 5, Level: 80, EvolvePhase: 2, PotentialRank: 4, Skills: []Skill{{Id: "ska", Level: 10}, {Id: "skb", Level: 9}, {Id: "skc", Level: 8}}, Equips: []Equip{{Id: "eqA", Level: 2}, {Id: "eqB", Level: 3}}},
	}
}

// renderDetailIcon 本地素材图标，缺失则跳过。
func renderDetailIcon(dc *gg.Context, rel string, x, y, w, h int) {
	if ic, err := LoadImage(AssetPath(rel)); err == nil {
		dc.DrawImage(ScaleExact(ic, w, h), x, y)
	}
}

func RenderBoxDetail(data []Detail) (*gg.Context, error) {
	const mainW, mainH = 722, 279
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 46, 48, 49)
	// 表头
	setFont(dc, 29)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, "干员", 55, 36)
	drawString(dc, "等级", 171, 36)
	drawString(dc, "潜能", 254, 36)
	drawString(dc, "技能", 414, 36)
	drawString(dc, "模组", 616, 36)
	for i, d := range data {
		rt := 44 + i*117
		// 头像 + 名字
		port := FetchImage(d.Id, amiyaPath)
		dc.DrawImage(smoothCover(port, 80, 80), 4, rt+20)
		setFont(dc, 25)
		dc.SetRGB255(255, 255, 255)
		drawString(dc, d.Name, 88, float64(rt+68))
		// 等级：精英徽章 + LV
		renderDetailIcon(dc, fmt.Sprintf("box/Evolve_%d.png", d.EvolvePhase), 162, rt+5, 58, 48)
		setFont(dc, 30)
		drawString(dc, "LV"+itoa(d.Level), 165, float64(rt+89))
		// 潜能
		renderDetailIcon(dc, fmt.Sprintf("box/Potential_%d.png", d.PotentialRank), 245, rt+8, 62, 52)
		// 技能：色块占位 + LV（组内居中于 x≈440）
		n := len(d.Skills)
		total := n*62 + (n-1)*23
		sx := 440 - total/2
		for _, s := range d.Skills {
			dc.SetRGB255(170, 40, 40)
			dc.DrawRoundedRectangle(float64(sx), float64(rt+2), 62, 57, 6)
			dc.Fill()
			setFont(dc, 22)
			dc.SetRGB255(255, 255, 255)
			drawStringAnchored(dc, "LV"+itoa(s.Level), float64(sx+31), float64(rt+95), 0.5, 0.5)
			sx += 85
		}
		// 模组：白色方框 + LV
		ex := 612
		for _, e := range d.Equips {
			dc.SetRGB255(255, 255, 255)
			dc.SetLineWidth(3)
			dc.DrawRectangle(float64(ex), float64(rt+15), 44, 42)
			dc.Stroke()
			drawStringAnchored(dc, "LV"+itoa(e.Level), float64(ex+22), float64(rt+90), 0.5, 0.5)
			ex += 70
		}
	}
	return dc, nil
}
