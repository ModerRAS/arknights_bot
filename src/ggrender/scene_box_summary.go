package ggrender

import (
	"github.com/fogleman/gg"
)

type MissingChar struct {
	SkinId, Name string
	Rarity       int
	Profession   string
}
type BoxSummary struct {
	Name                                                                              string
	AllCharCnt, Star6CharCnt, Star5CharCnt, Star4CharCnt                              string
	AllEvolvePhase2Cnt, Star6EvolvePhase2Cnt, Star5EvolvePhase2Cnt, Star4EvolvePhase2Cnt int
	AllSkill10Cnt, Star6Skill10Cnt, Star5Skill10Cnt, Star4Skill10Cnt                  int
	AllSkill9Cnt, Star6Skill9Cnt, Star5Skill9Cnt, Star4Skill9Cnt                      int
	AllSkill8Cnt, Star6Skill8Cnt, Star5Skill8Cnt, Star4Skill8Cnt                      int
	AllEquipStage3Cnt, Star6EquipStage3Cnt, Star5EquipStage3Cnt, Star4EquipStage3Cnt  int
	AllEquipStage2Cnt, Star6EquipStage2Cnt, Star5EquipStage2Cnt, Star4EquipStage2Cnt  int
	AllEquipStage1Cnt, Star6EquipStage1Cnt, Star5EquipStage1Cnt, Star4EquipStage1Cnt  int
	MissingChars                                                                      []MissingChar
}

func SampleBoxSummary() *BoxSummary {
	missing := make([]MissingChar, 0, 25)
	for i := 0; i < 25; i++ {
		missing = append(missing, MissingChar{SkinId: AmiyaAvatarURL, Name: "阿米娅", Rarity: 5, Profession: "PIONEER"})
	}
	return &BoxSummary{
		Name: "Dr 测试博士",
		AllCharCnt: "60/100", Star6CharCnt: "20/50", Star5CharCnt: "25/30", Star4CharCnt: "15/20",
		AllEvolvePhase2Cnt: 40, Star6EvolvePhase2Cnt: 18, Star5EvolvePhase2Cnt: 15, Star4EvolvePhase2Cnt: 7,
		AllSkill10Cnt: 30, Star6Skill10Cnt: 15, Star5Skill10Cnt: 10, Star4Skill10Cnt: 5,
		AllSkill9Cnt: 20, Star6Skill9Cnt: 10, Star5Skill9Cnt: 8, Star4Skill9Cnt: 2,
		AllSkill8Cnt: 10, Star6Skill8Cnt: 5, Star5Skill8Cnt: 4, Star4Skill8Cnt: 1,
		AllEquipStage3Cnt: 12, Star6EquipStage3Cnt: 6, Star5EquipStage3Cnt: 4, Star4EquipStage3Cnt: 2,
		AllEquipStage2Cnt: 8, Star6EquipStage2Cnt: 4, Star5EquipStage2Cnt: 3, Star4EquipStage2Cnt: 1,
		AllEquipStage1Cnt: 4, Star6EquipStage1Cnt: 2, Star5EquipStage1Cnt: 1, Star4EquipStage1Cnt: 1,
		MissingChars: missing,
	}
}

// ivals 四列数值转字符串。
func ivals(a, b, c, d int) [4]string { return [4]string{itoa(a), itoa(b), itoa(c), itoa(d)} }

func RenderBoxSummary(data *BoxSummary) (*gg.Context, error) {
	const mainW, mainH = 1350, 723
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 46, 48, 49)
	drawBoxFamilyHeader(dc, mainW, 86, data.Name, 57, 85)

	gx := [4]int{30, 374, 699, 1025}
	gw := [4]int{313, 295, 295, 295}
	// 列组标题
	setFont(dc, 30)
	dc.SetRGB255(255, 255, 255)
	titles := []string{"全部干员", "六星干员", "五星干员", "四星干员"}
	for i, t := range titles {
		drawStringAnchored(dc, t, float64(gx[i]+gw[i]/2), 116, 0.5, 0.5)
	}
	labels := []string{"招募干员数量", "精英阶段2干员", "专精三技能数量", "专精二技能数量", "专精一技能数量", "三级模组数量", "二级模组数量", "一级模组数量"}
	values := [8][4]string{
		{data.AllCharCnt, data.Star6CharCnt, data.Star5CharCnt, data.Star4CharCnt},
		ivals(data.AllEvolvePhase2Cnt, data.Star6EvolvePhase2Cnt, data.Star5EvolvePhase2Cnt, data.Star4EvolvePhase2Cnt),
		ivals(data.AllSkill10Cnt, data.Star6Skill10Cnt, data.Star5Skill10Cnt, data.Star4Skill10Cnt),
		ivals(data.AllSkill9Cnt, data.Star6Skill9Cnt, data.Star5Skill9Cnt, data.Star4Skill9Cnt),
		ivals(data.AllSkill8Cnt, data.Star6Skill8Cnt, data.Star5Skill8Cnt, data.Star4Skill8Cnt),
		ivals(data.AllEquipStage3Cnt, data.Star6EquipStage3Cnt, data.Star5EquipStage3Cnt, data.Star4EquipStage3Cnt),
		ivals(data.AllEquipStage2Cnt, data.Star6EquipStage2Cnt, data.Star5EquipStage2Cnt, data.Star4EquipStage2Cnt),
		ivals(data.AllEquipStage1Cnt, data.Star6EquipStage1Cnt, data.Star5EquipStage1Cnt, data.Star4EquipStage1Cnt),
	}
	for k := 0; k < 8; k++ {
		uy := 183 + 48*k
		dc.SetRGB255(210, 210, 210)
		for i := 0; i < 4; i++ {
			dc.DrawRectangle(float64(gx[i]), float64(uy), float64(gw[i]), 2)
			dc.Fill()
		}
		setFont(dc, 24)
		dc.SetRGB255(255, 255, 255)
		for i := 0; i < 4; i++ {
			drawString(dc, labels[k], float64(gx[i]+18), float64(uy-10))
			drawStringAnchored(dc, values[k][i], float64(gx[i]+gw[i]-17), float64(uy-7), 1, 1)
		}
	}
	// 未招募干员
	setFont(dc, 26)
	dc.SetRGB255(255, 255, 255)
	drawStringAnchored(dc, "未招募干员", 675, 553, 0.5, 0.5)
	// 迷你头像瓦片 60x58，pitch 68/71
	for i, c := range data.MissingChars {
		x := 31 + (i%19)*68
		y := 585 + (i/19)*71
		port := FetchImage(c.SkinId, amiyaPath)
		dc.DrawImage(smoothCover(port, 60, 58), x, y)
	}
	return dc, nil
}
