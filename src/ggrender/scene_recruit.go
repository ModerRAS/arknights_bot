package ggrender

import "github.com/fogleman/gg"

// Recruit — mirrors template/Recruit.tmpl rendered at 900x356 CSS, scale 1.5 -> 1350x534.
// White page, table: header row (标签|干员) with material shadow, one tr per tag-group:
// left cell = dark #313131 tag chips, right cell = inline avatar cards
// (avatar 100px, profession icon 30px top-left, rarity stars h=20 offset 30px).
type RecruitOp struct {
	Avatar, Profession string
	Rarity             int
}
type RecruitList struct {
	Tags      []string
	Operators []RecruitOp
}

const recruitAvatarURL = "https://media.prts.wiki/3/36/%E5%A4%B4%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85.png?image_process=format,webp/quality,Q_90"

func SampleRecruit() []RecruitList {
	op := func() RecruitOp { return RecruitOp{Avatar: recruitAvatarURL, Profession: "WARRIOR", Rarity: 5} }
	g1 := make([]RecruitOp, 0, 12)
	for i := 0; i < 12; i++ {
		g1 = append(g1, op())
	}
	g2 := make([]RecruitOp, 0, 2)
	for i := 0; i < 2; i++ {
		g2 = append(g2, op())
	}
	return []RecruitList{
		{Tags: []string{"高级资深干员", "输出"}, Operators: g1},
		{Tags: []string{"术师干员", "远程位"}, Operators: g2},
	}
}

func RenderRecruit(groups []RecruitList) (*gg.Context, error) {
	const W, H = 1350, 534
	const opsX = 280 // col2 left edge (col1 = tag cell)
	const opsPerLine = 6
	const linePitch = 163
	const lineTop0 = 45
	dc := gg.NewContext(W, H)
	FillBackground(dc, 255, 255, 255)

	// header row: 标签 | 干员 (th bold, centered), shadow under the tr
	setFont(dc, 24)
	dc.SetRGB255(0, 0, 0)
	drawStringBold(dc, "标签", 118, 30)
	drawStringBold(dc, "干员", 793, 30)
	dc.SetRGBA255(0, 0, 0, 70)
	dc.DrawRectangle(0, 42, W, 2.5)
	dc.Fill()
	dc.SetRGBA255(0, 0, 0, 30)
	dc.DrawRectangle(0, 44.5, W, 3)
	dc.Fill()

	lineTop := lineTop0
	for _, g := range groups {
		// tag chips vertically centered against the group's op lines
		lines := (len(g.Operators) + opsPerLine - 1) / opsPerLine
		if lines < 1 {
			lines = 1
		}
		centerY := float64(lineTop) + (float64(lines-1)*linePitch+150)/2 + 4
		if lines == 1 {
			centerY += 3
		}
		setFont(dc, 24)
		tx := 10.0
		for _, t := range g.Tags {
			tw, _ := measure(dc, t)
			bw := tw + 24
			dc.SetRGB255(49, 49, 49)
			RoundRect(dc, tx, centerY-21, bw, 42, 2)
			dc.SetRGB255(255, 255, 255)
			drawString(dc, t, tx+12, centerY+8)
			tx += bw + 30
		}
		// op cards
		for i, o := range g.Operators {
			col, row := i%opsPerLine, i/opsPerLine
			x := float64(opsX) + float64(col)*155.5
			y := lineTop + row*linePitch
			av := ScaleExact(FetchImage(o.Avatar, AssetPath("common/amiya.png")), 150, 150)
			dc.DrawImage(av, int(x), y)
			prof := tryLocal("box/" + o.Profession + ".png")
			dc.DrawImage(ScaleExact(prof, 45, 45), int(x), y)
			rar := tryLocal("box/Rarity_" + itoa(o.Rarity) + ".png")
			rw := rar.Bounds().Dx() * 30 / rar.Bounds().Dy()
			dc.DrawImage(ScaleExact(rar, rw, 30), int(x)+45, y)
		}
		lineTop += lines * linePitch
	}
	return dc, nil
}
