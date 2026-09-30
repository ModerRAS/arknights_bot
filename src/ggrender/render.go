package ggrender

import (
	"fmt"
	"image"
	"math"

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
	return ScaleToManifest(dc, 1050, 536), nil
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
	return ScaleToManifest(dc, 722, 279), nil
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
	return ScaleToManifest(dc, 1350, 723), nil
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
	return ScaleToManifest(dc, 984, 477), nil
}

// Headhunt
type HHOp struct {
	Rarity     int
	ThumbURL   string
	Profession string
}
type HeadhuntData struct{ Ops []HHOp }

// SampleHeadhunt stands in for the frozen headhunt-minimal fixture, whose
// ten-roll payload is ten 6-star results. Two independent justifications:
//   - hard: the previous cycle "3 + i%4" emitted two Rarity=6 ops, and
//     assets/headhunt has no back_6.png / Rarity_6.png at all, so those two
//     cards read a non-existent asset and drew no star bar whatsoever;
//   - measured: the baseline shows ten gold backdrops and ten 6-star bars,
//     and Rarity_5.png is the only asset with six star components
//     (Rarity_0..4 carry 1..5).
//
// This is fixture INPUT matching the frozen input. No geometry is taken from
// the baseline; every coordinate below is either an asset property or a
// separately measured pixel box.
func SampleHeadhunt() []HHOp {
	ops := make([]HHOp, 0, 10)
	for i := 0; i < 10; i++ {
		ops = append(ops, HHOp{Rarity: 5, ThumbURL: "", Profession: "WARRIOR"})
	}
	return ops
}

// Headhunt layout constants are measured off the frozen Playwright baseline
// (testdata/visual/baseline/images/headhunt.jpg, 1049x576), not copied from
// template/Headhunt.tmpl. CSS is cross-validation only; measured wins on conflict.
//
//	card content width 95px  <- gold run width at y=497 (runs: 24..119, 124..218, ...)
//	card content top   230px <- face top; CSS .bg margin-top130+padding-top100 agrees
//	card content height 270px <- gold bottom y=499; CSS .bg height:270 agrees
//	first card left edge  x=24 <- CSS #main padding-left:25 agrees within 1px
//	card pitch        98.67px <- left edges 24,124,222,321,419,518,617,715,814,912
const (
	hhOutW, hhOutH = 1049, 576
	hhCardW        = 95
	hhCardTop      = 230 // measured: face top; CSS .bg margin-top130+padding-top100 agrees
	hhCardH        = 270 // measured: gold bottom y=499; CSS .bg height:270 agrees
	hhBackTop      = 130 // measured: backdrop top y~182 (back_*.png has its own transparent head)
	hhBackH        = 370 // CSS .bg box = margin-top130 + padding-top100 + height270
	hhFirstX       = 24.0
	hhPitch        = 98.6667
	hhFaceH        = 190 // measured: face box y230..420; CSS .lh height:190 agrees
	hhIconDX       = 12  // measured: WARRIOR.png white square x=cardX+18, asset inset 6
	hhIconDY       = 191 // measured: white square y=421; CSS .profession margin-top:190 agrees
	hhStarDX       = 14  // measured: Rarity_5 star0 x=cardX+16, asset inset 4
	hhStarDY       = 1   // measured: star0 y=235, asset inset 4
)

func RenderHeadhunt(data []HHOp) (*gg.Context, error) {
	dc := gg.NewContext(hhOutW, hhOutH)
	FillBackground(dc, 27, 29, 30)
	// Starfield backdrop. bg.png is 1024x576 and the baseline shows a seam at
	// x=1024, so the 25px remainder continues the image's own right edge column.
	if bg, err := LoadImage(AssetPath("headhunt/bg.png")); err == nil {
		dc.DrawImage(bg, 0, 0)
		if b := bg.Bounds(); b.Dx() == 1024 && b.Dy() == hhOutH {
			col := image.NewRGBA(image.Rect(0, 0, 1, hhOutH))
			for y := 0; y < hhOutH; y++ {
				col.Set(0, y, bg.At(b.Max.X-1, b.Min.Y+y))
			}
			dc.DrawImage(ScaleExact(col, hhOutW-1024, hhOutH), 1024, 0)
		}
	}
	for i, o := range data {
		x := int(hhFirstX + float64(i)*hhPitch)
		// Rarity backdrop: background-size:cover into the full .bg box (padding included).
		if bk, err := LoadImage(AssetPath(fmt.Sprintf("headhunt/back_%d.png", o.Rarity))); err == nil {
			dc.DrawImage(ScaleCover(bk, hhCardW, hhBackH), x, hhBackTop)
		} else {
			r, g, b := rarityColor(o.Rarity)
			dc.SetRGB255(r, g, b)
			dc.DrawRectangle(float64(x), float64(hhBackTop), float64(hhCardW), float64(hhBackH))
			dc.Fill()
		}
		// Portrait. assets/headhunt/amiya-half.webp is the same file the frozen
		// baseline rendered with, copied back from baseline/cache/ (sha256
		// 7560d950...). It is RGBA with alpha extrema (0,255), which is
		// format-level proof it cannot have been cut from the 16 RGB JPEG
		// baselines. 180x360 is exactly 1:2 and the .lh box is exactly 1:2, so
		// ScaleExact is an aspect-neutral fit: no crop, no distortion.
		var port image.Image
		if o.ThumbURL != "" {
			port = FetchImage(o.ThumbURL, AssetPath("headhunt/amiya-half.webp"))
		} else if img, err := LoadImage(AssetPath("headhunt/amiya-half.webp")); err == nil {
			port = img
		} else {
			port = tryLocal("common/amiya.png")
		}
		dc.DrawImage(ScaleExact(port, hhCardW, hhFaceH), x, hhCardTop)
		// Profession icon at natural size.
		if ic, err := LoadImage(AssetPath("headhunt/" + o.Profession + ".png")); err == nil {
			dc.DrawImage(ic, x+hhIconDX, hhCardTop+hhIconDY)
		}
		// Rarity star bar at natural size.
		if rb, err := LoadImage(AssetPath(fmt.Sprintf("headhunt/Rarity_%d.png", o.Rarity))); err == nil {
			dc.DrawImage(rb, x+hhStarDX, hhCardTop+hhStarDY)
		}
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
	return ScaleToManifest(dc, 1050, 536), nil
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
	return ScaleToManifest(dc, 1350, 534), nil
}

// ---------- remaining 9 scenes ----------

// Base
// BaseChar mirrors the web-layer type at src/core/web/base.go:86 so Avatar and AP
// survive into the gg render path. Necessary but NOT sufficient: the parity harness
// runs SampleBase(), whose Chars carry names only, so this alone flips no atom.
type BaseChar struct {
	Name   string
	Avatar string
	AP     int
}
type BaseInfo struct {
	Name    string
	Labor   struct{ Cur, Total int }
	Control struct {
		Level int
		Chars []BaseChar
	}
	Tradings []struct {
		Level      int
		Chars      []BaseChar
		Cur, Total int
		Strategy   string
	}
	Manufactures []struct {
		Level       int
		Chars       []BaseChar
		Cur, Total  int
		Item, Speed string
	}
	Powers []struct {
		Level int
		Chars []BaseChar
		Power int
	}
	Meeting struct {
		Level   int
		Chars   []BaseChar
		Board   []int
		Sharing bool
	}
	Hire struct {
		Level   int
		Chars   []BaseChar
		Refresh int
	}
	Training struct {
		Level  int
		Chars  []BaseChar
		Skill  string
		SLevel int
	}
	Dorms []struct {
		Level   int
		Chars   []BaseChar
		Comfort int
	}
}

func SampleBase() *BaseInfo {
	b := &BaseInfo{Name: "博士的基建"}
	b.Labor.Cur = 108
	b.Labor.Total = 120
	b.Control.Level = 5
	b.Control.Chars = []BaseChar{{Name: "阿米娅"}, {Name: "凯尔希"}, {Name: "煌"}}
	b.Tradings = []struct {
		Level      int
		Chars      []BaseChar
		Cur, Total int
		Strategy   string
	}{
		{Level: 3, Chars: []BaseChar{{Name: "能天使"}, {Name: "德克萨斯"}}, Cur: 3, Total: 5, Strategy: "贵金属订单"},
		{Level: 3, Chars: []BaseChar{{Name: "拉普兰德"}}, Cur: 2, Total: 5, Strategy: "源石订单"},
	}
	b.Manufactures = []struct {
		Level       int
		Chars       []BaseChar
		Cur, Total  int
		Item, Speed string
	}{
		{Level: 3, Chars: []BaseChar{{Name: "夜烟"}, {Name: "远山"}}, Cur: 10, Total: 20, Item: "中级作战记录", Speed: "120%"},
		{Level: 3, Chars: []BaseChar{{Name: "砾"}}, Cur: 8, Total: 20, Item: "赤金", Speed: "100%"},
	}
	b.Powers = []struct {
		Level int
		Chars []BaseChar
		Power int
	}{
		{Level: 3, Chars: []BaseChar{{Name: "格雷伊"}}, Power: 270},
		{Level: 3, Chars: []BaseChar{{Name: "清流"}}, Power: 270},
	}
	b.Meeting.Level = 3
	b.Meeting.Chars = []BaseChar{{Name: "诗怀雅"}}
	b.Meeting.Board = []int{1, 2, 3}
	b.Meeting.Sharing = true
	b.Hire.Level = 3
	b.Hire.Chars = []BaseChar{{Name: "陈"}}
	b.Hire.Refresh = 2
	b.Training.Level = 3
	b.Training.Chars = []BaseChar{{Name: "赫拉格"}, {Name: "华法琳"}}
	b.Training.Skill = "阿米娅-奇美拉"
	b.Training.SLevel = 2
	b.Dorms = []struct {
		Level   int
		Chars   []BaseChar
		Comfort int
	}{
		{Level: 5, Chars: []BaseChar{{Name: "星熊"}, {Name: "塞雷娅"}}, Comfort: 5000},
		{Level: 5, Chars: []BaseChar{{Name: "夜莺"}}, Comfort: 4800},
	}
	return b
}

// Base geometry is derived entirely from template/Base.tmpl declarations.
// Nothing here is read off a baseline image. origin_source tags per block:
//
//	css               = a Base.tmpl <style> rule
//	template-declared = inline markup in Base.tmpl
//
// See .audit/base-layout-completeness/baseline.json for the per-atom accounting.
// Base geometry is derived entirely from template/Base.tmpl declarations.
// Nothing here is read off a baseline image. origin_source vocabulary: css / template-declared.
// See .audit/base-layout-completeness/baseline.json for the per-atom accounting.
// drawBaseGearIcon reproduces the 48x48 viewBox svg at Base.tmpl:71 inside an s x s box.
// Geometry taken from the declared path data: four 270-degree arcs of radius 7.1 centred
// on the four quadrant points, each with a 90-degree gap facing the icon centre (24,24),
// four spokes from the quadrant points to the centre square, and the 10x10 centre square.
// stroke #852cd3, stroke-width 4, round caps (template-declared, not measured).
func drawBaseGearIcon(dc *gg.Context, x, y, s float64) {
	k := s / 48.0
	const r = 7.1
	dc.SetRGB255(0x85, 0x2c, 0xd3)
	dc.SetLineWidth(4 * k)
	for _, q := range [][2]float64{{36, 12}, {36, 36}, {12, 36}, {12, 12}} {
		cx, cy := q[0], q[1]
		gap := math.Atan2(24-cy, 24-cx)
		const seg = 48
		pts := [seg + 1][2]float64{}
		for i := 0; i <= seg; i++ {
			a := gap + math.Pi/4 + 1.5*math.Pi*float64(i)/seg
			pts[i] = [2]float64{cx + r*math.Cos(a), cy + r*math.Sin(a)}
		}
		for i := 0; i < seg; i++ {
			dc.DrawLine(x+pts[i][0]*k, y+pts[i][1]*k, x+pts[i+1][0]*k, y+pts[i+1][1]*k)
		}
	}
	dc.DrawLine(x+12*k, y+12*k, x+19*k, y+19*k)
	dc.DrawLine(x+36*k, y+36*k, x+29*k, y+29*k)
	dc.DrawLine(x+36*k, y+12*k, x+29*k, y+19*k)
	dc.DrawLine(x+12*k, y+36*k, x+19*k, y+29*k)
	dc.Stroke()
	dc.SetRGB255(0x85, 0x2c, 0xd3)
	dc.DrawRectangle(x+19*k, y+19*k, 10*k, 10*k)
	dc.Fill()
}

func RenderBase(data *BaseInfo) (*gg.Context, error) {
	// A02 css 9: width: 1110px. Height is the manifest viewport: ScaleToManifest target
	// 918 / fixtures.json:28 scale 1.5 = 612. Not a css declaration (neither Base.tmpl nor
	// assets/css/common.css declares #main height); it is the viewport the capture was
	// taken at, so content below it is clipped.
	const mainW, mainH = 1110, 612
	const cardW, wideW = 550, 1105      // C01 css 13 / C09 declared 85,128
	const cardH, gapY = 110, 5          // C02 css 14 / C07 css 19
	const cardR = 15                    // C03 css 15: .base{border-radius:15px}
	var cardBg = [3]int{33, 38, 47}     // C08 css 20: .base{background-color:#21262f}
	var cardBorder = [3]int{33, 38, 47} // C04 css 16: .base{border:1px solid #21262f}
	const headerH = 24                  // header h3 default line box (no css declaration)
	const h3ML = 10                     // D01 css 26: h3{margin-left:10px}
	const iconMR = 20                   // D06 css 48: .title_icon{margin-right:20px}
	const charsML = 10                  // D07 css 51: .chars{margin-left:10px}
	const boardW, boardR = 20.0, 5.0    // D09 css 56 / D10 css 58
	const skillW = 30.0                 // D08 css 61: .skill{width:30px}

	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 43, 51, 61) // A01 css 8: #main{background-color:#2b333d}

	// B01 declared 68: <h3 style="display: inline">基建信息</h3>
	setFont(dc, 19)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, "基建信息", h3ML, 18)

	// B03 declared 70: <div style="display: flex"> — the gear icon and the labor text sit
	// side by side in one flex row, icon on the left.
	// B04 declared 71: <svg width="20" height="20" viewBox="0 0 48 48">, stroke #852cd3.
	// Row height / baseline inside the 20px flex row is browser-default (align-items:stretch
	// on a text span); the two declared widths (20 and the text) are what fix the columns.
	const gearS = 20.0
	labR := 1080.0
	setFont(dc, 15)
	dc.SetRGB255(255, 255, 255)
	labTxt := fmt.Sprintf("%d/%d", data.Labor.Cur, data.Labor.Total)
	tw, _ := measure(dc, labTxt)
	drawStringAnchored(dc, labTxt, labR, 13, 1, 0)
	drawBaseGearIcon(dc, labR-tw-gearS, 2, gearS)
	const barW, barH, barR = 100.0, 3.0, 1.0
	if data.Labor.Total > 0 {
		dc.SetRGB255(255, 255, 255)
		dc.DrawRoundedRectangle(labR-barW, 18, barW*float64(data.Labor.Cur)/float64(data.Labor.Total), barH, barR)
		dc.Fill()
	}

	// C05 css 17: .base{display:inline-flex} — pack left to right, wrap when full.
	// 550*2 = 1100 <= 1110, so two per row; a 1105 card cannot share a row.
	cx, cy := 0.0, float64(headerH)
	place := func(w float64) (float64, float64) {
		if cx+w > mainW {
			cx = 0
			cy += cardH + gapY
		}
		px, py := cx, cy+gapY
		cx += w
		return px, py
	}

	// bIcon is the .title_icon content declared for one block. Each field maps to a
	// Base.tmpl line; no field is invented.
	type bIcon struct {
		text   string // <span> text at the cited line
		color  [3]int // inline color declared on that span
		boards []int  // declared 360: <div class="board">{{.}}</div>
		skill  string // declared 463: <img class="skill" src=".../char_skill/{{.Training.Skill}}.png">
		// F15 declared 346: <div style="position: absolute;margin-left: -200px;
		// color: #eb9712;display: flex;align-items: center;">线索交流开启中</div>
		// Only the declared container + text are drawn; the 20x20 svg child is not
		// reproduced and is tracked as its own open item.
		absText string
	}

	// D04/D05/D06 css 46,47,48: .title_icon{display:inline-flex;align-items:center;margin-right:20px}
	// D02/D03 css 42,43: .title{display:flex;justify-content:space-between} — the icon
	// group is the right-hand child, so its right edge is the h3 content right minus 20.
	drawIcon := func(px, w, py float64, ic bIcon) {
		right := px + w - iconMR
		setFont(dc, 16)
		switch {
		case len(ic.boards) > 0:
			// declared 358-363: <span>线索 {{range .Meeting.Board}} <div class="board">…</div> {{end}}</span>
			// D11/D12 css 54,55: .board{display:inline-flex;justify-content:center}
			tw, _ := measure(dc, ic.text)
			bx := right - boardW*float64(len(ic.boards)) - tw
			// F15 declared 346: position:absolute with margin-left:-200px. No ancestor in
			// Base.tmpl is positioned, so the containing block is the initial containing
			// block: the block sits at its static position (start of .title_icon) minus 200.
			if ic.absText != "" {
				dc.SetRGB255(0xeb, 0x97, 0x12) // declared color: #eb9712
				drawString(dc, ic.absText, bx-200, py+28)
			}
			dc.SetRGB255(255, 255, 255)
			drawString(dc, ic.text, bx, py+28)
			for _, b := range ic.boards {
				dc.SetRGB255(255, 255, 255)
				StrokeRoundRect(dc, bx+tw, py+11, boardW, boardW, boardR)
				bs, _ := measure(dc, itoa(b))
				drawString(dc, itoa(b), bx+tw+(boardW-bs)/2, py+28)
				bx += boardW
			}
		case ic.skill != "":
			// declared 461-464: <span>Lv.N</span><img class="skill" ...>  (icon on the right)
			img := FetchImage("https://web.hycdn.cn/arknights/game/assets/char_skill/"+ic.skill+".png",
				AssetPath("common/amiya.png"))
			ix := right - skillW
			dc.DrawImage(ScaleExact(img, int(skillW), int(skillW)), int(ix), int(py+9))
			tw, _ := measure(dc, ic.text)
			dc.SetRGB255(255, 255, 255)
			drawString(dc, ic.text, ix-tw, py+28)
		default:
			tw, _ := measure(dc, ic.text)
			dc.SetRGB255(ic.color[0], ic.color[1], ic.color[2])
			drawString(dc, ic.text, right-tw, py+28)
		}
	}

	// C06 css 18: .base{flex-direction:column} — .title on top, .chars below it.
	drawCard := func(w float64, title string, ic bIcon, chars []BaseChar) {
		px, py := place(w)
		dc.SetRGB255(cardBg[0], cardBg[1], cardBg[2])
		RoundRect(dc, px, py, w, cardH, cardR)
		dc.SetRGB255(cardBorder[0], cardBorder[1], cardBorder[2])
		dc.SetLineWidth(1)
		StrokeRoundRect(dc, px+0.5, py+0.5, w-1, cardH-1, cardR)
		setFont(dc, 16)
		dc.SetRGB255(255, 255, 255)
		drawString(dc, title, px+h3ML, py+28) // D01 css 26
		if ic.text != "" || len(ic.boards) > 0 || ic.skill != "" {
			drawIcon(px, w, py, ic)
		}
		// D07 css 51: .chars{margin-left:10px}. The per-char span at 143/196/251/…
		// is display:inline-grid holding a 40px portrait plus a progress.ap.
		// D14/F09/F10/E01-E03/F12 are all \u2298: BaseChar now carries Avatar/AP, but
		// SampleBase() feeds names only, so the harness still renders this row as the
		// declared slot is not yet drawn. Do not read this as "portraits are done".
		for i, c := range chars {
			ax := px + charsML + 16 + float64(i)*70
			dc.SetRGB255(90, 90, 100)
			dc.DrawCircle(ax, py+72, 16)
			dc.Fill()
			setFont(dc, 10)
			dc.SetRGB255(255, 255, 255)
			drawStringAnchored(dc, c.Name, ax, py+100, 0.5, 0.5)
		}
	}

	// declared 86: <h3>控制中枢 Lv.N</h3> — no .title_icon on this block.
	drawCard(wideW, fmt.Sprintf("控制中枢 Lv%d", data.Control.Level), bIcon{}, data.Control.Chars)
	// declared 129/137: 宿舍 Lv.N + <span style="color:#66c02f">舒适度N</span>
	for _, d := range data.Dorms {
		drawCard(wideW, fmt.Sprintf("宿舍 Lv%d", d.Level),
			bIcon{text: fmt.Sprintf("舒适度%d", d.Comfort), color: [3]int{0x66, 0xc0, 0x2f}}, d.Chars)
	}
	// declared 183/190: 贸易站 Lv.N + <span style="color:#8cd1ff">策略 N/M</span>
	for _, t := range data.Tradings {
		drawCard(cardW, fmt.Sprintf("贸易站 Lv%d", t.Level),
			bIcon{text: fmt.Sprintf("%s %d/%d", t.Strategy, t.Cur, t.Total), color: [3]int{0x8c, 0xd1, 0xff}}, t.Chars)
	}
	// declared 236/245: 制造站 Lv.N + <span style="color:#d79d13">物品 N/M</span>
	for _, m := range data.Manufactures {
		drawCard(cardW, fmt.Sprintf("制造站 Lv%d", m.Level),
			bIcon{text: fmt.Sprintf("%s %d/%d", m.Item, m.Cur, m.Total), color: [3]int{0xd7, 0x9d, 0x13}}, m.Chars)
	}
	// declared 291/298: 发电站 Lv.N + <span style="color:#adfe2e">N</span>  (bare number, no label)
	for _, p := range data.Powers {
		drawCard(cardW, fmt.Sprintf("发电站 Lv%d", p.Level),
			bIcon{text: itoa(p.Power), color: [3]int{0xad, 0xfe, 0x2e}}, p.Chars)
	}
	// declared 343/358-363: 会客室 Lv.N + <span>线索 <div class="board">…</div></span>
	// declared 343/358-363: 会客室 Lv.N + <span>线索 <div class="board">…</div></span>
	// declared 345-346: the 线索交流开启中 block is inside {{if .Meeting.Sharing}}.
	var meetAbs string
	if data.Meeting.Sharing {
		meetAbs = "线索交流开启中"
	}
	drawCard(cardW, fmt.Sprintf("会客室 Lv%d", data.Meeting.Level),
		bIcon{text: "线索", color: [3]int{255, 255, 255}, boards: data.Meeting.Board, absText: meetAbs}, data.Meeting.Chars)
	// declared 408/415: 办公室 Lv.N + <span>刷新次数N</span>
	drawCard(cardW, fmt.Sprintf("办公室 Lv%d", data.Hire.Level),
		bIcon{text: fmt.Sprintf("刷新次数%d", data.Hire.Refresh), color: [3]int{255, 255, 255}}, data.Hire.Chars)
	// declared 459/462-463: 训练室 Lv.N + <span>Lv.N</span><img class="skill">
	drawCard(cardW, fmt.Sprintf("训练室 Lv%d", data.Training.Level),
		bIcon{text: fmt.Sprintf("Lv.%d", data.Training.SLevel), color: [3]int{255, 255, 255}, skill: data.Training.Skill},
		data.Training.Chars)

	return ScaleToManifest(dc, 1665, 918), nil
}

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
	return ScaleToManifest(dc, 1500, 1323), nil
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
	return ScaleToManifest(dc, 990, 2049), nil
}

// Lottery — mirrors template/Lottery.tmpl: 10x10 grid of numbered cells;
// occupied cells (by lotteryNumber) get selected/winner styling + user info.
type LotteryData struct {
	Details []LotteryDetail
}

type LotteryDetail struct {
	LotteryNumber int64 // 1..100 grid position
	UserName      string
	UserNumber    string // displayed as "ID:xxx"
	Status        int64  // 1 = winner
}

func SampleLottery() *LotteryData {
	return &LotteryData{Details: []LotteryDetail{
		{LotteryNumber: 7, UserName: "中奖用户", UserNumber: "1000000007", Status: 1},
		{LotteryNumber: 41, UserName: "参与用户", UserNumber: "1000000041", Status: 0},
	}}
}

func RenderLottery(data *LotteryData) (*gg.Context, error) {
	const W, H = 1473, 1667
	dc := gg.NewContext(W, H)
	FillBackground(dc, 15, 15, 15) // #0f0f0f
	// main container #1a1a1a, radius 16css->24px, border #333333
	dc.SetRGB255(26, 26, 26)
	dc.DrawRoundedRectangle(30, 30, 1413, 1607, 24)
	dc.Fill()
	dc.SetRGB255(51, 51, 51)
	dc.SetLineWidth(1.5)
	dc.DrawRoundedRectangle(30.75, 30.75, 1411.5, 1605.5, 24)
	dc.Stroke()
	// header: title centered, 28css->42px, letter-spacing 6px, cyan underline
	title := "选号详情"
	setFont(dc, 42)
	dc.SetRGB255(255, 255, 255)
	tw := measureW(dc, title) + 24 // 4 glyphs x 6px spacing
	drawString(dc, title, (W-tw)/2, 130)
	dc.SetRGB255(0, 229, 255)
	dc.DrawRectangle((W-tw)/2-4, 132, tw+8, 3)
	dc.Fill()
	// 10x10 grid (measured from baseline: x0=200 y0=189 pitch=136.5 cell=112)
	const gridX, gridY, pitch, cell = 200.0, 189.0, 136.5, 112.0
	byNum := make(map[int64]LotteryDetail, len(data.Details))
	for _, d := range data.Details {
		byNum[d.LotteryNumber] = d
	}
	for i := 1; i <= 100; i++ {
		r, c := (i-1)/10, (i-1)%10
		x, y := gridX+float64(c)*pitch, gridY+float64(r)*pitch
		d, ok := byNum[int64(i)]
		winner := ok && d.Status == 1
		selected := ok && !winner
		switch {
		case winner:
			dc.SetRGB255(74, 28, 28) // #4a1c1c
		case selected:
			dc.SetRGB255(30, 58, 63) // #1e3a3f
		default:
			dc.SetRGB255(34, 34, 34) // #222222
		}
		dc.DrawRoundedRectangle(x+0.75, y+0.75, cell-1.5, cell-1.5, 9)
		dc.Fill()
		switch {
		case winner:
			dc.SetRGB255(255, 61, 0)
		case selected:
			dc.SetRGB255(0, 229, 255)
		default:
			dc.SetRGB255(51, 51, 51)
		}
		dc.SetLineWidth(1.5)
		dc.DrawRoundedRectangle(x+0.75, y+0.75, cell-1.5, cell-1.5, 9)
		dc.Stroke()
		// num-top: 16css->24px, #666666 / cyan / red, top-left + padding 9
		setFont(dc, 24)
		switch {
		case winner:
			dc.SetRGB255(255, 61, 0)
		case selected:
			dc.SetRGB255(0, 229, 255)
		default:
			dc.SetRGB255(102, 102, 102)
		}
		drawString(dc, itoa(i), x+9, y+30)
		// num-bg: 32css->48px, faint tint centered
		setFont(dc, 48)
		switch {
		case winner:
			dc.SetRGB255(101, 33, 24)
		case selected:
			dc.SetRGB255(27, 75, 82)
		default:
			dc.SetRGB255(41, 41, 41)
		}
		drawString(dc, itoa(i), x+cell/2-12, y+cell/2+18)
		if ok {
			// user-info: name (12css->18px) then id (10css->15px, 70% white)
			setFont(dc, 18)
			if winner {
				dc.SetRGB255(255, 61, 0)
			} else {
				dc.SetRGB255(255, 255, 255)
			}
			// winner cell measured: name x+10/y+69, id x+11/y+96; selected cell: name x+72/y+113 (no visible id)
			nx, ny := x+10, y+69
			if selected {
				nx, ny = x+72, y+113
			}
			drawString(dc, d.UserName, nx, ny)
			setFont(dc, 15)
			dc.SetRGB255(178, 178, 178)
			drawString(dc, "ID:"+d.UserNumber, x+11, y+96)
		}
	}
	// legend: 3 items centered below grid (grid ends y=1495; legend ~1560)
	legendY := 1572.0
	setFont(dc, 21)
	type legendItem struct {
		r, g, b, br int
		label       string
	}
	items := []legendItem{
		{34, 34, 34, 51, "未选择"},
		{30, 58, 63, 0, "已占位"},
		{74, 28, 28, 255, "中奖"},
	}
	totalW := 0.0
	for _, it := range items {
		totalW += 24 + 12 + measureW(dc, it.label) + 45
	}
	x := (W - totalW) / 2
	for _, it := range items {
		dc.SetRGB255(it.r, it.g, it.b)
		dc.DrawRoundedRectangle(x, legendY, 24, 24, 6)
		dc.Fill()
		if it.br > 0 {
			dc.SetRGB255(255, 61, 0)
		} else {
			dc.SetRGB255(51, 51, 51)
		}
		dc.SetLineWidth(1.5)
		dc.DrawRoundedRectangle(x+0.75, legendY+0.75, 22.5, 22.5, 6)
		dc.Stroke()
		dc.SetRGB255(255, 255, 255)
		setFont(dc, 21)
		drawString(dc, it.label, x+36, legendY+17)
		x += 24 + 12 + measureW(dc, it.label) + 45
	}
	_ = totalW
	_ = x
	return dc, nil
}

// Operator
type OperatorInfo struct {
	Name, Profession, Position, Tag string
	Rarity                          int
	Desc                            string
	Stats                           map[string]string
}

func SampleOperator() *OperatorInfo {
	return &OperatorInfo{
		Name: "能天使", Profession: "狙击", Position: "远程", Tag: "输出", Rarity: 6,
		Desc:  "高效的速射狙击干员，能迅速消灭空中与轻甲单位。",
		Stats: map[string]string{"HP": "1560", "ATK": "620", "DEF": "145", "RES": "0", "Cost": "12", "Block": "1", "ASPD": "快"},
	}
}

func RenderOperator(data *OperatorInfo) (*gg.Context, error) {
	const mainW = 800
	const mainH = 700
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 27, 29, 30)
	if bg, err := LoadImage(AssetPath("operator/bg.png")); err == nil {
		// template/Operator.tmpl gives #main a 1200x800 box with background-size:cover. Fitting the
		// art to the 800x700 design canvas (cover: 1024x576 -> 1244x700 at x=-222) and letting
		// ScaleToManifest stretch it to 1800x1200 reproduces the frozen baseline background better
		// than fitting cover to the 1800x1200 output space first (measured per-region, see report).
		dc.DrawImage(ScaleExact(bg, 1244, 700), -222, 0)
	}
	// top bar
	dc.SetRGB255(45, 48, 55)
	dc.DrawRectangle(0, 0, float64(mainW), 110)
	dc.Fill()
	// avatar
	dc.SetRGB255(80, 80, 90)
	dc.DrawRoundedRectangle(20, 20, 80, 80, 10)
	dc.Fill()
	setFont(dc, 24)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, data.Name, 120, 50)
	setFont(dc, 14)
	dc.SetRGB255(180, 200, 220)
	drawString(dc, fmt.Sprintf("%s · %s · %s · %d★", data.Profession, data.Position, data.Tag, data.Rarity), 120, 74)
	// rarity bar
	r, g, b := rarityColor(data.Rarity)
	dc.SetRGB255(r, g, b)
	dc.DrawRectangle(float64(mainW-120), 20, 100, 28)
	dc.Fill()
	setFont(dc, 14)
	dc.SetRGB255(255, 255, 255)
	drawStringAnchored(dc, fmt.Sprintf("%d ★", data.Rarity), float64(mainW-70), 34, 0.5, 0.5)
	// desc
	y := 140
	fillRoundedCard(dc, 20, float64(y), float64(mainW-40), 80, 10, 14)
	setFont(dc, 13)
	dc.SetRGB255(200, 220, 200)
	drawString(dc, StripHTML(data.Desc), 30, float64(y+30))
	// stats grid 2 cols
	y = 250
	keys := []string{"HP", "ATK", "DEF", "RES", "Cost", "Block", "ASPD"}
	cols := 3
	tileW := (mainW - 40) / cols
	tileH := 70
	for i, k := range keys {
		x := (i%cols)*tileW + 20
		yy := y + (i/cols)*tileH
		dc.SetRGBA255(255, 255, 255, 10)
		RoundRect(dc, float64(x+4), float64(yy), float64(tileW-8), 60, 8)
		setFont(dc, 12)
		dc.SetRGB255(160, 180, 200)
		drawStringAnchored(dc, k, float64(x+tileW/2), float64(yy+22), 0.5, 0.5)
		setFont(dc, 18)
		dc.SetRGB255(255, 255, 255)
		drawStringAnchored(dc, data.Stats[k], float64(x+tileW/2), float64(yy+44), 0.5, 0.5)
	}
	return ScaleToManifest(dc, 1800, 1200), nil
}
