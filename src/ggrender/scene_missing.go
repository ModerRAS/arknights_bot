package ggrender

import (
	"github.com/fogleman/gg"
)

type MissingInfo struct {
	Name  string
	Chars []MissingChar
}

func SampleMissing() *MissingInfo {
	chars := make([]MissingChar, 0, 12)
	for i := 0; i < 12; i++ {
		chars = append(chars, MissingChar{SkinId: AmiyaPortraitURL, Name: "阿米娅", Rarity: 5, Profession: "PIONEER"})
	}
	return &MissingInfo{Name: "Dr 测试博士(未获取)", Chars: chars}
}

func RenderMissing(data *MissingInfo) (*gg.Context, error) {
	const mainW, mainH = 1050, 536
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 46, 48, 49)
	drawBoxFamilyHeader(dc, mainW, 106, data.Name, 81, 107)
	for i, c := range data.Chars {
		x := (i % 10) * 105
		y := 114 + (i/10)*210
		drawBoxTile(dc, x, y, c.SkinId, c.Profession, c.Rarity, 0, 0, c.Name, false)
	}
	return dc, nil
}
