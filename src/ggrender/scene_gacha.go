package ggrender

import (
	"github.com/fogleman/gg"
	"math"
)

// Gacha — mirrors template/Gacha.tmpl rendered at 1000x882 CSS, scale 1.5 -> 1500x1323.
// header.png banner with h1 overlays; three .item cards (pie / avg table / pool bars);
// #article with two .chars panels; footer.png with date. ECharts pie+bar drawn natively.
type GachaChar struct {
	PoolName, CharName, Avatar, Ts string
	IsNew                          bool
}
type GachaStar6 struct {
	PoolName, Name, Avatar, Ts string
	Count                      int
	IsNew                      bool
}
type GachaPool struct {
	PoolName string
	Count    int
}
type GachaData struct {
	Name                      string
	Total                     int
	Star6, Star5, Star4, Star3 int
	Avg6, Avg5, Avg4, Avg3    string
	BegTime, EndTime          string
	Chars                     []GachaChar
	Star6Info                 []GachaStar6
	Pools                     []GachaPool
	Now                       string
}

const (
	gachaAvatarAmiya  = "https://media.prts.wiki/3/36/%E5%A4%B4%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85.png?image_process=format,webp/quality,Q_90"
	gachaAvatarExusiai = "https://media.prts.wiki/a/ad/%E5%A4%B4%E5%83%8F_%E8%83%BD%E5%A4%A9%E4%BD%BF.png?image_process=format,webp/quality,Q_90"
)

func SampleGacha() *GachaData {
	return &GachaData{
		Name: "基线博士", Total: 10, Star6: 2, Star5: 2, Star4: 3, Star3: 3,
		Avg6: "5", Avg5: "5", Avg4: "3.33", Avg3: "3.33",
		BegTime: "2025-01-14 20:00:00", EndTime: "2025-01-15 20:00:00",
		Chars: []GachaChar{
			{PoolName: "测试卡池甲", CharName: "阿米娅", Avatar: gachaAvatarAmiya, Ts: "2025-01-14", IsNew: true},
			{PoolName: "测试卡池乙", CharName: "能天使", Avatar: gachaAvatarExusiai, Ts: "2025-01-15"},
		},
		Star6Info: []GachaStar6{
			{PoolName: "测试卡池甲", Name: "阿米娅", Avatar: gachaAvatarAmiya, Ts: "2025-01-14", Count: 5, IsNew: true},
			{PoolName: "测试卡池乙", Name: "能天使", Avatar: gachaAvatarExusiai, Ts: "2025-01-15", Count: 5},
		},
		Pools: []GachaPool{{PoolName: "测试卡池甲", Count: 6}, {PoolName: "测试卡池乙", Count: 4}},
		Now:   "2026年08月19日",
	}
}

// fillSector draws a pie sector from angle a0 to a1 (gg convention: +angle = clockwise on screen).
func fillSector(dc *gg.Context, cx, cy, r, a0, a1 float64) {
	dc.MoveTo(cx, cy)
	dc.LineTo(cx+r*cos(a0), cy+r*sin(a0))
	dc.DrawArc(cx, cy, r, a0, a1)
	dc.LineTo(cx, cy)
	dc.Fill()
}

func cos(a float64) float64 { return math.Cos(a) }
func sin(a float64) float64 { return math.Sin(a) }

func RenderGacha(data *GachaData) (*gg.Context, error) {
	const W, H = 1500, 1323
	dc := gg.NewContext(W, H)
	FillBackground(dc, 12, 13, 12)
	// header
	dc.DrawImage(ScaleExact(tryLocal("gacha/header.png"), W, 600), 0, 0)
	setFont(dc, 48)
	dc.SetRGB255(238, 238, 238)
	drawStringBoldW(dc, data.Name, 480, 129, 1.8)
	totalStr := "共" + itoa(data.Total) + "抽"
	drawStringBoldW(dc, totalStr, 386, 197, 1.8)
	dates := "(" + data.BegTime + "——" + data.EndTime + ")"
	setFont(dc, 34.5)
	drawString(dc, dates, 611, 193)

	// three item cards
	card := func(x0 float64) {
		dc.SetRGBA255(0, 0, 0, 255)
		dc.SetLineWidth(1.5)
		RoundRect(dc, x0, 225, 452, 377, 30)
		dc.Stroke()
		dc.SetRGB255(31, 30, 30)
		RoundRect(dc, x0+1.5, 226.5, 449, 374, 29)
		dc.Fill()
	}
	card(30)
	card(519)
	card(1006)
	title := func(s string, cx float64) {
		setFont(dc, 30)
		dc.SetRGB255(238, 238, 238)
		drawStringAnchored(dc, s, cx, 264.5, 0.5, 0.5)
	}
	title("星级分布", 256)
	title("星级分布", 745)
	title("卡池分布(最近10个)", 1232)

	// card1: pie chart, center (312,436.5) r=116.5, startAngle 90deg clockwise
	type slice struct {
		frac float64
		col  [3]int
	}
	slices := []slice{
		{float64(data.Star6) / float64(data.Total), [3]int{244, 110, 30}},
		{float64(data.Star5) / float64(data.Total), [3]int{247, 171, 55}},
		{float64(data.Star4) / float64(data.Total), [3]int{161, 53, 246}},
		{float64(data.Star3) / float64(data.Total), [3]int{109, 116, 126}},
	}
	const pcx, pcy, pr = 312.0, 436.5, 116.5
	a := -math.Pi / 2
	for _, sl := range slices {
		dc.SetRGB255(sl.col[0], sl.col[1], sl.col[2])
		a2 := a + sl.frac*2*math.Pi
		fillSector(dc, pcx, pcy, pr, a, a2)
		a = a2
	}
	// legend: squares x40 w10, rows center 369.5 + i*48
	legend := []struct {
		s   string
		col [3]int
	}{{"6星", [3]int{244, 110, 30}}, {"5星", [3]int{247, 171, 55}}, {"4星", [3]int{161, 53, 246}}, {"3星", [3]int{109, 116, 126}}}
	pcts := []string{"20.00%", "20.00%", "30.00%", "30.00%"}
	for i, le := range legend {
		cy := 370.5 + float64(i)*48
		dc.SetRGB255(le.col[0], le.col[1], le.col[2])
		dc.DrawRectangle(38, cy-5, 10, 10)
		dc.Fill()
		setFont(dc, 21)
		dc.SetRGB255(255, 255, 255)
		drawString(dc, le.s, 56, cy+7)
		pw, _ := measure(dc, le.s)
		drawString(dc, "  "+pcts[i], 56+pw+8, cy+7)
	}

	// card2: avg table rows center y 335+i*67.3
	avg := []struct{ a, b string }{
		{itoa(data.Star6) + "个6星", data.Avg6 + "抽/个"},
		{itoa(data.Star5) + "个5星", data.Avg5 + "抽/个"},
		{itoa(data.Star4) + "个4星", data.Avg4 + "抽/个"},
		{itoa(data.Star3) + "个3星", data.Avg3 + "抽/个"},
	}
	setFont(dc, 24)
	dc.SetRGB255(238, 238, 238)
	for i, r := range avg {
		cy := 335 + float64(i)*67.3
		drawStringAnchored(dc, r.a, 572, cy, 0, 0.5)
		drawStringAnchored(dc, r.b, 770, cy, 0, 0.5)
	}

	// card3: horizontal bars
	bars := []struct {
		name  string
		count int
		y0    float64
		x1    float64
	}{{"测试卡池乙", 4, 316, 1315}, {"测试卡池甲", 6, 462, 1411}}
	for _, b := range bars {
		dc.SetRGB255(84, 111, 198)
		dc.DrawRectangle(1120, b.y0, b.x1-1120, 101)
		dc.Fill()
	}
	// axis bracket
	dc.SetRGB255(100, 100, 100)
	dc.DrawRectangle(1118, 300, 2, 270)
	for _, yy := range []float64{300, 433.5, 570} {
		dc.DrawRectangle(1110, yy, 10, 2)
		dc.Fill()
	}
	dc.Fill()
	setFont(dc, 21)
	dc.SetRGB255(255, 255, 255)
	drawStringAnchored(dc, "测试卡池乙", 1105, 366, 1, 0.5)
	drawStringAnchored(dc, "测试卡池甲", 1105, 512, 1, 0.5)
	drawStringAnchored(dc, "4", 1330, 366, 0, 0.5)
	drawStringAnchored(dc, "6", 1425, 512, 0, 0.5)

	// article
	dc.SetRGB255(12, 13, 12)
	dc.DrawRectangle(0, 600, W, 319.5)
	dc.Fill()
	panel := func(x0 float64) {
		dc.SetRGBA255(0, 0, 0, 255)
		dc.SetLineWidth(1.5)
		RoundRect(dc, x0, 660, 700, 255, 30)
		dc.Stroke()
		dc.SetRGB255(31, 30, 30)
		RoundRect(dc, x0+1.5, 661.5, 697, 252, 29)
		dc.Fill()
	}
	panel(30)
	panel(766)
	setFont(dc, 30)
	dc.SetRGB255(238, 238, 238)
	drawStringAnchored(dc, "新获得干员(至多显示20个)", 380, 699.5, 0.5, 0.5)
	drawStringAnchored(dc, "获得六星干员(至多显示20个)", 1116, 699.5, 0.5, 0.5)

	entry := func(x0 float64, avatar string, name string, isNew bool, ts, pool string, yDate, yPool, yCost float64, cost string) {
		dc.DrawImage(ScaleExact(FetchImage(avatar, AssetPath("common/amiya.png")), 150, 150), int(x0), 756)
		setFont(dc, 24)
		dc.SetRGB255(238, 238, 238)
		drawString(dc, name, x0+152, 783)
		if isNew {
			setFont(dc, 15)
			dc.SetRGB255(255, 0, 0)
			nw, _ := measure(dc, name)
			drawString(dc, "New", x0+152+nw+8, 768)
		}
		setFont(dc, 22.5)
		dc.SetRGB255(238, 238, 238)
		drawString(dc, ts, x0+152, yDate)
		drawString(dc, pool, x0+152, yPool)
		if cost != "" {
			drawString(dc, cost, x0+152, yCost)
		}
	}
	ex := 36.0
	for _, c := range data.Chars {
		entry(ex, c.Avatar, c.CharName, c.IsNew, c.Ts, c.PoolName, 838, 885, 0, "")
		ex += 345
	}
	ex = 776.0
	for _, s := range data.Star6Info {
		entry(ex, s.Avatar, s.Name, s.IsNew, s.Ts, s.PoolName, 830, 876, 904, "花费"+itoa(s.Count)+"抽")
		ex += 345
	}

	// footer
	dc.DrawImage(ScaleExact(tryLocal("gacha/footer.png"), W, 404), 0, 919)
	setFont(dc, 48)
	dc.SetRGB255(238, 238, 238)
	drawStringBoldW(dc, data.Now, 810, 1258, 1.8)
	return dc, nil
}
