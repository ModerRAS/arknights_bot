package ggrender

import "github.com/fogleman/gg"

// Recruit
type RecruitOp struct{ Avatar, Profession string; Rarity int }
type RecruitList struct {
	Tags      []string
	Operators []RecruitOp
}

func SampleRecruit() *RecruitList {
	ops := make([]RecruitOp, 0, 18)
	for i := 0; i < 18; i++ {
		ops = append(ops, RecruitOp{Avatar: "", Profession: "WARRIOR", Rarity: 3 + i%4})
	}
	return &RecruitList{Tags: []string{"高级资深干员", "新手", "狙击干员", "输出", "治疗", "支援", "费用回复", "精英材料"}, Operators: ops}
}

func RenderRecruit(data *RecruitList) (*gg.Context, error) {
	const mainW = 1350
	tileW, tileH := 100, 120
	cols := mainW / tileW
	pad := 20
	m := gg.NewContext(mainW, 10)
	setFont(m, 14)
	tagX := float64(pad)
	tagY := float64(pad) + 14
	tagArea := 40.0
	for _, t := range data.Tags {
		w, _ := m.MeasureString(t)
		if tagX+w+20 > float64(mainW-pad) {
			tagX = float64(pad)
			tagY += 26
			tagArea += 26
		}
		tagX += w + 20
	}
	dc := gg.NewContext(mainW, 534)
	FillBackground(dc, 27, 29, 30)
	setFont(dc, 14)
	tx := float64(pad)
	ty := float64(pad) + 14
	for _, t := range data.Tags {
		w, _ := dc.MeasureString(t)
		dc.SetRGB255(60, 90, 110)
		RoundRect(dc, tx, ty-14, w+16, 22, 6)
		dc.SetRGB255(220, 230, 235)
		drawString(dc, t, tx+8, ty)
		tx += w + 20
		if tx > float64(mainW-pad) {
			tx = float64(pad)
			ty += 26
		}
	}
	gridTop := int(tagY) + 10
	for i, o := range data.Operators {
		x := (i%cols)*tileW + pad
		y := gridTop + (i/cols)*tileH
		DrawPortraitTile(dc, x, y, tileW-10, tileH, o.Avatar, o.Profession, o.Rarity, 0, "")
	}
	return dc, nil
}
