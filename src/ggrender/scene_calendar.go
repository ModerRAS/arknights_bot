package ggrender

import (
	"fmt"
	"time"

	"github.com/fogleman/gg"
)

// Calendar follows template/Calendar.tmpl at its manifest size. The source
// page used the capture clock recorded in capture-clock-erratum.json.
const (
	calendarYear  = 2026
	calendarMonth = time.August
	calendarToday = 19
	calendarCSSW  = 1920
	calendarCSSH  = 1080
	calendarScale = 1.5
)

// CalendarData is the stable, network-free projection consumed by RenderCalendar.
type CalendarData struct {
	Entries []struct{ Title, Begin, End string }
}

func SampleCalendar() *CalendarData {
	c := &CalendarData{}
	c.Entries = []struct{ Title, Begin, End string }{
		{"主线活动「破碎日冕」", "2025-08-01", "2025-08-14"},
		{"危机合约", "2025-08-05", "2025-08-19"},
		{"常驻寻访", "2025-08-01", "2025-08-30"},
		{"愚人号复刻", "2025-08-10", "2025-08-24"},
		{"感谢庆典", "2025-08-15", "2025-08-29"},
	}
	return c
}

var calendarWeekdays = [...]string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}

// calendarCell is one visible cell in the Monday-first grid.
type calendarCell struct {
	date   time.Time
	month  int // -1 previous, 0 current, 1 next
	labels []string
}

func calendarDaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// calendarGrid mirrors monthHTML: at least five rows, with a sixth row when
// the first weekday and month length require it.
func calendarGrid(year int, month time.Month) [][]calendarCell {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	mondayIndex := (int(first.Weekday()) + 6) % 7
	days := calendarDaysInMonth(year, month)
	rows := (mondayIndex + days + 6) / 7
	if rows < 5 {
		rows = 5
	}
	prevMonth := month - 1
	prevYear := year
	if prevMonth < time.January {
		prevMonth = time.December
		prevYear--
	}
	prevDays := calendarDaysInMonth(prevYear, prevMonth)

	grid := make([][]calendarCell, rows)
	for row := range grid {
		grid[row] = make([]calendarCell, 7)
		for col := range grid[row] {
			index := row*7 + col
			day := index - mondayIndex + 1
			cell := calendarCell{month: 0}
			switch {
			case day < 1:
				cell.date = time.Date(prevYear, prevMonth, prevDays+day, 0, 0, 0, 0, time.UTC)
				cell.month = -1
			case day > days:
				nextMonth := month + 1
				nextYear := year
				if nextMonth > time.December {
					nextMonth = time.January
					nextYear++
				}
				cell.date = time.Date(nextYear, nextMonth, day-days, 0, 0, 0, 0, time.UTC)
				cell.month = 1
			default:
				cell.date = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
			}
			grid[row][col] = cell
		}
	}
	return grid
}

func calendarEventLabels(data *CalendarData) map[string][]string {
	labels := make(map[string][]string)
	for _, entry := range data.Entries {
		if entry.Title == "" {
			continue
		}
		if entry.Begin != "" {
			labels[entry.Begin] = append(labels[entry.Begin], "开始 "+entry.Title)
		}
		if entry.End != "" {
			labels[entry.End] = append(labels[entry.End], "结束 "+entry.Title)
		}
	}
	return labels
}

// calendarWrappedLines applies the template's break-all/overflow behavior to
// the narrow day cell without allowing text to change the grid geometry.
func calendarWrappedLines(dc *gg.Context, text string, maxWidth float64) []string {
	if text == "" {
		return nil
	}
	var lines []string
	line := ""
	for _, r := range []rune(text) {
		candidate := line + string(r)
		if line != "" {
			width, _ := dc.MeasureString(candidate)
			if width > maxWidth {
				lines = append(lines, line)
				line = string(r)
				continue
			}
		}
		line = candidate
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

func calendarDrawLabels(dc *gg.Context, cell calendarCell, x, y, w, h float64, current bool) {
	if len(cell.labels) == 0 {
		return
	}
	setFont(dc, 14)
	if current {
		dc.SetRGB255(255, 255, 255)
	} else if cell.month != 0 {
		dc.SetRGB255(191, 191, 191)
	} else {
		dc.SetRGB255(97, 97, 97)
	}

	// The original almanac is a 14px line inside a flex column. Keep two
	// compact lines visible, matching the clipped cell rather than overflowing.
	lineY := y + h*0.63
	for _, label := range cell.labels {
		for _, line := range calendarWrappedLines(dc, label, w-20) {
			dc.DrawStringAnchored(line, x+w/2, lineY, 0.5, 0.5)
			lineY += 22
			if lineY > y+h-18 {
				return
			}
		}
	}
}

func RenderCalendar(data *CalendarData) (*gg.Context, error) {
	if data == nil {
		data = &CalendarData{}
	}

	// Allocate the exact manifest canvas. The transform only expresses the
	// template's CSS coordinate system; no design-sized canvas is resampled.
	dc := gg.NewContext(2880, 1620)
	dc.Scale(calendarScale, calendarScale)
	FillBackground(dc, 255, 255, 255)

	// Aside: width 10%, padding 15px 20px, background-size: contain.
	asideW := float64(calendarCSSW) * 0.10
	dc.SetRGB255(20, 21, 22)
	dc.DrawRectangle(0, 0, asideW, calendarCSSH)
	dc.Fill()
	if bg := tryLocal("calendar/bg.png"); bg != nil {
		dc.DrawImage(ScaleContain(bg, int(asideW), calendarCSSH), 0, 0)
	}

	captureDate := time.Date(calendarYear, calendarMonth, calendarToday, 0, 0, 0, 0, time.UTC)
	setFont(dc, 19.2)
	dc.SetRGB255(255, 255, 255)
	dc.DrawStringAnchored(fmt.Sprintf("%d年%d月%d日", captureDate.Year(), captureDate.Month(), captureDate.Day()), asideW/2, 30, 0.5, 0.5)
	dc.DrawStringAnchored("星期三", asideW/2, 55, 0.5, 0.5)

	// todayResource begins at 350px in the template.
	setFont(dc, 15)
	dc.DrawString("资源关卡开放", 20, 350+19)
	dc.DrawString("经验书、技能书、碳", 20, 350+19+35)
	dc.DrawString("芯片关卡开放", 20, 350+19+93)
	dc.DrawString("近卫、特种、辅助、先锋", 20, 350+19+93+35)

	// Main content has 15px horizontal padding and a 40px weekday header.
	mainX := asideW + 15
	mainW := float64(calendarCSSW)*0.90 - 30
	colW := mainW / 7
	setFont(dc, 19.2)
	for col, name := range calendarWeekdays {
		if col >= 5 {
			dc.SetRGB255(224, 45, 45)
		} else {
			dc.SetRGB255(44, 155, 179)
		}
		dc.DrawStringAnchored(name, mainX+(float64(col)+0.5)*colW, 20, 0.5, 0.5)
	}

	pickerH := float64(calendarCSSH) - 4.4*19.2
	bodyH := pickerH - 40
	grid := calendarGrid(calendarYear, calendarMonth)
	rowH := bodyH / float64(len(grid))
	labels := calendarEventLabels(data)
	for row := range grid {
		top := 40 + float64(row)*rowH
		dc.SetRGB255(200, 202, 204)
		dc.DrawRectangle(mainX, top, mainW, 1)
		dc.Fill()
		for col := range grid[row] {
			cell := grid[row][col]
			cell.labels = labels[cell.date.Format("2006-01-02")]
			x := mainX + float64(col)*colW
			isToday := cell.date.Year() == captureDate.Year() && cell.date.Month() == captureDate.Month() && cell.date.Day() == captureDate.Day()
			if isToday {
				dc.SetRGB255(100, 149, 237)
				dc.DrawRectangle(x, top, colW, rowH)
				dc.Fill()
			}

			if isToday {
				dc.SetRGB255(255, 255, 255)
			} else if cell.month != 0 {
				dc.SetRGB255(191, 191, 191)
			} else if col >= 5 {
				dc.SetRGB255(224, 45, 45)
			} else {
				dc.SetRGB255(0, 0, 0)
			}
			dc.DrawStringAnchored(fmt.Sprintf("%d", cell.date.Day()), x+colW/2, top+rowH*0.28, 0.5, 0.5)
			calendarDrawLabels(dc, cell, x, top, colW, rowH, isToday)
		}
	}
	return dc, nil
}
