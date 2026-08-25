package ggrender

import (
	"fmt"

	"github.com/fogleman/gg"
)

// Gacha
type GachaData struct {
	Name string; Total, Star6, Star5, Star4, Star3 int; Avg6, Avg5, Avg4, Avg3 float64
	Chars []struct{ PoolName, CharName string; Rarity int64; IsNew bool }
}

func SampleGacha() *GachaData {
	g:=&GachaData{Name:"博士的寻访记录", Total:120, Star6:6, Star5:18, Star4:50, Star3:46, Avg6:20.0, Avg5:6.6, Avg4:2.4, Avg3:2.6}
	names:=[]string{"能天使","银灰","艾雅法拉","星熊","塞雷娅","闪灵","夜莺","斯卡蒂","陈","推进之王"}
	for i,n:=range names {
		r:=int64(5)
		if i%3==0 {r=4}
		if i%5==0 {r=3}
		g.Chars=append(g.Chars, struct{ PoolName, CharName string; Rarity int64; IsNew bool }{PoolName:"常驻", CharName:n, Rarity:r, IsNew: i%4==0})
	}
	return g
}

func RenderGacha(data *GachaData) (*gg.Context, error) {
	const mainW=1500
	headerH:=90
	mainH:=1323
	dc:=gg.NewContext(mainW,mainH)
	FillBackground(dc,27,29,30)
	// header
	dc.SetRGB255(45,48,55)
	dc.DrawRectangle(0,0,float64(mainW),float64(headerH))
	dc.Fill()
	setFont(dc,24)
	dc.SetRGB255(255,255,255)
	drawString(dc,data.Name,25,52)
	setFont(dc,14)
	dc.SetRGB255(180,200,220)
	drawString(dc,fmt.Sprintf("共 %d 抽 · 6星%d 5星%d 4星%d 3星%d",data.Total,data.Star6,data.Star5,data.Star4,data.Star3),25,74)
	// stats avg
	y:=headerH+16
	stats:=[]struct{label string; val float64; cnt int}{
		{"6星",data.Avg6,data.Star6},{"5星",data.Avg5,data.Star5},{"4星",data.Avg4,data.Star4},{"3星",data.Avg3,data.Star3},
	}
	x:=20
	for _,s:=range stats {
		dc.SetRGBA255(255,255,255,12)
		RoundRect(dc,float64(x),float64(y),200,80,8)
		setFont(dc,14)
		dc.SetRGB255(180,200,220)
		drawStringAnchored(dc,s.label,float64(x+100),float64(y+24),0.5,0.5)
		setFont(dc,22)
		dc.SetRGB255(255,240,120)
		drawStringAnchored(dc,fmt.Sprintf("%.1f",s.val),float64(x+100),float64(y+52),0.5,0.5)
		setFont(dc,11)
		dc.SetRGB255(200,200,200)
		drawStringAnchored(dc,fmt.Sprintf("(%d)",s.cnt),float64(x+100),float64(y+68),0.5,0.5)
		x+=220
	}
	y+=110
	// chars list
	setFont(dc,14)
	dc.SetRGB255(200,220,200)
	drawString(dc,"最近获得",20,float64(y))
	y+=20
	for i,ch:=range data.Chars {
		if i>=20 {break}
		yy:=y+i*74
		fillRoundedCard(dc,20,float64(yy),float64(mainW-40),64,8,10)
		// avatar
		dc.SetRGB255(80,80,90)
		dc.DrawCircle(50,float64(yy+32),22)
		dc.Fill()
		setFont(dc,14)
		dc.SetRGB255(255,255,255)
		drawString(dc,ch.CharName,80,float64(yy+24))
		setFont(dc,12)
		dc.SetRGB255(180,200,220)
		drawString(dc,ch.PoolName,80,float64(yy+44))
		// rarity color bar
		r,g,b:=rarityColor(int(ch.Rarity+1))
		dc.SetRGB255(r,g,b)
		dc.DrawRectangle(float64(mainW-80),float64(yy+20),50,24)
		dc.Fill()
		setFont(dc,12)
		dc.SetRGB255(255,255,255)
		drawStringAnchored(dc,fmt.Sprintf("%d★",ch.Rarity+1),float64(mainW-55),float64(yy+32),0.5,0.5)
		if ch.IsNew {
			setFont(dc,10)
			dc.SetRGB255(255,80,80)
			drawString(dc,"NEW",float64(mainW-140),float64(yy+32))
		}
	}
	return dc,nil
}
