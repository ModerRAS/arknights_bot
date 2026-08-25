package ggrender

import (
	"fmt"
	"image"
	"image/color"
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
type BaseInfo struct {
	Name string
	Labor struct{ Cur, Total int }
	Control struct{ Level int; Chars []string }
	Tradings []struct{ Level int; Chars []string; Cur, Total int; Strategy string }
	Manufactures []struct{ Level int; Chars []string; Cur, Total int; Item, Speed string }
	Powers []struct{ Level int; Chars []string; Power int }
	Meeting struct{ Level int; Chars []string; Board []int; Sharing bool }
	Hire struct{ Level int; Chars []string; Refresh int }
	Training struct{ Level int; Chars []string; Skill string; SLevel int }
	Dorms []struct{ Level int; Chars []string; Comfort int }
}

func SampleBase() *BaseInfo {
	b := &BaseInfo{Name: "博士的基建"}
	b.Labor.Cur = 108
	b.Labor.Total = 120
	b.Control.Level = 5
	b.Control.Chars = []string{"阿米娅", "凯尔希", "煌"}
	b.Tradings = []struct{ Level int; Chars []string; Cur, Total int; Strategy string }{
		{Level: 3, Chars: []string{"能天使", "德克萨斯"}, Cur: 3, Total: 5, Strategy: "贵金属订单"},
		{Level: 3, Chars: []string{"拉普兰德"}, Cur: 2, Total: 5, Strategy: "源石订单"},
	}
	b.Manufactures = []struct{ Level int; Chars []string; Cur, Total int; Item, Speed string }{
		{Level: 3, Chars: []string{"夜烟", "远山"}, Cur: 10, Total: 20, Item: "中级作战记录", Speed: "120%"},
		{Level: 3, Chars: []string{"砾"}, Cur: 8, Total: 20, Item: "赤金", Speed: "100%"},
	}
	b.Powers = []struct{ Level int; Chars []string; Power int }{
		{Level: 3, Chars: []string{"格雷伊"}, Power: 270},
		{Level: 3, Chars: []string{"清流"}, Power: 270},
	}
	b.Meeting.Level = 3
	b.Meeting.Chars = []string{"诗怀雅"}
	b.Meeting.Board = []int{1, 2, 3}
	b.Meeting.Sharing = true
	b.Hire.Level = 3
	b.Hire.Chars = []string{"陈"}
	b.Hire.Refresh = 2
	b.Training.Level = 3
	b.Training.Chars = []string{"赫拉格", "华法琳"}
	b.Training.Skill = "阿米娅-奇美拉"
	b.Training.SLevel = 2
	b.Dorms = []struct{ Level int; Chars []string; Comfort int }{
		{Level: 5, Chars: []string{"星熊", "塞雷娅"}, Comfort: 5000},
		{Level: 5, Chars: []string{"夜莺"}, Comfort: 4800},
	}
	return b
}

func RenderBase(data *BaseInfo) (*gg.Context, error) {
	const mainW = 1100
	const pad = 16
	// estimate height: header 60 + labor 50 + control 100 + tradings*110 + manufactures*110 + powers*80 + meeting 90 + hire 80 + training 90 + dorms*90
	h := 60 + 50 + 100 + len(data.Tradings)*110 + len(data.Manufactures)*110 + len(data.Powers)*80 + 90 + 80 + 90 + len(data.Dorms)*90 + 40
	dc := gg.NewContext(mainW, h)
	FillBackground(dc, 30, 32, 33)
	// header
	dc.SetRGB255(50, 55, 60)
	dc.DrawRectangle(0, 0, float64(mainW), 60)
	dc.Fill()
	setFont(dc, 26)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, data.Name+" · 基建总览", 20, 38)
	y := 80
	// labor
	fillRoundedCard(dc, float64(pad), float64(y), float64(mainW-2*pad), 50, 8, 18)
	setFont(dc, 16)
	dc.SetRGB255(220, 220, 220)
	drawString(dc, fmt.Sprintf("无人机 %d/%d", data.Labor.Cur, data.Labor.Total), float64(pad+20), float64(y+30))
	ProgressBar(dc, float64(pad+250), float64(y+20), 300, 14, float64(data.Labor.Cur)/float64(data.Labor.Total), 90, 180, 255)
	y += 70
	// control
	fillRoundedCard(dc, float64(pad), float64(y), float64(mainW-2*pad), 100, 8, 12)
	setFont(dc, 16)
	dc.SetRGB255(180, 220, 255)
	drawString(dc, fmt.Sprintf("控制中枢 Lv%d", data.Control.Level), float64(pad+20), float64(y+28))
	for i, n := range data.Control.Chars {
		cx := float64(pad+20 + i*90)
		cy := float64(y+65)
		dc.SetRGB255(80, 80, 90)
		dc.DrawCircle(cx+22, cy, 22)
		dc.Fill()
		setFont(dc, 11)
		dc.SetRGB255(255, 255, 255)
		drawStringAnchored(dc, n, cx+22, cy+32, 0.5, 0.5)
	}
	y += 120
	// tradings
	for _, t := range data.Tradings {
		fillRoundedCard(dc, float64(pad), float64(y), float64(mainW-2*pad), 100, 8, 12)
		setFont(dc, 14)
		dc.SetRGB255(220, 200, 120)
		drawString(dc, fmt.Sprintf("贸易站 Lv%d · %s %d/%d", t.Level, t.Strategy, t.Cur, t.Total), float64(pad+20), float64(y+28))
		for i, n := range t.Chars {
			cx := float64(pad+20 + i*70)
			cy := float64(y+60)
			dc.SetRGB255(90, 90, 100)
			dc.DrawCircle(cx+16, cy, 16)
			dc.Fill()
			setFont(dc, 10)
			dc.SetRGB255(255, 255, 255)
			drawStringAnchored(dc, n, cx+16, cy+24, 0.5, 0.5)
		}
		y += 110
	}
	// manufactures
	for _, m := range data.Manufactures {
		fillRoundedCard(dc, float64(pad), float64(y), float64(mainW-2*pad), 100, 8, 12)
		setFont(dc, 14)
		dc.SetRGB255(120, 220, 160)
		drawString(dc, fmt.Sprintf("制造站 Lv%d · %s %d/%d %s", m.Level, m.Item, m.Cur, m.Total, m.Speed), float64(pad+20), float64(y+28))
		for i, n := range m.Chars {
			cx := float64(pad+20 + i*70)
			cy := float64(y+60)
			dc.SetRGB255(90, 90, 100)
			dc.DrawCircle(cx+16, cy, 16)
			dc.Fill()
			setFont(dc, 10)
			dc.SetRGB255(255, 255, 255)
			drawStringAnchored(dc, n, cx+16, cy+24, 0.5, 0.5)
		}
		y += 110
	}
	// powers
	for _, p := range data.Powers {
		fillRoundedCard(dc, float64(pad), float64(y), float64(mainW-2*pad), 80, 8, 12)
		setFont(dc, 14)
		dc.SetRGB255(120, 180, 255)
		drawString(dc, fmt.Sprintf("发电站 Lv%d · %d 电力", p.Level, p.Power), float64(pad+20), float64(y+28))
		for i, n := range p.Chars {
			cx := float64(pad+20 + i*70)
			cy := float64(y+55)
			dc.SetRGB255(90, 90, 100)
			dc.DrawCircle(cx+16, cy, 16)
			dc.Fill()
			setFont(dc, 10)
			dc.SetRGB255(255, 255, 255)
			drawStringAnchored(dc, n, cx+16, cy+20, 0.5, 0.5)
		}
		y += 90
	}
	// meeting
	fillRoundedCard(dc, float64(pad), float64(y), float64(mainW-2*pad), 80, 8, 12)
	setFont(dc, 14)
	dc.SetRGB255(220, 180, 220)
	drawString(dc, fmt.Sprintf("会客室 Lv%d · 线索 %v 共享:%v", data.Meeting.Level, data.Meeting.Board, data.Meeting.Sharing), float64(pad+20), float64(y+28))
	for i, n := range data.Meeting.Chars {
		cx := float64(pad+20 + i*70)
		cy := float64(y+55)
		dc.SetRGB255(90, 90, 100)
		dc.DrawCircle(cx+16, cy, 16)
		dc.Fill()
		setFont(dc, 10)
		dc.SetRGB255(255, 255, 255)
		drawStringAnchored(dc, n, cx+16, cy+20, 0.5, 0.5)
	}
	y += 90
	// hire
	fillRoundedCard(dc, float64(pad), float64(y), float64(mainW-2*pad), 70, 8, 12)
	setFont(dc, 14)
	dc.SetRGB255(220, 200, 180)
	drawString(dc, fmt.Sprintf("办公室 Lv%d · 刷新 %d", data.Hire.Level, data.Hire.Refresh), float64(pad+20), float64(y+28))
	for i, n := range data.Hire.Chars {
		cx := float64(pad+20 + i*70)
		cy := float64(y+50)
		dc.SetRGB255(90, 90, 100)
		dc.DrawCircle(cx+16, cy, 16)
		dc.Fill()
		setFont(dc, 10)
		dc.SetRGB255(255,255,255)
		drawStringAnchored(dc, n, cx+16, cy+20, 0.5,0.5)
	}
	y += 80
	// training
	fillRoundedCard(dc, float64(pad), float64(y), float64(mainW-2*pad), 80, 8, 12)
	setFont(dc, 14)
	dc.SetRGB255(180, 220, 200)
	drawString(dc, fmt.Sprintf("训练室 Lv%d · %s 专精%d", data.Training.Level, data.Training.Skill, data.Training.SLevel), float64(pad+20), float64(y+28))
	for i, n := range data.Training.Chars {
		cx := float64(pad+20 + i*70)
		cy := float64(y+55)
		dc.SetRGB255(90,90,100)
		dc.DrawCircle(cx+16,cy,16)
		dc.Fill()
		setFont(dc,10)
		dc.SetRGB255(255,255,255)
		drawStringAnchored(dc,n,cx+16,cy+20,0.5,0.5)
	}
	y += 90
	// dorms
	for _, d := range data.Dorms {
		fillRoundedCard(dc, float64(pad), float64(y), float64(mainW-2*pad), 80, 8, 12)
		setFont(dc,14)
		dc.SetRGB255(200,200,220)
		drawString(dc, fmt.Sprintf("宿舍 Lv%d · 舒适度 %d", d.Level, d.Comfort), float64(pad+20), float64(y+28))
		for i, n := range d.Chars {
			cx:=float64(pad+20+i*70)
			cy:=float64(y+55)
			dc.SetRGB255(90,90,100)
			dc.DrawCircle(cx+16,cy,16)
			dc.Fill()
			setFont(dc,10)
			dc.SetRGB255(255,255,255)
			drawStringAnchored(dc,n,cx+16,cy+20,0.5,0.5)
		}
		y+=90
	}
	return dc,nil
}

// Calendar
type CardInfo struct {
	Name, Uid, ServerName, Resume string
	Level, RegTime int
	MainStageProgress, Avatar, SecretaryName, SecretaryEnName string
	CharCnt, FurnitureCnt, SkinCnt, EquipCnt int
}

func SampleCard() *CardInfo {
	return &CardInfo{
		Name: "博士", Uid: "10000001", ServerName: "官服", Resume: "罗德岛的博士，今日也在努力。",
		Level: 120, RegTime: 1620000000, MainStageProgress: "12-18", Avatar: "", SecretaryName: "阿米娅", SecretaryEnName: "Amiya",
		CharCnt: 280, FurnitureCnt: 320, SkinCnt: 120, EquipCnt: 80,
	}
}

func RenderCard(data *CardInfo) (*gg.Context, error) {
	const mainW=1000
	const mainH=600
	dc:=gg.NewContext(mainW,mainH)
	FillBackground(dc,27,29,30)
	// top decor bar
	dc.SetRGB255(45,48,55)
	dc.DrawRectangle(0,0,float64(mainW),90)
	dc.Fill()
	// avatar circle
	dc.SetRGB255(80,80,90)
	dc.DrawCircle(60,45,38)
	dc.Fill()
	setFont(dc,22)
	dc.SetRGB255(255,255,255)
	drawString(dc,data.Name,110,36)
	setFont(dc,13)
	dc.SetRGB255(180,200,220)
	drawString(dc,fmt.Sprintf("UID %s · %s · Lv%d",data.Uid,data.ServerName,data.Level),110,58)
	drawString(dc,"主线 "+data.MainStageProgress,110,76)
	// stats row
	stats:=[]struct{label string; val int}{
		{"干员",data.CharCnt},{"家具",data.FurnitureCnt},{"时装",data.SkinCnt},{"模组",data.EquipCnt},
	}
	x:=20
	y:=140
	for _,s:=range stats {
		dc.SetRGBA255(255,255,255,12)
		RoundRect(dc,float64(x),float64(y),140,80,8)
		setFont(dc,15)
		dc.SetRGB255(180,200,220)
		drawStringAnchored(dc,s.label,float64(x+70),float64(y+24),0.5,0.5)
		setFont(dc,24)
		dc.SetRGB255(255,255,255)
		drawStringAnchored(dc,itoa(s.val),float64(x+70),float64(y+54),0.5,0.5)
		x+=160
	}
	// secretary
	y=260
	fillRoundedCard(dc,20,float64(y),float64(mainW-40),120,10,14)
	setFont(dc,16)
	dc.SetRGB255(255,230,160)
	drawString(dc,"秘书 "+data.SecretaryName+" / "+data.SecretaryEnName,30,float64(y+36))
	setFont(dc,13)
	dc.SetRGB255(200,200,200)
	drawString(dc,StripHTML(data.Resume),30,float64(y+64))
	// resume icon placeholder
	dc.SetRGB255(70,70,80)
	dc.DrawCircle(float64(mainW-80),float64(y+60),36)
	dc.Fill()
	return dc,nil
}

// ponytail: depot split to depot.go for per-scene ownership (renderContext routes there)
// DepotData/DepotItem/SampleDepot/RenderDepot moved to src/ggrender/depot.go

// Gacha
type OperatorInfo struct {
	Name, Profession, Position, Tag string
	Rarity int
	Desc string
	Stats map[string]string
}

func SampleOperator() *OperatorInfo {
	return &OperatorInfo{
		Name:"能天使", Profession:"狙击", Position:"远程", Tag:"输出", Rarity:6,
		Desc:"高效的速射狙击干员，能迅速消灭空中与轻甲单位。",
		Stats: map[string]string{"HP":"1560","ATK":"620","DEF":"145","RES":"0","Cost":"12","Block":"1","ASPD":"快"},
	}
}

func RenderOperator(data *OperatorInfo) (*gg.Context, error) {
	const mainW=800
	const mainH=700
	dc:=gg.NewContext(mainW,mainH)
	FillBackground(dc,27,29,30)
	// top bar
	dc.SetRGB255(45,48,55)
	dc.DrawRectangle(0,0,float64(mainW),110)
	dc.Fill()
	// avatar
	dc.SetRGB255(80,80,90)
	dc.DrawRoundedRectangle(20,20,80,80,10)
	dc.Fill()
	setFont(dc,24)
	dc.SetRGB255(255,255,255)
	drawString(dc,data.Name,120,50)
	setFont(dc,14)
	dc.SetRGB255(180,200,220)
	drawString(dc,fmt.Sprintf("%s · %s · %s · %d★",data.Profession,data.Position,data.Tag,data.Rarity),120,74)
	// rarity bar
	r,g,b:=rarityColor(data.Rarity)
	dc.SetRGB255(r,g,b)
	dc.DrawRectangle(float64(mainW-120),20,100,28)
	dc.Fill()
	setFont(dc,14)
	dc.SetRGB255(255,255,255)
	drawStringAnchored(dc,fmt.Sprintf("%d ★",data.Rarity),float64(mainW-70),34,0.5,0.5)
	// desc
	y:=140
	fillRoundedCard(dc,20,float64(y),float64(mainW-40),80,10,14)
	setFont(dc,13)
	dc.SetRGB255(200,220,200)
	drawString(dc,StripHTML(data.Desc),30,float64(y+30))
	// stats grid 2 cols
	y=250
	keys:=[]string{"HP","ATK","DEF","RES","Cost","Block","ASPD"}
	cols:=3
	tileW:= (mainW-40)/cols
	tileH:=70
	for i,k:=range keys {
		x:=(i%cols)*tileW+20
		yy:= y+(i/cols)*tileH
		dc.SetRGBA255(255,255,255,10)
		RoundRect(dc,float64(x+4),float64(yy),float64(tileW-8),60,8)
		setFont(dc,12)
		dc.SetRGB255(160,180,200)
		drawStringAnchored(dc,k,float64(x+tileW/2),float64(yy+22),0.5,0.5)
		setFont(dc,18)
		dc.SetRGB255(255,255,255)
		drawStringAnchored(dc,data.Stats[k],float64(x+tileW/2),float64(yy+44),0.5,0.5)
	}
	return dc,nil
}

// State
type StateInfo struct {
	Name string; Level int; ApCur, ApTotal int
	SanityExpire string
	Chars []struct{Name string; Ap int}
}

func SampleState() *StateInfo {
	s:=&StateInfo{Name:"博士", Level:120, ApCur:135, ApTotal:135, SanityExpire:"12:30"}
	s.Chars=[]struct{Name string; Ap int}{{"阿米娅", 22},{"凯尔希", 20},{"煌", 18},{"能天使", 16}}
	return s
}

func RenderState(data *StateInfo) (*gg.Context, error) {
	const mainW=1092
	const mainH=510
	dc:=gg.NewContext(mainW,mainH)
	FillBackground(dc,46,48,49)
	// try draw bg image if exists
	if bg,err:=LoadImage(AssetPath("state/bg.png")); err==nil {
		dc.DrawImage(ScaleCover(bg,mainW,mainH),0,0)
	}
	// avatar
	dc.SetRGB255(80,80,90)
	dc.DrawCircle(34+27,34+27,27)
	dc.Fill()
	setFont(dc,26)
	dc.SetRGB255(255,255,255)
	drawString(dc,data.Name,100,60)
	setFont(dc,16)
	dc.SetRGB255(220,220,220)
	drawString(dc,fmt.Sprintf("Lv%d",data.Level),100,84)
	// ap bar
	frac:=float64(data.ApCur)/float64(data.ApTotal)
	if data.ApTotal==0 {frac=1}
	ProgressBar(dc,146,146,410,11,frac,255,255,255)
	setFont(dc,28)
	dc.SetRGB255(255,255,255)
	drawString(dc,fmt.Sprintf("%d/%d",data.ApCur,data.ApTotal),146,140)
	setFont(dc,16)
	dc.SetRGB255(200,200,200)
	drawString(dc,"理智恢复 "+data.SanityExpire,146,170)
	// chars row
	x:=20
	y:=260
	for _,c:=range data.Chars {
		dc.SetRGBA255(255,255,255,18)
		RoundRect(dc,float64(x),float64(y),120,120,10)
		dc.SetRGB255(80,80,90)
		dc.DrawCircle(float64(x+60),float64(y+44),28)
		dc.Fill()
		setFont(dc,12)
		dc.SetRGB255(255,255,255)
		drawStringAnchored(dc,c.Name,float64(x+60),float64(y+88),0.5,0.5)
		// ap badge
		dc.SetRGBA255(0,0,0,160)
		dc.DrawCircle(float64(x+60),float64(y+108),14)
		dc.Fill()
		setFont(dc,11)
		dc.SetRGB255(255,230,80)
		drawStringAnchored(dc,itoa(c.Ap),float64(x+60),float64(y+112),0.5,0.5)
		x+=140
	}
	// campaign / tradings placeholders
	y=410
	for i:=0;i<3;i++ {
		dc.SetRGBA255(255,255,255,10)
		RoundRect(dc,float64(20+i*260),float64(y),240,60,8)
		setFont(dc,13)
		dc.SetRGB255(200,220,200)
		drawString(dc,fmt.Sprintf("区块 %d",i+1),float64(40+i*260),float64(y+36))
	}
	return dc,nil
}

// helper for state to keep image import used
var _ = math.Pi
var _ = color.RGBA{}
