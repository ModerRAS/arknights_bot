package ggrender

import (
	"fmt"

	"github.com/fogleman/gg"
)

// Lottery scene: manifest 1473x1667 px = 982x1111 CSS @1.5x (template/Lottery.tmpl).
// Dark page (#0f0f0f), centered #1a1a1a rounded container: title with cyan
// underline, 10x10 number grid (82.8px cells, gap 8), 3-item legend.

type LotteryDetail struct {
	Number     int
	UserName   string
	UserNumber string
	Status     int // 1 = winner
}

type LotteryData struct {
	Details []LotteryDetail
}

func SampleLottery() *LotteryData {
	return &LotteryData{Details: []LotteryDetail{
		{Number: 7, UserName: "中奖博士", UserNumber: "100000007", Status: 1},
		{Number: 42, UserName: "占位博士", UserNumber: "100000042", Status: 0},
	}}
}

func RenderLottery(data *LotteryData) (*gg.Context, error) {
	const cssW, cssH = 982, 1111.33
	dc := gg.NewContext(1473, 1667) // manifest pixels
	dc.Scale(1.5, 1.5)
	FillBackground(dc, 0x0f, 0x0f, 0x0f)

	// container
	cx0, cy0, cw, ch := 0.5, 20.0, 981.0, 1071.0
	dc.SetRGB255(0x1a, 0x1a, 0x1a)
	dc.DrawRoundedRectangle(cx0, cy0, cw, ch, 16)
	dc.Fill()
	dc.SetRGB255(0x33, 0x33, 0x33)
	dc.SetLineWidth(1)
	dc.DrawRoundedRectangle(cx0, cy0, cw, ch, 16)
	dc.Stroke()

	// title
	setFont(dc, 28)
	dc.SetRGB255(255, 255, 255)
	title := "选号详情"
	tw, _ := dc.MeasureString(title)
	// letter-spacing 4px adds 3*4
	tw += 12
	drawStringAnchored(dc, title, cssW/2, 68, 0.5, 0.5)
	dc.SetRGB255(0x00, 0xe5, 0xff)
	dc.DrawRectangle(cssW/2-tw/2, 86, tw, 2)
	dc.Fill()

	// grid 10x10
	const pad = 40.0
	cell := (cw - 2*pad - 9*8) / 10
	gap := 8.0
	gridTop := 120.0
	lookup := map[int]LotteryDetail{}
	for _, d := range data.Details {
		lookup[d.Number] = d
	}
	for i := 1; i <= 100; i++ {
		col := (i - 1) % 10
		row := (i - 1) / 10
		x := cx0 + pad + float64(col)*(cell+gap)
		y := gridTop + float64(row)*(cell+gap)
		d, picked := lookup[i]
		winner := picked && d.Status == 1
		// cell bg
		if picked {
			if winner {
				dc.SetRGB255(0x4a, 0x1c, 0x1c)
			} else {
				dc.SetRGB255(0x1e, 0x3a, 0x3f)
			}
		} else {
			dc.SetRGB255(0x22, 0x22, 0x22)
		}
		dc.DrawRoundedRectangle(x, y, cell, cell, 6)
		dc.Fill()
		if picked {
			if winner {
				dc.SetRGB255(0xff, 0x3d, 0x00)
			} else {
				dc.SetRGB255(0x00, 0xe5, 0xff)
			}
		} else {
			dc.SetRGB255(0x33, 0x33, 0x33)
		}
		dc.DrawRoundedRectangle(x, y, cell, cell, 6)
		dc.Stroke()
		// num-bg big faded number
		setFont(dc, 32)
		if winner {
			dc.SetRGBA255(255, 61, 0, 38)
		} else if picked {
			dc.SetRGBA255(0, 229, 255, 25)
		} else {
			dc.SetRGBA255(255, 255, 255, 8)
		}
		drawStringAnchored(dc, fmt.Sprintf("%d", i), x+cell/2, y+cell/2, 0.5, 0.5)
		// num-top
		setFont(dc, 16)
		if picked {
			if winner {
				dc.SetRGB255(0xff, 0x3d, 0x00)
			} else {
				dc.SetRGB255(0x00, 0xe5, 0xff)
			}
		} else {
			dc.SetRGB255(0x66, 0x66, 0x66)
		}
		drawString(dc, fmt.Sprintf("%d", i), x+6, y+22)
		if picked {
			nameC := [3]int{255, 255, 255}
			if winner {
				nameC = [3]int{0xff, 0x3d, 0x00}
			}
			setFont(dc, 12)
			dc.SetRGB255(nameC[0], nameC[1], nameC[2])
			drawString(dc, d.UserName, x+6, y+cell-26)
			setFont(dc, 10)
			dc.SetRGBA255(255, 255, 255, 178)
			drawString(dc, "ID:"+d.UserNumber, x+6, y+cell-10)
		}
	}

	// legend
	ly := cy0 + ch - 36.0
	legend := []struct {
		label  string
		bg     [3]int
		border [3]int
	}{
		{"未选择", [3]int{0x22, 0x22, 0x22}, [3]int{0x33, 0x33, 0x33}},
		{"已占位", [3]int{0x1e, 0x3a, 0x3f}, [3]int{0x00, 0xe5, 0xff}},
		{"中奖", [3]int{0x4a, 0x1c, 0x1c}, [3]int{0xff, 0x3d, 0x00}},
	}
	itemW := 16.0 + 8.0 + 42.0 // box + gap + label(3 chars @14)
	totalW := 3*itemW + 2*30
	lx := (cssW - totalW) / 2
	setFont(dc, 14)
	for _, lg := range legend {
		dc.SetRGB255(lg.bg[0], lg.bg[1], lg.bg[2])
		dc.DrawRoundedRectangle(lx, ly, 16, 16, 4)
		dc.Fill()
		dc.SetRGB255(lg.border[0], lg.border[1], lg.border[2])
		dc.DrawRoundedRectangle(lx, ly, 16, 16, 4)
		dc.Stroke()
		dc.SetRGB255(255, 255, 255)
		drawString(dc, lg.label, lx+24, ly+13)
		lx += itemW + 30
	}
	return dc, nil
}
