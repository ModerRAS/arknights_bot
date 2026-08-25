package ggrender

import "github.com/fogleman/gg"

// Enemy — mirrors template/Enemy.tmpl rendered at 656x318 CSS, scale 1.5 -> 984x477.
// The bare <tr>{{.Ability}}</tr> is foster-parented above the table: black text on
// white body strip. #main #323332 with bordered centered table rows; canvas cuts
// after the 能力 header row (level tables fall below the fold).
type EnemySkill struct{ Name, SpInit, SpCost, Desc string }
type EnemyLevel struct {
	Desc, AttackType, Motion, HpRecovery, HP, ATK, DEF, Res, ATKRadius, Weight, MoveSpeed, Interval, DamageRes, ElementRes, Ridicule, Point, Abnormal string
	Skills []EnemySkill
	Talent string
}
type Enemy struct {
	Name, Pic, Desc, EnemyRace, EnemyLevel, AttackType, Motion string
	Ability string
	Levels []EnemyLevel
}

const enemyPicURL = "https://media.prts.wiki/3/3e/%E5%A4%B4%E5%83%8F_%E6%95%8C%E4%BA%BA_%E6%BA%90%E7%9F%B3%E8%99%AB.png"

func SampleEnemy() *Enemy {
	return &Enemy{
		Name: "源石虫", Pic: enemyPicURL, Desc: "拥有较高防御力的感染生物。", EnemyRace: "感染生物", EnemyLevel: "普通", AttackType: "近战", Motion: "地面",
		Ability: "免疫沉默",
		Levels: []EnemyLevel{
			{Desc: "标准个体", AttackType: "近战", Motion: "地面", HpRecovery: "0", HP: "5000", ATK: "400", DEF: "500", Res: "20", ATKRadius: "1.1", Weight: "2", MoveSpeed: "0.8", Interval: "1.5", DamageRes: "0", ElementRes: "0", Ridicule: "0", Point: "1", Abnormal: "无",
				Skills: []EnemySkill{{Name: "啃噬", SpInit: "0", SpCost: "5", Desc: "对目标造成 物理伤害"}}, Talent: "生命低于50%时防御提升"},
			{Desc: "强化个体", AttackType: "近战", Motion: "地面", HpRecovery: "0", HP: "8000", ATK: "600", DEF: "700", Res: "30", ATKRadius: "1.1", Weight: "3", MoveSpeed: "0.8", Interval: "1.5", DamageRes: "10", ElementRes: "0", Ridicule: "1", Point: "2", Abnormal: "晕眩抗性",
				Skills: []EnemySkill{{Name: "啃噬+", SpInit: "0", SpCost: "4", Desc: "造成更高 物理伤害"}}, Talent: "攻击力提升"},
		},
	}
}

func RenderEnemy(data *Enemy) (*gg.Context, error) {
	const W, H = 984, 477
	dc := gg.NewContext(W, H)
	FillBackground(dc, 50, 51, 50)
	// foster-parented ability text: black on the #323332 strip
	setFont(dc, 24)
	dc.SetRGB255(0, 0, 0)
	drawString(dc, StripHTML(data.Ability), 3, 27)

	// table rows: [y0,y1] canvas
	rows := []struct{ y0, y1 float64 }{
		{36, 98},   // name
		{98, 342},  // pic | desc
		{342, 384}, // 种类 | 地位级别 | 攻击方式 | 行动方式
		{384, 426}, // values
		{426, 477}, // 能力
	}
	// 2px #595858 lines centered on the cell boundaries (canvas already #323332)
	border := func(x0, y0, x1, y1 float64) {
		dc.SetRGB255(89, 88, 88)
		dc.DrawRectangle(x0, y0, x1-x0, y1-y0)
		dc.Fill()
	}
	// horizontal borders
	for _, r := range rows {
		border(0, r.y0-1, W, r.y0+1)
	}
	border(0, 475, W, 477)
	// vertical borders: pic|desc split at 425 (rows 1-2), header/value splits in rows 2-3
	border(424, rows[1].y0-1, 426, rows[2].y1)
	for _, x := range []float64{425, 611.5, 797.5} {
		border(x-1, rows[2].y0-1, x+1, rows[2].y1)
	}
	border(981.5, rows[2].y0-1, 983.5, rows[2].y1)

	centerText := func(s string, cx, cy float64, size float64) {
		setFont(dc, size)
		dc.SetRGB255(255, 255, 255)
		drawStringAnchored(dc, s, cx, cy, 0.5, 0.5)
	}
	// name row
	setFont(dc, 37.5)
	dc.SetRGB255(255, 255, 255)
	nw, _ := measure(dc, data.Name)
	drawStringBoldW(dc, data.Name, W/2-nw/2-1, 81, 1.4)
	// pic cell: img 158 CSS = 237 canvas centered in 0..425
	pic := FetchImage(data.Pic, AssetPath("common/amiya.png"))
	dc.DrawImage(ScaleExact(pic, 237, 237), 95, 102)
	// desc cell centered
	centerText(StripHTML(data.Desc), (425+W)/2+1, 221, 24)
	// header row
	hdrs := []struct {
		s  string
		cx float64
	}{{"种类", 213.5}, {"地位级别", 519.5}, {"攻击方式", 705.5}, {"行动方式", 891.5}}
	for _, hd := range hdrs {
		centerText(hd.s, hd.cx, 363, 24)
	}
	// values row
	vals := []struct {
		s  string
		cx float64
	}{{data.EnemyRace, 213.5}, {data.EnemyLevel, 519.5}, {data.AttackType, 705.5}, {data.Motion, 891.5}}
	for _, v := range vals {
		centerText(v.s, v.cx, 405, 24)
	}
	// 能力 row
	setFont(dc, 30)
	dc.SetRGB255(255, 255, 255)
	aw, _ := measure(dc, "能力")
	drawStringBoldW(dc, "能力", W/2-aw/2, 463.5, 1.2)
	return dc, nil
}
