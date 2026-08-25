package ggrender

import (
	"fmt"
	"image"

	"github.com/fogleman/gg"
)

// Scenes canonical 16 must stay exact; harness fails if changed.
var Scenes = []string{
	"base", "box", "box-detail", "box-summary",
	"calendar", "card", "depot", "enemy",
	"gacha", "headhunt", "help", "lottery",
	"missing", "operator", "recruit", "state",
}

// SceneSet for validation.
var SceneSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(Scenes))
	for _, s := range Scenes {
		m[s] = struct{}{}
	}
	return m
}()

// RenderGG unified production renderer: scene -> image.
// data may be nil -> uses frozen Sample fixture for deterministic tests.
// ponytail: single dispatch, no extra abstraction.
func RenderGG(scene string, data interface{}) (image.Image, error) {
	dc, err := renderContext(scene, data)
	if err != nil {
		return nil, err
	}
	return dc.Image(), nil
}

// RenderGGContext returns gg Context directly (for EncodePNG convenience).
func RenderGGContext(scene string, data interface{}) (*gg.Context, error) {
	return renderContext(scene, data)
}

func normalizeScene(s string) string {
	// accepts "BoxDetail", "boxDetail", "box_detail", "box-detail", "box" etc -> canonical hyphen lower
	s = string([]rune(s))
	// simple lower
	lower := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			r = r - 'A' + 'a'
		}
		lower += string(r)
	}
	// replace '_' with '-'
	rep := ""
	for _, r := range lower {
		if r == '_' {
			rep += "-"
		} else {
			rep += string(r)
		}
	}
	// handle camelCase boxDetail -> box-detail via known aliases
	alias := map[string]string{
		"boxdetail": "box-detail", "boxsummary": "box-summary", "lotterydetail": "lottery",
	}
	if v, ok := alias[rep]; ok {
		return v
	}
	// also handle "box-detail" already
	return rep
}

func renderContext(scene string, data interface{}) (*gg.Context, error) {
	scene = normalizeScene(scene)
	switch scene {
	case "base":
		if d, ok := data.(*BaseInfo); ok && d != nil {
			return RenderBase(d)
		}
		return RenderBase(SampleBase())
	case "box":
		if d, ok := data.(*BoxInfo); ok && d != nil {
			return RenderBox(d)
		}
		return RenderBox(SampleBox())
	case "box-detail":
		if d, ok := data.(*BoxDetailList); ok && d != nil {
			return RenderBoxDetail(d.Items)
		}
		if arr, ok := data.([]Detail); ok {
			return RenderBoxDetail(arr)
		}
		return RenderBoxDetail(SampleBoxDetail())
	case "box-summary":
		if d, ok := data.(*BoxSummary); ok && d != nil {
			return RenderBoxSummary(d)
		}
		return RenderBoxSummary(SampleBoxSummary())
	case "calendar":
		if d, ok := data.(*CalendarData); ok && d != nil {
			return RenderCalendar(d)
		}
		return RenderCalendar(SampleCalendar())
	case "card":
		if d, ok := data.(*CardInfo); ok && d != nil {
			return RenderCard(d)
		}
		return RenderCard(SampleCard())
	case "depot":
		if d, ok := data.(*DepotData); ok && d != nil {
			return RenderDepot(d)
		}
		return RenderDepot(SampleDepot())
	case "enemy":
		if d, ok := data.(*Enemy); ok && d != nil {
			return RenderEnemy(d)
		}
		return RenderEnemy(SampleEnemy())
	case "gacha":
		if d, ok := data.(*GachaData); ok && d != nil {
			return RenderGacha(d)
		}
		return RenderGacha(SampleGacha())
	case "headhunt":
		if d, ok := data.(*HeadhuntData); ok && d != nil {
			return RenderHeadhunt(d.Ops)
		}
		if arr, ok := data.([]HHOp); ok {
			return RenderHeadhunt(arr)
		}
		return RenderHeadhunt(SampleHeadhunt())
	case "help":
		if d, ok := data.(*HelpData); ok && d != nil {
			return RenderHelp(d)
		}
		return RenderHelp(SampleHelp())
	case "lottery":
		if d, ok := data.(*LotteryData); ok && d != nil {
			return RenderLottery(d)
		}
		return RenderLottery(SampleLottery())
	case "missing":
		if d, ok := data.(*MissingInfo); ok && d != nil {
			return RenderMissing(d)
		}
		return RenderMissing(SampleMissing())
	case "operator":
		if d, ok := data.(*OperatorInfo); ok && d != nil {
			return RenderOperator(d)
		}
		return RenderOperator(SampleOperator())
	case "recruit":
		if d, ok := data.(*RecruitList); ok && d != nil {
			return RenderRecruit(d)
		}
		return RenderRecruit(SampleRecruit())
	case "state":
		if d, ok := data.(*StateInfo); ok && d != nil {
			return RenderState(d)
		}
		return RenderState(SampleState())
	default:
		return nil, fmt.Errorf("unknown scene %q", scene)
	}
}

// ---------- shared sample/fixture types ----------

// Char mirrors Box.
type Char struct {
	CharId, SkinId, Name              string
	Level, EvolvePhase, PotentialRank int
	FavorPercent, Rarity              int
	Profession                        string
}

type BoxInfo struct {
	Name  string
	Chars []Char
}

func SampleBox() *BoxInfo {
	profs := []string{"WARRIOR", "CASTER", "MEDIC", "PIONEER", "SNIPER", "SPECIAL", "SUPPORT", "TANK"}
	chars := make([]Char, 0, 18)
	for i := 0; i < 18; i++ {
		r := 3 + i%4
		chars = append(chars, Char{
			SkinId:        fmt.Sprintf("char_%03d_amiya_%d", i, i%3+1),
			Name:          fmt.Sprintf("干员%d", i+1),
			Rarity:        r,
			Profession:    profs[i%len(profs)],
			Level:         30 + i%60,
			EvolvePhase:   i % 3,
			PotentialRank: i % 6,
		})
	}
	return &BoxInfo{Name: "博士的作战档案", Chars: chars}
}

func RenderBox(data *BoxInfo) (*gg.Context, error) {
	const mainW = 700
	tileW, tileH := 70, 140
	cols := 10
	rows := (len(data.Chars) + cols - 1) / cols
	if rows < 1 {
		rows = 1
	}
	labelH := 60
	gridTop := labelH + 10
	mainH := gridTop + rows*tileH + 20
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 46, 48, 49)
	// label bar
	dc.SetRGB255(60, 62, 64)
	dc.DrawRectangle(0, 0, float64(mainW), float64(labelH))
	dc.Fill()
	setFont(dc, 28)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, data.Name, 25, float64(labelH)-20)
	for i, c := range data.Chars {
		x := (i % cols) * tileW
		y := gridTop + (i/cols)*tileH
		DrawPortraitTile(dc, x, y, tileW, tileH, "", c.Profession, c.Rarity, c.Level, c.Name)
		if c.EvolvePhase > 0 {
			ev := tryLocal(fmt.Sprintf("box/Evolve_%d.png", c.EvolvePhase))
			dc.DrawImage(ScaleExact(ev, 24, 24), x+tileW-26, y+tileH/2)
		}
	}
	return dc, nil
}

// BoxDetail

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
		{Name: "闪灵", Id: "char_1001_amiya_1", Rarity: 5, Level: 80, EvolvePhase: 2, PotentialRank: 5, Skills: []Skill{{Id: "sk1", Level: 10}, {Id: "sk2", Level: 10}}, Equips: []Equip{}},
		{Name: "史尔特尔", Id: "char_1002_amiya_1", Rarity: 6, Level: 90, EvolvePhase: 2, PotentialRank: 6, Skills: []Skill{{Id: "sk3", Level: 10}}, Equips: []Equip{{Id: "eq1", Level: 3}}},
		{Name: "能天使", Id: "char_1003_amiya_1", Rarity: 6, Level: 90, EvolvePhase: 2, PotentialRank: 6, Skills: []Skill{{Id: "sk4", Level: 10}, {Id: "sk5", Level: 10}}, Equips: []Equip{{Id: "eq2", Level: 3}, {Id: "eq3", Level: 2}}},
		{Name: "星熊", Id: "char_1004_amiya_1", Rarity: 6, Level: 85, EvolvePhase: 2, PotentialRank: 4, Skills: []Skill{{Id: "sk6", Level: 10}}, Equips: []Equip{}},
	}
}

func RenderBoxDetail(data []Detail) (*gg.Context, error) {
	const mainW = 900
	const cardH = 155
	const pad = 10
	mainH := pad + len(data)*cardH + pad
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 27, 29, 30)
	for i, d := range data {
		y := pad + i*cardH
		fillRoundedCard(dc, float64(pad), float64(y), float64(mainW-2*pad), float64(cardH-10), 8, 12)
		// avatar placeholder rect
		dc.SetRGB255(80, 80, 90)
		dc.DrawRoundedRectangle(float64(pad+10), float64(y+10), 90, 90, 6)
		dc.Fill()
		setFont(dc, 20)
		dc.SetRGB255(255, 255, 255)
		drawString(dc, d.Name, float64(pad+115), float64(y+30))
		ev := tryLocal(fmt.Sprintf("box/Evolve_%d.png", d.EvolvePhase))
		dc.DrawImage(ScaleExact(ev, 22, 22), pad+115, y+40)
		pot := tryLocal(fmt.Sprintf("box/Potential_%d.png", d.PotentialRank))
		dc.DrawImage(ScaleExact(pot, 22, 22), pad+145, y+40)

		sx := pad + 115
		sy := y + 78
		setFont(dc, 13)
		dc.SetRGB255(200, 210, 230)
		drawString(dc, "技能", float64(sx), float64(sy))
		skx := sx + 40
		for _, s := range d.Skills {
			dc.SetRGB255(60, 70, 80)
			dc.DrawRoundedRectangle(float64(skx), float64(sy-22), 30, 30, 4)
			dc.Fill()
			setFont(dc, 11)
			dc.SetRGB255(230, 230, 230)
			drawStringAnchored(dc, "Lv"+itoa(s.Level), float64(skx+15), float64(sy+14), 0.5, 0.5)
			skx += 42
		}
		ey := y + 118
		dc.SetRGB255(200, 210, 230)
		setFont(dc, 13)
		drawString(dc, "模组", float64(sx), float64(ey))
		ekx := sx + 40
		for _, e := range d.Equips {
			dc.SetRGB255(60, 70, 80)
			dc.DrawRoundedRectangle(float64(ekx), float64(ey-22), 30, 30, 4)
			dc.Fill()
			setFont(dc, 11)
			dc.SetRGB255(230, 230, 230)
			drawStringAnchored(dc, "Lv"+itoa(e.Level), float64(ekx+15), float64(ey+14), 0.5, 0.5)
			ekx += 42
		}
	}
	return dc, nil
}

// BoxSummary

type MissingChar struct {
	SkinId, Name string
	Rarity       int
	Profession   string
}
type BoxSummary struct {
	Name                                                                                 string
	AllCharCnt, Star6CharCnt, Star5CharCnt, Star4CharCnt                                 string
	AllEvolvePhase2Cnt, Star6EvolvePhase2Cnt, Star5EvolvePhase2Cnt, Star4EvolvePhase2Cnt int
	AllSkill10Cnt, Star6Skill10Cnt, Star5Skill10Cnt, Star4Skill10Cnt                     int
	AllSkill9Cnt, Star6Skill9Cnt, Star5Skill9Cnt, Star4Skill9Cnt                         int
	AllSkill8Cnt, Star6Skill8Cnt, Star5Skill8Cnt, Star4Skill8Cnt                         int
	AllEquipStage3Cnt, Star6EquipStage3Cnt, Star5EquipStage3Cnt, Star4EquipStage3Cnt     int
	AllEquipStage2Cnt, Star6EquipStage2Cnt, Star5EquipStage2Cnt, Star4EquipStage2Cnt     int
	AllEquipStage1Cnt, Star6EquipStage1Cnt, Star5EquipStage1Cnt, Star4EquipStage1Cnt     int
	MissingChars                                                                         []MissingChar
}

func SampleBoxSummary() *BoxSummary {
	return &BoxSummary{
		Name: "博士的干员总览", AllCharCnt: "120", Star6CharCnt: "40", Star5CharCnt: "50", Star4CharCnt: "30",
		AllEvolvePhase2Cnt: 80, Star6EvolvePhase2Cnt: 35, Star5EvolvePhase2Cnt: 35, Star4EvolvePhase2Cnt: 10,
		AllSkill10Cnt: 70, Star6Skill10Cnt: 30, Star5Skill10Cnt: 30, Star4Skill10Cnt: 10,
		AllSkill9Cnt: 90, Star6Skill9Cnt: 38, Star5Skill9Cnt: 40, Star4Skill9Cnt: 12,
		AllSkill8Cnt: 100, Star6Skill8Cnt: 40, Star5Skill8Cnt: 45, Star4Skill8Cnt: 15,
		AllEquipStage3Cnt: 50, Star6EquipStage3Cnt: 30, Star5EquipStage3Cnt: 18, Star4EquipStage3Cnt: 2,
		AllEquipStage2Cnt: 80, Star6EquipStage2Cnt: 38, Star5EquipStage2Cnt: 30, Star4EquipStage2Cnt: 12,
		AllEquipStage1Cnt: 100, Star6EquipStage1Cnt: 40, Star5EquipStage1Cnt: 40, Star4EquipStage1Cnt: 20,
		MissingChars: []MissingChar{
			{SkinId: "char_2001_x_1", Name: "乌有", Rarity: 6, Profession: "WARRIOR"},
			{SkinId: "char_2002_x_1", Name: "傀影", Rarity: 6, Profession: "CASTER"},
			{SkinId: "char_2003_x_1", Name: "温蒂", Rarity: 6, Profession: "PIONEER"},
			{SkinId: "char_2004_x_1", Name: "煌", Rarity: 6, Profession: "WARRIOR"},
			{SkinId: "char_2005_x_1", Name: "阿", Rarity: 5, Profession: "MEDIC"},
		},
	}
}

func RenderBoxSummary(data *BoxSummary) (*gg.Context, error) {
	const mainW = 900
	headerH := 60
	tableTop := headerH + 10
	rowH := 34
	rows := 8
	tableH := rows * rowH
	missingTop := tableTop + tableH + 20
	cols := 10
	mrows := (len(data.MissingChars) + cols - 1) / cols
	if mrows < 1 {
		mrows = 1
	}
	mtileH := 140
	mainH := missingTop + mrows*mtileH + 30
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 27, 29, 30)
	dc.SetRGB255(60, 62, 64)
	dc.DrawRectangle(0, 0, float64(mainW), float64(headerH))
	dc.Fill()
	setFont(dc, 26)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, data.Name, 25, 40)
	headers := []string{"指标", "总览", "6星", "5星", "4星"}
	colX := []float64{20, 300, 460, 600, 740}
	setFont(dc, 15)
	dc.SetRGB255(180, 200, 220)
	for i, h := range headers {
		drawString(dc, h, colX[i], float64(tableTop)+20)
	}
	metrics := []struct {
		name              string
		total, s6, s5, s4 int
	}{
		{"干员数", atoiSafe(data.AllCharCnt), atoiSafe(data.Star6CharCnt), atoiSafe(data.Star5CharCnt), atoiSafe(data.Star4CharCnt)},
		{"精英2", data.AllEvolvePhase2Cnt, data.Star6EvolvePhase2Cnt, data.Star5EvolvePhase2Cnt, data.Star4EvolvePhase2Cnt},
		{"技能10", data.AllSkill10Cnt, data.Star6Skill10Cnt, data.Star5Skill10Cnt, data.Star4Skill10Cnt},
		{"技能9", data.AllSkill9Cnt, data.Star6Skill9Cnt, data.Star5Skill9Cnt, data.Star4Skill9Cnt},
		{"技能8", data.AllSkill8Cnt, data.Star6Skill8Cnt, data.Star5Skill8Cnt, data.Star4Skill8Cnt},
		{"模组3", data.AllEquipStage3Cnt, data.Star6EquipStage3Cnt, data.Star5EquipStage3Cnt, data.Star4EquipStage3Cnt},
		{"模组2", data.AllEquipStage2Cnt, data.Star6EquipStage2Cnt, data.Star5EquipStage2Cnt, data.Star4EquipStage2Cnt},
		{"模组1", data.AllEquipStage1Cnt, data.Star6EquipStage1Cnt, data.Star5EquipStage1Cnt, data.Star4EquipStage1Cnt},
	}
	for ri, m := range metrics {
		ry := float64(tableTop) + float64(ri)*float64(rowH) + float64(rowH) - 6
		if ri%2 == 0 {
			fillRoundedCard(dc, 15, float64(tableTop)+float64(ri)*float64(rowH)+4, float64(mainW-30), float64(rowH)-4, 4, 10)
		}
		setFont(dc, 14)
		dc.SetRGB255(220, 220, 220)
		drawString(dc, m.name, colX[0], ry)
		dc.SetRGB255(230, 230, 230)
		drawString(dc, itoa(m.total), colX[1], ry)
		dc.SetRGB255(240, 180, 40)
		drawString(dc, itoa(m.s6), colX[2], ry)
		dc.SetRGB255(170, 110, 220)
		drawString(dc, itoa(m.s5), colX[3], ry)
		dc.SetRGB255(90, 160, 230)
		drawString(dc, itoa(m.s4), colX[4], ry)
	}
	SectionTitle(dc, "缺失干员", 20, float64(missingTop)-6)
	for i, c := range data.MissingChars {
		x := (i % cols) * 70
		y := missingTop + (i/cols)*mtileH
		DrawPortraitTile(dc, x, y, 70, mtileH, c.SkinId, c.Profession, c.Rarity, 0, c.Name)
	}
	return dc, nil
}

// Enemy
type EnemySkill struct{ Name, SpInit, SpCost, Desc string }
type EnemyLevel struct {
	Desc, AttackType, Motion, HpRecovery, HP, ATK, DEF, Res, ATKRadius, Weight, MoveSpeed, Interval, DamageRes, ElementRes, Ridicule, Point, Abnormal string
	Skills                                                                                                                                            []EnemySkill
	Talent                                                                                                                                            string
}
type Enemy struct {
	Name, Pic, Desc, EnemyRace, EnemyLevel, AttackType, Motion string
	Ability                                                    string
	Levels                                                     []EnemyLevel
}

func SampleEnemy() *Enemy {
	return &Enemy{
		Name: "霜星", Pic: "", Desc: "雪怪小队领袖，擅长冰属性法术。", EnemyRace: "人类", EnemyLevel: "精英", AttackType: "法术", Motion: "地面",
		Ability: "攻击造成法术伤害，并施加寒冷。",
		Levels:  []EnemyLevel{{HP: "12000", ATK: "850", DEF: "300", Res: "40", Talent: "攻击范围内我方单位移动速度降低。", Skills: []EnemySkill{{Name: "冰封", SpInit: "10", SpCost: "30", Desc: "对范围内单位造成大量法术伤害并冻结。"}}}},
	}
}

func RenderEnemy(data *Enemy) (*gg.Context, error) {
	const mainW = 656
	const pad = 16
	headerH := 160
	levelH := 260
	mainH := headerH + len(data.Levels)*levelH + 40
	if mainH < 400 {
		mainH = 400
	}
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 27, 29, 30)
	// pic placeholder
	dc.SetRGB255(70, 70, 80)
	dc.DrawRoundedRectangle(float64(pad), float64(pad), 120, 120, 8)
	dc.Fill()
	setFont(dc, 22)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, data.Name, 150, 40)
	setFont(dc, 14)
	dc.SetRGB255(200, 200, 200)
	drawString(dc, StripHTML(data.Desc), 150, 70)
	info := fmt.Sprintf("种族:%s  等级:%s  攻击:%s  移动:%s", data.EnemyRace, data.EnemyLevel, data.AttackType, data.Motion)
	drawString(dc, info, 150, 95)
	y := headerH
	for _, lv := range data.Levels {
		fillRoundedCard(dc, float64(pad), float64(y), float64(mainW-2*pad), float64(levelH-10), 8, 15)
		row := func(label, val string, ry float64) {
			setFont(dc, 14)
			dc.SetRGB255(150, 150, 150)
			drawString(dc, label, float64(pad+10), ry)
			dc.SetRGB255(230, 230, 230)
			drawString(dc, val, float64(pad+110), ry)
		}
		row("HP", lv.HP, float64(y+24))
		row("ATK", lv.ATK, float64(y+46))
		row("DEF", lv.DEF, float64(y+68))
		row("RES", lv.Res, float64(y+90))
		setFont(dc, 13)
		dc.SetRGB255(180, 220, 200)
		drawString(dc, "特性: "+StripHTML(lv.Talent), float64(pad+10), float64(y+120))
		drawString(dc, "能力: "+StripHTML(data.Ability), float64(pad+10), float64(y+142))
		sy := float64(y + 170)
		for _, sk := range lv.Skills {
			setFont(dc, 13)
			dc.SetRGB255(230, 210, 150)
			drawString(dc, fmt.Sprintf("技能 %s (技力%s/%s): %s", sk.Name, sk.SpInit, sk.SpCost, StripHTML(sk.Desc)), float64(pad+10), sy)
			sy += 22
		}
		y += levelH
	}
	return dc, nil
}

// Headhunt
type HHOp struct {
	Rarity     int
	ThumbURL   string
	Profession string
}
type HeadhuntData struct{ Ops []HHOp }

func SampleHeadhunt() []HHOp {
	ops := make([]HHOp, 0, 10)
	for i := 0; i < 10; i++ {
		ops = append(ops, HHOp{Rarity: 3 + i%4, ThumbURL: "", Profession: "WARRIOR"})
	}
	return ops
}

func RenderHeadhunt(data []HHOp) (*gg.Context, error) {
	const mainW, mainH = 1024, 576
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 27, 29, 30)
	n := len(data)
	if n < 1 {
		n = 1
	}
	tileW := mainW / n
	if tileW > 120 {
		tileW = 120
	}
	startX := (mainW - tileW*n) / 2
	cy := mainH/2 - 90
	for i, o := range data {
		x := startX + i*tileW
		// back per rarity color
		r, g, b := rarityColor(o.Rarity)
		dc.SetRGB255(r, g, b)
		dc.DrawRectangle(float64(x), float64(cy), float64(tileW), 180)
		dc.Fill()
		DrawPortraitTile(dc, x, cy, tileW, 180, o.ThumbURL, o.Profession, o.Rarity, 0, "")
	}
	return dc, nil
}

// Missing
type MissingInfo struct {
	Name  string
	Chars []MissingChar
}

func SampleMissing() *MissingInfo {
	chars := make([]MissingChar, 0, 12)
	for i := 0; i < 12; i++ {
		chars = append(chars, MissingChar{SkinId: fmt.Sprintf("char_%03d_x_%d", i, i%3+1), Name: fmt.Sprintf("缺失%d", i+1), Rarity: 3 + i%4, Profession: "WARRIOR"})
	}
	return &MissingInfo{Name: "博士 的未拥有干员", Chars: chars}
}

func RenderMissing(data *MissingInfo) (*gg.Context, error) {
	const mainW = 700
	tileW, tileH := 70, 140
	cols := 10
	rows := (len(data.Chars) + cols - 1) / cols
	if rows < 1 {
		rows = 1
	}
	labelH := 60
	gridTop := labelH + 10
	mainH := gridTop + rows*tileH + 20
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 46, 48, 49)
	dc.SetRGB255(60, 62, 64)
	dc.DrawRectangle(0, 0, float64(mainW), float64(labelH))
	dc.Fill()
	setFont(dc, 26)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, data.Name, 25, float64(labelH)-18)
	for i, c := range data.Chars {
		x := (i % cols) * tileW
		y := gridTop + (i/cols)*tileH
		DrawPortraitTile(dc, x, y, tileW, tileH, c.SkinId, c.Profession, c.Rarity, 0, c.Name)
	}
	return dc, nil
}

// Recruit
type RecruitOp struct {
	Avatar, Profession string
	Rarity             int
}
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
	const mainW = 900
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
	rows := (len(data.Operators) + cols - 1) / cols
	if rows < 1 {
		rows = 1
	}
	mainH := int(tagY+10) + rows*tileH + pad
	dc := gg.NewContext(mainW, mainH)
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

// ---------- remaining 9 scenes ----------

// Card
type CardInfo struct {
	Name, Uid, ServerName, Resume                             string
	Level, RegTime                                            int
	MainStageProgress, Avatar, SecretaryName, SecretaryEnName string
	CharCnt, FurnitureCnt, SkinCnt, EquipCnt                  int
}

func SampleCard() *CardInfo {
	return &CardInfo{
		Name: "博士", Uid: "10000001", ServerName: "官服", Resume: "罗德岛的博士，今日也在努力。",
		Level: 120, RegTime: 1620000000, MainStageProgress: "12-18", Avatar: "", SecretaryName: "阿米娅", SecretaryEnName: "Amiya",
		CharCnt: 280, FurnitureCnt: 320, SkinCnt: 120, EquipCnt: 80,
	}
}

func RenderCard(data *CardInfo) (*gg.Context, error) {
	const mainW = 1000
	const mainH = 600
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 27, 29, 30)
	// top decor bar
	dc.SetRGB255(45, 48, 55)
	dc.DrawRectangle(0, 0, float64(mainW), 90)
	dc.Fill()
	// avatar circle
	dc.SetRGB255(80, 80, 90)
	dc.DrawCircle(60, 45, 38)
	dc.Fill()
	setFont(dc, 22)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, data.Name, 110, 36)
	setFont(dc, 13)
	dc.SetRGB255(180, 200, 220)
	drawString(dc, fmt.Sprintf("UID %s · %s · Lv%d", data.Uid, data.ServerName, data.Level), 110, 58)
	drawString(dc, "主线 "+data.MainStageProgress, 110, 76)
	// stats row
	stats := []struct {
		label string
		val   int
	}{
		{"干员", data.CharCnt}, {"家具", data.FurnitureCnt}, {"时装", data.SkinCnt}, {"模组", data.EquipCnt},
	}
	x := 20
	y := 140
	for _, s := range stats {
		dc.SetRGBA255(255, 255, 255, 12)
		RoundRect(dc, float64(x), float64(y), 140, 80, 8)
		setFont(dc, 15)
		dc.SetRGB255(180, 200, 220)
		drawStringAnchored(dc, s.label, float64(x+70), float64(y+24), 0.5, 0.5)
		setFont(dc, 24)
		dc.SetRGB255(255, 255, 255)
		drawStringAnchored(dc, itoa(s.val), float64(x+70), float64(y+54), 0.5, 0.5)
		x += 160
	}
	// secretary
	y = 260
	fillRoundedCard(dc, 20, float64(y), float64(mainW-40), 120, 10, 14)
	setFont(dc, 16)
	dc.SetRGB255(255, 230, 160)
	drawString(dc, "秘书 "+data.SecretaryName+" / "+data.SecretaryEnName, 30, float64(y+36))
	setFont(dc, 13)
	dc.SetRGB255(200, 200, 200)
	drawString(dc, StripHTML(data.Resume), 30, float64(y+64))
	// resume icon placeholder
	dc.SetRGB255(70, 70, 80)
	dc.DrawCircle(float64(mainW-80), float64(y+60), 36)
	dc.Fill()
	return dc, nil
}

// ponytail: depot split to depot.go for per-scene ownership (renderContext routes there)
// DepotData/DepotItem/SampleDepot/RenderDepot moved to src/ggrender/depot.go

// Gacha
type GachaData struct {
	Name                              string
	Total, Star6, Star5, Star4, Star3 int
	Avg6, Avg5, Avg4, Avg3            float64
	Chars                             []struct {
		PoolName, CharName string
		Rarity             int64
		IsNew              bool
	}
}

func SampleGacha() *GachaData {
	g := &GachaData{Name: "博士的寻访记录", Total: 120, Star6: 6, Star5: 18, Star4: 50, Star3: 46, Avg6: 20.0, Avg5: 6.6, Avg4: 2.4, Avg3: 2.6}
	names := []string{"能天使", "银灰", "艾雅法拉", "星熊", "塞雷娅", "闪灵", "夜莺", "斯卡蒂", "陈", "推进之王"}
	for i, n := range names {
		r := int64(5)
		if i%3 == 0 {
			r = 4
		}
		if i%5 == 0 {
			r = 3
		}
		g.Chars = append(g.Chars, struct {
			PoolName, CharName string
			Rarity             int64
			IsNew              bool
		}{PoolName: "常驻", CharName: n, Rarity: r, IsNew: i%4 == 0})
	}
	return g
}

func RenderGacha(data *GachaData) (*gg.Context, error) {
	const mainW = 900
	headerH := 90
	statsH := 120
	charsH := 20 * 74
	mainH := headerH + statsH + charsH + 40
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 27, 29, 30)
	// header
	dc.SetRGB255(45, 48, 55)
	dc.DrawRectangle(0, 0, float64(mainW), float64(headerH))
	dc.Fill()
	setFont(dc, 24)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, data.Name, 25, 52)
	setFont(dc, 14)
	dc.SetRGB255(180, 200, 220)
	drawString(dc, fmt.Sprintf("共 %d 抽 · 6星%d 5星%d 4星%d 3星%d", data.Total, data.Star6, data.Star5, data.Star4, data.Star3), 25, 74)
	// stats avg
	y := headerH + 16
	stats := []struct {
		label string
		val   float64
		cnt   int
	}{
		{"6星", data.Avg6, data.Star6}, {"5星", data.Avg5, data.Star5}, {"4星", data.Avg4, data.Star4}, {"3星", data.Avg3, data.Star3},
	}
	x := 20
	for _, s := range stats {
		dc.SetRGBA255(255, 255, 255, 12)
		RoundRect(dc, float64(x), float64(y), 200, 80, 8)
		setFont(dc, 14)
		dc.SetRGB255(180, 200, 220)
		drawStringAnchored(dc, s.label, float64(x+100), float64(y+24), 0.5, 0.5)
		setFont(dc, 22)
		dc.SetRGB255(255, 240, 120)
		drawStringAnchored(dc, fmt.Sprintf("%.1f", s.val), float64(x+100), float64(y+52), 0.5, 0.5)
		setFont(dc, 11)
		dc.SetRGB255(200, 200, 200)
		drawStringAnchored(dc, fmt.Sprintf("(%d)", s.cnt), float64(x+100), float64(y+68), 0.5, 0.5)
		x += 220
	}
	y += 110
	// chars list
	setFont(dc, 14)
	dc.SetRGB255(200, 220, 200)
	drawString(dc, "最近获得", 20, float64(y))
	y += 20
	for i, ch := range data.Chars {
		if i >= 20 {
			break
		}
		yy := y + i*74
		fillRoundedCard(dc, 20, float64(yy), float64(mainW-40), 64, 8, 10)
		// avatar
		dc.SetRGB255(80, 80, 90)
		dc.DrawCircle(50, float64(yy+32), 22)
		dc.Fill()
		setFont(dc, 14)
		dc.SetRGB255(255, 255, 255)
		drawString(dc, ch.CharName, 80, float64(yy+24))
		setFont(dc, 12)
		dc.SetRGB255(180, 200, 220)
		drawString(dc, ch.PoolName, 80, float64(yy+44))
		// rarity color bar
		r, g, b := rarityColor(int(ch.Rarity + 1))
		dc.SetRGB255(r, g, b)
		dc.DrawRectangle(float64(mainW-80), float64(yy+20), 50, 24)
		dc.Fill()
		setFont(dc, 12)
		dc.SetRGB255(255, 255, 255)
		drawStringAnchored(dc, fmt.Sprintf("%d★", ch.Rarity+1), float64(mainW-55), float64(yy+32), 0.5, 0.5)
		if ch.IsNew {
			setFont(dc, 10)
			dc.SetRGB255(255, 80, 80)
			drawString(dc, "NEW", float64(mainW-140), float64(yy+32))
		}
	}
	return dc, nil
}

// Help
type HelpData struct {
	Private []Cmd
	Public  []Cmd
	Admin   []Cmd
}
type Cmd struct {
	Cmd, Desc, Param string
	IsBind           bool
}

func SampleHelp() *HelpData {
	h := &HelpData{}
	h.Private = []Cmd{{Cmd: "/bind", Desc: "绑定角色", Param: ""}, {Cmd: "/unbind", Desc: "解绑角色", Param: ""}, {Cmd: "/cancel", Desc: "取消操作", Param: ""}}
	h.Public = []Cmd{
		{Cmd: "/help", Desc: "使用说明", Param: ""},
		{Cmd: "/box", Desc: "我的干员", Param: ""},
		{Cmd: "/state", Desc: "当前状态", Param: ""},
		{Cmd: "/card", Desc: "我的名片", Param: ""},
		{Cmd: "/base", Desc: "基建信息", Param: ""},
		{Cmd: "/gacha", Desc: "抽卡记录", Param: ""},
		{Cmd: "/depot", Desc: "我的仓库", Param: ""},
		{Cmd: "/calendar", Desc: "活动日历", Param: ""},
		{Cmd: "/recruit", Desc: "公招计算", Param: ""},
		{Cmd: "/headhunt", Desc: "寻访模拟", Param: ""},
	}
	h.Admin = []Cmd{{Cmd: "/news", Desc: "动态推送", Param: ""}, {Cmd: "/birthday", Desc: "生日推送", Param: ""}}
	return h
}

func RenderHelp(data *HelpData) (*gg.Context, error) {
	const mainW = 990
	// heights: header 200 + sections
	privH := 40 + len(data.Private)*32
	pubH := 40 + len(data.Public)*32
	adminH := 40 + len(data.Admin)*32
	mainH := 200 + privH + pubH + adminH + 60
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 46, 48, 49)
	// banner placeholder
	dc.SetRGB255(60, 62, 80)
	dc.DrawRectangle(0, 0, float64(mainW), 140)
	dc.Fill()
	setFont(dc, 28)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, "Arknights Bot · 使用说明", 30, 80)
	setFont(dc, 14)
	dc.SetRGB255(200, 220, 255)
	drawString(dc, "基于森空岛数据的罗德岛助手", 30, 110)
	y := 160
	drawSection := func(title string, cmds []Cmd, yy int) int {
		setFont(dc, 16)
		dc.SetRGB255(120, 200, 220)
		drawString(dc, title, 20, float64(yy))
		yy += 20
		for _, c := range cmds {
			dc.SetRGBA255(255, 255, 255, 10)
			RoundRect(dc, 20, float64(yy), float64(mainW-40), 28, 6)
			setFont(dc, 13)
			dc.SetRGB255(255, 230, 120)
			drawString(dc, c.Cmd, 30, float64(yy+18))
			dc.SetRGB255(200, 200, 200)
			drawString(dc, c.Desc, 160, float64(yy+18))
			if c.Param != "" {
				dc.SetRGB255(160, 180, 200)
				drawString(dc, c.Param, 300, float64(yy+18))
			}
			yy += 32
		}
		return yy + 10
	}
	y = drawSection("私聊指令", data.Private, y)
	y = drawSection("群聊指令", data.Public, y)
	y = drawSection("管理员指令", data.Admin, y)
	return dc, nil
}

// State
type StateInfo struct {
	Name           string
	Level          int
	ApCur, ApTotal int
	SanityExpire   string
	Chars          []struct {
		Name string
		Ap   int
	}
}

func SampleState() *StateInfo {
	s := &StateInfo{Name: "博士", Level: 120, ApCur: 135, ApTotal: 135, SanityExpire: "12:30"}
	s.Chars = []struct {
		Name string
		Ap   int
	}{{"阿米娅", 22}, {"凯尔希", 20}, {"煌", 18}, {"能天使", 16}}
	return s
}

func RenderState(data *StateInfo) (*gg.Context, error) {
	const mainW = 1092
	const mainH = 510
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 46, 48, 49)
	// try draw bg image if exists
	if bg, err := LoadImage(AssetPath("state/bg.png")); err == nil {
		dc.DrawImage(ScaleCover(bg, mainW, mainH), 0, 0)
	}
	// avatar
	dc.SetRGB255(80, 80, 90)
	dc.DrawCircle(34+27, 34+27, 27)
	dc.Fill()
	setFont(dc, 26)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, data.Name, 100, 60)
	setFont(dc, 16)
	dc.SetRGB255(220, 220, 220)
	drawString(dc, fmt.Sprintf("Lv%d", data.Level), 100, 84)
	// ap bar
	frac := float64(data.ApCur) / float64(data.ApTotal)
	if data.ApTotal == 0 {
		frac = 1
	}
	ProgressBar(dc, 146, 146, 410, 11, frac, 255, 255, 255)
	setFont(dc, 28)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, fmt.Sprintf("%d/%d", data.ApCur, data.ApTotal), 146, 140)
	setFont(dc, 16)
	dc.SetRGB255(200, 200, 200)
	drawString(dc, "理智恢复 "+data.SanityExpire, 146, 170)
	// chars row
	x := 20
	y := 260
	for _, c := range data.Chars {
		dc.SetRGBA255(255, 255, 255, 18)
		RoundRect(dc, float64(x), float64(y), 120, 120, 10)
		dc.SetRGB255(80, 80, 90)
		dc.DrawCircle(float64(x+60), float64(y+44), 28)
		dc.Fill()
		setFont(dc, 12)
		dc.SetRGB255(255, 255, 255)
		drawStringAnchored(dc, c.Name, float64(x+60), float64(y+88), 0.5, 0.5)
		// ap badge
		dc.SetRGBA255(0, 0, 0, 160)
		dc.DrawCircle(float64(x+60), float64(y+108), 14)
		dc.Fill()
		setFont(dc, 11)
		dc.SetRGB255(255, 230, 80)
		drawStringAnchored(dc, itoa(c.Ap), float64(x+60), float64(y+112), 0.5, 0.5)
		x += 140
	}
	// campaign / tradings placeholders
	y = 410
	for i := 0; i < 3; i++ {
		dc.SetRGBA255(255, 255, 255, 10)
		RoundRect(dc, float64(20+i*260), float64(y), 240, 60, 8)
		setFont(dc, 13)
		dc.SetRGB255(200, 220, 200)
		drawString(dc, fmt.Sprintf("区块 %d", i+1), float64(40+i*260), float64(y+36))
	}
	return dc, nil
}
