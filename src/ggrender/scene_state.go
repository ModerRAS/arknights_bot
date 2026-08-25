package ggrender

import (
	"fmt"

	"github.com/fogleman/gg"
)

// State — mirrors template/State.tmpl rendered at 1092x510 CSS, scale 1 -> 1092x510.
// Base layer is the production page background assets/state/bg.png (1092x510);
// everything else is pure GG vector/text over it, matching the frozen
// state-minimal fixture and the browserDeterminism frozen clock
// Date.now()=1736942400000ms (fixtures.json) applied through State.js semantics.

const stateFrozenNow = int64(1736942400) // frozen capture clock, seconds

type StateMeter struct {
	Current, Max int
	RecoverTs    string // epoch seconds, "" when none
}

type StateTraining struct {
	CharIcon    string
	LeftSeconds string
}

type StateInfo struct {
	PlayerName  string
	AvatarURL   string
	Ap          StateMeter
	TowerLower  StateMeter
	TowerHigher StateMeter
	Reward      StateMeter
	Recruitment StateMeter
	Trading     StateMeter
	Manufacture StateMeter
	TiredChars  int
	CheckedIn   bool
	Training    StateTraining
}

const (
	stateAvatarURLFrozen = "https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/char_002_amiya%231.png"
	stateTraineeURLFroze = "https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/char_1001_amiya2%232.png"
)

func SampleState() *StateInfo {
	return &StateInfo{
		PlayerName:  "基线博士",
		AvatarURL:   stateAvatarURLFrozen,
		Ap:          StateMeter{95, 135, "1736951640"},
		TowerLower:  StateMeter{3, 6, "1737115200"},
		TowerHigher: StateMeter{4, 8, "1737201601"},
		Reward:      StateMeter{1, 3, "1737028801"},
		Recruitment: StateMeter{2, 4, ""},
		Trading:     StateMeter{6, 10, ""},
		Manufacture: StateMeter{7, 12, ""},
		TiredChars:  2,
		CheckedIn:   true,
		Training:    StateTraining{stateTraineeURLFroze, "93784"},
	}
}

// jsHoursMinutes mirrors State.js completeRecoveryTime hour/minute parts.
func jsHoursMinutes(ts string) (int, int) {
	diff := atoiSafe(ts) - int(stateFrozenNow)
	return floorMod(diff/3600, 24), floorMod(diff/60, 60)
}

// jsDays mirrors common.js daysUntil: ceil((ts-now)/86400).
func jsDays(ts string) int {
	diff := atoiSafe(ts) - int(stateFrozenNow)
	d := diff / 86400
	if diff%86400 != 0 {
		d++
	}
	return d
}

// jsFormatTime mirrors common.js formatTime HH:MM:SS.
func jsFormatTime(sec string) string {
	s := atoiSafe(sec)
	return pad2(s/3600) + ":" + pad2(s%3600/60) + ":" + pad2(s%60)
}

func floorMod(a, b int) int {
	m := a % b
	if m < 0 {
		m += b
	}
	return m
}

func pad2(n int) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

func RenderState(data *StateInfo) (*gg.Context, error) {
	const W, H = 1092, 510
	dc := gg.NewContext(W, H)

	// base: production page background, drawn 1:1 (asset is exactly WxH)
	if bg, err := LoadImage(AssetPath("state/bg.png")); err == nil {
		dc.DrawImage(bg, 0, 0)
	} else {
		FillBackground(dc, 46, 48, 49)
	}

	white := func() { dc.SetRGB255(255, 255, 255) }

	// avatar 54x54 @ (34,34)
	av := FetchImage(data.AvatarURL, AssetPath("state/avatar-amiya.png"))
	dc.DrawImage(ScaleExactCR(av, 54, 54), 34, 34)

	// name "Dr <name>" 30px white, abspos after avatar
	setFont(dc, 30)
	white()
	drawString(dc, "Dr "+data.PlayerName, 98, 133)

	// checked-in flag 25px, green when signed
	setFont(dc, 25)
	if data.CheckedIn {
		dc.SetRGB255(0x5d, 0x9a, 0x00)
		drawString(dc, "已签到", 75, 32.5)
	} else {
		dc.SetRGB255(0xcd, 0x28, 0x28)
		drawString(dc, "未签到", 75, 32.5)
	}

	// ap icon + counter + recovery label
	if ap, err := LoadImage(AssetPath("state/ap.png")); err == nil {
		dc.DrawImage(ap, 36, 161)
	}
	setFont(dc, 30)
	white()
	drawString(dc, fmt.Sprintf("%d/%d", data.Ap.Current, data.Ap.Max), 146, 177.9)
	setFont(dc, 21)
	white()
	if data.Ap.Current >= data.Ap.Max {
		drawString(dc, "理智已全部恢复", 145, 221.9)
	} else {
		hh, mm := jsHoursMinutes(data.Ap.RecoverTs)
		drawString(dc, fmt.Sprintf("%d时%d分后恢复", hh, mm), 145, 221.9)
	}

	// tower lower / higher rows + progress bars (fill white on native track)
	drawTowerRow(dc, "数据增补条", jsDays(data.TowerLower.RecoverTs),
		fmt.Sprintf("%d/%d", data.TowerLower.Current, data.TowerLower.Max),
		fracOf(data.TowerLower), 117.5)
	drawTowerRow(dc, "数据增补仪", jsDays(data.TowerHigher.RecoverTs),
		fmt.Sprintf("%d/%d", data.TowerHigher.Current, data.TowerHigher.Max),
		fracOf(data.TowerHigher), 193.5)

	// campaign cluster (upper right zone)
	campX, campY := 898.5, 83.5
	if ic, err := LoadImage(AssetPath("state/campaign.png")); err == nil {
		dc.DrawImage(ic, int(campX), int(campY))
	}
	setFont(dc, 20)
	white()
	drawString(dc, fmt.Sprintf("%d/%d", data.Reward.Current, data.Reward.Max), campX+75, campY+138)
	dc.SetRGBA255(0, 0, 0, 128)
	dc.DrawRectangle(campX-38.5, campY+140, 110, 25)
	dc.Fill()
	setFont(dc, 16)
	white()
	drawClockGlyph(dc, campX-30, campY+152.5, 16)
	drawString(dc, fmt.Sprintf("%d天", jsDays(data.Reward.RecoverTs)), campX-10, campY+158)

	// recruit / tired rows (lower-left column)
	drawIconRow(dc, AssetPath("state/recruit.png"), 49, 223.5, 65, "公开招募",
		fmt.Sprintf("%d/%d", data.Recruitment.Current, data.Recruitment.Max))
	drawIconRow(dc, AssetPath("state/tired_chars.png"), 49, 330.5, 55, "干员疲劳",
		itoa(data.TiredChars))

	// tradings / manufactures (right column, abspos icons)
	if ic, err := LoadImage(AssetPath("state/tradings.png")); err == nil {
		dc.DrawImage(ic, 460, 193)
	}
	setFont(dc, 25)
	white()
	drawString(dc, "订单进度", 520, 222.5)
	drawString(dc, fmt.Sprintf("%d/%d", data.Trading.Current, data.Trading.Max), 780, 222.5)
	if ic, err := LoadImage(AssetPath("state/manufactures.png")); err == nil {
		dc.DrawImage(ic, 460, 427)
	}
	white()
	drawString(dc, "制造进度", 520, 459.5)
	drawString(dc, fmt.Sprintf("%d/%d", data.Manufacture.Current, data.Manufacture.Max), 780, 459.5)

	// training room card (far right)
	if data.Training.CharIcon != "" || data.Training.LeftSeconds != "" {
		tr := FetchImage(data.Training.CharIcon, AssetPath("state/avatar-trainee.png"))
		dc.DrawImage(ScaleExactCR(tr, 130, 130), 922, 347)
		dc.SetRGBA255(0, 0, 0, 128)
		dc.DrawRectangle(923, 456, 133, 25)
		dc.Fill()
		setFont(dc, 16)
		white()
		drawClockGlyph(dc, 950, 477, 16)
		drawString(dc, jsFormatTime(data.Training.LeftSeconds), 970, 477)
		setFont(dc, 30)
		white()
		drawString(dc, "训练室", 945, 488.5+34.8)
	}
	return dc, nil
}

func fracOf(m StateMeter) float64 {
	if m.Max == 0 {
		return 1
	}
	f := float64(m.Current) / float64(m.Max)
	if f < 0 {
		f = 0
	}
	if f > 1 {
		f = 1
	}
	return f
}

// drawTowerRow renders 数据增补条/仪 flex row + progress bar.
// rowTop: block top of the flex paragraph; texts 25px, baseline = rowTop+29.
func drawTowerRow(dc *gg.Context, label string, days int, reward string, frac float64, rowTop float64) {
	const x0 = 460.0
	setFont(dc, 25)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, label, x0, rowTop+29)
	lw, _ := measure(dc, label)
	svgX := x0 + lw + 30
	drawClockGlyph(dc, svgX, rowTop+18.5, 16)
	tx := svgX + 16 + 10
	drawString(dc, fmt.Sprintf("%d天", days), tx, rowTop+29)
	tw, _ := measure(dc, fmt.Sprintf("%d天", days))
	drawString(dc, reward, tx+tw+130, rowTop+29)
	// progress: track + white value, 410x11 radius 1, top = rowTop+rowH-18
	barY := rowTop + 37 - 18 + 37 - 37 // keep simple: rowTop+19
	_ = barY
	barTop := rowTop + 19
	dc.SetRGB255(0xef, 0xef, 0xef)
	dc.DrawRoundedRectangle(x0, barTop, 410, 11, 1)
	dc.Fill()
	dc.SetRGB255(255, 255, 255)
	dc.DrawRoundedRectangle(x0, barTop, 410*frac, 11, 1)
	dc.Fill()
}

// drawIconRow renders recruit/tired style rows: icon 42x42 + title + value (25px).
func drawIconRow(dc *gg.Context, iconPath string, x, rowTop, iconMarginTop float64, title, value string) {
	rowH := 42 + iconMarginTop
	if ic, err := LoadImage(iconPath); err == nil {
		dc.DrawImage(ic, int(x), int(rowTop+iconMarginTop))
	}
	setFont(dc, 25)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, title, x+42+120, rowTop+62+29)
	drawString(dc, value, x+42+(330-120)+120, rowTop+62+29)
	_ = rowH
}

// drawClockGlyph draws a small white clock (svg placeholder) centered baseline-aware.
func drawClockGlyph(dc *gg.Context, x, cy float64, d float64) {
	r := d / 2
	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(1.6)
	dc.DrawCircle(x+r, cy, r-0.8)
	dc.Stroke()
	dc.DrawLine(x+r, cy, x+r, cy-r+2.8)
	dc.DrawLine(x+r, cy, x+r+3.2, cy+1)
	dc.Stroke()
}
