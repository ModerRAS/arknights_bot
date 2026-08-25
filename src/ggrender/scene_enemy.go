package ggrender

import (
	"fmt"

	"github.com/fogleman/gg"
)

// Enemy
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

func SampleEnemy() *Enemy {
	return &Enemy{
		Name: "霜星", Pic: "", Desc: "雪怪小队领袖，擅长冰属性法术。", EnemyRace: "人类", EnemyLevel: "精英", AttackType: "法术", Motion: "地面",
		Ability: "攻击造成法术伤害，并施加寒冷。",
		Levels: []EnemyLevel{{HP: "12000", ATK: "850", DEF: "300", Res: "40", Talent: "攻击范围内我方单位移动速度降低。", Skills: []EnemySkill{{Name: "冰封", SpInit: "10", SpCost: "30", Desc: "对范围内单位造成大量法术伤害并冻结。"}}}},
	}
}

func RenderEnemy(data *Enemy) (*gg.Context, error) {
	const mainW = 984
	const pad = 16
	headerH := 160
	levelH := 260
	dc := gg.NewContext(mainW, 477)
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
