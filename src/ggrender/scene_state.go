package ggrender

import (
	"fmt"

	"github.com/fogleman/gg"
)

// State mirrors State.tmpl at 1092x510 CSS pixels. All inputs are derived from
// the deterministic state-minimal fixture and all images come from local assets.
const stateFrozenNow = int64(1736942400)

type StateMeter struct {
	Current, Max int
	RecoverTs    string // epoch seconds, empty when no recovery is active
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
	stateAvatarURLFrozen  = "https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/char_002_amiya%231.png"
	stateTraineeURLFrozen = "https://web.hycdn.cn/arknights/game/assets/char_skin/avatar/char_1001_amiya2%232.png"
)

func SampleState() *StateInfo {
	return &StateInfo{
		PlayerName:  "基线博士",
		AvatarURL:   stateAvatarURLFrozen,
		Ap:          StateMeter{95, 135, "1736935440"},
		TowerLower:  StateMeter{3, 6, "1687003200"},
		TowerHigher: StateMeter{4, 8, "1687003200"},
		Reward:      StateMeter{1, 3, "1686916800"},
		Recruitment: StateMeter{2, 4, ""},
		Trading:     StateMeter{6, 10, ""},
		Manufacture: StateMeter{7, 12, ""},
		TiredChars:  2,
		CheckedIn:   true,
		Training:    StateTraining{stateTraineeURLFrozen, "93784"},
	}
}

func jsHoursMinutes(ts string) (int, int) {
	diff := atoiSafe(ts) - int(stateFrozenNow)
	return floorDiv(diff, 3600) % 24, floorDiv(diff, 60) % 60
}

func jsDays(ts string) int {
	diff := atoiSafe(ts) - int(stateFrozenNow)
	days := diff / 86400
	if diff > 0 && diff%86400 != 0 {
		days++
	}
	return days
}

// floorDiv matches Math.floor(a/b) for positive b while preserving Go's
// integer arithmetic for the negative recovery values in the fixture.
func floorDiv(a, b int) int {
	q, r := a/b, a%b
	if r != 0 && a < 0 {
		q--
	}
	return q
}

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

const (
	meterX       = 460.0
	meterW       = 410.0
	meterLowerY  = 158.0
	meterHigherY = 249.0
)

var stateTrackTone = [3]int{128, 130, 131}

func RenderState(data *StateInfo) (*gg.Context, error) {
	const width, height = 1092, 510
	dc := gg.NewContext(width, height)
	FillBackground(dc, 46, 48, 49)
	if bg, err := LoadImage(AssetPath("state/bg.png")); err == nil {
		dc.DrawImage(bg, 0, 0)
	}

	white := func() { dc.SetRGB255(255, 255, 255) }

	avatar := FetchImage(data.AvatarURL, AssetPath("state/avatar-amiya.png"))
	dc.DrawImage(ScaleExact(avatar, 54, 54), 34, 34)

	setFont(dc, 30)
	white()
	drawString(dc, "Dr "+data.PlayerName, 98, 78)

	setFont(dc, 25)
	if data.CheckedIn {
		dc.SetRGB255(0x5d, 0x9a, 0x00)
	} else {
		dc.SetRGB255(0xcd, 0x28, 0x28)
	}
	if data.CheckedIn {
		drawString(dc, "已签到", 945, 77)
	} else {
		drawString(dc, "未签到", 945, 77)
	}

	if ap, err := LoadImage(AssetPath("state/ap.png")); err == nil {
		dc.DrawImage(ap, 35, 146)
	}
	setFont(dc, 30)
	white()
	drawString(dc, fmt.Sprintf("%d/%d", data.Ap.Current, data.Ap.Max), 146, 171)
	setFont(dc, 21)
	if data.Ap.Current >= data.Ap.Max {
		drawString(dc, "理智已全部恢复", 145, 216)
	} else {
		hours, minutes := jsHoursMinutes(data.Ap.RecoverTs)
		drawString(dc, fmt.Sprintf("%d时%d分后恢复", hours, minutes), 145, 216)
	}

	lowerDays := jsDays(data.TowerLower.RecoverTs)
	drawMeter(dc, "数据增补条", lowerDays, data.TowerLower, 119, meterLowerY)
	// State.js derives both labels from lowerItemTermTime.
	drawMeter(dc, "数据增补仪", lowerDays, data.TowerHigher, 210, meterHigherY)

	if campaign, err := LoadImage(AssetPath("state/campaign.png")); err == nil {
		dc.DrawImage(campaign, 931, 127)
	}
	setFont(dc, 20)
	white()
	drawString(dc, fmt.Sprintf("%d/%d", data.Reward.Current, data.Reward.Max), 973, 265)
	dc.SetRGBA255(0, 0, 0, 128)
	dc.DrawRectangle(927, 213, 112, 21)
	dc.Fill()
	setFont(dc, 16)
	drawClockGlyph(dc, 943, 223.5, 14)
	drawString(dc, fmt.Sprintf("%d天", jsDays(data.Reward.RecoverTs)), 961, 229)

	drawRow(dc, "recruit.png", 49, 331, "公开招募", 120, 339,
		fmt.Sprintf("%d/%d", data.Recruitment.Current, data.Recruitment.Max), 330, 339)
	drawRow(dc, "tired_chars.png", 49, 431, "干员疲劳", 115, 439,
		itoa(data.TiredChars), 325, 439)
	drawRow(dc, "tradings.png", 460, 338, "订单进度", 520, 341,
		fmt.Sprintf("%d/%d", data.Trading.Current, data.Trading.Max), 780, 341)
	drawRow(dc, "manufactures.png", 460, 436, "制造进度", 520, 439,
		fmt.Sprintf("%d/%d", data.Manufacture.Current, data.Manufacture.Max), 780, 439)

	if data.Training.CharIcon != "" || data.Training.LeftSeconds != "" {
		trainee := FetchImage(data.Training.CharIcon, AssetPath("state/avatar-trainee.png"))
		dc.DrawImage(ScaleExactCR(trainee, 130, 130), 922, 307)
		dc.SetRGBA255(0, 0, 0, 128)
		dc.DrawRectangle(922, 412, 133, 25)
		dc.Fill()
		setFont(dc, 16)
		white()
		drawClockGlyph(dc, 930, 422, 16)
		drawString(dc, jsFormatTime(data.Training.LeftSeconds), 950, 427)
		setFont(dc, 30)
		drawString(dc, "训练室", 945, 481)
	}
	return dc, nil
}

func fracOf(m StateMeter) float64 {
	if m.Max == 0 {
		return 1
	}
	f := float64(m.Current) / float64(m.Max)
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

func drawMeter(dc *gg.Context, label string, days int, meter StateMeter, labelTop, barY float64) {
	setFont(dc, 25)
	dc.SetRGB255(255, 255, 255)
	baseline := labelTop + 21
	drawString(dc, label, meterX, baseline+1)
	drawClockGlyph(dc, meterX+measureW(dc, label)+30, baseline-11.5, 16)
	term := fmt.Sprintf("%d天", days)
	drawString(dc, term, meterX+measureW(dc, label)+56, baseline+1)
	reward := fmt.Sprintf("%d/%d", meter.Current, meter.Max)
	rewardWidth, _ := measure(dc, reward)
	// flex layout: reward sits 13px right of the right-aligned anchor (gap after term)
	drawString(dc, reward, meterX+meterW-rewardWidth+13, baseline+1)

	dc.SetRGB255(stateTrackTone[0], stateTrackTone[1], stateTrackTone[2])
	dc.DrawRectangle(meterX, barY, meterW, 11)
	dc.Fill()
	dc.SetRGB255(255, 255, 255)
	dc.DrawRectangle(meterX, barY, meterW*fracOf(meter), 11)
	dc.Fill()
}

func measureW(dc *gg.Context, s string) float64 {
	w, _ := measure(dc, s)
	return w
}

func drawRow(dc *gg.Context, icon string, ix, iy float64, title string, tx, ty float64, value string, vx, vy float64) {
	if image, err := LoadImage(AssetPath("state/" + icon)); err == nil {
		dc.DrawImage(image, int(ix), int(iy))
	}
	setFont(dc, 25)
	dc.SetRGB255(255, 255, 255)
	drawString(dc, title, tx, ty+22)
	drawString(dc, value, vx, vy+22)
}

func drawClockGlyph(dc *gg.Context, x, cy, diameter float64) {
	radius := diameter / 2
	dc.SetRGB255(255, 255, 255)
	dc.SetLineWidth(1.6)
	dc.DrawCircle(x+radius-2, cy, radius-0.8)
	dc.Stroke()
	dc.DrawLine(x+radius-3, cy, x+radius-3, cy-radius+2.8)
	dc.DrawLine(x+radius-2, cy, x+radius+1.2, cy+5)
	dc.Stroke()
}
