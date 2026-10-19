package ggrender

import (
	"fmt"
	"image"
	"math"
	"net/url"
	"time"

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
		if d, ok := data.(RecruitGroups); ok && d != nil {
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

// SampleBox carries the box-minimal fixture values. Provenance:
// feat/gg-render-canvas:src/core/web/static_props_export_test.go BoxFixture
// (in-repo declaration): 测试博士 + 11 identical Amiya chars.
func SampleBox() *BoxInfo {
	chars := make([]Char, 0, 11)
	for i := 0; i < 11; i++ {
		chars = append(chars, Char{
			CharId: "char_002_amiya", SkinId: "char_002_amiya#1", Name: "阿米娅",
			Level: 90, EvolvePhase: 2, PotentialRank: 5, FavorPercent: 100,
			Rarity: 5, Profession: "WARRIOR",
		})
	}
	return &BoxInfo{Name: "测试博士", Chars: chars}
}

// boxPortraitURL builds the URL exactly as Box.tmpl declares
// (char_skin/portrait/{{urlquery .SkinId}}.png); offline fallback per the
// template's onerror is assets/common/amiya.png. Both live URLs verified
// reachable (hycdn 200, prts thumb 200), so the fallback is a safety net.
func boxPortraitURL(skinID string) (string, string) {
	return "https://web.hycdn.cn/arknights/game/assets/char_skin/portrait/" + url.QueryEscape(skinID) + ".png", AssetPath("common/amiya.png")
}

// drawBoxTile draws one 70x140 .operator tile per Box/Missing.tmpl.
// Overlay geometry from display types + declared margins (all absolute
// insertions share the static position right after the 70px img, except the
// block-level .level/.name divs whose static position is below the img):
// profession 15px at (+3,+5); rarity 40px at (+20,+5); evolve 40px at
// (+20,+80); level 20px circle at (+5,+106); potential 30px at (+40,+100);
// name bar 70x15 at (+0,+127) with 10px centered text. Missing omits the
// evolve/level/potential row (its template has none).
func drawBoxTile(dc *gg.Context, x, y float64, portraitURL, profession string, rarity, evolve, potential, level int, name string, detail bool) {
	pURL, pFB := portraitURL, AssetPath("common/amiya.png")
	if portraitURL == "" || len(portraitURL) < 8 || portraitURL[:8] != "https://" {
		pURL, pFB = boxPortraitURL(portraitURL)
	}
	dc.DrawImage(ScaleExact(FetchImage(pURL, pFB), 70, 140), int(x), int(y))
	if prof := cardAsset("box/" + profession + ".png"); prof != nil {
		dc.DrawImage(ScaleExact(prof, 15, 15), int(x)+3, int(y)+5)
	}
	if rar := cardAsset("box/Rarity_" + itoa(rarity) + ".png"); rar != nil {
		rb := rar.Bounds()
		dc.DrawImage(ScaleExact(rar, 40, 40*rb.Dy()/rb.Dx()), int(x)+20, int(y)+5)
	}
	if detail {
		if evo := cardAsset("box/Evolve_" + itoa(evolve) + ".png"); evo != nil {
			eb := evo.Bounds()
			dc.DrawImage(ScaleExact(evo, 40, 40*eb.Dy()/eb.Dx()), int(x)+20, int(y)+80)
		}
		dc.SetRGB255(0, 0, 0)
		dc.DrawCircle(x+15, y+116, 10)
		dc.Fill()
		setFont(dc, 16)
		dc.SetRGB255(255, 255, 255)
		drawStringAnchored(dc, itoa(level), x+15, y+116, 0.5, 0.4)
		if pot := cardAsset("box/Potential_" + itoa(potential) + ".png"); pot != nil {
			dc.DrawImage(ScaleExact(pot, 30, 30), int(x)+40, int(y)+100)
		}
	}
	dc.SetRGBA255(0, 0, 0, 179) // .name{background-color:rgba(0,0,0,0.7)}
	dc.DrawRectangle(x, y+127, 70, 15)
	dc.Fill()
	setFont(dc, 10)
	dc.SetRGB255(255, 255, 255)
	drawStringAnchored(dc, name, x+35, y+134, 0.5, 0.5)
}

// drawBoxHeader draws the label.png bar (width 100%) with the 30px title
// pulled up by margin-top:-55. label.png is 644x70 -> 700x76.09.
func drawBoxHeader(dc *gg.Context, title string) float64 {
	if label := cardAsset("help/label.png"); label != nil {
		lb := label.Bounds()
		lh := 700 * lb.Dy() / lb.Dx()
		dc.DrawImage(ScaleExact(label, 700, lh), 0, 0)
	}
	labelH := 700.0 * 70 / 644
	setFont(dc, 30)
	dc.SetRGB255(255, 255, 255)
	cardTextTop(dc, title, 25, labelH-55)
	return labelH
}

func RenderBox(data *BoxInfo) (*gg.Context, error) {
	const mainW = 700
	dc := gg.NewContext(mainW, 357) // content 76.09+280 = 356.09, snapped 357
	FillBackground(dc, 46, 48, 49)  // #main{background-color:#2e3031}
	labelH := drawBoxHeader(dc, "Dr "+data.Name)
	for i, c := range data.Chars {
		x := float64((i % 10) * 70) // .operator{float:left}, 10 per 700px row
		y := labelH + float64(i/10)*140
		drawBoxTile(dc, x, y, c.SkinId, c.Profession, c.Rarity, c.EvolvePhase, c.PotentialRank, c.Level, c.Name, true)
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

// SampleBoxDetail carries the box-detail-minimal fixture values. Provenance:
// feat/gg-render-canvas:src/core/web/static_props_export_test.go
// BoxDetailFixture (in-repo declaration): 2 Amiya rows (90/80, evo2, pot5/4,
// 1+3 skills, 1+2 equips of type "original").
func SampleBoxDetail() []Detail {
	return []Detail{
		{Name: "阿米娅", Id: "char_002_amiya#1", Rarity: 5, Level: 90, EvolvePhase: 2, PotentialRank: 5, Skills: []Skill{{Id: "skcom_magic_rage[3]", Level: 10}}, Equips: []Equip{{Id: "original", Level: 1}}},
		{Name: "阿米娅", Id: "char_002_amiya#1", Rarity: 5, Level: 80, EvolvePhase: 2, PotentialRank: 4, Skills: []Skill{{Id: "skcom_magic_rage[3]", Level: 10}, {Id: "skchr_amiya_2", Level: 9}, {Id: "skchr_amiya_3", Level: 8}}, Equips: []Equip{{Id: "original", Level: 2}, {Id: "original", Level: 3}}},
	}
}

// boxDetailVendor loads a manifest-mapped capture file from assets/boxdetail.
// Provenance per file (shas verified on copy): avatar e95afb7b (BoxDetail
// avatar), skill 54ee21a6 (BoxDetail skill), equip 3680ff6e (BoxDetail equip).
// Same vendored-capture class as depot lmd.png; render never reads baseline/.
func boxDetailVendor(name string) (image.Image, error) {
	return LoadImage(AssetPath("boxdetail/" + name))
}

// boxDetailSkillIcon resolves a skill icon: the skillID value contains raw
// brackets no HTTP client sends, so it loads its manifest-mapped vendored
// file; other IDs go to hycdn; amiya is the template's onerror last resort.
func boxDetailSkillIcon(id string) image.Image {
	if id == "skcom_magic_rage[3]" {
		if img, err := boxDetailVendor("boxdetail-skill-amiya.png"); err == nil {
			return img
		}
	}
	if img, err := fetch("https://web.hycdn.cn/arknights/game/assets/char_skill/" + id + ".png"); err == nil {
		return img
	}
	img, _ := LoadImage(AssetPath("common/amiya.png"))
	return img
}

// boxDetailEquipIcon resolves a module icon: "original" loads its
// manifest-mapped vendored file; other IDs go to hycdn uniequip URLs.
func boxDetailEquipIcon(id string) image.Image {
	if id == "original" {
		if img, err := boxDetailVendor("boxdetail-equip-original.png"); err == nil {
			return img
		}
	}
	if img, err := fetch("https://web.hycdn.cn/arknights/game/assets/uniequip/type/icon/" + id + ".png"); err == nil {
		return img
	}
	img, _ := LoadImage(AssetPath("common/amiya.png"))
	return img
}

func RenderBoxDetail(data []Detail) (*gg.Context, error) {
	const mainW, mainH = 481, 186
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 46, 48, 49) // #main{background-color:#2e3031}
	// Column content widths from declarations: avatar 50 + name; evolve 50;
	// potential 50; skills n*50; equips n*50. Plus td 1px side padding each
	// and 2px cellspacing on 6 edges; excess split evenly (auto-lite).
	maxSkills, maxEquips := 0, 0
	for _, d := range data {
		if len(d.Skills) > maxSkills {
			maxSkills = len(d.Skills)
		}
		if len(d.Equips) > maxEquips {
			maxEquips = len(d.Equips)
		}
	}
	setFont(dc, 16)
	nameW, _ := measure(dc, "阿米娅")
	colContent := []float64{50 + nameW, 50, 50, float64(maxSkills) * 50, float64(maxEquips) * 50}
	sumMin := 10.0
	for _, w := range colContent {
		sumMin += w
	}
	excess := (float64(mainW) - 12 - sumMin) / 5
	colX := [6]float64{2}
	for i := 0; i < 5; i++ {
		colX[i+1] = colX[i] + colContent[i] + 2 + excess
	}
	// Header row: th bold centered (double-drawn, accepted 600 technique).
	headerH := 16*notoNormalLineBox + 2
	setFont(dc, 16)
	dc.SetRGB255(255, 255, 255)
	headers := []string{"干员", "等级", "潜能", "技能", "模组"}
	for i, h := range headers {
		w, _ := measure(dc, h)
		cx := (colX[i] + 1 + colX[i+1] - 1) / 2
		drawString(dc, h, cx-w/2+0.45, 2+1+(headerH-2-16)/2+16*0.8)
		drawString(dc, h, cx-w/2, 2+1+(headerH-2-16)/2+16*0.8)
	}
	y := 2 + headerH + 2
	for _, d := range data {
		rowH := 50.0 + 16*notoNormalLineBox + 2
		mid := func(top, h, ch float64) float64 { return top + (h-ch)/2 }
		// Avatar (vendored capture, remote fallback) + name, middle.
		av, aerr := boxDetailVendor("boxdetail-avatar-amiya.png")
		if aerr != nil {
			av = FetchImage(baseAvatarURL(d.Id), AssetPath("common/amiya.png"))
		}
		ax := colX[0] + 1
		ay := y + mid(0, rowH-2, 50)
		dc.DrawImage(ScaleExact(av, 50, 50), int(ax), int(ay))
		setFont(dc, 16)
		dc.SetRGB255(255, 255, 255)
		drawString(dc, d.Name, ax+50, ay+(50-16)/2+16*0.8)
		// Evolve icon + LV stacked column, middle.
		evx := colX[1] + 1 + (colX[2]-1-(colX[1]+1)-50)/2
		ey := y + mid(0, rowH-2, 50+16*notoNormalLineBox)
		if ev := cardAsset("box/Evolve_" + itoa(d.EvolvePhase) + ".png"); ev != nil {
			eb := ev.Bounds()
			dc.DrawImage(ScaleExact(ev, 50, 50*eb.Dy()/eb.Dx()), int(evx), int(ey))
		}
		drawString(dc, "LV"+itoa(d.Level), evx, ey+50+16*0.8)
		// Potential icon, middle.
		potx := colX[2] + 1 + (colX[3]-1-(colX[2]+1)-50)/2
		if pot := cardAsset("box/Potential_" + itoa(d.PotentialRank) + ".png"); pot != nil {
			dc.DrawImage(ScaleExact(pot, 50, 50), int(potx), int(y+mid(0, rowH-2, 50)))
		}
		// Skills + equips flow with whitespace gaps, middle bands.
		skillsW := float64(len(d.Skills))*50 + float64(len(d.Skills)-1)*3.584
		skx := colX[3] + 1 + (colX[4]-1-(colX[3]+1)-skillsW)/2
		sky := y + mid(0, rowH-2, 50+16*notoNormalLineBox)
		for k, s := range d.Skills {
			sx := skx + float64(k)*(50+3.584)
			dc.DrawImage(ScaleExact(boxDetailSkillIcon(s.Id), 50, 50), int(sx), int(sky))
			setFont(dc, 16)
			dc.SetRGB255(255, 255, 255)
			drawString(dc, "LV"+itoa(s.Level), sx, sky+50+16*0.8)
		}
		equipsW := float64(len(d.Equips))*50 + float64(len(d.Equips)-1)*3.584
		exx := colX[4] + 1 + (colX[5]-1-(colX[4]+1)-equipsW)/2
		exy := y + mid(0, rowH-2, 50+16*notoNormalLineBox)
		for k, e := range d.Equips {
			ex := exx + float64(k)*(50+3.584)
			eq := boxDetailEquipIcon(e.Id)
			eb := eq.Bounds()
			dc.DrawImage(ScaleExact(eq, 50, 50*eb.Dy()/eb.Dx()), int(ex), int(exy))
			setFont(dc, 16)
			dc.SetRGB255(255, 255, 255)
			drawString(dc, "LV"+itoa(e.Level), ex, exy+50+16*0.8)
		}
		y += rowH + 2
	}
	return ScaleToManifest(dc, 722, 279), nil
}

// BoxSummary

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

// SampleBoxSummary carries the box-summary-minimal fixture values.
// Provenance: feat/gg-render-canvas:src/core/web/static_props_export_test.go
// BoxSummaryFixture (in-repo declaration). Char counts stay STRINGS
// ("60/100" renders whole; atoi would zero them). 24 missing Amiyas.
func SampleBoxSummary() *BoxSummary {
	return &BoxSummary{
		Name: "测试博士", AllCharCnt: "60/100", Star6CharCnt: "20/50", Star5CharCnt: "25/30", Star4CharCnt: "15/20",
		AllEvolvePhase2Cnt: 40, Star6EvolvePhase2Cnt: 18, Star5EvolvePhase2Cnt: 15, Star4EvolvePhase2Cnt: 7,
		AllSkill10Cnt: 30, Star6Skill10Cnt: 15, Star5Skill10Cnt: 10, Star4Skill10Cnt: 5,
		AllSkill9Cnt: 20, Star6Skill9Cnt: 10, Star5Skill9Cnt: 8, Star4Skill9Cnt: 2,
		AllSkill8Cnt: 10, Star6Skill8Cnt: 5, Star5Skill8Cnt: 4, Star4Skill8Cnt: 1,
		AllEquipStage3Cnt: 12, Star6EquipStage3Cnt: 6, Star5EquipStage3Cnt: 4, Star4EquipStage3Cnt: 2,
		AllEquipStage2Cnt: 8, Star6EquipStage2Cnt: 4, Star5EquipStage2Cnt: 3, Star4EquipStage2Cnt: 1,
		AllEquipStage1Cnt: 4, Star6EquipStage1Cnt: 2, Star5EquipStage1Cnt: 1, Star4EquipStage1Cnt: 1,
		MissingChars: func() []MissingChar {
			m := make([]MissingChar, 0, 24)
			for i := 0; i < 24; i++ {
				m = append(m, MissingChar{SkinId: "https://media.prts.wiki/a/a0/%E5%8D%8A%E8%BA%AB%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85_1.png?image_process=format,webp/quality,Q_90", Name: "阿米娅", Rarity: 5, Profession: "WARRIOR"})
			}
			return m
		}(),
	}
}

// RenderBoxSummary draws BoxSummary.tmpl at design 900x482 (1350x723 @1.5).
// Structure (template-declared): label.png header (#label height 60, title
// 30px at margin-top -55; floats clear below the bar so the table starts at
// 60); table width 100% with border-spacing 20x5 (cols (900-5*20)/4 = 200);
// header th + 8 metric rows (flex div space-between, 10px side margins, 1px
// bottom borders); missing th + justified 40px portrait rows (last line
// natural gaps). Char counts stay strings ("60/100" renders whole).
// Portraits: the fixture URL names the half-body thumb, but rendered tiles
// behave square (row budget + face framing both prove it) and the manifest
// also maps amiya-avatar.webp (180x180, BoxSummary target) -- the vendored
// square file is drawn, remote thumb as fallback, amiya last.
func RenderBoxSummary(data *BoxSummary) (*gg.Context, error) {
	const mainW, mainH = 900, 482
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 46, 48, 49) // #main{background-color:#2e3031}
	if label := cardAsset("help/label.png"); label != nil {
		dc.DrawImage(ScaleExact(label, 900, 60), 0, 0)
	}
	setFont(dc, 30)
	dc.SetRGB255(255, 255, 255)
	cardTextTop(dc, "Dr "+data.Name, 25, 60-55)
	const colW = 200.0 // (900-5*20)/4
	colX := func(i int) float64 { return 20 + float64(i)*(colW+20) }
	setFont(dc, 16)
	dc.SetRGB255(255, 255, 255)
	y := 60.0 + 5 // top border-spacing, then header row
	ths := []string{"全部干员", "六星干员", "五星干员", "四星干员"}
	for i, h := range ths {
		w, _ := measure(dc, h)
		cx := colX(i) + 1 + (colW-2)/2
		drawString(dc, h, cx-w/2+0.45, y+1+(24-16)/2+16*0.8)
		drawString(dc, h, cx-w/2, y+1+(24-16)/2+16*0.8)
	}
	y += 16*notoNormalLineBox + 2 + 5
	metrics := [][5]string{
		{"招募干员数量", data.AllCharCnt, data.Star6CharCnt, data.Star5CharCnt, data.Star4CharCnt},
		{"精英阶段2干员", itoa(data.AllEvolvePhase2Cnt), itoa(data.Star6EvolvePhase2Cnt), itoa(data.Star5EvolvePhase2Cnt), itoa(data.Star4EvolvePhase2Cnt)},
		{"专精三技能数量", itoa(data.AllSkill10Cnt), itoa(data.Star6Skill10Cnt), itoa(data.Star5Skill10Cnt), itoa(data.Star4Skill10Cnt)},
		{"专精二技能数量", itoa(data.AllSkill9Cnt), itoa(data.Star6Skill9Cnt), itoa(data.Star5Skill9Cnt), itoa(data.Star4Skill9Cnt)},
		{"专精一技能数量", itoa(data.AllSkill8Cnt), itoa(data.Star6Skill8Cnt), itoa(data.Star5Skill8Cnt), itoa(data.Star4Skill8Cnt)},
		{"三级模组数量", itoa(data.AllEquipStage3Cnt), itoa(data.Star6EquipStage3Cnt), itoa(data.Star5EquipStage3Cnt), itoa(data.Star4EquipStage3Cnt)},
		{"二级模组数量", itoa(data.AllEquipStage2Cnt), itoa(data.Star6EquipStage2Cnt), itoa(data.Star5EquipStage2Cnt), itoa(data.Star4EquipStage2Cnt)},
		{"一级模组数量", itoa(data.AllEquipStage1Cnt), itoa(data.Star6EquipStage1Cnt), itoa(data.Star5EquipStage1Cnt), itoa(data.Star4EquipStage1Cnt)},
	}
	for _, m := range metrics {
		setFont(dc, 16)
		dc.SetRGB255(255, 255, 255)
		for i := 0; i < 4; i++ {
			drawString(dc, m[0], colX(i)+1+10, y+1+(24-16)/2+16*0.8)
			vw, _ := measure(dc, m[i+1])
			drawString(dc, m[i+1], colX(i)+colW-1-10-vw, y+1+(24-16)/2+16*0.8)
		}
		for i := 0; i < 4; i++ {
			dc.DrawRectangle(colX(i), y+24+2, colW, 1)
			dc.Fill()
		}
		y += 16*notoNormalLineBox + 2 + 1 + 5
	}
	// Missing th + portraits. td spans all columns; portraits flow with
	// justify spread on full lines, natural gaps on the last line.
	setFont(dc, 16)
	dc.SetRGB255(255, 255, 255)
	mw, _ := measure(dc, "未招募干员")
	drawString(dc, "未招募干员", 450-mw/2+0.45, y+1+(24-16)/2+16*0.8)
	drawString(dc, "未招募干员", 450-mw/2, y+1+(24-16)/2+16*0.8)
	y += 16*notoNormalLineBox + 2 + 5
	missImg, merr := LoadImage(AssetPath("boxsummary/amiya-avatar.webp"))
	if merr != nil {
		thumbURL := ""
		if len(data.MissingChars) > 0 {
			thumbURL = data.MissingChars[0].SkinId
		}
		missImg = FetchImage(thumbURL, AssetPath("common/amiya.png"))
	}
	lineW := 900.0 - 40 - 2
	py := y + 1
	perLine := int((lineW + 3.584) / (40 + 3.584)) // 19 by wrap math, not choice
	if perLine < 1 {
		perLine = 1
	}
	for start := 0; start < len(data.MissingChars); start += perLine {
		end := start + perLine
		if end > len(data.MissingChars) {
			end = len(data.MissingChars)
		}
		n := end - start
		gap := 3.584
		if end < len(data.MissingChars) && n > 1 {
			gap = (lineW - float64(n)*40) / float64(n-1)
		}
		xx := 21.0
		for k := start; k < end; k++ {
			_ = data.MissingChars[k]
			dc.DrawImage(ScaleExact(missImg, 40, 40), int(xx), int(py))
			xx += 40 + gap
		}
		py += 40 + inlineStrutDrop
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

// SampleEnemy carries the enemy-minimal fixture values. Provenance:
// feat/gg-render-canvas:src/core/web/complex_props_export_test.go
// visualComplexEnemyFixture (in-repo declaration): 源石虫 with prts pic URL,
// race 感染生物 / 普通 / 近战 / 地面, markup ability, and 2 full levels.
// Levels are kept for fidelity but fall beyond the 318-high manifest viewport
// (same clipping contract as base), so this renderer draws the #base table.
func SampleEnemy() *Enemy {
	return &Enemy{
		Name: "源石虫", Pic: "https://media.prts.wiki/3/3e/%E5%A4%B4%E5%83%8F_%E6%95%B5%E4%BA%BA_%E6%BA%90%E7%9F%B3%E8%99%AB.png", Desc: "拥有较高防御力的感染生物。", EnemyRace: "感染生物", EnemyLevel: "普通", AttackType: "近战", Motion: "地面",
		Ability: "<span>免疫沉默</span>",
		Levels: []EnemyLevel{
			{Desc: "标准个体", AttackType: "近战", Motion: "地面", HpRecovery: "0", HP: "5000", ATK: "400", DEF: "500", Res: "20", ATKRadius: "1.1", Weight: "2", MoveSpeed: "0.8", Interval: "1.5", DamageRes: "0", ElementRes: "0", Ridicule: "0", Point: "1", Abnormal: "无", Talent: "<b>生命低于50%时防御提升</b>", Skills: []EnemySkill{{Name: "啃噬", SpInit: "<span>0</span>", SpCost: "<span>5</span>", Desc: "对目标造成 <i>物理伤害</i>"}}},
			{Desc: "强化个体", AttackType: "近战", Motion: "地面", HpRecovery: "0", HP: "8000", ATK: "600", DEF: "700", Res: "30", ATKRadius: "1.1", Weight: "3", MoveSpeed: "0.8", Interval: "1.5", DamageRes: "10", ElementRes: "0", Ridicule: "1", Point: "2", Abnormal: "晕眩抗性", Talent: "<b>攻击力提升</b>", Skills: []EnemySkill{{Name: "啃噬+", SpInit: "<span>0</span>", SpCost: "<span>4</span>", Desc: "造成更高 <i>物理伤害</i>"}}},
		},
	}
}

// RenderEnemy draws template/Enemy.tmpl's #base table at the manifest design
// canvas 656x318 (984x477 @1.5). Per-level tables fall beyond the viewport and
// clip, same contract as base.
//
// Structure notes (all template-declared):
//   - {{.Ability}} sits in a bare <tr> with no <td>, so the HTML parser
//     foster-parents it BEFORE the table: plain 16px text at top-left.
//     Markup is stripped the way enemyProps/templateText does server-side.
//   - #main{width:656px}, #base{width:656px;border-spacing:0} with separate
//     1px cell borders: outer 1px, interior doubled 2px lines.
//   - No table-layout declaration, so columns follow auto layout. Implemented
//     as auto-lite: min from content (col1 = 158px img + 2 padding, others one
//     CJK char measured in the shipped font), excess split evenly. A stated
//     approximation of Blink's distribution, not a fitted constant.
//   - td padding is the 1px cellpadding default (no CSS override); img width
//     is the declared 158px; pic tops the cell, desc is middle-aligned.
//   - Line boxes come from the shipped font via measure(); baselines use the
//     in-repo 0.80 ascent fraction precedent (scene_card.go cardAscFrac).
func RenderEnemy(data *Enemy) (*gg.Context, error) {
	const mainW, mainH = 656, 318
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 0x32, 0x33, 0x32)
	setFont(dc, 16)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, StripHTML(data.Ability), 0, 17)
	tableTop := 24.0
	setFont(dc, 16)
	minCh, _ := measure(dc, "测")
	colW := [8]float64{160, minCh, minCh, minCh, minCh, minCh, minCh, minCh}
	sumMin := 0.0
	for _, w := range colW {
		sumMin += w
	}
	excess := (656 - 16 - sumMin) / 8
	for i := range colW {
		colW[i] += excess
	}
	colX := [9]float64{}
	colX[0] = 1
	for i := 0; i < 8; i++ {
		colX[i+1] = colX[i] + colW[i]
	}
	border := func() { dc.SetRGB255(0x59, 0x58, 0x58) }
	hline := func(yy float64, w float64) {
		border()
		dc.DrawRectangle(0, yy, 656, w)
		dc.Fill()
	}
	vlines := func(top, h float64, edges []float64) {
		border()
		dc.DrawRectangle(0, top, 1, h)
		dc.Fill()
		dc.DrawRectangle(655, top, 1, h)
		dc.Fill()
		for _, e := range edges {
			dc.DrawRectangle(e-1, top, 2, h)
			dc.Fill()
		}
	}
	centerText := func(s string, size float64, cx, top, boxH float64, bold bool) {
		setFont(dc, size)
		w, h := measure(dc, s)
		dc.SetRGB255(255, 255, 255)
		// 600-declared texts with only Regular shipped: Chromium synthesizes
		// the weight. Double-draw at +0.45 design px (calendar lesson: fixed
		// 0.6 overshoots small text, so this stays an experiment knob).
		if bold {
			drawString(dc, s, cx-w/2+0.45, top+boxH/2+h*(cardAscFrac-0.5))
		}
		drawString(dc, s, cx-w/2, top+boxH/2+h*(cardAscFrac-0.5))
	}
	// enemyLineBox lived here; hoisted to package-level notoNormalLineBox
	// (scene_card.go) so table scenes share one documented factor.
	lineH := func(size float64) float64 {
		return size * notoNormalLineBox
	}
	yy := tableTop
	hline(yy, 1)
	yy++
	nh := lineH(25) + 2
	centerText(data.Name, 25, 328, yy+1, nh-2, true)
	vlines(yy, nh, nil)
	yy += nh
	hline(yy, 2)
	yy += 2
	picH := 158.0
	// Portrait bytes: the fixture's prts URL 404s today (verified live), but
	// the capture manifest records serving these exact bytes for it
	// (targets:["Enemy pic"], sha34e239ab verified on copy), vendored to
	// assets/ per the depot lmd precedent. Template fallback (amiya) applies
	// only if the vendored file itself is missing.
	pic, perr := LoadImage(AssetPath("enemy/enemy-originium-slug-158.png"))
	if perr != nil {
		pic = FetchImage(data.Pic, AssetPath("common/amiya.png"))
	}
	dc.DrawImage(ScaleExact(pic, 158, 158), int(colX[0]+1+(colW[0]-2-158)/2), int(yy+1))
	setFont(dc, 16)
	descW, _ := measure(dc, data.Desc)
	descCx := (colX[1] + 1 + 653) / 2
	dc.SetRGB255(255, 255, 255)
	drawString(dc, data.Desc, descCx-descW/2, yy+1+picH/2+8)
	vlines(yy, picH+2, []float64{colX[1]})
	yy += picH + 2
	hline(yy, 2)
	yy += 2
	headers := []string{"种类", "地位级别", "攻击方式", "行动方式"}
	values := []string{data.EnemyRace, data.EnemyLevel, data.AttackType, data.Motion}
	pedges := []float64{colX[2], colX[4], colX[6]}
	hh := lineH(16) + 2
	for i := 0; i < 4; i++ {
		cx := (colX[2*i] + 1 + colX[2*i+2] - 1) / 2
		centerText(headers[i], 16, cx, yy+1, hh-2, false)
	}
	vlines(yy, hh, pedges)
	yy += hh
	hline(yy, 2)
	yy += 2
	for i := 0; i < 4; i++ {
		cx := (colX[2*i] + 1 + colX[2*i+2] - 1) / 2
		centerText(values[i], 16, cx, yy+1, hh-2, false)
	}
	vlines(yy, hh, pedges)
	yy += hh
	hline(yy, 2)
	yy += 2
	capH := lineH(20) + 2
	centerText("能力", 20, 328, yy+1, capH-2, true)
	vlines(yy, capH, nil)
	yy += capH
	hline(yy, 1)
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
// SampleHeadhunt emits exactly 10 entries, all Rarity 5, all with an empty
// ThumbURL, and carries no comment explaining where the 10 comes from.
// Headhunt.tmpl has no declaration fixing the entry count ({{range .}} iterates
// whatever is passed), so this count is a fixture fit to the baseline.
// OPEN. Revoking it needs a replacement fixture, and any replacement value would
// be invented rather than derived -- same class of problem as hhPitch above.
// Not changed in this commit. The commit text this fit was justified with is not
// present in the current tree (grep for it returns exit 1, empty output).
func SampleHeadhunt() []HHOp {
	ops := make([]HHOp, 0, 10)
	for i := 0; i < 10; i++ {
		ops = append(ops, HHOp{Rarity: 5, ThumbURL: "", Profession: "WARRIOR"})
	}
	return ops
}

// Geometry is derived from template/CSS declarations. The frozen baseline is used
// as cross-validation only; on conflict the declaration wins.
// (This reverses the earlier "measured wins on conflict", which was itself a
// policy defect, not only a red line. That sentence was introduced in the same
// commit that added these constants and has now been removed.)
//
//	card content width 95px  <- Headhunt.tmpl:25 .lh{width:95px}
//	card content top   230px <- .bg margin-top:130 + padding-top:100 (Headhunt.tmpl:20)
//	card content height 270px <- .bg height:270px (Headhunt.tmpl:21)
//	first card left edge  x=25 <- #main padding-left:25px (Headhunt.tmpl:10)
//	card pitch        98.6667 <- NOT derivable from Headhunt.tmpl. See hhPitch below.
const (
	hhOutW, hhOutH = 1049, 576 // #main width:1024 + padding-left:25 (Headhunt.tmpl:9,10); height:576 (Headhunt.tmpl:11)
	hhCardW        = 95        // .lh width:95px (Headhunt.tmpl:25)
	hhCardTop      = 230       // .bg margin-top:130 + padding-top:100 (Headhunt.tmpl:20)
	hhCardH        = 270       // .bg height:270px (Headhunt.tmpl:21)
	hhBackTop      = 130       // .bg margin-top:130px (Headhunt.tmpl:20)
	hhBackH        = 370       // .bg box = padding-top:100 + height:270 (margin-top is outside the box)
	hhFirstX       = 25.0      // #main padding-left:25px (Headhunt.tmpl:10)
	// FITTED VALUE, NOT A DECLARATION DERIVATION. Still OPEN.
	// Origin of the value: mean of the adjacent differences between the baseline's
	// 10 card left edges (888/9), not derived from Headhunt.tmpl.
	// The declaration-derived half is available: .bg has no width/margin, so the box
	// is shrink-to-fit and equals .lh width 95px (Headhunt.tmpl:25); .profession and
	// .rarity are position:absolute and do not contribute.
	// The other half is the collapsed whitespace between </div> and <div class="bg">,
	// i.e. one space at the inherited size (no font-size is declared on #main or on
	// body, so 16px by browser default). Measured from the font file, two ways:
	//   raw font units, fontTools 4.63.0: U+0020 advance 224/1000 em -> 3.584000px
	//   the stack this renderer actually uses, gg LoadFontFace at 16.0 (FreeType
	//   HintingFull, which rounds advances): -> 3.000000px
	// Neither reproduces 3.6667px, which is what 98.6667-95 implies. So the
	// "95 + space" derivation does NOT reproduce the fitted value, and the source of
	// 98.6667 stays unexplained. The two measurements were not tuned toward 3.6667.
	// OPEN -- to be replaced once the pitch can be derived from Headhunt.tmpl.
	hhPitch = 98.6667
	// Corrected back to the declaration, NOT a bugfix: the previous value was a
	// baseline-fitted number that happened to land 1px off the declared padding.
	// #main{padding-left:25px} and .bg is inline-block in #main's content box, so
	// the first card's left edge is 25.  (origin_source: css, Headhunt.tmpl:10)
	// Corrected back to the declaration, NOT a bugfix: .profession{margin-top:190px}
	// (Headhunt.tmpl:31). Previous 191 was baseline-fitted.
	hhFaceH  = 190 // .lh height:190 (Headhunt.tmpl:25)
	hhIconDX = 12  // asset inset, not template-declared
	hhIconDY = 190 // corrected back to .profession{margin-top:190px} (Headhunt.tmpl:31)
	hhStarDX = 14  // asset inset, not template-declared
	hhStarDY = 1   // asset inset, not template-declared
)

// OPEN (not restored, no declaration to restore from). Three elements that exist
// in the frozen baseline were removed in an earlier commit; Headhunt.tmpl declares
// no geometry for any of them, so the removal cannot be reversed from a declaration:
//  1. rarity "dots" (top right) -- the template has only
//     `.rarity{position:absolute;height:20px}`: no x, no y, and it is a single
//     Rarity_{{.Rarity}}.png image, not dots.
//  2. level badge "0" (bottom left) -- zero declaration; no element, class or text.
//  3. empty name bar -- zero declaration; no element, class or text.
//
// An unreliable baseline only invalidates the removal; it does not supply the
// geometry back. Inferring a value from "the baseline might be wrong" would be a
// second kind of fit, so these stay OPEN rather than be reconstructed.
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

// SampleMissing carries the missing-minimal fixture values. Provenance:
// feat/gg-render-canvas:src/core/web/static_props_export_test.go
// MissingFixture (in-repo declaration): 测试博士 + 11 Amiya thumbs.
func SampleMissing() *MissingInfo {
	chars := make([]MissingChar, 0, 11)
	for i := 0; i < 11; i++ {
		chars = append(chars, MissingChar{SkinId: "https://media.prts.wiki/a/a0/%E5%8D%8A%E8%BA%AB%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85_1.png?image_process=format,webp/quality,Q_90", Name: "阿米娅", Rarity: 5, Profession: "WARRIOR"})
	}
	return &MissingInfo{Name: "测试博士", Chars: chars}
}

func RenderMissing(data *MissingInfo) (*gg.Context, error) {
	const mainW = 700
	dc := gg.NewContext(mainW, 357)
	FillBackground(dc, 46, 48, 49) // #main{background-color:#2e3031}
	labelH := drawBoxHeader(dc, "Dr "+data.Name+"(未获取)")
	for i, c := range data.Chars {
		x := float64((i % 10) * 70)
		y := labelH + float64(i/10)*140
		drawBoxTile(dc, x, y, c.SkinId, c.Profession, c.Rarity, 0, 0, 0, c.Name, false)
	}
	return ScaleToManifest(dc, 1050, 536), nil
}

// Recruit
type RecruitOp struct {
	Avatar, Profession string
	Rarity             int
}

// RecruitGroups is the template's {{range .}} shape: one entry per table row,
// each with its own tags cell and operator strip. The frozen recruit-minimal
// fixture carries two groups (see canvas static_props_export_test.go
// RecruitFixture: 12 ops under 高级资深干员/输出, 2 under 术师干员/远程位),
// so a single flat group cannot express it (retires the A-1 note below).
type RecruitGroups []RecruitList
type RecruitList struct {
	Tags      []string
	Operators []RecruitOp
}

// SampleRecruit carries the recruit-minimal fixture values. Provenance:
// feat/gg-render-canvas:src/core/web/static_props_export_test.go
// RecruitFixture (in-repo declaration): two groups, all 14 operators 阿米娅 /
// WARRIOR / Rarity 5 / operatorAvatarURL. Retires the A-1 note: the count is
// 12+2 in two groups, not a flat 14.
func SampleRecruit() RecruitGroups {
	ops := func(n int) []RecruitOp {
		list := make([]RecruitOp, 0, n)
		for i := 0; i < n; i++ {
			list = append(list, RecruitOp{Avatar: "https://media.prts.wiki/3/36/%E5%A4%B4%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85.png?image_process=format,webp/quality,Q_90", Profession: "WARRIOR", Rarity: 5})
		}
		return list
	}
	return RecruitGroups{
		{Tags: []string{"高级资深干员", "输出"}, Operators: ops(12)},
		{Tags: []string{"术师干员", "远程位"}, Operators: ops(2)},
	}
}

// RenderRecruit draws Recruit.tmpl's tag/operator table at the manifest design
// canvas 900x356 (1350x534 @1.5). Structure (template-declared): header row
// (th 标签/干员）, one tr per group with a nowrap centered tags td and an ops
// td where .op tiles (100px cells, inline-flex with template-newline
// whitespace gaps) wrap by width. The ops column origin is the measured tags
// column, so the artX/cols/rowPitch/gridTop measured constants retire.
// .tag pills: height 28, padding 0 8, margin 4 at body-default 16px (no
// font-size declared anywhere on the tag path). Table cellspacing is the 2px
// default (no border-spacing declaration).
func RenderRecruit(groups RecruitGroups) (*gg.Context, error) {
	const mainW = 900
	const mainH = 356
	const cellW = 100.0
	const tileGap = 3.584 // inline .op whitespace nodes, same figure as depot
	const spacing = 2.0   // default table cellspacing, no CSS override
	// subPitch: wrapped lines stack at 100 + the line strut's drop below the
	// avatar baseline. Derivation (all declared/font-side): avatar 100 tall;
	// td font is the body-default 16px with Noto normal 1.5em = 24px line
	// (notoNormalLineBox); Noto descent 120/1000*16 = 1.92, half-leading
	// (24-16)/2 = 4, so the strut hangs 1.92+4 = 5.92 below the baseline:
	// pitch = 100+5.92 = 105.92. Cross-check only: (205-46)/1.5 = 106.0.
	const subPitch = 105.92
	dc := gg.NewContext(mainW, mainH)
	// D1: Recruit.tmpl declares no background on #main or body, so the frozen
	// baseline background is the browser default page colour #ffffff.
	FillBackground(dc, 255, 255, 255)
	tagSlot := func(t string) float64 {
		setFont(dc, 16)
		w, _ := measure(dc, t)
		return w + 16 + 8 // padding 0 8 + margin 4 each side
	}
	// Tags column = widest group row (td nowrap; UA 1px side padding each).
	tagsCol := 0.0
	for _, g := range groups {
		w := 0.0
		for _, t := range g.Tags {
			w += tagSlot(t)
		}
		if w > tagsCol {
			tagsCol = w
		}
	}
	tagsCol += 2
	tableX0 := spacing // border-spacing before the first cell
	tagsCellX := tableX0
	opsTdX := tagsCellX + tagsCol + spacing
	opsX0 := opsTdX + 1 // td 1px padding
	opsW := float64(mainW) - spacing - 1 - opsX0
	perRow := int((opsW + tileGap) / (cellW + tileGap))
	if perRow < 1 {
		perRow = 1
	}
	// Header row: th bold centered 16px; line box via notoNormalLineBox.
	// Table top edge carries border-spacing too, so the header starts at y=2.
	headerH := 16*notoNormalLineBox + 2
	setFont(dc, 16)
	dc.SetRGB255(0, 0, 0)
	hdrBase := spacing + 1 + (headerH-2-16)/2 + 16*0.8
	hdrDraw := func(s string, cx float64) {
		w, _ := measure(dc, s)
		drawString(dc, s, cx-w/2+0.45, hdrBase)
		drawString(dc, s, cx-w/2, hdrBase)
	}
	hdrDraw("标签", tagsCellX+1+tagsCol/2)
	hdrDraw("干员", opsX0+opsW/2)
	y := spacing + headerH + spacing
	// Retired: variants D2a/D2b1/D2b2 (gridTop/rowPitch/cols/artX measured
	// constants) and the full-width tags pill row, superseded by the table
	// structure above. Measurement records stay in _recruit_measure/.
	// (D2b2/gridTop/rowPitch/cols/artX discussion folded into the retired
	// note above; records in _recruit_measure/.)
	// blit 放到浮点设计坐标上：gg 的 DrawImage 只收 int，直接截断会在落点上
	// 白吃 1 个输出像素，所以用 Translate 交给仿射做小数定位。
	// 目标宽高只能取整（ScaleExact 签名是 int），量化误差 <= 0.5 设计像素，在此声明而非藏起。
	blit := func(img image.Image, x, y, w, h float64) {
		if img == nil {
			return
		}
		dc.Push()
		dc.Translate(x, y)
		dc.DrawImage(ScaleExact(img, int(w+0.5), int(h+0.5)), 0, 0)
		dc.Pop()
	}
	// 全表共用的两个素材在循环外只读一次。portrait 由 step-2 三向对照定为
	// assets/state/avatar-amiya.png（9.252 vs Q 版 66.951 vs 纯白 104.033，编码地板 6.515）。
	// 重采样用 ScaleExact：它与原 DrawPortraitTile 在方形素材上的 ScaleCover 等价，
	// 因此重采样口径不是本批次新引入的变量。
	port, _ := LoadImage(AssetPath("state/avatar-amiya.png"))
	prof, _ := LoadImage(AssetPath("box/WARRIOR.png"))
	for _, g := range groups {
		subRows := (len(g.Operators) + perRow - 1) / perRow
		if subRows < 1 {
			subRows = 1
		}
		opsH := float64(subRows) * subPitch
		rowH := opsH
		if rowH < 36 {
			rowH = 36
		}
		// Tags cell: nowrap pill line, vertically middle. Box-shadow on
		// .tag/.tr is declared but not drawn (open item, same as before).
		setFont(dc, 16)
		dc.SetRGB255(255, 255, 255)
		pillsTop := y + (rowH-36)/2 + 4
		px := tagsCellX + 1
		for _, t := range g.Tags {
			w, _ := measure(dc, t)
			dc.SetRGB255(0x31, 0x31, 0x31) // .tag{background-color:#313131}
			RoundRect(dc, px, pillsTop, w+16, 28, 6)
			dc.SetRGB255(255, 255, 255)
			drawString(dc, t, px+8, pillsTop+6+16*0.8)
			px += w + 16 + 8
		}
		// Ops tiles flow left to right, wrapping by width.
		gy := y + (rowH-opsH)/2
		for i, o := range g.Operators {
			x := opsX0 + float64(i%perRow)*(cellW+tileGap)
			yy := gy + float64(i/perRow)*subPitch
			// Overlay geometry from template declarations: 100x100 cell;
			// .profession{width:30px} at (0,0); .rarity{height:20px;
			// margin-left:30px}. No level badges or name strips in baseline.
			blit(port, x, yy, cellW, cellW)
			blit(prof, x, yy, 30, 30)
			// rarity width is auto: intrinsic aspect x 20 height.
			if r, err := LoadImage(AssetPath("box/Rarity_" + itoa(o.Rarity) + ".png")); err == nil {
				b := r.Bounds()
				blit(r, x+30, yy, float64(b.Dx())*20/float64(b.Dy()), 20)
			}
		}
		y += rowH + spacing
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

// baseAvatarURL builds the portrait URL exactly as Base.tmpl declares
// (char_skin/avatar/{{urlquery .Avatar}}.png). Offline it resolves to the
// template's own onerror fallback assets/common/amiya.png -- the same image
// the browser would show, not a gray placeholder the template never draws.
func baseAvatarURL(avatar string) string {
	return "https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/" + url.QueryEscape(avatar) + ".png"
}

// drawBaseMood reproduces the 20x20 AP mood svg declared on every char row in
// Base.tmpl: circle + face in #e42e20 (AP<=0), #f0ab22 (AP<100), #3cd627
// (AP>=100). Geometry from the declared path data (48 viewBox), not measured.
func drawBaseMood(dc *gg.Context, ap int, x, y, s float64) {
	var r, g, b int
	switch {
	case ap <= 0:
		r, g, b = 0xe4, 0x2e, 0x20
	case ap < 100:
		r, g, b = 0xf0, 0xab, 0x22
	default:
		r, g, b = 0x3c, 0xd6, 0x27
	}
	k := s / 48.0
	dc.SetRGB255(r, g, b)
	dc.SetLineWidth(4 * k)
	const seg = 40
	for i := 0; i < seg; i++ {
		a0 := 2 * math.Pi * float64(i) / seg
		a1 := 2 * math.Pi * float64(i+1) / seg
		dc.DrawLine(x+(24+20*math.Cos(a0))*k, y+(24+20*math.Sin(a0))*k, x+(24+20*math.Cos(a1))*k, y+(24+20*math.Sin(a1))*k)
	}
	line := func(x0, y0, x1, y1 float64) {
		dc.DrawLine(x+x0*k, y+y0*k, x+x1*k, y+y1*k)
	}
	if ap <= 0 {
		line(19, 18, 15, 22)
		line(15, 18, 19, 22)
		line(33, 18, 29, 22)
		line(29, 18, 33, 22)
	} else {
		line(31, 18, 31, 19)
		line(17, 18, 17, 19)
	}
	dc.Stroke()
	if ap <= 0 {
		dc.DrawRoundedRectangle(x+20*k, y+28*k, 8*k, 8*k, 4*k)
		dc.Stroke()
	} else {
		var mouth [5][2]float64
		if ap < 100 {
			mouth = [5][2]float64{{31, 31}, {27, 28}, {24, 27.5}, {21, 28}, {17, 31}}
		} else {
			mouth = [5][2]float64{{31, 31}, {27, 34}, {24, 35}, {21, 34}, {17, 31}}
		}
		for i := 0; i < 4; i++ {
			line(mouth[i][0], mouth[i][1], mouth[i+1][0], mouth[i+1][1])
		}
		dc.Stroke()
	}
}

// SampleBase carries the base-minimal fixture values. Provenance: the payload
// mirrors feat/gg-render-canvas:src/core/web/complex_props_export_test.go
// visualComplexBaseFixture ("mirror the approved Temp complex_fixtures.go
// payloads"). Every value below is identical to that in-repo declaration:
// labor 42/99, 1 dorm (Lv5/20000/Exusiai), 1 trading (Lv3/贵金属订单 3/3/Texas),
// 1 manufacture (Lv3/赤金 1/3/Siege), 1 power (Lv3/270/Saria), meeting board
// [1,7] sharing (Kal'tsit), hire refresh 3 (SilverAsh), training skill
// skchr_amiya_3 SLevel 10 (Eyjafjalla). Previous values were a different,
// larger invented dataset (2 dorms, 108/120 labor, ...) that overflowed the
// 612-high viewport and matched nothing.
func SampleBase() *BaseInfo {
	b := &BaseInfo{Name: "博士的基建"}
	b.Labor.Cur = 42
	b.Labor.Total = 99
	b.Control.Level = 5
	b.Control.Chars = []BaseChar{{Name: "阿米娅", Avatar: "char_002_amiya#1", AP: 0}}
	b.Dorms = []struct {
		Level   int
		Chars   []BaseChar
		Comfort int
	}{
		{Level: 5, Chars: []BaseChar{{Name: "能天使", Avatar: "char_103_angel#1", AP: 50}}, Comfort: 20000},
	}
	b.Tradings = []struct {
		Level      int
		Chars      []BaseChar
		Cur, Total int
		Strategy   string
	}{
		{Level: 3, Chars: []BaseChar{{Name: "德克萨斯", Avatar: "char_102_texas#1", AP: 100}}, Cur: 3, Total: 3, Strategy: "贵金属订单"},
	}
	b.Manufactures = []struct {
		Level       int
		Chars       []BaseChar
		Cur, Total  int
		Item, Speed string
	}{
		{Level: 3, Chars: []BaseChar{{Name: "推进之王", Avatar: "char_112_siege#1", AP: 0}}, Cur: 1, Total: 3, Item: "赤金", Speed: "130%"},
	}
	b.Powers = []struct {
		Level int
		Chars []BaseChar
		Power int
	}{
		{Level: 3, Chars: []BaseChar{{Name: "塞雷娅", Avatar: "char_202_demkni#1", AP: 50}}, Power: 270},
	}
	b.Meeting.Level = 3
	b.Meeting.Chars = []BaseChar{{Name: "凯尔希", Avatar: "char_003_kalts#1", AP: 100}}
	b.Meeting.Board = []int{1, 7}
	b.Meeting.Sharing = true
	b.Hire.Level = 3
	b.Hire.Chars = []BaseChar{{Name: "银灰", Avatar: "char_172_svrash#1", AP: 0}}
	b.Hire.Refresh = 3
	b.Training.Level = 3
	b.Training.Chars = []BaseChar{{Name: "艾雅法拉", Avatar: "char_180_amgoat#1", AP: 50}}
	b.Training.Skill = "skchr_amiya_3"
	b.Training.SLevel = 10
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
	const cardW, wideW = 550, 1105 // C01 css 13 / C09 declared 85,128
	// C02 css 14 declares .base{height:110px} with the UA default content-box:
	// the 1px border (C04 css 16) sits OUTSIDE the content box, so the border
	// box is 112 tall and the row pitch is 112+5=117, not 115. Drawing the
	// border inside a 110 box shrank every card 2px and drifted the bottom
	// rows ~10px up (edge-measured pitch 175 device = 117 design).
	const cardH, gapY = 112, 5          // C02 css 14 + content-box border / C07 css 19
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
		// D07 css 51: .chars{margin-left:10px}. Per-char span is
		// display:inline-grid: an inline-flex row (40px portrait + 20px AP
		// mood svg + name, vertically centered) with the progress.ap bar
		// below spanning the row width. No gray placeholder circle exists in
		// the template; portraits resolve per the declared onerror fallback.
		setFont(dc, 16)
		for i, c := range chars {
			x0 := px + charsML + float64(i)*70
			rowTop := py + 52
			port := FetchImage(baseAvatarURL(c.Avatar), AssetPath("common/amiya.png"))
			// Square slot: avatar sources are square thumbs (cf. .avatar 180x180
			// use elsewhere); cover-crop the fallback into the slot instead of
			// stretching it.
			dc.DrawImage(ScaleCover(port, 40, 40), int(x0), int(rowTop))
			drawBaseMood(dc, c.AP, x0+40, rowTop+10, 20)
			dc.SetRGB255(255, 255, 255)
			drawString(dc, c.Name, x0+60, rowTop+26)
			// progress.ap: Base.tmpl css declares progress{width:150px;height:3px}
			// with a white -webkit-progress-value. Fixed 150x3, not row width.
			dc.SetRGB255(85, 85, 85)
			dc.DrawRectangle(x0, rowTop+44, 150, 3)
			dc.Fill()
			if c.AP > 0 {
				frac := float64(c.AP) / 100.0
				if frac > 1 {
					frac = 1
				}
				dc.SetRGB255(255, 255, 255)
				dc.DrawRectangle(x0, rowTop+44, 150*frac, 3)
				dc.Fill()
			}
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
type CardNation struct {
	Name string
	Flag int64 // 1 highlighted, 0 plain, -1 dimmed
}
type CardAssist struct {
	Name        string
	SkinID      string
	Level       int
	EvolvePhase int
	IsSpecMax   bool
}
type CardInfo struct {
	Name, Uid, ServerName, Resume              string
	Level, RegTime                             int
	MainStageProgress, Avatar, Secretary       string
	SecretaryName, SecretaryEnName             string
	CharCnt, FurnitureCnt, SkinCnt             int
	EquipCnt, EquipStage3Cnt, EquipOperatorCnt int
	NationList                                 []CardNation
	AssistChars                                []CardAssist
}

// SampleCard carries the card-minimal fixture values. Provenance:
// feat/gg-render-canvas:src/core/web/complex_props_export_test.go
// visualComplexCardFixture (in-repo declaration). RegTime 1704067200 renders
// 2024-01-01; assists carry real SkinIDs (remote, amiya fallback offline).
func SampleCard() *CardInfo {
	return &CardInfo{
		Name: "冻结博士", Uid: "10000001", ServerName: "官服", Resume: "稳定基线签名",
		Level: 120, RegTime: 1704067200, MainStageProgress: "14-21",
		Avatar:        "https://fixture-cache.invalid/card/player-avatar.png",
		Secretary:     "https://fixture-cache.invalid/card/secretary-painting.png",
		SecretaryName: "阿米娅", SecretaryEnName: "Amiya",
		CharCnt: 289, FurnitureCnt: 100, SkinCnt: 88,
		EquipCnt: 46, EquipStage3Cnt: 12, EquipOperatorCnt: 23,
		NationList: []CardNation{{Name: "rhodes", Flag: 1}, {Name: "lungmen", Flag: 0}, {Name: "yan", Flag: -1}},
		AssistChars: []CardAssist{
			{Name: "阿米娅", SkinID: "char_002_amiya#1", Level: 90, EvolvePhase: 2, IsSpecMax: true},
			{Name: "能天使", SkinID: "char_103_angel#1", Level: 90, EvolvePhase: 2, IsSpecMax: false},
			{Name: "德克萨斯", SkinID: "char_102_texas#1", Level: 80, EvolvePhase: 2, IsSpecMax: true},
		},
	}
}

// ponytail: depot split to depot.go for per-scene ownership (renderContext routes there)
// DepotData/DepotItem/SampleDepot/RenderDepot moved to src/ggrender/depot.go

// Gacha
type GachaPool struct {
	PoolName  string
	PoolCount int
}
type GachaChar struct {
	PoolName, CharName, Avatar string
	Rarity                     int64
	IsNew                      bool
	Ts                         int64
}
type GachaStar6 struct {
	Name, Avatar, PoolName string
	Ts                     int64
	Count                  int
	IsNew                  bool
}
type GachaData struct {
	Name                              string
	Total, Star6, Star5, Star4, Star3 int
	Avg6, Avg5, Avg4, Avg3            float64
	BegTime, EndTime                  int64 // ms, rendered via timestampToTime
	FooterDate                        string
	PoolCount                         []GachaPool
	Chars                             []GachaChar
	Star6Info                         []GachaStar6
}

// SampleGacha carries the gacha-minimal fixture values. Provenance:
// feat/gg-render-canvas legacyGachaFixture (in-repo declaration): 基线博士,
// 10 pulls, 2/2/3/3 stars, avgs 5/5/3.33/3.33, ms timestamps, pools 甲6/乙4,
// 2 new chars + 2 star6 infos. FooterDate is the frozen capture clock
// (capture-clock-erratum.json gacha entry 1787070547 -> 2026年08月19日
// Asia/Shanghai); getNow() must never run in the harness (gacha advance).
func SampleGacha() *GachaData {
	return &GachaData{
		Name: "基线博士", Total: 10, Star6: 2, Star5: 2, Star4: 3, Star3: 3,
		Avg6: 5, Avg5: 5, Avg4: 3.33, Avg3: 3.33,
		BegTime: 1736856000000, EndTime: 1736942400000,
		FooterDate: "2026年08月19日",
		PoolCount:  []GachaPool{{PoolName: "测试卡池甲", PoolCount: 6}, {PoolName: "测试卡池乙", PoolCount: 4}},
		Chars: []GachaChar{
			{PoolName: "测试卡池甲", CharName: "阿米娅", Avatar: "https://media.prts.wiki/3/36/%E5%A4%B4%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85.png?image_process=format,webp/quality,Q_90", Rarity: 5, IsNew: true, Ts: 1736856000000},
			{PoolName: "测试卡池乙", CharName: "能天使", Avatar: "https://media.prts.wiki/a/ad/%E5%A4%B4%E5%83%8F_%E8%83%BD%E5%A4%A9%E4%BD%BF.png?image_process=format,webp/quality,Q_90", Rarity: 4, IsNew: false, Ts: 1736942400000},
		},
		Star6Info: []GachaStar6{
			{Name: "阿米娅", Avatar: "https://media.prts.wiki/3/36/%E5%A4%B4%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85.png?image_process=format,webp/quality,Q_90", Ts: 1736856000000, Count: 5, IsNew: true, PoolName: "测试卡池甲"},
			{Name: "能天使", Avatar: "https://media.prts.wiki/a/ad/%E5%A4%B4%E5%83%8F_%E8%83%BD%E5%A4%A9%E4%BD%BF.png?image_process=format,webp/quality,Q_90", Ts: 1736942400000, Count: 5, IsNew: false, PoolName: "测试卡池乙"},
		},
	}
}

// gachaClock formats ms timestamps in Asia/Shanghai exactly like the
// template's timestampToTime (full) / formatDate (date) JS helpers.
func gachaClock(ms int64, withTime bool) string {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*60*60)
	}
	t := time.UnixMilli(ms).In(loc)
	if withTime {
		return t.Format("2006-01-02 15:04:05")
	}
	return t.Format("2006-01-02")
}

func RenderGacha(data *GachaData) (*gg.Context, error) {
	const mainW, mainH = 1000, 882
	dc := gg.NewContext(mainW, mainH)
	// Header bg exact asset.
	if header := cardAsset("gacha/header.png"); header != nil {
		dc.DrawImage(header, 0, 0)
	}
	dc.SetRGB255(0xee, 0xee, 0xee)
	setFont(dc, 32)
	cardTextTop(dc, data.Name, 320, 50)
	// Second h1 is absolute too: sequential from x=250 at the same baseline.
	period := fmt.Sprintf("共%d抽", data.Total)
	setFont(dc, 32)
	pw, _ := measure(dc, period)
	drawString(dc, period, 250, 96+38)
	dates := fmt.Sprintf("(%s——%s)", gachaClock(data.BegTime, true), gachaClock(data.EndTime, true))
	setFont(dc, 23)
	drawString(dc, dates, 250+pw, 96+38)
	// Three .item boxes: 300x250, inline-block with ml 20 + newline gaps.
	itemX := func(i int) float64 { return 20 + float64(i)*(300+20+3.584) }
	itemTitles := []string{"星级分布", "星级分布", "卡池分布(最近10个)"}
	for i := 0; i < 3; i++ {
		dc.SetRGB255(0x1f, 0x1e, 0x1e)
		dc.DrawRoundedRectangle(itemX(i), 150, 300, 250, 20)
		dc.Fill()
		setFont(dc, 20)
		dc.SetRGB255(0xee, 0xee, 0xee)
		drawStringAnchored(dc, itemTitles[i], itemX(i)+150, 150+10+15, 0.5, 0.5)
	}
	// Pie (item 0): slices from echarts data (start 90deg, clockwise) with
	// declared colors; labels off; legend vertical at left/top:45.
	pieCx, pieCy, pieR := itemX(0)+187.5, 150.0+140.0, 79.0
	pieVals := []struct {
		v       int
		r, g, b int
	}{
		{data.Star6, 244, 110, 30}, {data.Star5, 247, 171, 55},
		{data.Star4, 161, 53, 246}, {data.Star3, 109, 116, 126},
	}
	pieTotal := 0
	for _, p := range pieVals {
		pieTotal += p.v
	}
	ang := -math.Pi / 2
	for _, p := range pieVals {
		if p.v <= 0 || pieTotal <= 0 {
			continue
		}
		sweep := 2 * math.Pi * float64(p.v) / float64(pieTotal)
		dc.MoveTo(pieCx, pieCy)
		dc.DrawArc(pieCx, pieCy, pieR, ang, ang+sweep)
		dc.ClosePath()
		dc.SetRGB255(p.r, p.g, p.b)
		dc.Fill()
		ang += sweep
	}
	setFont(dc, 14)
	legends := []struct {
		label   string
		r, g, b int
	}{
		{"6星", 244, 110, 30}, {"5星", 247, 171, 55},
		{"4星", 161, 53, 246}, {"3星", 109, 116, 126},
	}
	for i, lg := range legends {
		ly := 150.0 + 40 + 45 + float64(i)*(14*1.5+18)
		dc.SetRGB255(lg.r, lg.g, lg.b)
		dc.DrawRectangle(itemX(0)+8, ly, 8, 8)
		dc.Fill()
		dc.SetRGB255(255, 255, 255)
		pct := 0.0
		if pieTotal > 0 {
			pct = float64(pieVals[i].v) * 100 / float64(pieTotal)
		}
		drawString(dc, fmt.Sprintf("%s  %.2f%%", lg.label, pct), itemX(0)+20, ly+12)
	}
	// Avg table (item 1): rows block-margin 15 apart (collapsed), 150px tds.
	setFont(dc, 16)
	dc.SetRGB255(0xee, 0xee, 0xee)
	avgRows := [][2]string{
		{fmt.Sprintf("%d个6星", data.Star6), fmt.Sprintf("%v抽/个", data.Avg6)},
		{fmt.Sprintf("%d个5星", data.Star5), fmt.Sprintf("%v抽/个", data.Avg5)},
		{fmt.Sprintf("%d个4星", data.Star4), fmt.Sprintf("%v抽/个", data.Avg4)},
		{fmt.Sprintf("%d个3星", data.Star3), fmt.Sprintf("%v抽/个", data.Avg3)},
	}
	ay := 150.0 + 40 + 15
	for _, r := range avgRows {
		drawString(dc, r[0], itemX(1)+30, ay+19)
		drawString(dc, r[1], itemX(1)+30+150, ay+19)
		ay += 16 * notoNormalLineBox
	}
	// Bar chart (item 2): horizontal bars, first pool at bottom (echarts
	// category default), bar color is the echarts default blue (no itemStyle
	// declared), white value labels right of bars. Positions approximate.
	bx := itemX(2)
	maxV := 1
	for _, p := range data.PoolCount {
		if p.PoolCount > maxV {
			maxV = p.PoolCount
		}
	}
	plotX0, plotX1 := bx+110, bx+270
	setFont(dc, 14)
	for i, p := range data.PoolCount {
		by := 150.0 + 40 + 30 + float64(len(data.PoolCount)-1-i)*90
		dc.SetRGB255(255, 255, 255)
		drawStringAnchored(dc, p.PoolName, plotX0-8, by+20, 1, 0.5)
		blen := float64(p.PoolCount) / float64(maxV) * (plotX1 - plotX0)
		dc.SetRGB255(0x54, 0x70, 0xc6)
		dc.DrawRectangle(plotX0, by, blen, 44)
		dc.Fill()
		dc.SetRGB255(255, 255, 255)
		drawString(dc, itoa(p.PoolCount), plotX0+blen+6, by+30)
	}
	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(1.5)
	dc.DrawLine(plotX0, 150+40+10, plotX0, 150+40+200)
	dc.Stroke()
	// Article: bg band with two .chars panels (465 wide, inline-table with
	// newline gaps, margin-top 40). Each panel: 20px title + 230px tables
	// side by side (avatar 100 + text col), footer pinned at 613.
	dc.SetRGB255(0x0c, 0x0d, 0x0c) // #article{background-color:#0c0d0c}
	dc.DrawRectangle(0, 400, mainW, 613-400)
	dc.Fill()
	panelX := func(i int) float64 { return 20 + float64(i)*(465+3.584) }
	panelTitles := []string{"新获得干员(至多显示20个)", "获得六星干员(至多显示20个)"}
	charLists := []struct {
		name, avatar, pool string
		ts                 int64
		isNew              bool
		count              int // 花费抽, star6 panel only (-1 = empty row)
	}{}
	_ = charLists
	drawCharTable := func(tx, ty float64, name, avatar, pool string, ts int64, isNew bool, count int) {
		av := FetchImage(avatar, AssetPath("common/amiya.png"))
		dc.DrawImage(ScaleExact(av, 100, 100), int(tx), int(ty))
		setFont(dc, 16)
		dc.SetRGB255(0xee, 0xee, 0xee)
		drawString(dc, name, tx+100, ty+19)
		if isNew {
			setFont(dc, 10)
			dc.SetRGB255(255, 0, 0) // .new{color:red}
			nw, _ := measure(dc, name)
			drawString(dc, "New", tx+100+nw, ty+8)
		}
		setFont(dc, 15)
		dc.SetRGB255(0xee, 0xee, 0xee)
		drawString(dc, gachaClock(ts, false), tx+100, ty+19+22)
		drawString(dc, pool, tx+100, ty+19+44)
		if count >= 0 {
			drawString(dc, fmt.Sprintf("花费%d抽", count), tx+100, ty+19+66)
		}
	}
	for pi := 0; pi < 2; pi++ {
		px := panelX(pi)
		dc.SetRGB255(0x1f, 0x1e, 0x1e)
		dc.DrawRoundedRectangle(px, 440, 465, 140, 20)
		dc.Fill()
		setFont(dc, 20)
		dc.SetRGB255(0xee, 0xee, 0xee)
		drawStringAnchored(dc, panelTitles[pi], px+465.0/2, 440+10+15, 0.5, 0.5)
		for k := 0; k < 2; k++ {
			tx := px + float64(k)*(230+3.584)
			ty := 440.0 + 40
			if pi == 0 {
				if k >= len(data.Chars) {
					continue
				}
				c := data.Chars[k]
				drawCharTable(tx, ty, c.CharName, c.Avatar, c.PoolName, c.Ts, c.IsNew, -1)
			} else {
				if k >= len(data.Star6Info) {
					continue
				}
				s := data.Star6Info[k]
				drawCharTable(tx, ty, s.Name, s.Avatar, s.PoolName, s.Ts, s.IsNew, s.Count)
			}
		}
	}
	// Footer: exact asset at manifest-pinned top + frozen clock date text.
	if footer := cardAsset("gacha/footer.png"); footer != nil {
		dc.DrawImage(footer, 0, 613)
	}
	setFont(dc, 32)
	dc.SetRGB255(0xee, 0xee, 0xee)
	drawString(dc, data.FooterDate, 530, 613+190+26)
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

// SampleHelp carries the help-minimal fixture values. Provenance:
// feat/gg-render-canvas:src/core/web/static_props_export_test.go HelpFixture
// (in-repo declaration): 6 private + 36 public + 9 admin commands.
func SampleHelp() *HelpData {
	return &HelpData{
		Private: []Cmd{
			{Cmd: "/bind", Desc: "绑定角色", Param: "", IsBind: false},
			{Cmd: "/unbind", Desc: "解绑角色", Param: "", IsBind: true},
			{Cmd: "/cancel", Desc: "取消操作", Param: "", IsBind: false},
			{Cmd: "/reset_token", Desc: "重设token", Param: "", IsBind: true},
			{Cmd: "/import_gacha", Desc: "导入抽卡记录", Param: "", IsBind: true},
			{Cmd: "/export_gacha", Desc: "导出抽卡记录", Param: "", IsBind: true},
		},
		Public: []Cmd{
			{Cmd: "/help", Desc: "使用说明", Param: "", IsBind: false},
			{Cmd: "/ping", Desc: "存活测试", Param: "", IsBind: false},
			{Cmd: "/tag", Desc: "自定义群标签", Param: "标签", IsBind: false},
			{Cmd: "/sign", Desc: "签到", Param: "", IsBind: true},
			{Cmd: "/sign", Desc: "开启自动签到", Param: "auto", IsBind: true},
			{Cmd: "/sign", Desc: "关闭自动签到", Param: "stop", IsBind: true},
			{Cmd: "/sign", Desc: "全部通知", Param: "notify_all", IsBind: true},
			{Cmd: "/sign", Desc: "仅失败时通知", Param: "notify_fail", IsBind: true},
			{Cmd: "/sign", Desc: "仅成功时通知", Param: "notify_success", IsBind: true},
			{Cmd: "/ap", Desc: "开启理智提醒", Param: "on", IsBind: true},
			{Cmd: "/ap", Desc: "关闭理智提醒", Param: "off", IsBind: true},
			{Cmd: "/ap", Desc: "设理智提醒阈值", Param: "thr [1-100]", IsBind: true},
			{Cmd: "/state", Desc: "当前状态", Param: "", IsBind: true},
			{Cmd: "/box", Desc: "我的干员(默认6星)", Param: "", IsBind: true},
			{Cmd: "/box", Desc: "所有干员", Param: "all", IsBind: true},
			{Cmd: "/box", Desc: "对应星级干员", Param: "5,6", IsBind: true},
			{Cmd: "/box_detail", Desc: "干员详情(默认6星)", Param: "", IsBind: true},
			{Cmd: "/box_detail", Desc: "对应星级干员", Param: "5", IsBind: true},
			{Cmd: "/box_summary", Desc: "干员信息汇总", Param: "", IsBind: true},
			{Cmd: "/missing", Desc: "未获取干员(默认6星)", Param: "", IsBind: true},
			{Cmd: "/missing", Desc: "所有未获取干员", Param: "all", IsBind: true},
			{Cmd: "/missing", Desc: "对应星级未获取干员", Param: "5,6", IsBind: true},
			{Cmd: "/card", Desc: "我的名片", Param: "", IsBind: true},
			{Cmd: "/base", Desc: "基建信息", Param: "", IsBind: true},
			{Cmd: "/gacha", Desc: "抽卡记录", Param: "", IsBind: true},
			{Cmd: "/operator", Desc: "干员查询", Param: "", IsBind: false},
			{Cmd: "/skin", Desc: "干员皮肤查询", Param: "", IsBind: false},
			{Cmd: "/enemy", Desc: "敌人查询", Param: "", IsBind: false},
			{Cmd: "/report", Desc: "举报", Param: "", IsBind: false},
			{Cmd: "/quiz", Desc: "云玩家检测", Param: "", IsBind: false},
			{Cmd: "/quiz", Desc: "云玩家检测(困难)", Param: "h", IsBind: false},
			{Cmd: "/redeem", Desc: "CDK兑换", Param: "[CDK]", IsBind: true},
			{Cmd: "/headhunt", Desc: "寻访模拟", Param: "", IsBind: false},
			{Cmd: "/recruit", Desc: "公招计算(图片附带)", Param: "", IsBind: false},
			{Cmd: "/calendar", Desc: "活动日历", Param: "", IsBind: false},
			{Cmd: "/depot", Desc: "我的仓库", Param: "", IsBind: true},
		},
		Admin: []Cmd{
			{Cmd: "/news", Desc: "开启/关闭动态推送", Param: "", IsBind: false},
			{Cmd: "/birthday", Desc: "开启/关闭生日推送", Param: "", IsBind: false},
			{Cmd: "/request_mode", Desc: "切换群验证模式", Param: "", IsBind: false},
			{Cmd: "/quiz", Desc: "开启云玩家检测", Param: "on", IsBind: false},
			{Cmd: "/quiz", Desc: "关闭云玩家检测", Param: "off", IsBind: false},
			{Cmd: "/headhunt", Desc: "开启寻访模拟", Param: "on", IsBind: false},
			{Cmd: "/headhunt", Desc: "关闭寻访模拟", Param: "off", IsBind: false},
			{Cmd: "/reg", Desc: "回复消息设置为群规", Param: "", IsBind: false},
			{Cmd: "/welcome", Desc: "设置入群欢迎信息", Param: "文本", IsBind: false},
		},
	}
}

// drawHelpPerson approximates the 16px bootstrap person-circle glyphs used
// for bind markers: outer ring + head dot, currentColor white.
func drawHelpPerson(dc *gg.Context, x, y float64) {
	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(1.5)
	dc.DrawCircle(x+8, y+8, 7)
	dc.Stroke()
	dc.DrawCircle(x+8, y+6, 2.2)
	dc.Fill()
}

// RenderHelp draws Help.tmpl at design 660x1366 (990x2049 @1.5). Structure
// (template-declared): banner.png (100% width, margin-top 10) with h1/p
// overlays; .bg dark gradient + amiya cover block; per group a 40px label
// bar with cmdType overlay, then 150px .cmd pills (15px/600, border 1px,
// radius 10) floating 4 per row with two text lines + optional 16px bind
// glyph. tr/pill box-shadows declared but not drawn (open item). The
// person-circle glyphs render as ring+dot (stated approximation of the
// bootstrap path data, illegible at 16px either way).
func RenderHelp(data *HelpData) (*gg.Context, error) {
	const mainW, mainH = 660, 1366
	dc := gg.NewContext(mainW, mainH)
	if bg := cardAsset("help/bg.jpg"); bg != nil {
		dc.DrawImage(ScaleCover(bg, mainW, mainH), 0, 0)
	}
	// Banner: img 660x239 at y=10; h1 32px bold + bind-note p overlaid.
	if banner := cardAsset("help/banner.png"); banner != nil {
		bb := banner.Bounds()
		bh := 660 * bb.Dy() / bb.Dx()
		dc.DrawImage(ScaleExact(banner, 660, bh), 0, 10)
	}
	setFont(dc, 32)
	dc.SetRGB255(255, 255, 255)
	cardTextTop(dc, "使用说明", 20, 249-110)
	drawHelpPerson(dc, 20, 196)
	setFont(dc, 16)
	drawString(dc, "为需要绑定角色的指令", 40, 196+15)
	// .bg starts below the banner floats. CSS paints first-listed background
	// on top: amiya.png first, then the rgba(0,0,0,0.8) gradient over it.
	bgTop := 249.0
	if amiya := cardAsset("help/amiya.png"); amiya != nil {
		ab := amiya.Bounds()
		ah := 660 * ab.Dy() / ab.Dx()
		dc.DrawImage(ScaleExact(amiya, 660, ah), 0, int(bgTop))
	}
	dc.SetRGBA255(0, 0, 0, 204)
	dc.DrawRectangle(0, bgTop, mainW, mainH-bgTop)
	dc.Fill()
	y := bgTop
	drawGroup := func(title string, cmds []Cmd) {
		y += 10
		if label := cardAsset("help/label.png"); label != nil {
			dc.DrawImage(ScaleExact(label, 660, 40), 0, int(y))
		}
		setFont(dc, 16)
		dc.SetRGB255(255, 255, 255)
		drawString(dc, title, 25, y+8+13)
		y += 40
		for i, c := range cmds {
			px := 10.0 + float64(i%4)*160
			py := y + 15 + float64(i/4)*68
			dc.SetRGB255(255, 255, 255)
			dc.SetLineWidth(1.5)
			dc.DrawRoundedRectangle(px+0.75, py+0.75, 150-1.5, 53-1.5, 10)
			dc.Stroke()
			setFont(dc, 15)
			line1 := c.Cmd
			if c.Param != "" {
				line1 += " " + c.Param
			}
			drawString(dc, line1, px+5, py+1+3+15.75)
			if c.IsBind {
				drawHelpPerson(dc, px+150-10-16, py+1+3+3)
			}
			drawString(dc, c.Desc, px+5, py+1+3+15.75+22.5)
		}
		rows := (len(cmds) + 3) / 4
		if rows < 1 {
			rows = 1
		}
		y += 15 + float64(rows)*53 + float64(rows-1)*15
	}
	drawGroup("私聊指令", data.Private)
	drawGroup("普通指令", data.Public)
	drawGroup("管理员指令", data.Admin)
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
	// 10x10 grid from Lottery.tmpl declarations (scale 1.5, device px).
	// Screenshot is the #main element (manifest bbox x=149 = page centering),
	// so in-image x starts at the border box: border 1 + padding 40.
	// gridX = (1+40)*1.5 = 61.5. Grid width 900: 10 cells + 9 gaps of 8 ->
	// cell = (900-9*8)/10 = 82.8 CSS = 124.2 device; pitch = (82.8+8)*1.5.
	// gridY: padding-top 40 + h1 line 28*1.5 + padding-bottom 10 + border 2
	// + .header margin-bottom 30 = 124 CSS = 186 device. The old values
	// (200/189/136.5/112, measured off the baseline image) pushed the whole
	// grid ~one pitch right and clipped column 10 at the canvas edge.
	const gridX, gridY, pitch, cell = 61.5, 186.0, 136.2, 124.2
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
// Operator shapes mirror the enemy-tested fixture vocabulary.
type OperatorPotential struct {
	Rank int
	Desc string
}
type OperatorTalent struct {
	Evolve, Name, Desc string
}
type OperatorBuilding struct {
	Evolve, Icon, Name, Desc string
}
type OperatorSkill struct {
	Icon, Name, Desc, SkillRange string
	SpType                       []string
	SpInit, SpCost, Duration     string
}
type OperatorBranch struct {
	Name, Pic, Desc string
}
type OperatorStat struct {
	Name, Code, NameEn, Profession, Logo, Tags string
	Rarity                                     int
	HP, ATK, DEF, Res, Interval, ReDeploy      string
	Block, Cost                                string
}
type OperatorInfo struct {
	Painting, AttackRange string
	Op                    OperatorStat
	Branch                OperatorBranch
	Potentials            []OperatorPotential
	Talents               []OperatorTalent
	Buildings             []OperatorBuilding
	Skills                []OperatorSkill
}

// SampleOperator carries the operator-minimal fixture values. Provenance:
// feat/gg-render-canvas:src/core/web/complex_props_export_test.go
// visualComplexOperatorFixture (in-repo declaration): Amiya, markup
// stripped the way operatorProps/templateText does server-side.
func SampleOperator() *OperatorInfo {
	return &OperatorInfo{
		Painting: "https://fixture-cache.invalid/operator/amiya-painting.png", AttackRange: "□",
		Op: OperatorStat{
			Name: "阿米娅", Code: "R001", NameEn: "Amiya", Profession: "CASTER",
			Logo: "罗德岛", Tags: "远程位 输出", Rarity: 5,
			HP: "1742", ATK: "699", DEF: "121", Res: "10", Interval: "1.6s",
			ReDeploy: "70s", Block: "1", Cost: "18",
		},
		Branch:     OperatorBranch{Name: "中坚术师", Pic: "https://fixture-cache.invalid/operator/profession-branch.png", Desc: "攻击造成法术伤害"},
		Potentials: []OperatorPotential{{Rank: 1, Desc: "部署费用-1"}, {Rank: 2, Desc: "再部署时间-4秒"}},
		Talents:    []OperatorTalent{{Evolve: "精英化1", Name: "精神融合", Desc: "攻击力 +10%"}},
		Buildings:  []OperatorBuilding{{Evolve: "精英化0", Icon: "https://fixture-cache.invalid/operator/building-skill-icon.png", Name: "合作协议", Desc: "控制中枢内线索搜集速度提升"}},
		Skills:     []OperatorSkill{{Icon: "https://fixture-cache.invalid/operator/skill-icon.png", Name: "战术咏唱", Desc: "攻击力提升", SkillRange: "□", SpType: []string{"自动回复"}, SpInit: "0", SpCost: "45", Duration: "30"}},
	}
}

// RenderOperator draws Operator.tmpl at design 1200x800 (1800x1200 @1.5).
// Structure (template-declared): bg cover; painting (height 650, bottom 0,
// left 5%) from the shared capture bytes (manifest maps the same 立绘 file
// to Operator Painting as to Card Secretary: one vendored file, no copy);
// left attr table (b 100 black / w 70 #efeeef, spacing 0) + potential table;
// bottom-left pb/names block anchored at bottom 5%; right-floated 600px
// talent/bsk/skill tables (12px translucent rows). Skill/building icons load
// manifest-mapped vendored files (fixture URLs dead); template onerror amiya
// applies to the skill icon only. Painting/tables order keeps translucents
// blending over the art.
func RenderOperator(data *OperatorInfo) (*gg.Context, error) {
	const mainW, mainH = 1200, 800
	dc := gg.NewContext(mainW, mainH)
	// bg cover (replaces a fitted draw: pure cover math from asset dims).
	if bg, err := LoadImage(AssetPath("operator/bg.png")); err == nil {
		b := bg.Bounds()
		s := math.Max(float64(mainW)/float64(b.Dx()), float64(mainH)/float64(b.Dy()))
		dc.DrawImage(ScaleExact(bg, int(float64(b.Dx())*s+0.5), int(float64(b.Dy())*s+0.5)), 0, 0)
	}
	if paint, err := LoadImage(AssetPath("card/card-secretary-1024.png")); err == nil {
		dc.DrawImage(ScaleExact(paint, 650, 650), 60, mainH-650)
	}
	rowText := func(s string, size float64, x, y, rowH float64, r, g, b int, bold bool) {
		setFont(dc, size)
		w, _ := measure(dc, s)
		dc.SetRGB255(r, g, b)
		if bold {
			drawString(dc, s, x+0.45, y+(rowH-size)/2+size*0.8)
		}
		drawString(dc, s, x, y+(rowH-size)/2+size*0.8)
		_ = w
	}
	// Left: attr table (3x6, rows 26) + potential table.
	y := 20.0
	attrs := [][2]string{
		{"最大生命值", data.Op.HP}, {"攻击力", data.Op.ATK}, {"防御力", data.Op.DEF},
		{"法抗", data.Op.Res}, {"攻击间隔", data.Op.Interval}, {"再部署时间", data.Op.ReDeploy},
		{"阻挡数", data.Op.Block}, {"部署费用", data.Op.Cost}, {"所属", data.Op.Logo},
	}
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			a := attrs[r*3+c]
			bx := float64(c) * 170
			// template .b/.w carry opacity:0.8 (Operator.tmpl); blend over bg.
			dc.SetRGBA255(0, 0, 0, 204)
			dc.DrawRectangle(bx, y, 100, 26)
			dc.Fill()
			dc.SetRGBA255(0xef, 0xee, 0xef, 204)
			dc.DrawRectangle(bx+100, y, 70, 26)
			dc.Fill()
			rowText(a[0], 16, bx+1, y, 26, 255, 255, 255, false)
			rowText(a[1], 16, bx+101, y, 26, 0, 0, 0, false)
		}
		y += 26
	}
	// Potential table (tinted) with black header.
	y += 20
	potW := 300.0
	dc.SetRGBA255(239, 238, 239, 102)
	dc.DrawRectangle(0, y, potW, 26+float64(len(data.Potentials))*26)
	dc.Fill()
	dc.SetRGB255(0, 0, 0)
	dc.DrawRectangle(0, y, potW, 26)
	dc.Fill()
	rowText("潜能提升", 16, 1, y, 26, 255, 255, 255, false)
	y += 26
	for _, p := range data.Potentials {
		if pot := cardAsset("box/Potential_" + itoa(p.Rank) + ".png"); pot != nil {
			dc.DrawImage(ScaleExact(pot, 20, 20), 1, int(y+3))
		}
		rowText(p.Desc, 16, 24, y, 26, 0, 0, 0, false)
		y += 26
	}
	// Bottom-left block anchored at bottom 5% (40px).
	pbTop := float64(mainH-40) - (75 + 30 + 25 + 40) - (48 + 30)
	if prof := cardAsset("box/" + data.Op.Profession + ".png"); prof != nil {
		dc.DrawImage(ScaleExact(prof, 75, 75), 10, int(pbTop))
	}
	if rar := cardAsset("box/Rarity_" + itoa(data.Op.Rarity) + ".png"); rar != nil {
		rb := rar.Bounds()
		dc.DrawImage(ScaleExact(rar, 150, 150*rb.Dy()/rb.Dx()), 10+75, int(pbTop))
	}
	setFont(dc, 24)
	dc.SetRGB255(0, 0, 0) // pb cells carry no color: body-default black
	drawString(dc, data.Branch.Name, 10+75, pbTop+75+24)
	rowText(data.Op.Tags, 16, 10, pbTop+75+30, 26, 0, 0, 0, false)
	rowText(data.Branch.Desc, 14, 10, pbTop+75+30+26, 26, 0, 0, 0, false)
	namesTop := pbTop + 75 + 30 + 26 + 26
	setFont(dc, 32)
	dc.SetRGB255(0, 0, 0)
	drawString(dc, data.Op.Name, 10, namesTop+26)
	setFont(dc, 20)
	drawString(dc, data.AttackRange, 200, namesTop+22)
	dc.SetRGB255(0, 0, 0)
	dc.DrawRectangle(10, namesTop+48, 70, 24)
	dc.Fill()
	dc.SetRGB255(255, 255, 255)
	drawString(dc, data.Op.Code, 14, namesTop+48+19)
	setFont(dc, 20)
	dc.SetRGB255(0, 0, 0)
	drawString(dc, data.Op.NameEn, 120, namesTop+48+19)
	// Right float div: 600px tables stacked with mt 20 on talent.
	rx := float64(mainW - 600)
	ry := 20.0
	// Header tds declare opacity:0.8 inline (Operator.tmpl talent/bsk
	// headers), so both fill and title text composite at alpha 204.
	// Geometry matches rowText(title,14,rx+4,ry,24): only alpha changes.
	tableHead := func(title string) {
		dc.SetRGBA255(0, 0, 0, 204)
		dc.DrawRectangle(rx, ry, 600, 24)
		dc.Fill()
		setFont(dc, 14)
		dc.SetRGBA255(255, 255, 255, 204)
		drawString(dc, title, rx+4, ry+(24-14)/2+14*0.8)
		ry += 24
	}
	tintRow := func(h float64) {
		dc.SetRGBA255(176, 177, 177, 179)
		dc.DrawRectangle(rx, ry, 600, h)
		dc.Fill()
	}
	tableHead("天赋")
	for _, t := range data.Talents {
		tintRow(30)
		rowText(t.Evolve, 12, rx+4, ry, 30, 0, 0, 0, false)
		rowText(t.Name, 12, rx+120, ry, 30, 0, 0, 0, true)
		rowText(t.Desc, 12, rx+260, ry, 30, 0, 0, 0, false)
		ry += 30
	}
	tableHead("基建技能")
	for _, bs := range data.Buildings {
		tintRow(40)
		rowText(bs.Evolve, 12, rx+4, ry, 40, 0, 0, 0, false)
		if bicon, err := LoadImage(AssetPath("operator/operator-building-36.png")); err == nil {
			dc.DrawImage(ScaleExact(bicon, 36, 36), int(rx+100), int(ry+2))
		}
		rowText(bs.Name, 12, rx+150, ry, 40, 0, 0, 0, false)
		rowText(bs.Desc, 12, rx+260, ry, 40, 0, 0, 0, false)
		ry += 40
	}
	for _, sk := range data.Skills {
		dc.SetRGBA255(63, 63, 62, 252)
		dc.DrawRectangle(rx, ry, 600, 110)
		dc.Fill()
		if sicon, err := LoadImage(AssetPath("operator/operator-skill-128.png")); err == nil {
			dc.DrawImage(ScaleExact(sicon, 50, 50), int(rx+10), int(ry+8))
		}
		setFont(dc, 12)
		sx := rx + 80
		for _, sp := range sk.SpType {
			w, _ := measure(dc, sp)
			dc.SetRGB255(255, 255, 255)
			drawString(dc, sp, sx, ry+22)
			sx += w + 10
		}
		pill := func(s string, x float64) float64 {
			setFont(dc, 12)
			w, _ := measure(dc, s)
			dc.SetRGB255(0x80, 0x80, 0x80)
			RoundRect(dc, x, ry+8, w+10, 20, 3)
			dc.SetRGB255(255, 255, 255)
			drawString(dc, s, x+5, ry+23)
			return w + 10 + 5
		}
		sx += pill(fmt.Sprintf("技力%s/%s", sk.SpInit, sk.SpCost), sx)
		if sk.Duration != "" {
			pill(fmt.Sprintf("持续时间%ss", sk.Duration), sx)
		}
		rowText(sk.Desc, 12, rx+80, ry+40, 30, 255, 255, 255, false)
		rowText(sk.Name, 12, rx+80, ry+72, 26, 255, 255, 255, false)
		rowText(sk.SkillRange, 16, rx+520, ry, 110, 255, 255, 255, false)
		ry += 110
	}
	return ScaleToManifest(dc, 1800, 1200), nil
}
