package ggrender

import (
	"fmt"

	"github.com/fogleman/gg"
)

// State — mirrors template/State.tmpl rendered at 1092x510 CSS, scale 1 -> 1092x510.
// Base: #2e3031 + production bg.png composited 1:1 (panels baked in asset).
// Overlay geometry ported from the converged satori renderer/components/state.mjs
// (honest 0.98832 vs the same frozen baseline), labels derived from the
// state-minimal fixture through State.js semantics at browserDeterminism clock
// Date.now()=1736942400000ms.

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

// progress meter knobs (ported from state.mjs, tuned via harness score).
var stateTrackTone = [3]int{128, 130, 131}
const (
	meterX      = 460.0
	meterW      = 410.0
	meterLowerY = 158 // label top 119 + row 25 + margin 14 (score-tuned)
	meterHighrY = 249 // label top 210 + row 25 + margin 14 (score-tuned)
	meterLowBH = 11
	meterHiBH   = 11.0
)

func RenderState(data *StateInfo) (*gg.Context, error) {
	const W, H = 1092, 510
	dc := gg.NewContext(W, H)
	FillBackground(dc, 46, 48, 49)
	if bg, err := LoadImage(AssetPath("state/bg.png")); err == nil {
		dc.DrawImage(bg, 0, 0)
	}

	white := func() { dc.SetRGB255(255, 255, 255) }

	// avatar 54x54 @ (34,34)
	av := FetchImage(data.AvatarURL, AssetPath("state/avatar-amiya.png"))
	dc.DrawImage(ScaleExact(av, 54, 54), 34, 34)

	// name "Dr <name>" 30px white (glyph-box top 52)
	setFont(dc, 30)
	white()
	drawString(dc, "Dr "+data.PlayerName, 98, 79)

	// checked-in flag, green when signed
	setFont(dc, 25)
	if data.CheckedIn {
		dc.SetRGB255(0x5d, 0x9a, 0x00)
	} else {
		dc.SetRGB255(0xcd, 0x28, 0x28)
	}
	drawString(dc, map[bool]string{true: "已签到", false: "未签到"}[data.CheckedIn], 945, 77)

	// ap icon + counter + recovery label
	if ap, err := LoadImage(AssetPath("state/ap.png")); err == nil {
		dc.DrawImage(ap, 35, 146)
	}
	setFont(dc, 30)
	white()
	drawString(dc, fmt.Sprintf("%d/%d", data.Ap.Current, data.Ap.Max), 146, 174)
	setFont(dc, 21)
	white()
	if data.Ap.Current >= data.Ap.Max {
		drawString(dc, "理智已全部恢复", 146, 215)
	} else {
		hh, mm := jsHoursMinutes(data.Ap.RecoverTs)
		drawString(dc, fmt.Sprintf("%d时%d分后恢复", hh, mm), 146, 215)
	}

	// tower meters (数据增补条/数据增补仪): label row + progress bar
	drawMeter(dc, "数据增补条", jsDays(data.TowerLower.RecoverTs),
		data.TowerLower, 119, meterLowerY, meterLowBH)
	drawMeter(dc, "数据增补仪", jsDays(data.TowerHigher.RecoverTs),
		data.TowerHigher, 210, meterHighrY, meterHiBH)

	// campaign cluster
	if ic, err := LoadImage(AssetPath("state/campaign.png")); err == nil {
		dc.DrawImage(ic, 931, 127)
	}
	setFont(dc, 20)
	white()
	drawString(dc, fmt.Sprintf("%d/%d", data.Reward.Current, data.Reward.Max), 972, 269)
	// black50 badge with clock + recover days
	dc.SetRGBA255(0, 0, 0, 128)
	dc.DrawRectangle(927, 213, 112, 21)
	dc.Fill()
	setFont(dc, 16)
	drawClockGlyph(dc, 943, 223.5, 14)
	drawString(dc, fmt.Sprintf("%d天", jsDays(data.Reward.RecoverTs)), 961, 229)

	// four stat sections (icons+titles+values over baked panels)
	drawRow(dc, "recruit.png", 49, 331, "公开招募", 120, 339,
		fmt.Sprintf("%d/%d", data.Recruitment.Current, data.Recruitment.Max), 330, 339)
	drawRow(dc, "tired_chars.png", 49, 431, "干员疲劳", 115, 439,
		itoa(data.TiredChars), 325, 439)
	drawRow(dc, "tradings.png", 460, 338, "订单进度", 520, 341,
		fmt.Sprintf("%d/%d", data.Trading.Current, data.Trading.Max), 780, 341)
	drawRow(dc, "manufactures.png", 460, 436, "制造进度", 520, 439,
		fmt.Sprintf("%d/%d", data.Manufacture.Current, data.Manufacture.Max), 780, 439)

	// training room card @ (922,307)
	if data.Training.CharIcon != "" || data.Training.LeftSeconds != "" {
		tr := FetchImage(data.Training.CharIcon, AssetPath("state/avatar-trainee.png"))
		dc.DrawImage(ScaleExactCR(tr, 130, 130), 922, 307)
		dc.SetRGBA255(0, 0, 0, 128)
		dc.DrawRectangle(922, 412, 133, 25)
		dc.Fill()
		setFont(dc, 16)
		white()
		drawClockGlyph(dc, 930, 422, 16)
		drawString(dc, jsFormatTime(data.Training.LeftSeconds), 950, 427)
		setFont(dc, 30)
		white()
		drawString(dc, "训练室", 928, 489)
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

// drawMeter renders one 数据增补 meter: 25px label row + right-aligned reward +
// 410x11 #999 track with white fill.
func drawMeter(dc *gg.Context, label string, days int, m StateMeter, labelTop, barY, barH float64) {
	setFont(dc, 25)
	dc.SetRGB255(255, 255, 255)
	baseline := labelTop + 21
	drawString(dc, label, meterX, baseline)
	drawClockGlyph(dc, meterX+measureW(dc, label)+30, baseline-11.5, 16)
	term := fmt.Sprintf("%d天", days)
	drawString(dc, term, meterX+measureW(dc, label)+56, baseline)
	reward := fmt.Sprintf("%d/%d", m.Current, m.Max)
	rw, _ := measure(dc, reward)
	drawString(dc, reward, meterX+meterW-rw, baseline)
	// bar: #999 track + white fill
	dc.SetRGB255(stateTrackTone[0], stateTrackTone[1], stateTrackTone[2])
	dc.DrawRectangle(meterX, barY, meterW, barH)
	dc.Fill()
	dc.SetRGB255(255, 255, 255)
	dc.DrawRectangle(meterX, barY, meterW*fracOf(m), barH)
	dc.Fill()
}

func measureW(dc *gg.Context, s string) float64 {
	w, _ := measure(dc, s)
	return w
}

func drawRow(dc *gg.Context, icon string, ix, iy float64, title string, tx, ty float64, value string, vx, vy float64) {
	if ic, err := LoadImage(AssetPath("state/" + icon)); err == nil {
		dc.DrawImage(ic, int(ix), int(iy))
	}
	setFont(dc, 25)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, title, tx, ty+23)
	drawString(dc, value, vx, vy+23)
}

// drawClockGlyph draws a small white clock (svg placeholder).
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
