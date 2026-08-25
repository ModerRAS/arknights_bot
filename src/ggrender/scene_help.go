package ggrender

import "github.com/fogleman/gg"

// Help
type HelpData struct {
	Private []Cmd; Public []Cmd; Admin []Cmd
}
type Cmd struct{ Cmd, Desc, Param string; IsBind bool }

func SampleHelp() *HelpData {
	h:=&HelpData{}
	h.Private=[]Cmd{{Cmd:"/bind",Desc:"绑定角色",Param:""},{Cmd:"/unbind",Desc:"解绑角色",Param:""},{Cmd:"/cancel",Desc:"取消操作",Param:""}}
	h.Public=[]Cmd{
		{Cmd:"/help",Desc:"使用说明",Param:""},
		{Cmd:"/box",Desc:"我的干员",Param:""},
		{Cmd:"/state",Desc:"当前状态",Param:""},
		{Cmd:"/card",Desc:"我的名片",Param:""},
		{Cmd:"/base",Desc:"基建信息",Param:""},
		{Cmd:"/gacha",Desc:"抽卡记录",Param:""},
		{Cmd:"/depot",Desc:"我的仓库",Param:""},
		{Cmd:"/calendar",Desc:"活动日历",Param:""},
		{Cmd:"/recruit",Desc:"公招计算",Param:""},
		{Cmd:"/headhunt",Desc:"寻访模拟",Param:""},
	}
	h.Admin=[]Cmd{{Cmd:"/news",Desc:"动态推送",Param:""},{Cmd:"/birthday",Desc:"生日推送",Param:""}}
	return h
}

func RenderHelp(data *HelpData) (*gg.Context, error) {
	const mainW=990
	const mainH=2049
	dc:=gg.NewContext(mainW,mainH)
	FillBackground(dc,46,48,49)
	// banner placeholder
	dc.SetRGB255(60,62,80)
	dc.DrawRectangle(0,0,float64(mainW),140)
	dc.Fill()
	setFont(dc,28)
	dc.SetRGB255(255,255,255)
	drawString(dc,"Arknights Bot · 使用说明",30,80)
	setFont(dc,14)
	dc.SetRGB255(200,220,255)
	drawString(dc,"基于森空岛数据的罗德岛助手",30,110)
	y:=160
	drawSection:=func(title string, cmds []Cmd, yy int) int {
		setFont(dc,16)
		dc.SetRGB255(120,200,220)
		drawString(dc,title,20,float64(yy))
		yy+=20
		for _,c:=range cmds {
			dc.SetRGBA255(255,255,255,10)
			RoundRect(dc,20,float64(yy),float64(mainW-40),28,6)
			setFont(dc,13)
			dc.SetRGB255(255,230,120)
			drawString(dc,c.Cmd,30,float64(yy+18))
			dc.SetRGB255(200,200,200)
			drawString(dc,c.Desc,160,float64(yy+18))
			if c.Param!="" {
				dc.SetRGB255(160,180,200)
				drawString(dc,c.Param,300,float64(yy+18))
			}
			yy+=32
		}
		return yy+10
	}
	y=drawSection("私聊指令",data.Private,y)
	y=drawSection("群聊指令",data.Public,y)
	y=drawSection("管理员指令",data.Admin,y)
	return dc,nil
}
