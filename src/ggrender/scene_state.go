package ggrender

import (
	"fmt"

	"github.com/fogleman/gg"
)

// State — moved verbatim from render.go (atomic-dedup); parity rewrite follows.

// State
type StateInfo struct {
	Name string; Level int; ApCur, ApTotal int
	SanityExpire string
	Chars []struct{Name string; Ap int}
}

func SampleState() *StateInfo {
	s:=&StateInfo{Name:"博士", Level:120, ApCur:135, ApTotal:135, SanityExpire:"12:30"}
	s.Chars=[]struct{Name string; Ap int}{{"阿米娅", 22},{"凯尔希", 20},{"煌", 18},{"能天使", 16}}
	return s
}

func RenderState(data *StateInfo) (*gg.Context, error) {
	const mainW=1092
	const mainH=510
	dc:=gg.NewContext(mainW,mainH)
	FillBackground(dc,46,48,49)
	// try draw bg image if exists
	if bg,err:=LoadImage(AssetPath("state/bg.png")); err==nil {
		dc.DrawImage(ScaleCover(bg,mainW,mainH),0,0)
	}
	// avatar
	dc.SetRGB255(80,80,90)
	dc.DrawCircle(34+27,34+27,27)
	dc.Fill()
	setFont(dc,26)
	dc.SetRGB255(255,255,255)
	drawString(dc,data.Name,100,60)
	setFont(dc,16)
	dc.SetRGB255(220,220,220)
	drawString(dc,fmt.Sprintf("Lv%d",data.Level),100,84)
	// ap bar
	frac:=float64(data.ApCur)/float64(data.ApTotal)
	if data.ApTotal==0 {frac=1}
	ProgressBar(dc,146,146,410,11,frac,255,255,255)
	setFont(dc,28)
	dc.SetRGB255(255,255,255)
	drawString(dc,fmt.Sprintf("%d/%d",data.ApCur,data.ApTotal),146,140)
	setFont(dc,16)
	dc.SetRGB255(200,200,200)
	drawString(dc,"理智恢复 "+data.SanityExpire,146,170)
	// chars row
	x:=20
	y:=260
	for _,c:=range data.Chars {
		dc.SetRGBA255(255,255,255,18)
		RoundRect(dc,float64(x),float64(y),120,120,10)
		dc.SetRGB255(80,80,90)
		dc.DrawCircle(float64(x+60),float64(y+44),28)
		dc.Fill()
		setFont(dc,12)
		dc.SetRGB255(255,255,255)
		drawStringAnchored(dc,c.Name,float64(x+60),float64(y+88),0.5,0.5)
		// ap badge
		dc.SetRGBA255(0,0,0,160)
		dc.DrawCircle(float64(x+60),float64(y+108),14)
		dc.Fill()
		setFont(dc,11)
		dc.SetRGB255(255,230,80)
		drawStringAnchored(dc,itoa(c.Ap),float64(x+60),float64(y+112),0.5,0.5)
		x+=140
	}
	// campaign / tradings placeholders
	y=410
	for i:=0;i<3;i++ {
		dc.SetRGBA255(255,255,255,10)
		RoundRect(dc,float64(20+i*260),float64(y),240,60,8)
		setFont(dc,13)
		dc.SetRGB255(200,220,200)
		drawString(dc,fmt.Sprintf("区块 %d",i+1),float64(40+i*260),float64(y+36))
	}
	return dc,nil
}
