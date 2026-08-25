package ggrender

import (
	"github.com/fogleman/gg"
	"math"
)

// Help — mirrors template/Help.tmpl rendered at 660x1366 CSS, scale 1.5 -> 990x2049.
// bg.jpg cover; banner.png (660x239 CSS) at y=10 with h1 + person-line overlay;
// dark .bg panel (amiya.png + 0.8 black) holds label bars + 4-per-row cmd chips.
type Cmd struct {
	Cmd, Desc, Param string
	IsBind           bool
}
type HelpData struct {
	Private []Cmd
	Public  []Cmd
	Admin   []Cmd
}

func SampleHelp() *HelpData {
	h := &HelpData{}
	h.Private = []Cmd{
		{Cmd: "/bind", Desc: "绑定角色"},
		{Cmd: "/unbind", Desc: "解绑角色", IsBind: true},
		{Cmd: "/cancel", Desc: "取消操作"},
		{Cmd: "/reset_token", Desc: "重设token", IsBind: true},
		{Cmd: "/import_gacha", Desc: "导入抽卡记录", IsBind: true},
		{Cmd: "/export_gacha", Desc: "导出抽卡记录", IsBind: true},
	}
	h.Public = []Cmd{
		{Cmd: "/help", Desc: "使用说明"},
		{Cmd: "/ping", Desc: "存活测试"},
		{Cmd: "/tag", Desc: "自定义群标签", Param: "标签"},
		{Cmd: "/sign", Desc: "签到", IsBind: true},
		{Cmd: "/sign", Desc: "开启自动签到", Param: "auto", IsBind: true},
		{Cmd: "/sign", Desc: "关闭自动签到", Param: "stop", IsBind: true},
		{Cmd: "/sign", Desc: "全部通知", Param: "notify_all", IsBind: true},
		{Cmd: "/sign", Desc: "仅失败时通知", Param: "notify_fail", IsBind: true},
		{Cmd: "/sign", Desc: "仅成功时通知", Param: "notify_success", IsBind: true},
		{Cmd: "/ap", Desc: "开启理智提醒", Param: "on", IsBind: true},
		{Cmd: "/ap", Desc: "关闭理智提醒", Param: "off", IsBind: true},
		{Cmd: "/ap", Desc: "设理智提醒阈值", Param: "thr [1-100]", IsBind: true},
		{Cmd: "/state", Desc: "当前状态", IsBind: true},
		{Cmd: "/box", Desc: "我的干员(默认6星)", IsBind: true},
		{Cmd: "/box", Desc: "所有干员", Param: "all", IsBind: true},
		{Cmd: "/box", Desc: "对应星级干员", Param: "5,6", IsBind: true},
		{Cmd: "/box_detail", Desc: "干员详情(默认6星)", IsBind: true},
		{Cmd: "/box_detail", Desc: "对应星级干员", Param: "5", IsBind: true},
		{Cmd: "/box_summary", Desc: "干员信息汇总", IsBind: true},
		{Cmd: "/missing", Desc: "未获取干员(默认6星)", IsBind: true},
		{Cmd: "/missing", Desc: "所有未获取干员", Param: "all", IsBind: true},
		{Cmd: "/missing", Desc: "对应星级未获取干员", Param: "5,6", IsBind: true},
		{Cmd: "/card", Desc: "我的名片", IsBind: true},
		{Cmd: "/base", Desc: "基建信息", IsBind: true},
		{Cmd: "/gacha", Desc: "抽卡记录", IsBind: true},
		{Cmd: "/operator", Desc: "干员查询"},
		{Cmd: "/skin", Desc: "干员皮肤查询"},
		{Cmd: "/enemy", Desc: "敌人查询"},
		{Cmd: "/report", Desc: "举报"},
		{Cmd: "/quiz", Desc: "云玩家检测"},
		{Cmd: "/quiz", Desc: "云玩家检测(困难)", Param: "h"},
		{Cmd: "/redeem", Desc: "CDK兑换", Param: "[CDK]", IsBind: true},
		{Cmd: "/headhunt", Desc: "寻访模拟"},
		{Cmd: "/recruit", Desc: "公招计算(图片附带)"},
		{Cmd: "/calendar", Desc: "活动日历"},
		{Cmd: "/depot", Desc: "我的仓库", IsBind: true},
	}
	h.Admin = []Cmd{
		{Cmd: "/news", Desc: "开启/关闭动态推送"},
		{Cmd: "/birthday", Desc: "开启/关闭生日推送"},
		{Cmd: "/request_mode", Desc: "切换群验证模式"},
		{Cmd: "/quiz", Desc: "开启云玩家检测", Param: "on"},
		{Cmd: "/quiz", Desc: "关闭云玩家检测", Param: "off"},
		{Cmd: "/headhunt", Desc: "开启寻访模拟", Param: "on"},
		{Cmd: "/headhunt", Desc: "关闭寻访模拟", Param: "off"},
		{Cmd: "/reg", Desc: "回复消息设置为群规"},
		{Cmd: "/welcome", Desc: "设置入群欢迎信息", Param: "文本"},
	}
	return h
}

// drawPersonIcon approximates the bootstrap person-circle svg (16px grid scaled).
func drawPersonIcon(dc *gg.Context, x, y, s float64) {
	cx, cy, r := x+s/2, y+s/2, s/2
	lw := s * 0.0833
	dc.SetRGBA255(255, 255, 255, 255)
	dc.SetLineWidth(lw)
	dc.DrawCircle(cx, cy, r-lw/2)
	dc.Stroke()
	dc.DrawCircle(cx, cy-s*0.125, s*0.1875) // head
	dc.Fill()
	// shoulders: filled dome (upper half of circle), no clipping needed
	dc.DrawArc(cx, cy+s*0.42, s*0.34, math.Pi, 2*math.Pi)
	dc.ClosePath()
	dc.Fill()
	dc.Stroke()
}

func drawHelpChips(dc *gg.Context, cmds []Cmd, yTop float64) float64 {
	const chipW, chipH = 227.0, 80.0
	const chipMarginX, chipMarginTop = 14.0, 21.7
	const perRow = 4
	y := yTop
	for i, c := range cmds {
		col := i % perRow
		if col == 0 && i > 0 {
			y += chipH + 23.55
		}
		x := chipMarginX + float64(col)*243
		dc.SetRGBA255(255, 255, 255, 255)
		dc.SetLineWidth(1.5)
		StrokeRoundRect(dc, x, y, chipW, chipH, 15)
		// p1: cmd + param, person icon right when IsBind
		setFont(dc, 22.5)
		dc.SetRGB255(255, 255, 255)
		line1 := c.Cmd
		if c.Param != "" {
			line1 += " " + c.Param
		}
		drawStringBoldW(dc, line1, x+10, y+31, 0.9)
		if c.IsBind {
			drawPersonIcon(dc, x+chipW-40.5, y+6, 24)
		}
		// p2: desc
		drawStringBoldW(dc, c.Desc, x+10, y+70, 0.9)
	}
	return y + chipH
}

// drawStringBold fakes bold by symmetric double-draw (only regular face available).
func drawStringBold(dc *gg.Context, s string, x, y float64) {
	drawStringBoldW(dc, s, x, y, 1.2)
}

func RenderHelp(data *HelpData) (*gg.Context, error) {
	const W, H = 990, 2049
	dc := gg.NewContext(W, H)
	FillBackground(dc, 255, 255, 255)
	// #main bg.jpg cover
	dc.DrawImage(ScaleCover(tryLocal("help/bg.jpg"), W, H), 0, 0)
	// banner.png 100% width at y=15
	dc.DrawImage(ScaleExact(tryLocal("help/banner.png"), W, 359), 0, 15)
	// h1 使用说明 (32px bold white) over banner bottom area
	setFont(dc, 48)
	dc.SetRGB255(255, 255, 255)
	drawStringBold(dc, "使用说明", 29, 271)
	// person line: svg + 为需要绑定角色的指令
	drawPersonIcon(dc, 26, 309, 24)
	setFont(dc, 24)
	drawString(dc, "为需要绑定角色的指令", 54, 333)

	// .bg panel: amiya.png (660px auto, center top) + rgba(0,0,0,0.8) overlay
	bgTop := 373.6
	dc.DrawImage(ScaleExact(tryLocal("help/amiya.png"), W, 1018), -4, int(bgTop))
	dc.SetRGBA255(0, 0, 0, 204)
	dc.DrawRectangle(0, bgTop, W, H-bgTop)
	dc.Fill()

	// sections: label bar (990x60) + chips
	sections := []struct {
		title string
		cmds  []Cmd
	}{
		{"私聊指令", data.Private},
		{"普通指令", data.Public},
		{"管理员指令", data.Admin},
	}
	y := bgTop
	for _, sec := range sections {
		y += 15
		dc.DrawImage(ScaleExact(tryLocal("help/label.png"), W, 60), 0, int(y))
		setFont(dc, 24)
		dc.SetRGB255(255, 255, 255)
		drawString(dc, sec.title, 37.5, y+38)
		y += 60
		y = drawHelpChips(dc, sec.cmds, y+23.7)
	}
	return dc, nil
}
