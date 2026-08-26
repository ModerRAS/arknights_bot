package ggrender

import (
	"time"

	"github.com/fogleman/gg"
)

// Calendar scene: manifest 2880x1620 px = 1920x1080 CSS @1.5x (template/Calendar.tmpl).
// Left aside (10%, #141516 + bg.png contain) with date + resource info;
// white main area: weekday header + 6x7 month grid (Monday first).
// Frozen baseline state: 2026-08-19 (Wednesday), today cell cornflowerblue.

type CalendarData struct {
	Year, Month, Today int // Today = day of month
	Resource           string
	Chip               string
}

func SampleCalendar() *CalendarData {
	return &CalendarData{Year: 2026, Month: 8, Today: 19,
		Resource: "经验书、技能书、碳",
		Chip:     "近卫、特种、辅助、先锋"}
}

var weekHeads = []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}

// monthGrid returns 42 cells (6x7, Monday-first) day numbers with month offset
// (-1 prev, 0 current, +1 next) for year/month.
func monthGrid(year, month int) ([]int, []int) {
	first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	startIdx := (int(first.Weekday()) + 6) % 7 // Monday=0
	daysIn := func(y, m int) int {
		return time.Date(y, time.Month(m)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	}
	prevDays := daysIn(year, month-1)
	curDays := daysIn(year, month)
	nums := make([]int, 0, 42)
	offs := make([]int, 0, 42)
	for i := 0; i < 42; i++ {
		d := i - startIdx + 1
		switch {
		case d < 1:
			nums = append(nums, prevDays+d)
			offs = append(offs, -1)
		case d > curDays:
			nums = append(nums, d-curDays)
			offs = append(offs, 1)
		default:
			nums = append(nums, d)
			offs = append(offs, 0)
		}
	}
	return nums, offs
}

func RenderCalendar(data *CalendarData) (*gg.Context, error) {
	const cssW, cssH = 1920, 1080
	dc := gg.NewContext(2880, 1620) // manifest pixels
	dc.Scale(1.5, 1.5)
	FillBackground(dc, 255, 255, 255)

	// ---- aside ----
	dc.SetRGB255(0x14, 0x15, 0x16)
	dc.DrawRectangle(0, 0, 192, cssH)
	dc.Fill()
	bg := tryLocal("calendar/bg.png")
	bgi := ScaleContain(bg, 192, 1080)
	dc.DrawImage(bgi, 0, 0)
	// timeNow (JS-filled): date + weekday
	setFont(dc, 19)
	dc.SetRGB255(255, 255, 255)
	drawStringAnchored(dc, "2026年8月19日", 96, 30, 0.5, 0.5)
	drawStringAnchored(dc, "星期三", 96, 55, 0.5, 0.5)
	// todayResource
	setFont(dc, 16)
	drawString(dc, "资源关卡开放", 20, 437)
	setFont(dc, 15)
	drawString(dc, data.Resource, 20, 472)
	setFont(dc, 16)
	drawString(dc, "芯片关卡开放", 20, 530)
	setFont(dc, 15)
	// wrap like CSS width 152px
	drawString(dc, data.Chip, 20, 565)

	// ---- main ----
	gridX := 207.0  // 192 + padding 15
	gridW := 1698.0 // 1920-192-30
	colW := gridW / 7
	// thead y 0..40
	setFont(dc, 19)
	for i, h := range weekHeads {
		if i >= 5 {
			dc.SetRGB255(0xe0, 0x2d, 0x2d)
		} else {
			dc.SetRGB255(0x2c, 0x9b, 0xb3)
		}
		drawStringAnchored(dc, h, gridX+float64(i)*colW+colW/2, 20, 0.5, 0.5)
	}
	// tbody rows y 40..995.5, 6 rows
	rowH := 955.5 / 6
	nums, offs := monthGrid(data.Year, data.Month)
	todayIdx := -1
	{
		first := time.Date(data.Year, time.Month(data.Month), 1, 0, 0, 0, 0, time.UTC)
		startIdx := (int(first.Weekday()) + 6) % 7
		todayIdx = startIdx + data.Today - 1
	}
	for r := 0; r < 6; r++ {
		ry := 40 + float64(r)*rowH
		dc.SetRGB255(0xc8, 0xca, 0xcc)
		dc.DrawRectangle(gridX, ry, gridW, 1)
		dc.Fill()
		for c := 0; c < 7; c++ {
			i := r*7 + c
			n, off := nums[i], offs[i]
			cx := gridX + float64(c)*colW
			cy := ry + rowH*0.28
			isToday := i == todayIdx
			if isToday {
				dc.SetRGB255(0x64, 0x95, 0xed) // cornflowerblue
				dc.DrawRectangle(cx, ry, colW, rowH)
				dc.Fill()
			}
			setFont(dc, 19)
			switch {
			case isToday:
				dc.SetRGB255(255, 255, 255)
			case off != 0:
				dc.SetRGB255(0xbf, 0xbf, 0xbf)
			case c >= 5:
				dc.SetRGB255(0xe0, 0x2d, 0x2d)
			default:
				dc.SetRGB255(0, 0, 0)
			}
			drawStringAnchored(dc, itoa(n), cx+colW/2, cy, 0.5, 0.5)
		}
	}
	return dc, nil
}
