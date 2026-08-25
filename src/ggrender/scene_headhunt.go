package ggrender

import "github.com/fogleman/gg"

// Headhunt — mirrors template/Headhunt.tmpl rendered at 1049x576 scale=1.
// #main: 1024px content + 25px padding-left, bg.png (1024x576) repeats after x=1024.
// Cards: .bg back_{rarity} cover 95x270 top y=130 padding-top 100; thumb 95x190;
// rarity stars h=20 and profession icon centered, profession offset +190.
type HHOp struct {
	Rarity     int
	ThumbURL   string
	Profession string
}
type HeadhuntData struct{ Ops []HHOp }

const headhuntThumbURL = "https://media.prts.wiki/a/a0/%E5%8D%8A%E8%BA%AB%E5%83%8F_%E9%98%BF%E7%B1%B3%E5%A8%85_1.png?image_process=format,webp/quality,Q_90"

func SampleHeadhunt() []HHOp {
	ops := make([]HHOp, 0, 10)
	for i := 0; i < 10; i++ {
		ops = append(ops, HHOp{Rarity: 5, ThumbURL: headhuntThumbURL, Profession: "WARRIOR"})
	}
	return ops
}

func RenderHeadhunt(data []HHOp) (*gg.Context, error) {
	const mainW, mainH = 1049, 576
	dc := gg.NewContext(mainW, mainH)
	FillBackground(dc, 255, 255, 255)

	// #main background-image: url(bg.png) contain -> exact 1024x576, repeats at x>=1024
	bg := ScaleExact(tryLocal("headhunt/bg.png"), 1024, 576)
	dc.DrawImage(bg, 0, 0)
	dc.DrawImage(bg, 1024, 0)

	for i, o := range data {
		x := 26 + i*98
		// .bg panel: back_{rarity}.png cover 95x270 at y=130
		back := ScaleCover(tryLocal("headhunt/back_"+itoa(o.Rarity)+".png"), 95, 270)
		dc.DrawImage(back, x, 130)
		// thumb 95x190 at y=230 (padding-top 100)
		thumb := ScaleExact(FetchImage(o.ThumbURL, AssetPath("common/amiya.png")), 95, 190)
		dc.DrawImage(thumb, x, 230)
		// rarity stars: height 20, horizontally centered on card
		rar := tryLocal("headhunt/Rarity_" + itoa(o.Rarity) + ".png")
		rw := rar.Bounds().Dx() * 20 / rar.Bounds().Dy()
		dc.DrawImage(ScaleExact(rar, rw, 20), x+(95-rw)/2, 230)
		// profession icon: natural size, centered, static top y=230 + margin-top 190
		prof := tryLocal("headhunt/" + o.Profession + ".png")
		pw, ph := prof.Bounds().Dx(), prof.Bounds().Dy()
		dc.DrawImage(ScaleExact(prof, pw, ph), x+(95-pw)/2, 420)
	}
	return dc, nil
}
